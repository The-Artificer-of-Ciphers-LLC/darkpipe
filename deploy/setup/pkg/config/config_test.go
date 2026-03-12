package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig_Values(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version != CurrentVersion {
		t.Errorf("Version = %q, want %q", cfg.Version, CurrentVersion)
	}
	if cfg.MailServer != "stalwart" {
		t.Errorf("MailServer = %q, want %q", cfg.MailServer, "stalwart")
	}
	if cfg.Webmail != "snappymail" {
		t.Errorf("Webmail = %q, want %q", cfg.Webmail, "snappymail")
	}
	if cfg.Calendar != "builtin" {
		t.Errorf("Calendar = %q, want %q", cfg.Calendar, "builtin")
	}
	if cfg.Transport != "wireguard" {
		t.Errorf("Transport = %q, want %q", cfg.Transport, "wireguard")
	}
	if !cfg.QueueEnabled {
		t.Error("QueueEnabled should be true by default")
	}
	if cfg.StrictMode {
		t.Error("StrictMode should be false by default")
	}
	if cfg.ExtraDomains == nil {
		t.Error("ExtraDomains should not be nil")
	}
}

func TestSaveAndLoadConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-config.yml")

	original := DefaultConfig()
	original.MailDomain = "example.com"
	original.RelayHostname = "relay.example.com"
	original.AdminEmail = "admin@example.com"
	original.ExtraDomains = []string{"other.com", "third.com"}

	if err := SaveConfig(original, path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Compare fields
	checks := []struct {
		field string
		got   string
		want  string
	}{
		{"Version", loaded.Version, original.Version},
		{"MailDomain", loaded.MailDomain, original.MailDomain},
		{"RelayHostname", loaded.RelayHostname, original.RelayHostname},
		{"MailServer", loaded.MailServer, original.MailServer},
		{"Webmail", loaded.Webmail, original.Webmail},
		{"Calendar", loaded.Calendar, original.Calendar},
		{"AdminEmail", loaded.AdminEmail, original.AdminEmail},
		{"Transport", loaded.Transport, original.Transport},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.field, c.got, c.want)
		}
	}

	if loaded.QueueEnabled != original.QueueEnabled {
		t.Errorf("QueueEnabled = %v, want %v", loaded.QueueEnabled, original.QueueEnabled)
	}
	if loaded.StrictMode != original.StrictMode {
		t.Errorf("StrictMode = %v, want %v", loaded.StrictMode, original.StrictMode)
	}
	if len(loaded.ExtraDomains) != len(original.ExtraDomains) {
		t.Errorf("ExtraDomains length = %d, want %d", len(loaded.ExtraDomains), len(original.ExtraDomains))
	}
}

func TestLoadConfig_NonexistentFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yml")
	if err == nil {
		t.Error("LoadConfig on nonexistent file should return error")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yml")

	if err := os.WriteFile(path, []byte("{{invalid yaml content"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("LoadConfig on invalid YAML should return error")
	}
}

func TestSaveConfig_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new-config.yml")

	cfg := DefaultConfig()
	if err := SaveConfig(cfg, path); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("saved config file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("saved config file is empty")
	}
}
