// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"fmt"
	"os"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/setupwizard"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "darkpipe-setup",
		Short: "DarkPipe interactive setup with live validation",
		Long:  "Interactive setup tool for DarkPipe mail server deployment",
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("darkpipe-setup version %s\n", version)
		},
	}

	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Run interactive setup wizard",
		Long:  "Interactive wizard that collects configuration and generates docker-compose.yml",
		Run:   runSetup,
	}

	rootCmd.AddCommand(versionCmd, setupCmd, migrateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runSetup(cmd *cobra.Command, args []string) {
	setupwizard.NewFlow().Run()
}
