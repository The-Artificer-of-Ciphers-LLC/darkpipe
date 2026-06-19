// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package delivery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogIngestionRecordsDeliveryLine(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "mail.log")
	writeLogLine(t, logPath, postfixLine("ABCDEF1234", "sent", "alice@example.com"))

	tracker := NewDeliveryTracker(10)
	cancel, done := startTestLogIngestion(logPath, tracker)
	defer stopTestLogIngestion(t, cancel, done)

	waitForDeliveryCount(t, tracker, 1)
	recent := tracker.GetRecent(1)
	if recent[0].QueueID != "ABCDEF1234" {
		t.Fatalf("QueueID = %q, want ABCDEF1234", recent[0].QueueID)
	}
	if recent[0].Status != "delivered" {
		t.Fatalf("Status = %q, want delivered", recent[0].Status)
	}
}

func TestLogIngestionStopsWhileWaitingForMissingFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "missing.log")

	tracker := NewDeliveryTracker(10)
	cancel, done := startTestLogIngestion(logPath, tracker)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("log ingestion did not stop after context cancellation")
	}
}

func TestLogIngestionIgnoresNonDeliveryAndParseFailures(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "mail.log")
	writeLogLine(t, logPath, "not a delivery line\n")
	appendLogLine(t, logPath, "bad timestamp postfix/smtp[1234]: BAD1234567: to=<alice@example.com>, status=sent\n")
	appendLogLine(t, logPath, postfixLine("BCDEF12345", "bounced", "bob@example.com"))

	tracker := NewDeliveryTracker(10)
	cancel, done := startTestLogIngestion(logPath, tracker)
	defer stopTestLogIngestion(t, cancel, done)

	waitForDeliveryCount(t, tracker, 1)
	recent := tracker.GetRecent(1)
	if recent[0].QueueID != "BCDEF12345" {
		t.Fatalf("QueueID = %q, want BCDEF12345", recent[0].QueueID)
	}
	if recent[0].Status != "bounced" {
		t.Fatalf("Status = %q, want bounced", recent[0].Status)
	}
}

func TestLogIngestionReopensReplacedLogFile(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "mail.log")
	writeLogLine(t, logPath, postfixLine("CDEF123456", "sent", "alice@example.com"))

	tracker := NewDeliveryTracker(10)
	cancel, done := startTestLogIngestion(logPath, tracker)
	defer stopTestLogIngestion(t, cancel, done)

	waitForDeliveryCount(t, tracker, 1)

	if err := os.Rename(logPath, filepath.Join(dir, "mail.log.1")); err != nil {
		t.Fatalf("Rename() error: %v", err)
	}
	writeLogLine(t, logPath, postfixLine("DEF1234567", "deferred", "bob@example.com"))

	waitForDeliveryCount(t, tracker, 2)
	recent := tracker.GetRecent(1)
	if recent[0].QueueID != "DEF1234567" {
		t.Fatalf("QueueID = %q, want DEF1234567", recent[0].QueueID)
	}
	if recent[0].Status != "deferred" {
		t.Fatalf("Status = %q, want deferred", recent[0].Status)
	}
}

func startTestLogIngestion(logPath string, tracker *DeliveryTracker) (context.CancelFunc, <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	ingestion := newLogIngestion(logPath, tracker)
	ingestion.openRetryInterval = time.Millisecond
	ingestion.reopenRetryInterval = time.Millisecond
	ingestion.readPollInterval = time.Millisecond

	done := make(chan struct{})
	go func() {
		defer close(done)
		ingestion.run(ctx)
	}()
	return cancel, done
}

func stopTestLogIngestion(t *testing.T, cancel context.CancelFunc, done <-chan struct{}) {
	t.Helper()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("log ingestion did not stop")
	}
}

func waitForDeliveryCount(t *testing.T, tracker *DeliveryTracker, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if tracker.GetStats().Total >= want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("delivery count = %d, want at least %d", tracker.GetStats().Total, want)
}

func writeLogLine(t *testing.T, path, line string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(line), 0o600); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}
}

func appendLogLine(t *testing.T, path, line string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatalf("OpenFile() error: %v", err)
	}
	defer f.Close()
	if _, err := f.WriteString(line); err != nil {
		t.Fatalf("WriteString() error: %v", err)
	}
}

func postfixLine(queueID, status, to string) string {
	return fmt.Sprintf("Feb 14 10:23:45 mail postfix/smtp[1234]: %s: to=<%s>, from=<sender@example.com>, relay=mx.example.com[192.0.2.10]:25, delay=1.2, dsn=2.0.0, status=%s (250 OK)\n", queueID, to, status)
}
