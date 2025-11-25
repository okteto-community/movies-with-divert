package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewProxy creates a reverse proxy handler for a given target URL.
func NewProxy(target string) http.Handler {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Error parsing target URL: %v", err)
	}
	return httputil.NewSingleHostReverseProxy(targetURL)
}
