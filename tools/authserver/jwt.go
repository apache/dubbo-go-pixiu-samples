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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"time"
)

import (
	"github.com/pkg/errors"
)

const (
	keyID    = "demo-key-1"
	tokenTTL = time.Hour
)

var (
	// privKey is the RSA private key generated at startup.
	privKey *rsa.PrivateKey
)

// initJWT generates an ephemeral RSA key for signing tokens.
// In a production environment, keys should be loaded from a secure vault.
func initJWT() {
	var err error
	privKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("failed to generate RSA key: %v", err)
	}
}

// issueJWT creates a new JWT with the given audience and scope.
func issueJWT(audience, scope string) (string, error) {
	header := map[string]string{
		"alg": "RS256",
		"typ": "JWT",
		"kid": keyID,
	}
	headerBytes, _ := json.Marshal(header)
	headerEnc := base64.RawURLEncoding.EncodeToString(headerBytes)

	claims := map[string]interface{}{
		"iss":   issuerBaseURL, // Use shared constant
		"aud":   audience,
		"scope": scope,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(tokenTTL).Unix(),
	}
	claimsBytes, _ := json.Marshal(claims)
	claimsEnc := base64.RawURLEncoding.EncodeToString(claimsBytes)

	signingInput := headerEnc + "." + claimsEnc

	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	digest := hasher.Sum(nil)

	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, digest)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign token")
	}

	sigEnc := base64.RawURLEncoding.EncodeToString(sig)

	return signingInput + "." + sigEnc, nil
}
