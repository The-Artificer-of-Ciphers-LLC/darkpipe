// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package delivery

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"time"
)

// StartLogIngestion reads mail delivery log lines and records delivery events.
func StartLogIngestion(ctx context.Context, logPath string, tracker *DeliveryTracker) {
	newLogIngestion(logPath, tracker).run(ctx)
}

type logIngestion struct {
	logPath             string
	tracker             *DeliveryTracker
	parser              *Parser
	openRetryInterval   time.Duration
	reopenRetryInterval time.Duration
	readPollInterval    time.Duration
	open                func(string) (*os.File, error)
	wait                func(context.Context, time.Duration) bool
}

func newLogIngestion(logPath string, tracker *DeliveryTracker) *logIngestion {
	return &logIngestion{
		logPath:             logPath,
		tracker:             tracker,
		parser:              NewParser(),
		openRetryInterval:   30 * time.Second,
		reopenRetryInterval: 5 * time.Second,
		readPollInterval:    1 * time.Second,
		open:                os.Open,
		wait:                waitContext,
	}
}

func (i *logIngestion) run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		f, err := i.open(i.logPath)
		if err != nil {
			log.Printf("Waiting for mail log at %s: %v", i.logPath, err)
			if !i.wait(ctx, i.openRetryInterval) {
				return
			}
			continue
		}

		i.ingestFile(ctx, f)
		if err := f.Close(); err != nil {
			log.Printf("Error closing mail log %s: %v", i.logPath, err)
		}

		if !i.wait(ctx, i.reopenRetryInterval) {
			return
		}
	}
}

func (i *logIngestion) ingestFile(ctx context.Context, f *os.File) {
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if line != "" {
			i.recordLine(line)
		}
		if err == nil {
			continue
		}
		if !errors.Is(err, io.EOF) {
			log.Printf("Error reading mail log %s: %v", i.logPath, err)
			return
		}
		if ctx.Err() != nil || i.fileWasReplacedOrTruncated(f) {
			return
		}
		if !i.wait(ctx, i.readPollInterval) {
			return
		}
	}
}

func (i *logIngestion) recordLine(line string) {
	entry, err := i.parser.ParseLine(line)
	if err != nil || entry == nil {
		return
	}
	i.tracker.Record(entry)
}

func (i *logIngestion) fileWasReplacedOrTruncated(f *os.File) bool {
	current, err := f.Stat()
	if err != nil {
		return true
	}
	latest, err := os.Stat(i.logPath)
	if err != nil {
		return true
	}
	if !os.SameFile(current, latest) {
		return true
	}
	offset, err := f.Seek(0, io.SeekCurrent)
	if err != nil {
		return true
	}
	return latest.Size() < offset
}

func waitContext(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
