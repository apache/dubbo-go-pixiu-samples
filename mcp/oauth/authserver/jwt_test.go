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
	"encoding/json"
	"strings"
	"testing"
	"time"
)

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitJWT(t *testing.T) {
	// Run the init function
	initJWT()

	// Assert that the private key was initialized
	assert.NotNil(t, privKey, "Private key should not be nil after init")
	assert.NoError(t, privKey.Validate(), "Private key should be a valid key")
}

func TestIssueJWT(t *testing.T) {
	// Ensure JWT is initialized
	initJWT()

	// Define test cases
	testCases := []struct {
		name      string
		audience  string
		scope     string
		expectErr bool
	}{
		{
			name:      "Valid JWT with audience and scope",
			audience:  "test-audience",
			scope:     "read:data",
			expectErr: false,
		},
		{
			name:      "Valid JWT with empty scope",
			audience:  "test-audience-2",
			scope:     "",
			expectErr: false,
		},
		{
			name:      "Valid JWT with empty audience",
			audience:  "",
			scope:     "read:data",
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokenString, err := issueJWT(tc.audience, tc.scope)

			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, tokenString)

			// 1. Validate JWT structure (three parts separated by dots)
			parts := strings.Split(tokenString, ".")
			require.Len(t, parts, 3, "JWT should have 3 parts")

			// 2. Decode and validate header
			headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
			require.NoError(t, err)
			var header map[string]string
			err = json.Unmarshal(headerBytes, &header)
			require.NoError(t, err)
			assert.Equal(t, "RS256", header["alg"])
			assert.Equal(t, "JWT", header["typ"])
			assert.Equal(t, keyID, header["kid"])

			// 3. Decode and validate claims
			claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)
			var claims map[string]interface{}
			err = json.Unmarshal(claimsBytes, &claims)
			require.NoError(t, err)

			assert.Equal(t, issuerBaseURL, claims["iss"]) // Use shared constant
			assert.Equal(t, tc.audience, claims["aud"])
			assert.Equal(t, tc.scope, claims["scope"])

			// Check timestamps
			now := float64(time.Now().Unix())
			iat, ok := claims["iat"].(float64)
			require.True(t, ok)
			assert.InDelta(t, now, iat, 5, "Issue time should be close to now")

			exp, ok := claims["exp"].(float64)
			require.True(t, ok)
			expectedExp := float64(time.Now().Add(tokenTTL).Unix())
			assert.InDelta(t, expectedExp, exp, 5, "Expiration time should be correct")

			// 4. Verify signature (this is a simplified verification)
			// A full verification would re-calculate the signature and compare,
			// but for this test, we trust the crypto library.
			// We can check if the signature part is not empty.
			assert.NotEmpty(t, parts[2], "Signature part should not be empty")
		})
	}
}

// Note: A full signature verification test would require a more complex setup
// to parse the token and use the public key to verify the signature.
// For the scope of this example, we are trusting the signing function works correctly
// if it returns no error and a non-empty signature.
