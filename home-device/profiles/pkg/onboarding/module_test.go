// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package onboarding

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

type testAppPasswordStore struct {
	created        []apppassword.AppPassword
	plainPasswords []string
	revoked        []string
}

func (s *testAppPasswordStore) Create(email, deviceName, plainPassword string) (*apppassword.AppPassword, error) {
	id := fmt.Sprintf("pw-%d", len(s.created)+1)
	ap := apppassword.AppPassword{
		ID:         id,
		Email:      email,
		DeviceName: deviceName,
	}
	s.created = append(s.created, ap)
	s.plainPasswords = append(s.plainPasswords, plainPassword)
	return &ap, nil
}

func (s *testAppPasswordStore) List(email string) ([]apppassword.AppPassword, error) {
	var out []apppassword.AppPassword
	for _, password := range s.created {
		if password.Email == email {
			out = append(out, password)
		}
	}
	return out, nil
}

func (s *testAppPasswordStore) Revoke(id string) error {
	s.revoked = append(s.revoked, id)
	return nil
}

func (s *testAppPasswordStore) Verify(email, plainPassword string) (bool, error) {
	return true, nil
}

func newTestModule(store *testAppPasswordStore, profileGen *mobileconfig.ProfileGenerator) (Module, qrcode.TokenStore) {
	tokens := qrcode.NewMemoryTokenStore()
	module := New(profileGen, tokens, store, Config{
		Domain:      "example.com",
		Hostname:    "mail.example.com",
		CalDAVURL:   "https://mail.example.com/caldav",
		CardDAVURL:  "https://mail.example.com/carddav",
		CalDAVPort:  443,
		CardDAVPort: 443,
	})
	return module, tokens
}

func TestConsumeSetupIntentGeneratesPasswordByDefaultAndConsumesToken(t *testing.T) {
	store := &testAppPasswordStore{}
	module, _ := newTestModule(store, &mobileconfig.ProfileGenerator{})

	intent, err := module.IssueSetupIntent(SetupIntentInput{
		Email:      "test@example.com",
		DeviceName: "Laptop",
		Platform:   PlatformOther,
	})
	if err != nil {
		t.Fatalf("IssueSetupIntent failed: %v", err)
	}

	setup, err := module.ConsumeSetupIntent(ConsumeSetupInput{Token: intent.Token})
	if err != nil {
		t.Fatalf("ConsumeSetupIntent failed: %v", err)
	}
	if err := apppassword.ValidateAppPasswordFormat(setup.AppPassword); err != nil {
		t.Fatalf("generated password failed format policy: %v", err)
	}
	if len(store.created) != 1 {
		t.Fatalf("created passwords = %d, want 1", len(store.created))
	}
	if store.created[0].Email != "test@example.com" || store.created[0].DeviceName != "Laptop" {
		t.Fatalf("created password = %+v, want email test@example.com and device Laptop", store.created[0])
	}
	if store.plainPasswords[0] != setup.AppPassword {
		t.Fatalf("stored plaintext input = %q, want generated setup password", store.plainPasswords[0])
	}

	_, err = module.ConsumeSetupIntent(ConsumeSetupInput{Token: intent.Token})
	var usedErr ErrUsedToken
	if !errors.As(err, &usedErr) {
		t.Fatalf("second consume err = %T, want ErrUsedToken", err)
	}
}

func TestConsumeSetupIntentAcceptsSuppliedPasswordAndConsumesToken(t *testing.T) {
	store := &testAppPasswordStore{}
	module, _ := newTestModule(store, &mobileconfig.ProfileGenerator{})

	intent, err := module.IssueSetupIntent(SetupIntentInput{
		Email:      "test@example.com",
		DeviceName: "Tablet",
		Platform:   PlatformAndroid,
	})
	if err != nil {
		t.Fatalf("IssueSetupIntent failed: %v", err)
	}

	setup, err := module.ConsumeSetupIntent(ConsumeSetupInput{
		Token:               intent.Token,
		SuppliedAppPassword: "ABCD-EFGH-JKLM-NPQR",
	})
	if err != nil {
		t.Fatalf("ConsumeSetupIntent failed: %v", err)
	}
	if setup.AppPassword != "ABCD-EFGH-JKLM-NPQR" {
		t.Fatalf("AppPassword = %q, want supplied password", setup.AppPassword)
	}
	if len(store.plainPasswords) != 1 || store.plainPasswords[0] != "ABCD-EFGH-JKLM-NPQR" {
		t.Fatalf("plainPasswords = %v, want supplied password recorded once", store.plainPasswords)
	}

	_, err = module.ConsumeSetupIntent(ConsumeSetupInput{Token: intent.Token})
	var usedErr ErrUsedToken
	if !errors.As(err, &usedErr) {
		t.Fatalf("second consume err = %T, want ErrUsedToken", err)
	}
}

