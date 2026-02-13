// main.go — Chi server for inventory management API
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// Initialize store
	store := NewStore()

	// Routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/items", GetItems(store))
		r.Post("/items", CreateItem(store))
		r.Get("/items/{id}", GetItem(store))
		r.Patch("/items/{id}", UpdateItem(store))
		r.Delete("/items/{id}", DeleteItem(store))
	})

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
