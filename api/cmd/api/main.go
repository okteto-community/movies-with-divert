package main

import (
	"log"
	"net/http"

	"fmt"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/okteto/movies/handlers"
	"github.com/okteto/movies/middleware"
)

func main() {
	muxRouter := mux.NewRouter().StrictSlash(true)
	muxRouter.Use(middleware.BaggageMiddleware)
	muxRouter.Use(middleware.LoggingMiddleware)

	// define the public API
	muxRouter.Handle("/api/rent/return", http.StripPrefix("/api", handlers.NewProxy("http://rent:8080"))).Methods(http.MethodPost)
	muxRouter.Handle("/api/rent", http.StripPrefix("/api", handlers.NewProxy("http://rent:8080"))).Methods(http.MethodPost)
	muxRouter.HandleFunc("/api/rent", handlers.GetRentalsWithCatalogInfo).Methods(http.MethodGet)

	muxRouter.Handle("/api/catalog/healthz", http.StripPrefix("/api", handlers.NewProxy("http://catalog:8080/healthz"))).Methods(http.MethodGet)
	muxRouter.Handle("/api/catalog", http.StripPrefix("/api", handlers.NewProxy("http://catalog:8080"))).Methods(http.MethodGet)

	fmt.Println("Running server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", muxRouter))
}
