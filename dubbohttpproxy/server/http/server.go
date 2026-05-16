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
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/apache/dubbo-go-pixiu/pkg/common/constant"
)

func main() {
	http.HandleFunc("/com.dubbogo.pixiu.TripleUserService/GetUserById", user)
	log.Println("Starting sample server ...")
	log.Fatal(http.ListenAndServe("127.0.0.1:20001", nil))
}

func user(w http.ResponseWriter, r *http.Request) {
	if r.Method != constant.Post {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	byts, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	// Pixiu's dgp.filter.dubbo.http marshals generic invocation arguments as an array,
	// so the request body is shaped like ["0001"] and should be decoded into []string.
	var args []string
	if err := json.Unmarshal(byts, &args); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if len(args) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing id argument"))
		return
	}
	u, ok := cache.Get(args[0])
	if !ok {
		w.Header().Set(constant.HeaderKeyContextType, constant.HeaderValueJsonUtf8)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"user not found"}`))
		return
	}
	b, _ := json.Marshal(u)
	w.Header().Set(constant.HeaderKeyContextType, constant.HeaderValueJsonUtf8)
	w.Write(b)
}
