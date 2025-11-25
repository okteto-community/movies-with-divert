package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"fmt"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("Running server on port 8080...")

	muxRouter := mux.NewRouter().StrictSlash(true)
	muxRouter.Use(PropagateBaggageHeaderMiddleware)
	muxRouter.Use(LoggingMiddleware)

	muxRouter.Handle("/rent", NewProxy("http://rent:8080"))
	muxRouter.Handle("/catalog", NewProxy("http://catalog:8080"))

	log.Fatal(http.ListenAndServe(":8080", muxRouter))
}

// NewProxy creates a reverse proxy handler for a given target URL.
func NewProxy(target string) http.Handler {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Error parsing target URL: %v", err)
	}
	return httputil.NewSingleHostReverseProxy(targetURL)
}

// LoggingMiddleware logs incoming requests.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// PropagateBaggageHeaderMiddleware propages the baggage header
func PropagateBaggageHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := r.Header.Get("baggage")
		if b != "" {
			w.Header().Add("baggage", b)
		}

		next.ServeHTTP(w, r)
	})
}
