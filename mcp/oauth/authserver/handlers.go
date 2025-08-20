package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

// tokenResponse defines the structure of the JSON response from the token endpoint.
type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

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

// dynamicClientRegistrationRequest represents the RFC 7591 minimal client metadata accepted by /register.
type dynamicClientRegistrationRequest struct {
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
}

// handleDynamicClientRegistration implements a minimal RFC 7591 dynamic client registration endpoint.
func handleDynamicClientRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	var req dynamicClientRegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	// Basic validation: require at least one redirect URI and simple scheme check.
	if len(req.RedirectURIs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uris", "error_description": "redirect_uris must be provided"})
		return
	}
	for _, ru := range req.RedirectURIs {
		if !(strings.HasPrefix(ru, "http://") || strings.HasPrefix(ru, "https://")) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uri", "error_description": "redirect_uris must be absolute http(s) URLs"})
			return
		}
	}

	// Default token endpoint auth method
	if req.TokenEndpointAuthMethod == "" {
		req.TokenEndpointAuthMethod = "none"
	}

	clientID := generateRandomString(16)
	var clientSecret string
	if req.TokenEndpointAuthMethod != "none" {
		clientSecret = generateRandomString(32)
	}

	now := time.Now().Unix()
	client := ClientInfo{
		ID:                      clientID,
		Secret:                  clientSecret,
		RedirectURIs:            req.RedirectURIs,
		TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
		ClientIDIssuedAt:        now,
	}

	clients[clientID] = client

	issuer := "http://localhost" + listenAddr
	regURI := issuer + "/register/" + clientID

	resp := map[string]interface{}{
		"client_id":                  client.ID,
		"redirect_uris":              client.RedirectURIs,
		"client_id_issued_at":        client.ClientIDIssuedAt,
		"token_endpoint_auth_method": client.TokenEndpointAuthMethod,
		"registration_client_uri":    regURI,
	}
	if client.Secret != "" {
		resp["client_secret"] = client.Secret
	}

	w.Header().Set("Location", regURI)
	writeJSON(w, http.StatusCreated, resp)
}

func handleMetadata(w http.ResponseWriter, r *http.Request) {
	issuer := "http://localhost" + listenAddr
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

func handleAuthorize(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	clientID := query.Get("client_id")
	redirectURI := query.Get("redirect_uri")
	responseType := query.Get("response_type")
	codeChallenge := query.Get("code_challenge")
	codeChallengeMethod := query.Get("code_challenge_method")
	resource := query.Get("resource")
	state := query.Get("state") // Preserve state parameter

	// Validate client
	client, ok := clients[clientID]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_client"})
		return
	}
	// Validate matching redirect URI against registered redirect_uris
	matched := false
	for _, ru := range client.RedirectURIs {
		if ru == redirectURI {
			matched = true
			break
		}
	}
	if !matched {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_redirect_uri"})
		return
	}

	// Validate request parameters
	if responseType != "code" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_response_type"})
		return
	}
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "error_description": "code_challenge required and must be S256"})
		return
	}

	// Require resource parameter
	if resource == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "error_description": "resource parameter required"})
		return
	}

	// In a real server, this is where you would authenticate the user and ask for consent.
	// For this demo, we auto-approve.

	// Generate and store authorization code
	code := generateRandomString(32)
	authCodes[code] = AuthCodeInfo{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Resource:            resource,
		Expiry:              time.Now().Add(10 * time.Minute),
	}

	// Redirect back to the client
	redirectURL := fmt.Sprintf("%s?code=%s", redirectURI, code)
	if state != "" {
		redirectURL += "&state=" + state
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	// Validate grant type
	grantType := r.PostForm.Get("grant_type")
	if grantType != "authorization_code" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
		return
	}

	// Validate authorization code
	code := r.PostForm.Get("code")
	authCode, ok := authCodes[code]
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant"})
		return
	}
	if time.Now().After(authCode.Expiry) {
		delete(authCodes, code) // Clean up expired code
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant", "error_description": "authorization code expired"})
		return
	}

	// Validate client and redirect URI
	clientID := r.PostForm.Get("client_id")
	if clientID != authCode.ClientID {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_client"})
		return
	}

	// Require resource parameter and verify it matches the one associated with the auth code
	resource := r.PostForm.Get("resource")
	if resource == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request", "error_description": "resource parameter required"})
		return
	}
	if resource != authCode.Resource {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant", "error_description": "resource mismatch"})
		return
	}

	// Perform PKCE validation
	codeVerifier := r.PostForm.Get("code_verifier")
	if !validatePKCE(authCode.CodeChallenge, codeVerifier) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_grant", "error_description": "PKCE verification failed"})
		return
	}

	// All checks passed, clean up the auth code
	delete(authCodes, code)

	// Issue JWT
	accessToken, err := issueJWT(authCode.Resource, "") // Scope can be added here if needed
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server_error", "error_description": "failed to issue token"})
		return
	}

	// Return the token
	resp := tokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(tokenTTL.Seconds()),
	}
	writeJSON(w, http.StatusOK, resp)
}

// validatePKCE performs the S256 PKCE challenge verification.
func validatePKCE(challenge, verifier string) bool {
	hasher := sha256.New()
	hasher.Write([]byte(verifier))
	calculatedChallenge := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))
	return calculatedChallenge == challenge
}

// writeJSON is a helper to write JSON responses.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// generateRandomString creates a secure random string of a given length.
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// In a real application, this should be handled more gracefully.
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

// corsMiddleware wraps an http.Handler and sets permissive CORS headers.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
