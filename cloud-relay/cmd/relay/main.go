// Package main provides the entrypoint for the cloud relay daemon.
//
// The relay daemon listens on localhost:10025 for SMTP connections from
// Postfix and forwards mail to the home device via WireGuard or mTLS transport.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/config"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/smtp"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting DarkPipe cloud relay daemon...")

	// Load configuration from environment
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config: listen=%s transport=%s home=%s", cfg.ListenAddr, cfg.TransportType, cfg.HomeDeviceAddr)

	// Create appropriate forwarder based on transport type
	var forwarder forward.Forwarder
	if cfg.TransportType == "mtls" {
		log.Println("Initializing mTLS forwarder...")
		forwarder, err = forward.NewMTLSForwarder(
			cfg.CACertPath,
			cfg.ClientCertPath,
			cfg.ClientKeyPath,
			cfg.HomeDeviceAddr,
		)
		if err != nil {
			log.Fatalf("Failed to create mTLS forwarder: %v", err)
		}
	} else {
		log.Println("Initializing WireGuard forwarder...")
		forwarder = forward.NewWireGuardForwarder(cfg.HomeDeviceAddr)
	}
	defer forwarder.Close()

	// Create and start SMTP server
	server := smtp.NewServer(forwarder, cfg)
	log.Printf("Relay daemon listening on %s (forwarding to %s via %s)", cfg.ListenAddr, cfg.HomeDeviceAddr, cfg.TransportType)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutdown signal received, stopping server...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server (blocks until shutdown)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Relay daemon stopped")
}
