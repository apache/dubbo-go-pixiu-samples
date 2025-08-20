package main

import (
	"log"
	"net/http"
)

const (
	listenAddr = ":9000"
)

func main() {
	// Initialize data stores and JWT keys.
	initStore()
	initJWT()

	// Setup HTTP routes.
	http.HandleFunc("/register", handleDynamicClientRegistration)
	http.HandleFunc("/.well-known/oauth-authorization-server", handleMetadata)
	http.HandleFunc("/.well-known/jwks.json", handleJwks)
	http.HandleFunc("/oauth/authorize", handleAuthorize)
	http.HandleFunc("/oauth/token", handleToken)

	log.Printf("OAuth Authorization Server listening on %s", listenAddr)

	// Start the server.
	if err := http.ListenAndServe(listenAddr, corsMiddleware(http.DefaultServeMux)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
