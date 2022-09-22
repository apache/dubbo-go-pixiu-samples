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
	"fmt"
	"log"
	"net/http"
)

func main() {
	router := []string{"/user1", "/user2", "/user3", "/user4"}
	data := GetRequests()
	for k, v := range data.Requests {
		http.HandleFunc(router[k], func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(fmt.Sprintf(`{"method":"%s","url":%s}`, v.Method, v.URL)))
		})
	}
	log.Println("Starting sample server ...")
	log.Fatal(http.ListenAndServe(":1314", nil))
}

func GetRequests() APIRequests {
	return APIRequests{
		Requests: []Request{
			{
				Method: "POST",
				URL:    "/mock/test",
			},
			{
				Method: "GET",
				URL:    "/mock/test",
			},
			{
				Method: "PUT",
				URL:    "/mock/test",
			},
			{
				Method: "DELETE",
				URL:    "/mock/test",
			},
		},
	}
}

type APIRequests struct {
	Requests []Request `json:"requests"`
}

type Request struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}
