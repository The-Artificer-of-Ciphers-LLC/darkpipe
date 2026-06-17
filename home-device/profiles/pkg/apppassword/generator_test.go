// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package apppassword

import (
	"strings"
	"testing"
)

func TestGenerateAppPassword_Format(t *testing.T) {
	password, err := GenerateAppPassword()
	if err != nil {
		t.Fatalf("GenerateAppPassword failed: %v", err)
	}

	if err := ValidateAppPasswordFormat(password); err != nil {
		t.Fatalf("generated password failed format policy: %v", err)
	}

	// Check length (19 chars including hyphens)
	if len(password) != 19 {
		t.Errorf("Password length incorrect: got %d, want 19", len(password))
	}

	// Check hyphen positions
	if password[4] != '-' || password[9] != '-' || password[14] != '-' {
		t.Errorf("Hyphens in wrong positions: %s", password)
	}
}

func TestGenerateAppPassword_Charset(t *testing.T) {
	password, err := GenerateAppPassword()
	if err != nil {
		t.Fatalf("GenerateAppPassword failed: %v", err)
	}

	// Remove hyphens for charset check
	chars := strings.ReplaceAll(password, "-", "")

	// Check all characters are in charset
	for _, ch := range chars {
		if !strings.ContainsRune(charset, ch) {
			t.Errorf("Invalid character in password: %c", ch)
		}
	}

	// Check no confusing characters (0, O, 1, I)
	confusing := "01OI"
	for _, ch := range chars {
		if strings.ContainsRune(confusing, ch) {
			t.Errorf("Password contains confusing character: %c", ch)
		}
	}
}

func TestGenerateAppPassword_Uniqueness(t *testing.T) {
	// Generate 100 passwords and check for duplicates
	passwords := make(map[string]bool)
	for i := 0; i < 100; i++ {
		password, err := GenerateAppPassword()
		if err != nil {
			t.Fatalf("GenerateAppPassword failed on iteration %d: %v", i, err)
		}
		if passwords[password] {
			t.Errorf("Duplicate password generated: %s", password)
		}
		passwords[password] = true
	}
}

func TestValidateAppPasswordFormat(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid format",
			password: "ABCD-EFGH-JKLM-NPQR",
		},
		{
			name:     "empty",
			password: "",
			wantErr:  true,
		},
		{
			name:     "one group short",
			password: "ABCD-EFGH-JKLM",
			wantErr:  true,
		},
		{
			name:     "one character short",
			password: "ABCD-EFGH-JKLM-NPQ",
			wantErr:  true,
		},
		{
			name:     "one character long",
			password: "ABCD-EFGH-JKLM-NPQRR",
			wantErr:  true,
		},
		{
			name:     "wrong separator",
			password: "ABCD_EFGH_JKLM_NPQR",
			wantErr:  true,
		},
		{
			name:     "excluded zero",
			password: "ABCD-EFGH-JKLM-NPQ0",
			wantErr:  true,
		},
		{
			name:     "excluded capital O",
			password: "ABCD-EFGH-JKLM-NPQO",
			wantErr:  true,
		},
		{
			name:     "lowercase",
			password: "abcd-efgh-jklm-npqr",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAppPasswordFormat(tt.password)
			if tt.wantErr && err == nil {
				t.Fatalf("ValidateAppPasswordFormat(%q) succeeded, want error", tt.password)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("ValidateAppPasswordFormat(%q) failed: %v", tt.password, err)
			}
		})
	}
}

func TestHashPassword_Roundtrip(t *testing.T) {
	password := "TEST-PASS-WORD-1234"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash is empty")
	}

	// Verify correct password
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword failed for correct password")
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	password := "TEST-PASS-WORD-1234"
	wrongPassword := "WRONG-PASS-WORD-5678"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify wrong password fails
	if VerifyPassword(wrongPassword, hash) {
		t.Error("VerifyPassword succeeded for wrong password")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	password := "TEST-PASS-WORD-1234"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("First HashPassword failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	// Bcrypt should produce different salts/hashes
	if hash1 == hash2 {
		t.Error("Two hashes of same password are identical (salt issue)")
	}

	// But both should verify
	if !VerifyPassword(password, hash1) || !VerifyPassword(password, hash2) {
		t.Error("Hash verification failed")
	}
}