func TestConsumeSetupIntentRejectsInvalidSuppliedPasswordAndAllowsRetry(t *testing.T) {
	store := &testAppPasswordStore{}
	module, _ := newTestModule(store, &mobileconfig.ProfileGenerator{})

	intent, err := module.IssueSetupIntent(SetupIntentInput{
		Email:      "test@example.com",
		DeviceName: "Laptop",
		Platform:   PlatformOther,
	})
	if err != nil {
		t.Fatalf("IssueSetupIntent failed: %v", err)
	}

	_, err = module.ConsumeSetupIntent(ConsumeSetupInput{
		Token:               intent.Token,
		SuppliedAppPassword: "not-valid",
	})
	var formatErr ErrInvalidAppPasswordFormat
	if !errors.As(err, &formatErr) {
		t.Fatalf("err = %T, want ErrInvalidAppPasswordFormat", err)
	}
	if len(store.created) != 0 {
		t.Fatalf("created passwords = %d, want 0", len(store.created))
	}

	setup, err := module.ConsumeSetupIntent(ConsumeSetupInput{
		Token:               intent.Token,
		SuppliedAppPassword: "ABCD-EFGH-JKLM-NPQR",
	})
	if err != nil {
		t.Fatalf("retry ConsumeSetupIntent failed: %v", err)
	}
	if setup.AppPassword != "ABCD-EFGH-JKLM-NPQR" {
		t.Fatalf("retry AppPassword = %q, want supplied password", setup.AppPassword)
	}
}

func TestConsumeSetupIntentRollsBackPasswordWhenArtifactGenerationFailsAndAllowsRetry(t *testing.T) {
	store := &testAppPasswordStore{}
	module, _ := newTestModule(store, nil)

	intent, err := module.IssueSetupIntent(SetupIntentInput{
		Email:      "test@example.com",
		DeviceName: "Phone",
		Platform:   PlatformIOS,
	})
	if err != nil {
		t.Fatalf("IssueSetupIntent failed: %v", err)
	}

	for attempt := 1; attempt <= 2; attempt++ {
		_, err = module.ConsumeSetupIntent(ConsumeSetupInput{Token: intent.Token})
		var genErr ErrGenerationFailure
		if !errors.As(err, &genErr) {
			t.Fatalf("attempt %d err = %T, want ErrGenerationFailure", attempt, err)
		}
	}
	if len(store.created) != 2 {
		t.Fatalf("created passwords = %d, want 2 failed attempts", len(store.created))
	}
	if got, want := store.revoked, []string{"pw-1", "pw-2"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("revoked = %v, want %v", got, want)
	}
}

func TestConsumeSetupIntentMapsInvalidAndExpiredTokensWithoutProvisioning(t *testing.T) {
	tests := []struct {
		name    string
		token   func(t *testing.T, tokens qrcode.TokenStore) string
		wantErr func(error) bool
	}{
		{
			name: "invalid token",
			token: func(t *testing.T, tokens qrcode.TokenStore) string {
				return "missing-token"
			},
			wantErr: func(err error) bool {
				var target ErrInvalidToken
				return errors.As(err, &target)
			},
		},
		{
			name: "expired token",
			token: func(t *testing.T, tokens qrcode.TokenStore) string {
				token, err := tokens.CreateIntent("test@example.com", "Phone", string(PlatformOther), time.Now().Add(-time.Minute))
				if err != nil {
					t.Fatalf("CreateIntent failed: %v", err)
				}
				return token
			},
			wantErr: func(err error) bool {
				var target ErrExpiredToken
				return errors.As(err, &target)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &testAppPasswordStore{}
			module, tokens := newTestModule(store, &mobileconfig.ProfileGenerator{})

			_, err := module.ConsumeSetupIntent(ConsumeSetupInput{Token: tt.token(t, tokens)})
			if !tt.wantErr(err) {
				t.Fatalf("err = %T, want mapped token error", err)
			}
			if len(store.created) != 0 {
				t.Fatalf("created passwords = %d, want 0", len(store.created))
			}
		})
	}
}

