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
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

import (
	"gopkg.in/yaml.v3"
)

const (
	V1AddrEnv      = "TRAFFIC_V1_ADDR"
	V2AddrEnv      = "TRAFFIC_V2_ADDR"
	V3AddrEnv      = "TRAFFIC_V3_ADDR"
	GatewayURLEnv  = "TRAFFIC_GATEWAY_URL"
	defaultV1Addr  = ":1315"
	defaultV2Addr  = ":1316"
	defaultV3Addr  = ":1317"
	defaultGateway = "http://localhost:8888"
)

type Backend struct {
	ListenAddr string
	Endpoint   string
	Port       int
}

type Config struct {
	V1          Backend
	V2          Backend
	V3          Backend
	GatewayURL  string
	GatewayPort int
}

// LoadConfig parses and validates the traffic sample's shared environment.
func LoadConfig() (Config, error) {
	return loadConfig(os.LookupEnv)
}

func loadConfig(lookup func(string) (string, bool)) (Config, error) {
	var config Config
	var err error
	if config.V1, err = parseBackend(lookup, V1AddrEnv, defaultV1Addr); err != nil {
		return Config{}, err
	}
	if config.V2, err = parseBackend(lookup, V2AddrEnv, defaultV2Addr); err != nil {
		return Config{}, err
	}
	if config.V3, err = parseBackend(lookup, V3AddrEnv, defaultV3Addr); err != nil {
		return Config{}, err
	}
	if listenAddressesConflict(config.V1, config.V2) || listenAddressesConflict(config.V1, config.V3) || listenAddressesConflict(config.V2, config.V3) {
		return Config{}, fmt.Errorf("traffic backend listen addresses must be distinct: %q, %q, %q", config.V1.ListenAddr, config.V2.ListenAddr, config.V3.ListenAddr)
	}

	config.GatewayURL = defaultGateway
	if value, ok := lookup(GatewayURLEnv); ok {
		config.GatewayURL = value
	}
	config.GatewayPort, err = parseGatewayURL(config.GatewayURL)
	if err != nil {
		return Config{}, fmt.Errorf("invalid %s: %w", GatewayURLEnv, err)
	}
	return config, nil
}

func listenAddressesConflict(a, b Backend) bool {
	if a.Port != b.Port {
		return false
	}
	hostA, _, _ := net.SplitHostPort(a.ListenAddr)
	hostB, _, _ := net.SplitHostPort(b.ListenAddr)
	wildcard := func(host string) bool { return host == "" || host == "0.0.0.0" || host == "::" }
	return hostA == hostB || wildcard(hostA) || wildcard(hostB)
}

func parseBackend(lookup func(string) (string, bool), key, fallback string) (Backend, error) {
	addr := fallback
	if value, ok := lookup(key); ok {
		addr = value
	}
	host, port, err := parseHostPort(addr)
	if err != nil {
		return Backend{}, fmt.Errorf("invalid %s: %w", key, err)
	}
	endpointHost := host
	if endpointHost == "" || endpointHost == "0.0.0.0" || endpointHost == "::" || endpointHost == "[::]" {
		endpointHost = "127.0.0.1"
	}
	return Backend{ListenAddr: addr, Endpoint: net.JoinHostPort(endpointHost, strconv.Itoa(port)), Port: port}, nil
}

func parseHostPort(value string) (string, int, error) {
	if strings.TrimSpace(value) != value || value == "" || strings.Contains(value, "://") || strings.ContainsAny(value, "/?#") {
		return "", 0, fmt.Errorf("expected a host:port address, got %q", value)
	}
	host, portText, err := net.SplitHostPort(value)
	if err != nil {
		return "", 0, fmt.Errorf("expected a host:port address: %w", err)
	}
	if !validHost(host) {
		return "", 0, fmt.Errorf("invalid host in %q", value)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("port must be an integer from 1 to 65535 in %q", value)
	}
	return strings.Trim(host, "[]"), port, nil
}

func validHost(host string) bool {
	if host == "" || net.ParseIP(host) != nil {
		return true
	}
	if address, zone, ok := strings.Cut(host, "%"); ok {
		return zone != "" && !strings.ContainsAny(zone, " \t\r\n/%?#@[]") && strings.Contains(address, ":") && net.ParseIP(address) != nil
	}
	if len(host) > 253 {
		return false
	}
	for _, label := range strings.Split(strings.TrimSuffix(host, "."), ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, char := range label {
			switch {
			case char >= 'a' && char <= 'z', char >= 'A' && char <= 'Z', char >= '0' && char <= '9', char == '-':
			default:
				return false
			}
		}
	}
	return true
}

