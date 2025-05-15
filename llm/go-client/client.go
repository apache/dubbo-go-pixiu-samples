/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

import (
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("go-client/.env")
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	url := "http://localhost:8888/chat/completions"
	data := `{"model":"deepseek-chat","messages":[{"role": "user", "content": "3+5=?"}],"temperature":0.8,"stream":true}`
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	key := os.Getenv("API_KEY")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+key)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	var eventData strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading response:", err)
				return
			}
			return
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "data:") {
			dataLine := strings.TrimPrefix(line, "data:")
			eventData.WriteString(dataLine)
		} else if line == "" {
			if eventData.Len() > 0 {
				dataStr := eventData.String()
				var currentParsedData struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}
				err := json.Unmarshal([]byte(dataStr), &currentParsedData)
				if err != nil {
					continue
				} else if len(currentParsedData.Choices) > 0 && len(currentParsedData.Choices[0].Delta.Content) > 0 {
					fmt.Print(currentParsedData.Choices[0].Delta.Content)
					// logger.Info()
				}
				eventData.Reset()
			}
		}
	}
}
