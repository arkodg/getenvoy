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
	"os"
	"os/exec"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Config is the configuration for setting up the API Server
type Config struct {
	// ListenAddr is the address the API Server is listening on
	ListenAddr *url.URL
	// ServeAddr is the address of the backend application being served
	ServeAddr *url.URL
	// SwaggerFile is the input swagger file used to configure the API Server
	SwaggerFile string
	// GenConfigDir is the directory where the generated istio configuration is saved into
	GenConfigDir string
	// MeshConfigFile is the file where the istio mesh configuration is saved into
	MeshConfigFile string
	// DryRun generates the API gateway istio configuration without applying it
	DryRun bool
	// SkipRevert skips removing the applied configuration from the environment
	SkipRevert bool
}

type Server struct {
	Config
	ScratchDir string
}

// Run starts an API Server that generates the Istio configuration and applies it
// to implement the intended API Gateway configuration in the specified environment
func (s *Server) Run(e Environment) error {
	log.Info("generating istio based configuration from the swagger file")
	if err := s.createScratchDir(); err != nil {
		errors.Wrapf(err, "unable to create scratch dir")
	}
	// Generate istio configuration
	if err := s.generateIstioConfig(); err != nil {
		return errors.Wrapf(err, "failed to generate istio configuration")
	}

	// Generate environment specific configuration
	if err := e.GenerateEnvConfig(s.GenConfigDir, s.MeshConfigFile, s.ListenAddr, s.ServeAddr); err != nil {
		return errors.Wrapf(err, "failed to generate environment configuration")
	}

	// Return early if this is a dry run
	if s.DryRun {
		return nil
	}

	log.Info("applying the generated configuration")
	// Apply the istio configuration to implement the API Gateway spec
	if err := e.ApplyConfig(s.GenConfigDir, s.MeshConfigFile, s.ListenAddr, s.ServeAddr); err != nil {
		e.RevertConfig(s.GenConfigDir)
		s.deleteScratchDir()
		return errors.Wrapf(err, "failed to apply configuration")
	}

	// Wait until the user terminates the process
	log.Infof("Press Ctrl+C to end")
	waitForCtrlC()

	log.Info("reverting the generated configuration")
	// Revert the applied configuration
	if !s.SkipRevert {
		if err := e.RevertConfig(s.GenConfigDir); err != nil {
			return errors.Wrapf(err, "failed to revert configuration")
		}
	}
	if err := s.deleteScratchDir(); err != nil {
		errors.Wrapf(err, "unable to delete scratch dir")
	}

	return nil
}

// generateIstioConfig takes a swagger input file, generates the istio configuration
// to implement the API Gateway spec and saves it in the genConfigDir directory
func (s *Server) generateIstioConfig() error {
	if s.GenConfigDir == "" {
		s.GenConfigDir = s.ScratchDir + "/gen-istio-config"
		err := os.MkdirAll(s.GenConfigDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	s.MeshConfigFile = s.ScratchDir + "/mesh-config"
	// TODO gen-istio-swagger --in s.SwaggerFile --out-dir s.GenConfigDir
	return exec.Command("/usr/local/bin/gen-istio-swagger", "-i", s.SwaggerFile, "-o", s.GenConfigDir, "-s", s.ServeAddr.String()).Run()
}

func (s *Server) createScratchDir() error {
	// Setup a directory to save the configuration if it doesnt exist
	s.ScratchDir = getTempConfigDir()
	err := os.MkdirAll(s.ScratchDir, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) deleteScratchDir() error {
	err := os.Remove(s.ScratchDir)
	if err != nil {
		return err
	}
	return nil
}
