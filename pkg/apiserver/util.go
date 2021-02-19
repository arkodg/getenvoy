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
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
)

const (
	DefaultGenConfigDir = "/tmp/getenvoy/api-server/<pid>"
)

// getTempConfigDir generates a temporary directory name to save the generated istio configuration
func getTempConfigDir() string {
	dirFormat := DefaultGenConfigDir
	pid := strconv.Itoa(os.Getpid())
	dir := strings.Replace(dirFormat, "<pid>", pid, -1)
	return dir
}

func waitForCtrlC() {
	var endWaiter sync.WaitGroup
	endWaiter.Add(1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		endWaiter.Done()
	}()
	endWaiter.Wait()
}
