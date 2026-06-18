// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
)

func TestLoadProfileRuntimeConfig(t *testing.T) {
	t.Setenv("MAIL_DOMAIN", "example.com")
	t.Setenv("MAIL_HOSTNAME", "mail.example.com")
	t.Setenv("CALDAV_URL", "https://mail.example.com/caldav")
	t.Setenv("CARDDAV_URL", "https://mail.example.com/carddav")
	t.Setenv("CALDAV_PORT", "8443")
	t.Setenv("CARDDAV_PORT", "9443")
	t.Setenv("ADMIN_USER", "root")
	t.Setenv("ADMIN_PASSWORD", "secret")
	t.Setenv("PROFILE_SERVER_PORT", "9090")
	t.Setenv("MAIL_SERVER_TYPE", "maddy")
	t.Setenv("APP_PASSWORD_STORE_PATH", "/tmp/app-passwords.json")
	t.Setenv("MONITOR_CERT_PATHS", "/certs/mail.pem,/certs/relay.pem")
	t.Setenv("MONITOR_HEALTHCHECK_URL", "https://hc.example.com/ping")
	t.Setenv("MONITOR_LOG_PATH", "/logs/mail.log")

	cfg := loadProfileRuntimeConfig()

	if cfg.Server.Domain != "example.com" {
		t.Fatalf("Domain = %q, want example.com", cfg.Server.Domain)
	}
	if cfg.Server.CalDAVPort != 8443 {
		t.Fatalf("CalDAVPort = %d, want 8443", cfg.Server.CalDAVPort)
	}
	if cfg.Server.CardDAVPort != 9443 {
		t.Fatalf("CardDAVPort = %d, want 9443", cfg.Server.CardDAVPort)
	}
	if cfg.Port != "9090" {
		t.Fatalf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.MailServerType != "maddy" {
		t.Fatalf("MailServerType = %q, want maddy", cfg.MailServerType)
	}
	wantCerts := []string{"/certs/mail.pem", "/certs/relay.pem"}
	if !reflect.DeepEqual(cfg.MonitorCertPaths, wantCerts) {
		t.Fatalf("MonitorCertPaths = %v, want %v", cfg.MonitorCertPaths, wantCerts)
	}
	if cfg.MonitorHealthcheckURL != "https://hc.example.com/ping" {
		t.Fatalf("MonitorHealthcheckURL = %q", cfg.MonitorHealthcheckURL)
	}
	if cfg.MonitorLogPath != "/logs/mail.log" {
		t.Fatalf("MonitorLogPath = %q, want /logs/mail.log", cfg.MonitorLogPath)
	}
}

func TestNewAppPasswordStore(t *testing.T) {
	storePath := filepath.Join(t.TempDir(), "app-passwords.json")
	tests := []struct {
		name           string
		mailServerType string
		wantType       string
	}{
		{name: "stalwart", mailServerType: "stalwart", wantType: "*apppassword.StalwartStore"},
		{name: "dovecot", mailServerType: "dovecot", wantType: "*apppassword.DovecotStore"},
		{name: "maddy", mailServerType: "maddy", wantType: "*apppassword.MaddyStore"},
		{name: "postfix dovecot", mailServerType: "postfix-dovecot", wantType: "*apppassword.DovecotStore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, label, err := newAppPasswordStore(tt.mailServerType, storePath)
			if err != nil {
				t.Fatalf("newAppPasswordStore failed: %v", err)
			}
			if gotType := reflect.TypeOf(store).String(); gotType != tt.wantType {
				t.Fatalf("store type = %s, want %s", gotType, tt.wantType)
			}
			if label == "" {
				t.Fatal("label should not be empty")
			}
			assertStorePath(t, store, storePath)
		})
	}
}

func TestNewAppPasswordStoreUnknownType(t *testing.T) {
	_, _, err := newAppPasswordStore("unknown", filepath.Join(t.TempDir(), "app-passwords.json"))
	if err == nil {
		t.Fatal("err = nil, want unknown mail server error")
	}
}

func TestNewProfileRuntimeRegistersRoutes(t *testing.T) {
	runtime, err := newProfileRuntime(testRuntimeConfig(t, "maddy"))
	if err != nil {
		t.Fatalf("newProfileRuntime failed: %v", err)
	}

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{name: "profile download", method: http.MethodGet, path: "/profile/download", wantStatus: http.StatusBadRequest},
		{name: "qr generate auth", method: http.MethodGet, path: "/qr/generate", wantStatus: http.StatusUnauthorized},
		{name: "device list auth", method: http.MethodGet, path: "/devices", wantStatus: http.StatusUnauthorized},
		{name: "liveness", method: http.MethodGet, path: "/health/live", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			runtime.server.Handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("%s %s status = %d, want %d", tt.method, tt.path, w.Code, tt.wantStatus)
			}
		})
	}
}

func testRuntimeConfig(t *testing.T, mailServerType string) profileRuntimeConfig {
	t.Helper()
	return profileRuntimeConfig{
		Server: ServerConfig{
			Domain:      "example.com",
			Hostname:    "mail.example.com",
			CalDAVURL:   "https://mail.example.com/caldav",
			CardDAVURL:  "https://mail.example.com/carddav",
			CalDAVPort:  443,
			CardDAVPort: 443,
			AdminUser:   "admin",
			AdminPass:   "secret",
		},
		Port:                 "0",
		MailServerType:       mailServerType,
		AppPasswordStorePath: filepath.Join(t.TempDir(), "app-passwords.json"),
		MonitorLogPath:       filepath.Join(t.TempDir(), "mail.log"),
	}
}

func assertStorePath(t *testing.T, store apppassword.Store, wantPath string) {
	t.Helper()
	switch s := store.(type) {
	case *apppassword.DovecotStore:
		if s.FilePath != wantPath {
			t.Fatalf("DovecotStore.FilePath = %q, want %q", s.FilePath, wantPath)
		}
	case *apppassword.MaddyStore:
		if s.FilePath != wantPath {
			t.Fatalf("MaddyStore.FilePath = %q, want %q", s.FilePath, wantPath)
		}
	}
}
