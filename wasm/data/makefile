#
# Licensed to Apache Software Foundation (ASF) under one or more contributor
# license agreements. See the NOTICE file distributed with
# this work for additional information regarding copyright
# ownership. Apache Software Foundation (ASF) licenses this file to you under
# the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

.PHONY: wasm

wasm:
	docker run -it -v $(shell pwd):/tmp/proxy-wasm-go -e GOPROXY=https://goproxy.cn tinygo/tinygo-dev:a304fb738f2376548e008fbc5c02abf3bdff7156 \
		/bin/bash -c \
		"git clone --depth 1 --branch v0.0.13 https://github.com/tetratelabs/proxy-wasm-go-sdk.git /tmp/proxy-wasm-go/sdk &> /dev/null; \
		cp /tmp/proxy-wasm-go/${name}.go /tmp/proxy-wasm-go/sdk; \
		cd /tmp/proxy-wasm-go/sdk; \
		tinygo build -o /tmp/proxy-wasm-go/${name}.wasm -scheduler=none -target=wasi -wasm-abi=generic -tags 'abi_010' ./${name}.go; \
		rm -rf /tmp/proxy-wasm-go/sdk"