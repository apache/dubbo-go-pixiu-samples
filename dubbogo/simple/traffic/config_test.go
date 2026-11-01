/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package traffic

import (
	"os"
	"path/filepath"
	"testing"
)

import (
	"gopkg.in/yaml.v3"
)

func TestLoadConfigDefaultsAndWildcardEndpoint(t *testing.T) {
	config, err := loadConfig(func(string) (string, bool) { return "", false })
	if err != nil {
		t.Fatal(err)
	}
	if config.V1.ListenAddr != ":1315" || config.V1.Endpoint != "127.0.0.1:1315" || config.GatewayURL != defaultGateway || config.GatewayPort != 8888 {
		t.Fatalf("unexpected defaults: %+v", config)
	}
	config, err = loadConfig(func(key string) (string, bool) {
		if key == V1AddrEnv {
			return "0.0.0.0:11315", true
		}
		return "", false
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.V1.ListenAddr != "0.0.0.0:11315" || config.V1.Endpoint != "127.0.0.1:11315" {
		t.Fatalf("wildcard listener must render as a loopback endpoint: %+v", config.V1)
	}
}

func TestLoadConfigRejectsInvalidAddresses(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "address without port", key: V1AddrEnv, value: "127.0.0.1"},
		{name: "URL used as address", key: V1AddrEnv, value: "http://127.0.0.1:11315"},
		{name: "port out of range", key: V1AddrEnv, value: "127.0.0.1:65536"},
		{name: "zero port", key: V1AddrEnv, value: ":0"},
		{name: "malformed backend host", key: V1AddrEnv, value: "bad@host:11315"},
		{name: "malformed backend IPv6", key: V1AddrEnv, value: "[not:ipv6]:11315"},
		{name: "duplicate listener", key: V2AddrEnv, value: ":1315"},
		{name: "overlapping wildcard listener", key: V2AddrEnv, value: "0.0.0.0:1315"},
		{name: "gateway path", key: GatewayURLEnv, value: "http://127.0.0.1:18888/user"},
		{name: "gateway missing port", key: GatewayURLEnv, value: "http://127.0.0.1"},
		{name: "gateway non-http scheme", key: GatewayURLEnv, value: "https://127.0.0.1:18888"},
		{name: "gateway query", key: GatewayURLEnv, value: "http://127.0.0.1:18888?"},
		{name: "gateway fragment", key: GatewayURLEnv, value: "http://127.0.0.1:18888#"},
		{name: "malformed gateway IPv6", key: GatewayURLEnv, value: "http://[not:ipv6]:18888"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadConfig(func(key string) (string, bool) {
				if key == test.key {
					return test.value, true
				}
				return "", false
			})
			if err == nil {
				t.Fatalf("expected %s=%q to be rejected", test.key, test.value)
			}
		})
	}
}

func TestPrepareConfigKeepsDefaultsAndRendersOverrides(t *testing.T) {
	t.Setenv(V1AddrEnv, defaultV1Addr)
	t.Setenv(V2AddrEnv, defaultV2Addr)
	t.Setenv(V3AddrEnv, defaultV3Addr)
	t.Setenv(GatewayURLEnv, defaultGateway)

	type pixiuConfig struct {
		StaticResources struct {
			Listeners []struct {
				Address struct {
					Socket struct {
						Port int `yaml:"port"`
					} `yaml:"socket_address"`
				} `yaml:"address"`
			} `yaml:"listeners"`
			Clusters []struct {
				Name      string `yaml:"name"`
				Endpoints []struct {
					Socket struct {
						Address string `yaml:"address"`
						Port    int    `yaml:"port"`
					} `yaml:"socket_address"`
				} `yaml:"endpoints"`
			} `yaml:"clusters"`
		} `yaml:"static_resources"`
	}

	for _, configName := range []string{"conf.yaml", "header-conf.yaml"} {
		t.Run(configName, func(t *testing.T) {
			source := filepath.Join("pixiu", configName)
			original, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			defaultOutput := filepath.Join(t.TempDir(), "default", configName)
			if err := PrepareConfig(source, defaultOutput); err != nil {
				t.Fatal(err)
			}
			defaultBytes, err := os.ReadFile(defaultOutput)
			if err != nil {
				t.Fatal(err)
			}
			if string(defaultBytes) != string(original) {
				t.Fatal("default configuration should be copied byte-for-byte")
			}

			t.Setenv(V1AddrEnv, "127.0.0.1:11315")
			t.Setenv(V2AddrEnv, "127.0.0.1:11316")
			t.Setenv(V3AddrEnv, "127.0.0.1:11317")
			t.Setenv(GatewayURLEnv, "http://127.0.0.1:18888")
			output := filepath.Join(t.TempDir(), "build", configName)
			if err := PrepareConfig(source, output); err != nil {
				t.Fatal(err)
			}
			if got, err := os.ReadFile(source); err != nil {
				t.Fatal(err)
			} else if string(got) != string(original) {
				t.Fatal("rendering must not modify the source YAML")
			}
			rendered, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			var parsed pixiuConfig
			if err := yaml.Unmarshal(rendered, &parsed); err != nil {
				t.Fatalf("rendered YAML is invalid: %v", err)
			}
			if len(parsed.StaticResources.Listeners) != 1 || parsed.StaticResources.Listeners[0].Address.Socket.Port != 18888 {
				t.Fatalf("gateway listener port was not rendered: %+v", parsed.StaticResources.Listeners)
			}
			wantPorts := map[string]int{"user-v1": 11315, "user-v2": 11316, "user-v3": 11317}
			for _, cluster := range parsed.StaticResources.Clusters {
				if wantPort, ok := wantPorts[cluster.Name]; ok {
					if len(cluster.Endpoints) != 1 || cluster.Endpoints[0].Socket.Address != "127.0.0.1" || cluster.Endpoints[0].Socket.Port != wantPort {
						t.Errorf("cluster %s endpoint not rendered correctly: %+v", cluster.Name, cluster.Endpoints)
					}
					delete(wantPorts, cluster.Name)
				}
			}
			if len(wantPorts) != 0 {
				t.Errorf("missing rendered clusters: %v", wantPorts)
			}
		})
	}
}

func TestPrepareConfigRejectsInvalidConfiguration(t *testing.T) {
	t.Setenv(V1AddrEnv, "localhost:1315/path")
	if err := PrepareConfig(filepath.Join("pixiu", "conf.yaml"), filepath.Join(t.TempDir(), "conf.yaml")); err == nil {
		t.Fatal("invalid backend address must prevent config generation")
	}
}
