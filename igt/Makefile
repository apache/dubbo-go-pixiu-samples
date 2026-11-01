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

PROJECT_NAME = $(shell basename "$(PWD)")
PID = $(OUT_DIR)/server.pid
PROJECT_DIR=$(shell pwd)
BASE_DIR := $(PROJECT_DIR)/dist
PIXIU_DIR := $(PIXIU_DIR)
PIXIU_CONFIG ?= conf.yaml
TRAFFIC_CONFIG_GENERATOR ?=
TRAFFIC_MODULE_DIR ?= $(PIXIU_DIR)
TEST_TAGS ?= integration
TEST_RUN ?=
PIXIU_PID = $(OUT_DIR)/pixiu.pid

SOURCES = $(wildcard $(PROJECT_DIR)/server/app/*.go)
ifneq ($(wildcard $(PIXIU_DIR)/cmd/pixiu/*.go),)
PIXIU_BUILD_PACKAGE = ./cmd/pixiu
else
PIXIU_BUILD_PACKAGE = ./pixiu
endif

export GO111MODULE ?= on
export GOPROXY ?= https://goproxy.io,direct
export GOSUMDB ?= sum.golang.org
export GOARCH ?= $(shell go env GOARCH)

export DOCKER_HOST_IP = $(shell hostname)

OS := $(shell uname)
ifeq ($(OS), Linux)
	export GOOS ?= linux
else ifeq ($(OS), Darwin)
	export GOOS ?= darwin
else
	export GOOS ?= windows
endif

ifeq ($(GOOS), windows)
	export EXT_NAME ?= .exe
else
	export EXT_NAME ?=
endif

CGO ?= 0
ifeq ($(DEBUG), true)
	BUILD_TYPE := debug
	GCFLAGS := -gcflags="all=-N -l"
	LCFLAGS :=
else
	BUILD_TYPE := release
	LDFLAGS := "-s -w"
endif

OUT_DIR := $(abspath $(BASE_DIR)/$(GOOS)_$(GOARCH))
LOG_FILE := $(OUT_DIR)/$(PROJECT_NAME).log
API_CONFIG_PATH := $(OUT_DIR)/pixiuconf/api_config.yaml
CONFIG_PATH := $(OUT_DIR)/pixiuconf/conf.yaml

export APP_LOG_CONF_FILE ?= $(OUT_DIR)/conf/log.yml

# Both callers use a private PID file and verify the executable before signaling.
define stop_owned_process
if [ -f "$(1)" ]; then \
  pid=$$(cat "$(1)"); expected="$(2)"; \
  if kill -0 "$$pid" 2>/dev/null; then \
    actual=$$(ps -p "$$pid" -o command=); \
    case "$$actual" in "$$expected"|"$$expected "*) kill "$$pid" || exit 1;; *) echo "refusing to stop unrelated PID $$pid ($$actual)" >&2; exit 1;; esac; \
    for attempt in $$(seq 1 40); do \
      state=$$(ps -p "$$pid" -o stat= 2>/dev/null || true); \
      case "$$state" in ""|*Z*) break;; esac; \
      sleep 0.25; \
    done; \
    state=$$(ps -p "$$pid" -o stat= 2>/dev/null || true); \
    case "$$state" in ""|*Z*) ;; *) \
      actual=$$(ps -p "$$pid" -o command=); \
      case "$$actual" in "$$expected"|"$$expected "*) kill -KILL "$$pid" || exit 1;; *) echo "refusing to force-stop unrelated PID $$pid ($$actual)" >&2; exit 1;; esac; \
    esac; \
  fi; \
  rm -f "$(1)"; \
fi
endef

.PHONY: all
all: help
help: $(realpath $(firstword $(MAKEFILE_LIST)))
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## build: Build application's binaries
.PHONY: build
build: $(OUT_DIR)/$(PROJECT_NAME)$(EXT_NAME) config

.PHONY: $(OUT_DIR)/$(PROJECT_NAME)$(EXT_NAME)
$(OUT_DIR)/$(PROJECT_NAME)$(EXT_NAME):
	$(info   >  Buiding application binary: $(OUT_DIR)/$(PROJECT_NAME)$(EXT_NAME))
	@CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GCFLAGS) -ldflags=$(LDFLAGS) -o $(OUT_DIR)/server/$(PROJECT_NAME)$(EXT_NAME) $(SOURCES)


## config: Setup config files
.PHONY: config
config:
	$(info   >  Setting up config files)
	@mkdir -p $(OUT_DIR)/server
	@mkdir -p $(OUT_DIR)/pixiuconf
	@mkdir -p $(OUT_DIR)/proto
	@if [ -f "$(PROJECT_DIR)/server/profiles/dev/log.yml" ]; then sed "s#\$$HOST_IP#$(DOCKER_HOST_IP)#g" "$(PROJECT_DIR)/server/profiles/dev/log.yml" > "$(OUT_DIR)/server/log.yml"; fi
	@if [ -f "$(PROJECT_DIR)/server/profiles/dev/server.yml" ]; then sed "s#\$$HOST_IP#$(DOCKER_HOST_IP)#g" "$(PROJECT_DIR)/server/profiles/dev/server.yml" > "$(OUT_DIR)/server/server.yml"; fi
	@if [ -f "$(PROJECT_DIR)/pixiu/api_config.yaml" ]; then sed "s#\$$HOST_IP#$(DOCKER_HOST_IP)#g" "$(PROJECT_DIR)/pixiu/api_config.yaml" > "$(OUT_DIR)/pixiuconf/api_config.yaml"; fi
	@if [ -n "$(TRAFFIC_CONFIG_GENERATOR)" ]; then \
	  cd "$(TRAFFIC_MODULE_DIR)" && $(TRAFFIC_CONFIG_GENERATOR) -input "$(PROJECT_DIR)/pixiu/$(PIXIU_CONFIG)" -output "$(CONFIG_PATH)"; \
	elif [ -f "$(PROJECT_DIR)/pixiu/$(PIXIU_CONFIG)" ]; then \
	  sed -e "s#\$$HOST_IP#$(DOCKER_HOST_IP)#g" -e "s#\$$PROJECT_DIR#$(PROJECT_DIR)#g" "$(PROJECT_DIR)/pixiu/$(PIXIU_CONFIG)" > "$(CONFIG_PATH)"; \
	else \
	  echo "missing Pixiu config: $(PROJECT_DIR)/pixiu/$(PIXIU_CONFIG)" >&2; exit 1; \
	fi

## docker-up: Shutdown dependency services on docker
.PHONY: docker-up
docker-up:
	$(info   >  Starting dependency services with $(PROJECT_DIR)/docker/docker-compose.yml)
	@if [ -f "$(PROJECT_DIR)/docker/docker-compose.yml" ]; then docker compose -f "$(PROJECT_DIR)/docker/docker-compose.yml" up -d; fi

## docker-health-check: check services health on docker
.PHONY: docker-health-check
docker-health-check:
	$(info   >  run docker health check with $(PROJECT_DIR)/docker/docker-health-check.sh)
	@if [ -f "$(PROJECT_DIR)/docker/docker-health-check.sh" ]; then bash -f "$(PROJECT_DIR)/docker/docker-health-check.sh"; fi

## docker-down: Shutdown dependency services on docker
.PHONY: docker-down
docker-down:
	$(info   >  Stopping dependency services with $(PROJECT_DIR)/docker/docker-compose.yml)
	@if [ -f "$(PROJECT_DIR)/docker/docker-compose.yml" ]; then docker compose -f "$(PROJECT_DIR)/docker/docker-compose.yml" down; fi

## clean: Clean up the output and the binary of the application
.PHONY: clean
clean: stop
	$(info   >  Cleanning up $(OUT_DIR))
	@$(call stop_owned_process,$(PIXIU_PID),$(OUT_DIR)/dubbo-go-pixiu$(EXT_NAME))
	@rm -rf "$(OUT_DIR)"

## start: Start the application (for server)
.PHONY: start
start: export DUBBO_GO_CONFIG_PATH ?= $(OUT_DIR)/server/server.yml
start: build
	$(info   >  Starting application $(PROJECT_NAME), output is redirected to $(LOG_FILE))
	@if [ -f "$(PID)" ] && kill -0 "$$(cat "$(PID)")" 2>/dev/null; then echo "server PID file already refers to a running process" >&2; exit 1; fi
	@$(OUT_DIR)/server/$(PROJECT_NAME)$(EXT_NAME) > $(LOG_FILE) 2>&1 & echo $$! > $(PID)
	@cat $(PID) | sed "/^/s/^/  \>  PID: /"


## start: Start the application (for server) foreground
.PHONY: startServer
startServer: export DUBBO_GO_CONFIG_PATH ?= $(OUT_DIR)/server/server.yml
startServer: build
	$(info   >  Starting application $(PROJECT_NAME), output is redirected to $(LOG_FILE))
	@$(OUT_DIR)/server/$(PROJECT_NAME)$(EXT_NAME)  2>&1

## run: Run the application (for client)
.PHONY: run
run: buildPixiu
	$(info   >  Running application PIXIU, output is redirected to $(LOG_FILE))

## buildPixiuBinary: Build Pixiu without starting it
.PHONY: buildPixiuBinary
buildPixiuBinary:
	@echo "buildPixiu CGO=$(CGO)"
	@cd "$(PIXIU_DIR)" && CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GCFLAGS) -ldflags=$(LDFLAGS) -o "$(OUT_DIR)/dubbo-go-pixiu$(EXT_NAME)" $(PIXIU_BUILD_PACKAGE)

## buildPixiu: start pixiu
.PHONY: buildPixiu
buildPixiu: buildPixiuBinary
	@if [ -f "$(PIXIU_PID)" ] && kill -0 "$$(cat "$(PIXIU_PID)")" 2>/dev/null; then echo "Pixiu PID file already refers to a running process" >&2; exit 1; fi
	@$(OUT_DIR)/dubbo-go-pixiu$(EXT_NAME) gateway start -a $(API_CONFIG_PATH) -c $(CONFIG_PATH) & echo $$! > $(PIXIU_PID)
	@cat $(PIXIU_PID) | sed "/^/s/^/  \> PIXIU_PID: /"


## startPixiu: start pixiu foreground
.PHONY: startPixiu
startPixiu: buildPixiuBinary
	@$(OUT_DIR)/dubbo-go-pixiu$(EXT_NAME) gateway start -a $(API_CONFIG_PATH) -c $(CONFIG_PATH)

## stop: Stop running the application (for server)
.PHONY: stop
stop:
	$(info   >  Stopping the application $(PROJECT_NAME))
	@$(call stop_owned_process,$(PID),$(OUT_DIR)/server/$(PROJECT_NAME)$(EXT_NAME))

## integration: Run integration test for this application
.PHONY: integration
integration:
	$(info   >  Running integration test for application $(PROJECT_NAME))
	@go clean -testcache
	@go test -tags "$(TEST_TAGS)" -run "$(TEST_RUN)" -v $(PROJECT_DIR)/test/...
