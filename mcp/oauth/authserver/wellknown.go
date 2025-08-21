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
	"encoding/base64"
	"log"
	"math/big"
	"net/http"
)

// jwks represents a JSON Web Key Set.
type jwks struct {
	Keys []jwk `json:"keys"`
}

// jwk represents a single JSON Web Key.
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func handleMetadata(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	issuer := issuerBaseURL // Use shared constant
	meta := map[string]interface{}{
		"issuer":                                issuer,
		"authorization_endpoint":                issuer + "/oauth/authorize",
		"token_endpoint":                        issuer + "/oauth/token",
		"jwks_uri":                              issuer + "/.well-known/jwks.json",
		"registration_endpoint":                 issuer + "/register",
		"grant_types_supported":                 []string{"authorization_code"},
		"response_types_supported":              []string{"code"},
		"token_endpoint_auth_methods_supported": []string{"none"}, // PKCE does not require a client secret
		"code_challenge_methods_supported":      []string{"S256"},
	}
	writeJSON(w, http.StatusOK, meta)
}

func handleJwks(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
	pubKey := privKey.PublicKey
	key := jwk{
		Kty: "RSA",
		Kid: keyID,
		Use: "sig",
		Alg: "RS256",
		N:   base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes()),
	}
	writeJSON(w, http.StatusOK, jwks{Keys: []jwk{key}})
}
