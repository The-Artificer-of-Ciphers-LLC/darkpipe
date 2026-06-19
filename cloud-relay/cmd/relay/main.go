// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

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
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting DarkPipe cloud relay daemon...")

	// Load configuration from environment
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config: listen=%s transport=%s home=%s strict_mode=%v webhook=%v",
		cfg.ListenAddr, cfg.TransportType, cfg.HomeDeviceAddr, cfg.StrictModeEnabled, cfg.WebhookURL != "")

	runtime, err := newRelayRuntime(cfg)
	if err != nil {
		log.Fatalf("Failed to assemble relay runtime: %v", err)
	}
	defer func() {
		if err := runtime.Close(); err != nil {
			log.Printf("Runtime cleanup error: %v", err)
		}
	}()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutdown signal received, stopping server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := runtime.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	// Start server (blocks until shutdown)
	if err := runtime.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Relay daemon stopped")
}
