// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/onboarding"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

// RunQRCommand implements the CLI QR code generation command.
func RunQRCommand(args []string) {
	fs := flag.NewFlagSet("qr", flag.ExitOnError)
	outPath := fs.String("out", "", "Write QR PNG to file")
	pngPath := fs.String("png", "", "Deprecated alias for --out")
	profileServerURL := fs.String("server", os.Getenv("PROFILE_SERVER_URL"), "Profile server URL (default: from PROFILE_SERVER_URL env var)")
	jsonOutput := fs.Bool("json", false, "Output JSON metadata")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s qr <email@domain> [--out <file>] [--json] [--server <url>]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generates a QR code onboarding artifact.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  email       Email address for device setup\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}
	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(1)
	}

	email := fs.Arg(0)
	if *outPath == "" && *pngPath != "" {
		*outPath = *pngPath
	}

	if *profileServerURL != "" {
		log.Println("HTTP API mode not yet implemented, using standalone mode")
	}

	hostname := os.Getenv("MAIL_HOSTNAME")
	if hostname == "" {
		hostname = "mail.example.com"
		log.Printf("WARNING: MAIL_HOSTNAME not set, using default: %s", hostname)
	}

	tokenStore := qrcode.NewMemoryTokenStore()
	module := onboarding.New(nil, tokenStore, nil, onboarding.Config{Hostname: hostname})

	profileURL, tokenExpiry, err := module.GenerateQRURL(email)
	if err != nil {
		log.Fatalf("Failed to generate QR URL: %v", err)
	}

	if *outPath != "" {
		pngData, err := qrcode.GenerateQRCodePNG(profileURL, 256)
		if err != nil {
			log.Fatalf("Failed to generate QR code PNG: %v", err)
		}
		if err := os.WriteFile(*outPath, pngData, 0o644); err != nil {
			log.Fatalf("Failed to write QR code file: %v", err)
		}
	}

	if *jsonOutput {
		payload := map[string]any{
			"email":       email,
			"url":         profileURL,
			"expires_at":  tokenExpiry.Format(time.RFC3339),
			"expires_in":  time.Until(tokenExpiry).Round(time.Second).String(),
			"output_path": *outPath,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
		return
	}

	if *outPath != "" {
		fmt.Printf("QR code saved to: %s\n", *outPath)
		fmt.Printf("URL: %s\n", profileURL)
		fmt.Printf("Token expires: %s (%s)\n", tokenExpiry.Format(time.RFC3339), time.Until(tokenExpiry).Round(time.Second))
		return
	}

	terminalQR, err := qrcode.GenerateQRCodeTerminal(profileURL)
	if err != nil {
		log.Fatalf("Failed to generate terminal QR code: %v", err)
	}
	fmt.Println()
	fmt.Println(terminalQR)
	fmt.Println()
	fmt.Printf("URL: %s\n", profileURL)
	fmt.Printf("Token expires: %s (%s)\n", tokenExpiry.Format(time.RFC3339), time.Until(tokenExpiry).Round(time.Second))
	fmt.Println()
	fmt.Println("Instructions:")
	fmt.Println("  1. Scan this QR code with your phone camera")
	fmt.Println("  2. Follow the prompts to install the mail profile")
	fmt.Println("  3. Your device will be configured automatically")
	fmt.Println()
	fmt.Println("Note: This token can only be used once and expires in 15 minutes")
}

func init() {}
