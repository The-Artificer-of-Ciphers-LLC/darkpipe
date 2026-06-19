// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"log"
	"os"

	"github.com/darkpipe/darkpipe/monitoring/status"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "qr":
			RunQRCommand(os.Args[2:])
			return
		case "status":
			status.RunStatusCommand(os.Args[2:])
			return
		}
	}

	if err := runProfileRuntime(); err != nil {
		log.Fatalf("Profile server error: %v", err)
	}
}
