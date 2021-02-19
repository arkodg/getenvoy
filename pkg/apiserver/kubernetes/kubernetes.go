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

package kubernetes

import (
	"net/url"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type Environment struct {
	DeployIstio bool
	Namespace   string
}

func (e *Environment) GenerateEnvConfig(genConfigDir, meshConfigFile string, listenAddr, serveAddr *url.URL) error {
	if e.DeployIstio {
		istioctlCmd := "istioctl install --set profile=preview"
		return runCmd(istioctlCmd)
	}
	return nil
}

func (e *Environment) ApplyConfig(genConfigDir, meshConfigFile string, listenAddr, serveAddr *url.URL) error {
	kubectlCmd := "kubectl apply -f " + genConfigDir
	return runCmd(kubectlCmd)
}

func (e *Environment) RevertConfig(genConfigDir string) error {
	kubectlCmd := "kubectl delete -f " + genConfigDir
	err := runCmd(kubectlCmd)
	if e.DeployIstio {
		rmIstioCmd := "istioctl manifest generate --set profile=preview | kubectl delete -f -"
		return runCmd(rmIstioCmd)
	}
	return err
}

func runCmd(cmd string) error {
	log.Infof("running `%s`", cmd)
	return exec.Command("bash", "-c", cmd).Run()
}
