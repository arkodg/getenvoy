#!/usr/bin/env bash

# Copyright 2020 Tetrate
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

extension_build()  {
	# TODO: remove abi_010 tag after upgrading wasm:nightly or Envoy 1.17 release
	exec tinygo build -o "$1" -tags=abi_010 -scheduler=none -target wasi main.go
}

extension_test()  {
	exec go test -tags=proxytest -v ./...
}

extension_clean()  {
	rm /source/extension.wasm || true
	go clean -modcache
	rm -rf "${GOCACHE}" "${XDG_CACHE_HOME}" || true
}
