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
	"github.com/spf13/cobra"
)

// NewCmd returns a cobra command for all `api-server` subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-server [COMMAND]",
		Short: "Run a Envoy based API Server.",
		Long: `Run a preconfigured API Gateway by specifying a Swagger file as an input.
		        This file is used to generate a configuration that can be consumed by Istio and programmed into Envoy,
			allowing it to function as a gateway.`,
	}
	cmd.AddCommand(newLocalCmd())
	cmd.AddCommand(newKubernetesCmd())
	return cmd
}
