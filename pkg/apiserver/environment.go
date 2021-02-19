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
)

// Environment is the interface for generating and applying different
// APIServer configurations
type Environment interface {
	// GenerateEnvConfig creates any environment specific configuration if needed
	GenerateEnvConfig(genConfigDir, meshConfigFile string, listenAddr, serveAddr *url.URL) error
	// ApplyConfig consumes the generated configuration to create the
	// intended API Gateway configuration
	ApplyConfig(genConfigDir, meshConfigFile string, listenAddr, serveAddr *url.URL) error
	// RevertConfig reverts the config applied
	RevertConfig(genConfigDir string) error
}
