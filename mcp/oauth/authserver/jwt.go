package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"
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
		"iss":   "http://localhost:9000", // Hardcoded issuer
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
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	sigEnc := base64.RawURLEncoding.EncodeToString(sig)

	return signingInput + "." + sigEnc, nil
}
