// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package main

import "testing"

func TestProfileServerQRURL(t *testing.T) {
	got, err := profileServerQRURL("https://profiles.example.com/base/", "test+device@example.com")
	if err != nil {
		t.Fatalf("profileServerQRURL failed: %v", err)
	}
	want := "https://profiles.example.com/base/qr/generate?email=test%2Bdevice%40example.com"
	if got != want {
		t.Fatalf("URL = %q, want %q", got, want)
	}
}

func TestProfileServerQRURLRequiresAbsoluteURL(t *testing.T) {
	_, err := profileServerQRURL("profiles.example.com", "test@example.com")
	if err == nil {
		t.Fatal("err = nil, want invalid URL error")
	}
}
