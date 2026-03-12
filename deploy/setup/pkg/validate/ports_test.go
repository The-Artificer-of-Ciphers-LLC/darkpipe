package validate

import (
	"net"
	"testing"
	"time"
)

func TestValidatePort_OpenPort(t *testing.T) {
	// Start a TCP listener on a random port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port

	if err := ValidatePort("127.0.0.1", port, 2*time.Second); err != nil {
		t.Errorf("ValidatePort on open port %d returned error: %v", port, err)
	}
}

func TestValidatePort_ClosedPort(t *testing.T) {
	// Use a random high port that is almost certainly not listening
	// Bind and immediately close to get an unused port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get unused port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close() // port is now closed

	err = ValidatePort("127.0.0.1", port, 500*time.Millisecond)
	if err == nil {
		t.Errorf("ValidatePort on closed port %d should return error", port)
	}
}

func TestCheckLocalPorts_Mixed(t *testing.T) {
	// Start a listener on one port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer ln.Close()
	openPort := ln.Addr().(*net.TCPAddr).Port

	// Get a closed port
	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to get unused port: %v", err)
	}
	closedPort := ln2.Addr().(*net.TCPAddr).Port
	ln2.Close()

	inUse, err := CheckLocalPorts([]int{openPort, closedPort})
	if err != nil {
		t.Fatalf("CheckLocalPorts failed: %v", err)
	}

	// Only the open port should be reported
	if len(inUse) != 1 {
		t.Fatalf("CheckLocalPorts returned %d ports in use, want 1: %v", len(inUse), inUse)
	}
	if inUse[0] != openPort {
		t.Errorf("CheckLocalPorts reported port %d, want %d", inUse[0], openPort)
	}
}

func TestCheckLocalPorts_Empty(t *testing.T) {
	inUse, err := CheckLocalPorts([]int{})
	if err != nil {
		t.Fatalf("CheckLocalPorts failed: %v", err)
	}
	if len(inUse) != 0 {
		t.Errorf("CheckLocalPorts on empty input returned %d, want 0", len(inUse))
	}
}

func TestDetectRAM_ReturnsError(t *testing.T) {
	// DetectRAM is a stub that always returns an error
	_, err := DetectRAM()
	if err == nil {
		t.Error("DetectRAM should return an error (not yet implemented)")
	}
}
