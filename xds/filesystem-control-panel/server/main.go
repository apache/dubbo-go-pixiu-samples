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

package main

import (
	"context"
	"flag"
	"github.com/dubbo-go-pixiu/pixiu-api/pkg/xds"
	pixiupb "github.com/dubbo-go-pixiu/pixiu-api/pkg/xds/model"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"github.com/fsnotify/fsnotify"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

var (
	l       Logger
	port    = uint(18000)
	nodeID  = "test-id"
	version = 100
)

func init() {
	l = Logger{}
	l.Debug = true
}

func main() {
	flag.Parse()

	// Create a snaphost
	snaphost := cache.NewSnapshotCache(false, cache.IDHash{}, l)

	go func() {
		// Create the config that we'll serve to Envoy
		config := GenerateSnapshotPixiuFromFile()

		if err := config.Consistent(); err != nil {
			l.Errorf("config inconsistency: %+v\n%+v", config, err)
			os.Exit(1)
		}
		l.Debugf("will serve config %+v", config)

		// Add the config to the snaphost
		if err := snaphost.SetSnapshot(context.Background(), nodeID, config); err != nil {
			l.Errorf("config error %q for %+v", err, config)
			os.Exit(1)
		}

		go func() {
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				log.Fatal(err)
			}
			defer watcher.Close()
			done := make(chan bool)
			go func() {
				for {
					select {
					case event, ok := <-watcher.Events:
						if !ok {
							return
						}
						log.Println("event:", event)
						if event.Op&fsnotify.Write == fsnotify.Write {
							log.Println("modified file:", event.Name)
							conf := GenerateSnapshotPixiuFromFile()
							// Add the config to the snaphost
							if err := snaphost.SetSnapshot(context.Background(), nodeID, conf); err != nil {
								l.Errorf("config error %q for %+v", err, config)
								os.Exit(1)
							}
						}
					case err, ok := <-watcher.Errors:
						if !ok {
							return
						}
						log.Println("error:", err)
					}
				}
			}()

			err = watcher.Add("../pixiu")
			if err != nil {
				log.Fatal(err)
			}
			<-done
		}()
	}()

	// Run the xDS server
	ctx := context.Background()
	cb := &test.Callbacks{Debug: l.Debug}
	srv := server.NewServer(ctx, snaphost, cb)
	RunServer(ctx, srv, port)
}

func GenerateSnapshotPixiuFromFile() cache.Snapshot {
	cdsStr, err := ioutil.ReadFile("../pixiu/cds.json")
	if err != nil {
		l.Errorf("%s", err)
	}
	cds := &pixiupb.PixiuExtensionClusters{}
	err = protojson.Unmarshal(cdsStr, cds)
	if err != nil {
		l.Errorf("%s", err)
	}

	ldsStr, err := ioutil.ReadFile("../pixiu/lds.json")
	if err != nil {
		l.Errorf("%s", err)
	}

	lds := &pixiupb.PixiuExtensionListeners{}
	err = protojson.Unmarshal(ldsStr, lds)
	if err != nil {
		l.Errorf("%s", err)
	}

	version++
	ldsResource, _ := anypb.New(lds)
	cdsResource, _ := anypb.New(cds)
	snap, _ := cache.NewSnapshot(strconv.Itoa(version),
		map[resource.Type][]types.Resource{
			resource.ExtensionConfigType: {
				&core.TypedExtensionConfig{
					Name:        xds.ClusterType,
					TypedConfig: cdsResource,
				},
				&core.TypedExtensionConfig{
					Name:        xds.ListenerType,
					TypedConfig: ldsResource,
				},
			},
		},
	)
	return snap
}
