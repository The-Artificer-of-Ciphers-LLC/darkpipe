// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/darkpipe/darkpipe/monitoring/delivery"
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

// startLogParser reads Postfix log lines and feeds them to the delivery tracker.
func startLogParser(ctx context.Context, logPath string, tracker *delivery.DeliveryTracker) {
	parser := &delivery.Parser{}

	// Try to open the log file, retry periodically if not available yet
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		f, err := os.Open(logPath)
		if err != nil {
			log.Printf("Waiting for mail log at %s: %v", logPath, err)
			time.Sleep(30 * time.Second)
			continue
		}

		// Tail the log file
		scanner := newLineScanner(f)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				f.Close()
				return
			default:
			}

			entry, err := parser.ParseLine(scanner.Text())
			if err == nil && entry != nil {
				tracker.Record(entry)
			}
		}

		f.Close()
		// File rotated or closed, wait and retry
		time.Sleep(5 * time.Second)
	}
}

// newLineScanner wraps bufio.Scanner for line-by-line reading.
func newLineScanner(f *os.File) *lineScanner {
	return &lineScanner{file: f}
}

type lineScanner struct {
	file *os.File
	buf  [4096]byte
	line string
	pos  int64
}

func (s *lineScanner) Scan() bool {
	// Seek to current position and read new data
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := s.file.Read(buf)
		if n == 0 || err != nil {
			// No more data, wait for more
			time.Sleep(1 * time.Second)
			continue
		}
		if buf[0] == '\n' {
			s.line = string(line)
			return true
		}
		line = append(line, buf[0])
	}
}

func (s *lineScanner) Text() string {
	return s.line
}