func TestIssueOnboardingDefersIOSPasswordCreationUntilProfileDownload(t *testing.T) {
	store := &testAppPasswordStore{}
	module, _ := newTestModule(store, &mobileconfig.ProfileGenerator{})

	issued, err := module.IssueOnboarding(OnboardingIssueInput{
		Email:      "test@example.com",
		DeviceName: "iPhone",
		Platform:   PlatformIOS,
	})
	if err != nil {
		t.Fatalf("IssueOnboarding failed: %v", err)
	}
	if !issued.Deferred {
		t.Fatal("Deferred = false, want true")
	}
	if issued.AppPassword != "" {
		t.Fatalf("AppPassword = %q, want empty before token consumption", issued.AppPassword)
	}
	if len(store.created) != 0 {
		t.Fatalf("created passwords = %d, want 0 before token consumption", len(store.created))
	}
	if issued.Token == "" || issued.Setup == nil || issued.Setup.ProfileURL == "" || issued.Setup.QRCodeData == "" {
		t.Fatalf("issued setup missing token/profile URL/QR data: %+v", issued)
	}

	consumed, err := module.ConsumeSetupIntent(ConsumeSetupInput{Token: issued.Token})
	if err != nil {
		t.Fatalf("ConsumeSetupIntent failed: %v", err)
	}
	if consumed.AppPassword == "" {
		t.Fatal("AppPassword after token consumption is empty")
	}
	if len(store.created) != 1 {
		t.Fatalf("created passwords = %d, want 1 after token consumption", len(store.created))
	}
}

func TestIssueOnboardingConsumesNonAppleSetupImmediately(t *testing.T) {
	store := &testAppPasswordStore{}
	module, _ := newTestModule(store, &mobileconfig.ProfileGenerator{})

	issued, err := module.IssueOnboarding(OnboardingIssueInput{
		Email:               "test@example.com",
		DeviceName:          "Android",
		Platform:            PlatformAndroid,
		SuppliedAppPassword: "ABCD-EFGH-JKLM-NPQR",
	})
	if err != nil {
		t.Fatalf("IssueOnboarding failed: %v", err)
	}
	if issued.Deferred {
		t.Fatal("Deferred = true, want false")
	}
	if issued.AppPassword != "ABCD-EFGH-JKLM-NPQR" {
		t.Fatalf("AppPassword = %q, want supplied password", issued.AppPassword)
	}
	if len(store.plainPasswords) != 1 || store.plainPasswords[0] != "ABCD-EFGH-JKLM-NPQR" {
		t.Fatalf("plainPasswords = %v, want supplied password stored once", store.plainPasswords)
	}

	_, err = module.ConsumeSetupIntent(ConsumeSetupInput{Token: issued.Token})
	var usedErr ErrUsedToken
	if !errors.As(err, &usedErr) {
		t.Fatalf("second consume err = %T, want ErrUsedToken", err)
	}
}

func TestIssueOnboardingRejectsSuppliedPasswordForProfileDownload(t *testing.T) {
	store := &testAppPasswordStore{}
	module, _ := newTestModule(store, &mobileconfig.ProfileGenerator{})

	_, err := module.IssueOnboarding(OnboardingIssueInput{
		Email:               "test@example.com",
		DeviceName:          "iPhone",
		Platform:            PlatformIOS,
		SuppliedAppPassword: "ABCD-EFGH-JKLM-NPQR",
	})
	var unsupported ErrSuppliedAppPasswordUnsupported
	if !errors.As(err, &unsupported) {
		t.Fatalf("err = %T, want ErrSuppliedAppPasswordUnsupported", err)
	}
	if len(store.created) != 0 {
		t.Fatalf("created passwords = %d, want 0", len(store.created))
	}
}

func TestIssueOnboardingRejectsInvalidSuppliedPasswordBeforeProvisioning(t *testing.T) {
	store := &testAppPasswordStore{}
	module, _ := newTestModule(store, &mobileconfig.ProfileGenerator{})

	_, err := module.IssueOnboarding(OnboardingIssueInput{
		Email:               "test@example.com",
		DeviceName:          "Android",
		Platform:            PlatformAndroid,
		SuppliedAppPassword: "not-valid",
	})
	var formatErr ErrInvalidAppPasswordFormat
	if !errors.As(err, &formatErr) {
		t.Fatalf("err = %T, want ErrInvalidAppPasswordFormat", err)
	}
	if len(store.created) != 0 {
		t.Fatalf("created passwords = %d, want 0", len(store.created))
	}
}