func parseGatewayURL(value string) (int, error) {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "http" || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || strings.Contains(value, "#") || parsed.Fragment != "" || parsed.RawFragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return 0, fmt.Errorf("expected an http://host:port URL without a path, query, or fragment, got %q", value)
	}
	if !validHost(parsed.Hostname()) {
		return 0, fmt.Errorf("invalid URL host in %q", value)
	}
	port, err := strconv.Atoi(parsed.Port())
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("URL port must be an integer from 1 to 65535 in %q", value)
	}
	return port, nil
}

// PrepareConfig copies the directly usable defaults or renders validated,
// non-default values into the build directory without changing source YAML.
func PrepareConfig(inputPath, outputPath string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	input, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read traffic Pixiu config %q: %w", inputPath, err)
	}
	if config.isDefault() {
		return writeConfig(outputPath, input)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(input, &root); err != nil {
		return fmt.Errorf("parse traffic Pixiu config %q: %w", inputPath, err)
	}
	resources, err := mappingValue(&root, "static_resources")
	if err != nil {
		return err
	}
	listeners, err := mappingValue(resources, "listeners")
	if err != nil || listeners.Kind != yaml.SequenceNode || len(listeners.Content) == 0 {
		return fmt.Errorf("traffic Pixiu config must contain a static_resources listener")
	}
	listenerSocket, err := nestedMapping(listeners.Content[0], "address", "socket_address")
	if err != nil {
		return err
	}
	setScalar(listenerSocket, "port", strconv.Itoa(config.GatewayPort))

	clusters, err := mappingValue(resources, "clusters")
	if err != nil || clusters.Kind != yaml.SequenceNode {
		return fmt.Errorf("traffic Pixiu config must contain static_resources clusters")
	}
	backends := map[string]Backend{"user-v1": config.V1, "user-v2": config.V2, "user-v3": config.V3}
	seen := make(map[string]bool, len(backends))
	for _, cluster := range clusters.Content {
		nameNode, err := mappingValue(cluster, "name")
		if err != nil {
			return err
		}
		backend, ok := backends[nameNode.Value]
		if !ok {
			continue
		}
		socket, err := nestedMapping(cluster, "endpoints")
		if err != nil {
			return err
		}
		if socket.Kind != yaml.SequenceNode || len(socket.Content) == 0 {
			return fmt.Errorf("traffic cluster %q must contain an endpoint", nameNode.Value)
		}
		socket, err = nestedMapping(socket.Content[0], "socket_address")
		if err != nil {
			return err
		}
		host, portText, err := net.SplitHostPort(backend.Endpoint)
		if err != nil {
			return fmt.Errorf("invalid rendered endpoint %q: %w", backend.Endpoint, err)
		}
		setScalar(socket, "address", host)
		setScalar(socket, "port", portText)
		seen[nameNode.Value] = true
	}
	for name := range backends {
		if !seen[name] {
			return fmt.Errorf("traffic Pixiu config is missing cluster %q", name)
		}
	}
	rendered, err := yaml.Marshal(&root)
	if err != nil {
		return fmt.Errorf("render traffic Pixiu config: %w", err)
	}
	return writeConfig(outputPath, rendered)
}

func (c Config) isDefault() bool {
	return c.V1.ListenAddr == defaultV1Addr && c.V2.ListenAddr == defaultV2Addr && c.V3.ListenAddr == defaultV3Addr && c.GatewayURL == defaultGateway
}

func mappingValue(node *yaml.Node, key string) (*yaml.Node, error) {
	if node != nil && node.Kind == yaml.DocumentNode && len(node.Content) == 1 {
		node = node.Content[0]
	}
	if node == nil || node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected YAML mapping while looking for %q", key)
	}
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1], nil
		}
	}
	return nil, fmt.Errorf("traffic Pixiu YAML is missing %q", key)
}

func nestedMapping(node *yaml.Node, keys ...string) (*yaml.Node, error) {
	for _, key := range keys {
		var err error
		node, err = mappingValue(node, key)
		if err != nil {
			return nil, err
		}
	}
	return node, nil
}

func setScalar(node *yaml.Node, key, value string) {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			node.Content[i+1].Value = value
			node.Content[i+1].Tag = scalarTag(key)
			return
		}
	}
	node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, &yaml.Node{Kind: yaml.ScalarNode, Tag: scalarTag(key), Value: value})
}

func scalarTag(key string) string {
	if key == "port" {
		return "!!int"
	}
	return "!!str"
}

func writeConfig(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create traffic config directory: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write traffic Pixiu config %q: %w", path, err)
	}
	return nil
}
