package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/getsavvyinc/savvy-cli/extension"
)

func main() {
	// Create a context that we'll cancel on signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create and start the server
	processor := func(items []extension.HistoryItem) error {
		for _, item := range items {
			fmt.Printf("Processing item: %s\n", item.URL)
		}
		return nil
	}

	server := extension.New(processor)
	if err := server.Start(ctx); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}

	// Wait for signal
	sig := <-sigChan
	fmt.Printf("\nReceived signal: %v\n", sig)
	fmt.Println("Shutting down server...")

	// Cancel context to trigger graceful shutdown
	cancel()

	// Wait for server to close
	if err := server.Close(); err != nil {
		fmt.Printf("Error during shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server shutdown complete")
}
