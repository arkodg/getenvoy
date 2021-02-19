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

package apiserver

import (
	"net/url"

	"github.com/tetratelabs/getenvoy/pkg/apiserver"
	"github.com/tetratelabs/getenvoy/pkg/apiserver/kubernetes"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// newKubernetesCmd returns a cobra command for the `kubernetes` subcommand.
func newKubernetesCmd() *cobra.Command {
	var k kubernetes.Environment
	var s apiserver.Server
	var listenAddr, serveAddr string
	var err error
	cmd := &cobra.Command{
		Use:   "kubernetes [OPTIONS]",
		Short: "Run the API Server in kubernetes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			s.ListenAddr, err = url.Parse(listenAddr)
			if err != nil {
				log.Errorf("unable to parse listenAddr %s", err)
				return err
			}
			s.ServeAddr, err = url.Parse(serveAddr)
			if err != nil {
				log.Errorf("unable to parse serveAddr %s", err)
				return err
			}
			return runKubernetes(s, &k)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&listenAddr, "listen-addr", "http://0.0.0.0:8080", "Address on which the API Server is listening on")
	flags.StringVar(&serveAddr, "serve-addr", "http://0.0.0.0:9080", "Address of the backend application being served")
	flags.StringVar(&s.SwaggerFile, "swagger-file", "swagger.json", "Location of the Swagger file")
	flags.StringVar(&s.GenConfigDir, "config-dir", apiserver.DefaultGenConfigDir, "Location of the directory where the generated config is stored")
	flags.StringVar(&k.Namespace, "namespace", "istio-system", "kubernetes namespace to deploy the istio based api-gateway deployment and service")
	flags.BoolVar(&k.DeployIstio, "deploy-istio", false, "Deploy the istio control plane component istiod as well as the dataplane component ingress-gateway to the kubernetes cluster")
	return cmd
}

func runKubernetes(s apiserver.Server, k *kubernetes.Environment) error {
	return s.Run(k)
}
