// cmd/server/main.go — Entry point with architectural violations.
package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"super-lint-test/internal/handler"
	"super-lint-test/internal/repository/postgres"  // <-- VIOLATION: cmd imports internal subdirectory
	"super-lint-test/pkg/logging"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	logger := logging.NewLogger()

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// main.go directly instantiating repository - should use service layer
	userRepo := postgres.NewUserRepo(db, logger)  // <-- VIOLATION: Bypassing service layer

	// Handler wired directly to repository (no service layer!)
	userHandler := handler.NewUserHandler(userRepo, logger)  // Type mismatch, but shows intent

	mux := http.NewServeMux()
	mux.HandleFunc("/users", userHandler.CreateUser)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		logger.Info("server starting", "port", 8080)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
}
