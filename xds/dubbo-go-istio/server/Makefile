#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.
#

IMAGE = ccr.ccs.tencentyun.com/pixiu/demo-app#$(your_repo)/$(namespace)/$(image_name)
TAG = dubbo-go-sample-dubbo-1.1
APPNAME = dubbo-go-app

build-amd64-app:
	GOOS=linux GOARCH=amd64 go build -o build/app ./cmd

build: proto-gen tidy build-amd64-app
	cp ./conf/dubbogo.yaml ./build/dubbogo.yaml
	docker build --platform linux/amd64 ./build  -t ${IMAGE}:${TAG}
	docker push ${IMAGE}:${TAG}
	make clean

buildx-publish: proto-gen tidy build-amd64-app
	cp ./conf/dubbogo.yaml ./build/dubbogo.yaml
	docker buildx build \
    	 --platform linux/amd64 \
    	 -t ${IMAGE}:${TAG} \
    	 ./build --push
	make clean

#FROM ccr.ccs.tencentyun.com/pixiu/demo-app:dubbo-go-pixiu
build-pixiu-gateway:
	cp ../pixiu/conf.yaml ./build/
	docker buildx build \
    	 --platform linux/amd64 \
    	 -t ccr.ccs.tencentyun.com/pixiu/demo-app:dubbo-go-pixiu-gateway-0.5.1-rc \
    	 ./build -f ./build/DockerfilePixiu --push

remove:
	helm uninstall ${APPNAME}

deploy:
	helm install ${APPNAME} ./chart/app

deploy-nacos-env:
	helm install nacos ./chart/nacos_env

deploy-pixiu-gateway:
	helm install pixiu ./chart/pixiu-gateway

remove-pixiu-gateway:
	helm uninstall pixiu

remove-nacos-env:
	helm uninstall nacos

proto-gen:
	protoc --go_out=./api --go-triple_out=./api ./api/api.proto

clean:
	rm ./build/dubbogo.yaml
	rm ./build/app

tidy:
	go mod tidy

test:
	go test ./...
