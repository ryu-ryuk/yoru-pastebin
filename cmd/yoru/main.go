package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryu-ryuk/yoru-pastebin/internal/config"
	"github.com/ryu-ryuk/yoru-pastebin/internal/database"
	"github.com/ryu-ryuk/yoru-pastebin/internal/server"
)

// main function initlializes the application, loads configuration, connects to the database, starts the HTTP server,
// and handles proper shutdown.

// This is the entry point of the Yoru Pastebin application.
// it uses the config package to load configuration settings, the database package to establish a connection to the database,
// and the server package to start the HTTP server.
func main() {
	// load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// log the configuration for debugging purposes
	// log.Printf("Configuration loaded: %+v", cfg)

	// connecting to the database
	db, err := database.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// defer db.Close() // will close db pool explicitly on shutdown

	// start the HTTP server
	httpServer := server.NewServer(cfg, db)

	// start the HTTP server in a goroutine
	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed to start: %v", err)
		}
	}()

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) //  listen for Ctrl+C and termination signals

	<-sigChan // block until a signal is received

	log.Println("Received shutdown signal. Shutting down gracefully...")

	// build a context with a timeout for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP server graceful shutdown failed: %v", err)
	}
	db.Close() // close database pool after server shutdown
	log.Println("Yoru Pastebin shut down.")
}
