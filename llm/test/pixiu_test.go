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

package test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/apache/dubbo-go-pixiu/pkg/common/constant"

	"github.com/joho/godotenv"

	"github.com/stretchr/testify/assert"
)

func TestPostFixed(t *testing.T) {
	err := godotenv.Load(".env")
	assert.NoError(t, err)
	url := "http://localhost:8888/chat/completions"
	data := `{"model":"deepseek-chat","messages":[{"role": "user", "content": "3+5=?"}],"temperature":0.8,"stream":true}`
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	assert.NoError(t, err)

	key := os.Getenv("API_KEY")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+key)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get(constant.HeaderKeyContextType), "text/event-stream")

	reader := bufio.NewReader(resp.Body)
	var eventData strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			assert.NoError(t, err)
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
	fmt.Println()
}
