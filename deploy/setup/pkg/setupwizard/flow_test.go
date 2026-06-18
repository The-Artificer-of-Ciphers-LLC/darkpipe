// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package setupwizard

import (
	"strings"
	"testing"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/config"
)

func TestIsQuickSetupMode(t *testing.T) {
	if !isQuickSetupMode(quickSetupMode) {
		t.Fatal("quick setup mode should be quick")
	}
	if isQuickSetupMode(advancedSetupMode) {
		t.Fatal("advanced setup mode should not be quick")
	}
}

func TestCalendarOptionsForStalwart(t *testing.T) {
	cfg := &config.Config{
		MailServer: "stalwart",
		Calendar:   "builtin",
	}

	options := calendarOptions(cfg)
	want := []string{"none", "radicale", "builtin"}
	if !sameStrings(options, want) {
		t.Fatalf("calendar options = %v, want %v", options, want)
	}
	if cfg.Calendar != "builtin" {
		t.Fatalf("calendar = %q, want builtin", cfg.Calendar)
	}
}

func TestCalendarOptionsForNonStalwart(t *testing.T) {
	cfg := &config.Config{
		MailServer: "maddy",
		Calendar:   "builtin",
	}

	options := calendarOptions(cfg)
	want := []string{"none", "radicale"}
	if !sameStrings(options, want) {
		t.Fatalf("calendar options = %v, want %v", options, want)
	}
	if cfg.Calendar != "radicale" {
		t.Fatalf("calendar = %q, want radicale", cfg.Calendar)
	}
}

func TestSetupCompleteMessage(t *testing.T) {
	message := setupCompleteMessage("example.com")

	if !strings.Contains(message, "darkpipe-dns-setup --domain example.com") {
		t.Fatalf("message should include DNS setup command, got %q", message)
	}
	if !strings.Contains(message, "docker compose up -d") {
		t.Fatalf("message should include compose start command, got %q", message)
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
