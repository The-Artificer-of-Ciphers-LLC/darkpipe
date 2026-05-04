// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package qrcode

import (
	"sync"
	"testing"
	"time"
)

func TestGenerateSecureToken(t *testing.T) {
	token1, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	if token1 == "" {
		t.Fatal("GenerateSecureToken returned empty string")
	}

	if len(token1) < 43 {
		t.Errorf("Token too short: got %d, want >= 43", len(token1))
	}

	token2, err := GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed on second call: %v", err)
	}

	if token1 == token2 {
		t.Error("GenerateSecureToken produced duplicate tokens")
	}
}

func TestMemoryTokenStoreCreate(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	token, err := store.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if token == "" {
		t.Fatal("Create returned empty token")
	}

	store.mu.RLock()
	stored, exists := store.tokens[token]
	store.mu.RUnlock()

	if !exists {
		t.Fatal("Token not found in store")
	}

	if stored.Email != email {
		t.Errorf("Email mismatch: got %s, want %s", stored.Email, email)
	}

	if stored.Used {
		t.Error("New token should not be marked as used")
	}
}

func TestMemoryTokenStoreValidate(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	token, err := store.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	validEmail, state, err := store.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if state != ValidationStateValid {
		t.Errorf("Token state mismatch: got %s, want %s", state, ValidationStateValid)
	}
	if validEmail != email {
		t.Errorf("Email mismatch: got %s, want %s", validEmail, email)
	}

	_, state, err = store.Validate(token)
	if err != nil {
		t.Fatalf("Second Validate failed: %v", err)
	}
	if state != ValidationStateUsed {
		t.Errorf("Token state mismatch after first use: got %s, want %s", state, ValidationStateUsed)
	}
}

func TestMemoryTokenStoreValidateExpired(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(-1 * time.Minute)

	token, err := store.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, state, err := store.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if state != ValidationStateExpired {
		t.Errorf("Token state mismatch: got %s, want %s", state, ValidationStateExpired)
	}
}

func TestMemoryTokenStoreValidateNonExistent(t *testing.T) {
	store := NewMemoryTokenStore()

	_, state, err := store.Validate("nonexistent-token")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if state != ValidationStateInvalid {
		t.Errorf("Token state mismatch: got %s, want %s", state, ValidationStateInvalid)
	}
}

func TestMemoryTokenStoreInvalidate(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	token, err := store.Create(email, expiresAt)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = store.Invalidate(token)
	if err != nil {
		t.Fatalf("Invalidate failed: %v", err)
	}

	_, state, err := store.Validate(token)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if state != ValidationStateUsed {
		t.Errorf("Token state mismatch: got %s, want %s", state, ValidationStateUsed)
	}
}

func TestMemoryTokenStoreCleanup(t *testing.T) {
	store := NewMemoryTokenStore()

	expiredEmail := "expired@example.com"
	expiredToken, err := store.Create(expiredEmail, time.Now().Add(-1*time.Minute))
	if err != nil {
		t.Fatalf("Create expired token failed: %v", err)
	}

	validEmail := "valid@example.com"
	validToken, err := store.Create(validEmail, time.Now().Add(15*time.Minute))
	if err != nil {
		t.Fatalf("Create valid token failed: %v", err)
	}

	store.Cleanup()

	store.mu.RLock()
	_, expiredExists := store.tokens[expiredToken]
	_, validExists := store.tokens[validToken]
	store.mu.RUnlock()

	if expiredExists {
		t.Error("Cleanup did not remove expired token")
	}
	if !validExists {
		t.Error("Cleanup removed valid token")
	}
}

func TestMemoryTokenStoreConcurrency(t *testing.T) {
	store := NewMemoryTokenStore()
	email := "test@example.com"
	expiresAt := time.Now().Add(15 * time.Minute)

	var wg sync.WaitGroup
	numGoroutines := 10

	tokens := make([]string, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			token, err := store.Create(email, expiresAt)
			if err != nil {
				t.Errorf("Concurrent Create failed: %v", err)
				return
			}
			tokens[idx] = token
		}(i)
	}
	wg.Wait()

	results := make([]ValidationState, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, state, err := store.Validate(tokens[idx])
			if err != nil {
				t.Errorf("Concurrent Validate failed: %v", err)
				return
			}
			results[idx] = state
		}(i)
	}
	wg.Wait()

	for i, state := range results {
		if state != ValidationStateValid {
			t.Errorf("Token %d validation state mismatch: got %s", i, state)
		}
	}
}
