// Package tests provides end-to-end integration tests for the cloud relay.
//
// These tests verify the complete SMTP pipeline: SMTP client -> go-smtp server
// -> session -> forwarder -> mock home device.
package tests

import (
	"net"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/cloud-relay/relay/config"
	"github.com/darkpipe/darkpipe/cloud-relay/relay/forward"
	smtpserver "github.com/darkpipe/darkpipe/cloud-relay/relay/smtp"
)

// Helper function to start a test SMTP server
func startTestServer(t *testing.T, mockFwd *forward.MockForwarder) string {
	t.Helper()

	cfg := &config.Config{
		ListenAddr:      "127.0.0.1:0",
		MaxMessageBytes: 10 * 1024 * 1024,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
	}

	server := smtpserver.NewServer(mockFwd, cfg)

	// Create listener to get actual port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	addr := listener.Addr().String()

	// Start server in background
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Clean up server on test completion
	t.Cleanup(func() {
		server.Close()
		listener.Close()
	})

	return addr
}

func TestIntegration_SMTPRelayFlow(t *testing.T) {
	t.Parallel()

	// Create mock forwarder to capture forwarded messages
	mockFwd := forward.NewMockForwarder()

	// Start test server
	serverAddr := startTestServer(t, mockFwd)

	// Send a test email via SMTP client
	from := "sender@example.com"
	to := []string{"recipient@example.com"}
	message := []byte("Subject: Integration Test\r\n\r\nThis is a test message from the integration test.\r\n")

	err := smtp.SendMail(serverAddr, nil, from, to, message)
	if err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	// Give forwarder time to process
	time.Sleep(100 * time.Millisecond)

	// Verify mock forwarder received the message
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1", len(calls))
	}

	call := calls[0]

	// Verify envelope
	if call.From != from {
		t.Errorf("forwarded from = %q, want %q", call.From, from)
	}

	if len(call.To) != 1 {
		t.Fatalf("forwarded to count = %d, want 1", len(call.To))
	}

	if call.To[0] != to[0] {
		t.Errorf("forwarded to = %q, want %q", call.To[0], to[0])
	}

	// Verify message body
	if !strings.Contains(call.Data, "Integration Test") {
		t.Errorf("message data doesn't contain subject: %s", call.Data)
	}

	if !strings.Contains(call.Data, "test message from the integration test") {
		t.Errorf("message data doesn't contain body: %s", call.Data)
	}
}

func TestIntegration_MultipleRecipients(t *testing.T) {
	t.Parallel()

	// Create mock forwarder
	mockFwd := forward.NewMockForwarder()

	// Start test server
	serverAddr := startTestServer(t, mockFwd)

	// Send email with multiple recipients
	from := "sender@example.com"
	to := []string{
		"recipient1@example.com",
		"recipient2@example.com",
		"recipient3@example.com",
	}
	message := []byte("Subject: Multiple Recipients\r\n\r\nTest with multiple recipients.\r\n")

	err := smtp.SendMail(serverAddr, nil, from, to, message)
	if err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify all recipients were forwarded
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1", len(calls))
	}

	call := calls[0]

	if len(call.To) != 3 {
		t.Fatalf("forwarded to count = %d, want 3", len(call.To))
	}

	for i, expectedTo := range to {
		if call.To[i] != expectedTo {
			t.Errorf("forwarded to[%d] = %q, want %q", i, call.To[i], expectedTo)
		}
	}
}

func TestIntegration_LargeMessage(t *testing.T) {
	t.Parallel()

	// Create mock forwarder
	mockFwd := forward.NewMockForwarder()

	// Start test server
	serverAddr := startTestServer(t, mockFwd)

	// Create a moderately sized message with proper SMTP formatting
	// Note: SMTP has line length limits (typically 1000 chars), so we need to format properly
	bodyLines := make([]string, 100)
	for i := range bodyLines {
		bodyLines[i] = strings.Repeat("X", 70) // 70 chars per line, well under SMTP limit
	}
	largeBody := strings.Join(bodyLines, "\r\n")
	message := []byte("Subject: Large Message\r\n\r\n" + largeBody + "\r\n")

	from := "sender@example.com"
	to := []string{"recipient@example.com"}

	err := smtp.SendMail(serverAddr, nil, from, to, message)
	if err != nil {
		t.Fatalf("SendMail: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	// Verify message was forwarded
	calls := mockFwd.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("forwarder calls = %d, want 1", len(calls))
	}

	call := calls[0]

	// Verify large body was received (check for presence of X characters)
	if !strings.Contains(call.Data, strings.Repeat("X", 70)) {
		t.Errorf("large message body not fully received (got %d bytes)", len(call.Data))
	}

	// Verify reasonable size
	if len(call.Data) < 7000 { // 100 lines * 70 chars + headers and CRLF
		t.Errorf("message body too small: got %d bytes, want at least 7000", len(call.Data))
	}
}

func TestIntegration_EphemeralBehavior(t *testing.T) {
	t.Parallel()

	// This test verifies that after forwarding, no data remains in the session
	// (ephemeral behavior per RELAY-02).

	// Create mock forwarder
	mockFwd := forward.NewMockForwarder()

	// Start test server
	serverAddr := startTestServer(t, mockFwd)

	// Send first message
	from1 := "sender1@example.com"
	to1 := []string{"recipient1@example.com"}
	message1 := []byte("Subject: First\r\n\r\nFirst message\r\n")

	err := smtp.SendMail(serverAddr, nil, from1, to1, message1)
	if err != nil {
		t.Fatalf("SendMail(1): %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Send second message
	from2 := "sender2@example.com"
	to2 := []string{"recipient2@example.com"}
	message2 := []byte("Subject: Second\r\n\r\nSecond message\r\n")

	err = smtp.SendMail(serverAddr, nil, from2, to2, message2)
	if err != nil {
		t.Fatalf("SendMail(2): %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify both messages were forwarded correctly (no cross-contamination)
	calls := mockFwd.GetCalls()
	if len(calls) != 2 {
		t.Fatalf("forwarder calls = %d, want 2", len(calls))
	}

	// First message should contain only first sender/recipient/body
	if calls[0].From != from1 {
		t.Errorf("first call from = %q, want %q", calls[0].From, from1)
	}

	if calls[0].To[0] != to1[0] {
		t.Errorf("first call to = %q, want %q", calls[0].To[0], to1[0])
	}

	if strings.Contains(calls[0].Data, "Second message") {
		t.Error("first call data contains second message (not ephemeral)")
	}

	// Second message should contain only second sender/recipient/body
	if calls[1].From != from2 {
		t.Errorf("second call from = %q, want %q", calls[1].From, from2)
	}

	if calls[1].To[0] != to2[0] {
		t.Errorf("second call to = %q, want %q", calls[1].To[0], to2[0])
	}

	if strings.Contains(calls[1].Data, "First message") {
		t.Error("second call data contains first message (not ephemeral)")
	}
}
