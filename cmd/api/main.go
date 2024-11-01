package main

import (
	"errors"
	"log"
	"net/http"

	"pdf-maker/internal/server"
	"pdf-maker/internal/utils"
	// "pdf-maker/internal/server"
)

func main() {
	utils.LoadDbEnv(".env") // Load database credentials to env
	srv := server.NewServer()

	shutdownDone := make(chan bool, 1)            // Create a done channel to signal when the shutdown is complete
	go server.GracefulShutdown(srv, shutdownDone) // Run graceful shutdown in a separate goroutine
	defer func() {
		<-shutdownDone // Wait for the graceful shutdown to complete
		log.Println("Graceful shutdown complete.")
	}()

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server error: %s", err)
		}
	}()
}
