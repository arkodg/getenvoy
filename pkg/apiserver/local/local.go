// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package local

import (
	"context"
	"io"
	"net/url"
	"os"
	"text/template"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	log "github.com/sirupsen/logrus"
)

const (
	PilotDiscoveryDockerImage = "docker.io/istio/pilot:latest"
	PilotAgentDockerImage     = "docker.io/istio/proxyv2:latest"
	PilotDiscoveryHostname    = "pilot-discovery"
	PilotAgentHostname        = "pilot-agent"
	MeshConfig                = `
enableAutoMtls: false
accessLogFile: /dev/stdout
defaultConfig:
  discoveryAddress: pilot-discovery:15010
  controlPlaneAuthPolicy: NONE
  terminationDrainDuration: 0s
  sds:
    enabled: false
defaultServiceExportTo:
  - '*'
defaultVirtualServiceExportTo:
  - '*'
defaultDestinationRuleExportTo:
  - '*'
configSources:
- address: fs:///var/lib/istio/config/data
`
	ServiceEntryTemplate = `
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: {{.Host}} 
spec:
  exportTo:
  - "*"
  hosts:
  - "{{.Host}}"
  ports:
  - number: {{.Port}}
    name: http
    protocol: HTTP 
  resolution: DNS
`
)

type Environment struct {
	PilotDiscoveryContainerID string
	PilotAgentContainerID     string
}

type ServiceEntryHost struct {
	Host string
	Port string
}

func (e *Environment) GenerateEnvConfig(genConfigDir, meshConfigFile string, listenAddr, serveAddr *url.URL) error {
	// Write Mesh Config into a temp file that can be later mounted into the pilot container
	f, err := os.Create(meshConfigFile)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(MeshConfig); err != nil {
		return err
	}
	// Create Service Entry
	serviceEntryFile := genConfigDir + "/service-entry.yaml"
	w, err := os.Create(serviceEntryFile)
	if err != nil {
		return err
	}
	s := ServiceEntryHost{
		Host: serveAddr.Hostname(),
		Port: serveAddr.Port(),
	}
	t := template.Must(template.New("serviceEntryTemplate").Parse(ServiceEntryTemplate))
	if err = t.Execute(w, s); err != nil {
		return err
	}
	return nil
}

// ApplyConfig runs the pilot-discovery and pilot-agent containers
func (e *Environment) ApplyConfig(genConfigDir, meshConfigFile string, listenAddr, serveAddr *url.URL) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	log.Debugf("pulling %s docker image", PilotDiscoveryDockerImage)
	reader, err := cli.ImagePull(ctx, PilotDiscoveryDockerImage, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)
	log.Debugf("starting pilot-discovery container")
	serveHost := serveAddr.Hostname()
	resp, err := cli.ContainerCreate(ctx,
		&container.Config{
			Image: PilotDiscoveryDockerImage,
			Cmd: []string{"discovery",
				"--registries=",
				"--log_output_level=debug"},
			Hostname: PilotDiscoveryHostname,
		},
		&container.HostConfig{
			AutoRemove:  true,
			NetworkMode: container.NetworkMode("bridge"),
			Links:       []string{serveHost + ":" + serveHost},
			Binds: []string{
				genConfigDir + ":/var/lib/istio/config/data",
				meshConfigFile + ":/etc/istio/config/mesh",
			},
		}, nil, nil, PilotDiscoveryHostname)

	if err != nil {
		return err
	}
	e.PilotDiscoveryContainerID = resp.ID
	if err := cli.ContainerStart(ctx, e.PilotDiscoveryContainerID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	log.Infof("started pilot-discovery container")

	log.Debugf("pulling %s docker image", PilotAgentDockerImage)
	reader, err = cli.ImagePull(ctx, PilotAgentDockerImage, types.ImagePullOptions{})
	io.Copy(os.Stdout, reader)
	if err != nil {
		return err
	}

	containerPort, err := nat.NewPort("tcp", listenAddr.Port())
	if err != nil {
		log.Errorf("unable to get container port for pilot agent %s", err)
		return err
	}
	resp, err = cli.ContainerCreate(ctx,
		&container.Config{
			Image: PilotAgentDockerImage,
			Cmd: []string{"proxy",
				"router",
				"--proxyLogLevel=debug"},
			Env: []string{"ENABLE_CA_SERVER=false",
				"FILE_MOUNTED_CERTS=true",
				"ISTIO_META_NAMESPACE=default",
				`ISTIO_METAJSON_LABELS={"app": "istio-ingressgateway", "api": "demo"}`},
			ExposedPorts: nat.PortSet{
				containerPort: {},
			},
		},
		&container.HostConfig{
			AutoRemove:  true,
			NetworkMode: container.NetworkMode("bridge"),
			Links: []string{
				PilotDiscoveryHostname + ":" + PilotDiscoveryHostname,
				serveHost + ":" + serveHost},
			Binds: []string{
				meshConfigFile + ":/etc/istio/config/mesh",
			},
			PortBindings: nat.PortMap{
				containerPort: []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: listenAddr.Port(),
					},
				},
			},
		}, nil, nil, PilotAgentHostname)
	if err != nil {
		return err
	}
	e.PilotAgentContainerID = resp.ID
	if err := cli.ContainerStart(ctx, e.PilotAgentContainerID, types.ContainerStartOptions{}); err != nil {
		return err
	}
	log.Infof("started pilot-agent container")

	return nil
}

// RevertConfig stops the pilot-discovery and pilot-agent containers
func (e *Environment) RevertConfig(configDir string) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	log.Debugf("stopping pilot discovery container")
	if err := cli.ContainerStop(ctx, e.PilotDiscoveryContainerID, nil); err != nil {
		log.Errorf("unable to stop pilot discovery")
	}
	log.Debugf("stopping pilot agent container")
	if err := cli.ContainerStop(ctx, e.PilotAgentContainerID, nil); err != nil {
		log.Errorf("unable to stop pilot agent")
	}
	// Delete Service Entry
	serviceEntryFile := configDir + "/service-entry.yaml"
	if err := os.Remove(serviceEntryFile); err != nil {
		return err
	}
	return nil
}
