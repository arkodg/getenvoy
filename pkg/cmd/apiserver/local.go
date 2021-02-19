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
	"path/filepath"

	"github.com/tetratelabs/getenvoy/pkg/apiserver"
	"github.com/tetratelabs/getenvoy/pkg/apiserver/local"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// newLocalCmd returns a cobra command for the `local` subcommand.
func newLocalCmd() *cobra.Command {
	var l local.Environment
	var s apiserver.Server
	var configDir, listenAddr, serveAddr string
	var err error
	cmd := &cobra.Command{
		Use:   "local [OPTIONS]",
		Short: "Run the API Server locally to access an application running in a docker container",
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
			if s.GenConfigDir != "" {
				s.GenConfigDir, err = filepath.Abs(configDir)
				if err != nil {
					log.Errorf("unable to get absolute path %s", err)
				}
			}
			return runLocal(s, &l)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&listenAddr, "listen-addr", "", "Address on which the API Server is listening on e.g. http://0.0.0.0:8080")
	cmd.MarkPersistentFlagRequired("listen-addr")
	flags.StringVar(&serveAddr, "serve-addr", "", "URL of the backend application being served e.g. http://demo.com:9080. Make sure the application container's name matches the hostname (demo.com in this case) and is attached to the default docker bridge.")
	cmd.MarkPersistentFlagRequired("serve-addr")
	flags.StringVar(&s.SwaggerFile, "swagger-file", "", "Location of the Swagger file")
	cmd.MarkPersistentFlagRequired("swagger-file")
	flags.StringVar(&configDir, "config-dir", "", "Location of the directory where the generated config is stored")
	return cmd
}

func runLocal(s apiserver.Server, l *local.Environment) error {
	return s.Run(l)
}
