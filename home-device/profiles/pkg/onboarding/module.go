// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package onboarding

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/profiles/pkg/apppassword"
	"github.com/darkpipe/darkpipe/profiles/pkg/mobileconfig"
	"github.com/darkpipe/darkpipe/profiles/pkg/qrcode"
)

// Config contains onboarding generation settings.
type Config struct {
	Domain      string
	Hostname    string
	CalDAVURL   string
	CardDAVURL  string
	CalDAVPort  int
	CardDAVPort int
}

// Module is the onboarding seam for profile and QR artifacts.
type Platform string

const (
	PlatformIOS         Platform = "ios"
	PlatformMacOS       Platform = "macos"
	PlatformAndroid     Platform = "android"
	PlatformThunderbird Platform = "thunderbird"
	PlatformOutlook     Platform = "outlook"
	PlatformOther       Platform = "other"
)

type InstructionField struct {
	Key   string
	Value string
}

type PlatformSetup struct {
	Platform      Platform
	Title         string
	Summary       string
	Steps         []string
	Fields        []InstructionField
	DownloadLabel string
	QRCodeData    string
	ProfileURL    string
	TokenExpiry   string
	TokenTTL      string
}

type SetupIntentInput struct {
	Email      string
	DeviceName string
	Platform   Platform
}

type SetupIntent struct {
	Token     string
	URL       string
	ExpiresAt time.Time
	Platform  Platform
}

type ConsumeSetupInput struct {
	Token               string
	SuppliedAppPassword string
}

type ConsumedSetup struct {
	Email       string
	AppPassword string
	Platform    Platform
	Setup       *PlatformSetup
	ProfileData []byte
}

type Module interface {
	IssueSetupIntent(in SetupIntentInput) (*SetupIntent, error)
	ConsumeSetupIntent(in ConsumeSetupInput) (*ConsumedSetup, error)
	GenerateMobileConfigFromToken(token string) ([]byte, string, error)
	GenerateQRURL(email string) (string, time.Time, error)
	GenerateQRPNG(email string, size int) ([]byte, error)
	GeneratePlatformSetup(platform Platform, email, appPassword string) (*PlatformSetup, error)
}

type ErrUnauthorized struct{ Cause error }

func (e ErrUnauthorized) Error() string { return "unauthorized" }

type ErrInvalidToken struct{}

func (e ErrInvalidToken) Error() string { return "invalid token" }

type ErrExpiredToken struct{}

func (e ErrExpiredToken) Error() string { return "expired token" }

type ErrUsedToken struct{}

func (e ErrUsedToken) Error() string { return "used token" }

type ErrStoreFailure struct{ Cause error }

func (e ErrStoreFailure) Error() string { return fmt.Sprintf("store failure: %v", e.Cause) }

type ErrGenerationFailure struct{ Cause error }

func (e ErrGenerationFailure) Error() string { return fmt.Sprintf("generation failure: %v", e.Cause) }

type ErrUnknownPlatform struct{ Platform string }

func (e ErrUnknownPlatform) Error() string { return fmt.Sprintf("unknown platform: %s", e.Platform) }

type ErrInvalidAppPasswordFormat struct{ Cause error }

func (e ErrInvalidAppPasswordFormat) Error() string {
	return fmt.Sprintf("invalid app password format: %v", e.Cause)
}

type DefaultModule struct {
	profileGen *mobileconfig.ProfileGenerator
	tokens     qrcode.TokenStore
	appPass    apppassword.Store
	config     Config
	now        func() time.Time
}

func New(profileGen *mobileconfig.ProfileGenerator, tokens qrcode.TokenStore, appPass apppassword.Store, cfg Config) Module {
	return &DefaultModule{
		profileGen: profileGen,
		tokens:     tokens,
		appPass:    appPass,
		config:     cfg,
		now:        time.Now,
	}
}

func (m *DefaultModule) IssueSetupIntent(in SetupIntentInput) (*SetupIntent, error) {
	platform := in.Platform
	if platform == "" {
		platform = PlatformIOS
	}
	expiresAt := m.now().Add(qrcode.DefaultTokenExpiry)
	token, err := m.tokens.CreateIntent(in.Email, in.DeviceName, string(platform), expiresAt)
	if err != nil {
		return nil, ErrGenerationFailure{Cause: err}
	}
	url := fmt.Sprintf("https://%s/setup?token=%s", m.config.Hostname, token)
	return &SetupIntent{Token: token, URL: url, ExpiresAt: expiresAt, Platform: platform}, nil
}

func (m *DefaultModule) ConsumeSetupIntent(in ConsumeSetupInput) (*ConsumedSetup, error) {
	var out *ConsumedSetup
	_, state, err := m.tokens.Consume(in.Token, func(intent qrcode.Token) error {
		platform := Platform(intent.Platform)
		if platform == "" {
			platform = PlatformIOS
		}
		deviceName := intent.DeviceName
		if deviceName == "" {
			deviceName = fmt.Sprintf("QR-%d", m.now().Unix())
		}

		plainPassword := in.SuppliedAppPassword
		if plainPassword == "" {
			var err error
			plainPassword, err = apppassword.GenerateAppPassword()
			if err != nil {
				return ErrGenerationFailure{Cause: err}
			}
		}
		if err := apppassword.ValidateAppPasswordFormat(plainPassword); err != nil {
			return ErrInvalidAppPasswordFormat{Cause: err}
		}

		appPassword, err := m.appPass.Create(intent.Email, deviceName, plainPassword)
		if err != nil {
			return ErrStoreFailure{Cause: err}
		}
		rollback := func(cause error) error {
			if revokeErr := m.appPass.Revoke(appPassword.ID); revokeErr != nil {
				return ErrStoreFailure{Cause: fmt.Errorf("%v; rollback failed: %w", cause, revokeErr)}
			}
			return cause
		}

		setup, err := m.generatePlatformSetup(platform, intent.Email, plainPassword, false)
		if err != nil {
			return rollback(err)
		}

		var profileData []byte
		if platform == PlatformIOS || platform == PlatformMacOS {
			if m.profileGen == nil {
				return rollback(ErrGenerationFailure{Cause: fmt.Errorf("mobileconfig generator not configured")})
			}
			profileCfg := mobileconfig.ProfileConfig{
				Domain:       m.config.Domain,
				MailHostname: m.config.Hostname,
				Email:        intent.Email,
				AppPassword:  plainPassword,
				CalDAVURL:    m.config.CalDAVURL,
				CardDAVURL:   m.config.CardDAVURL,
				CalDAVPort:   m.config.CalDAVPort,
				CardDAVPort:  m.config.CardDAVPort,
			}
			profileData, err = m.profileGen.GenerateProfile(profileCfg)
			if err != nil {
				return rollback(ErrGenerationFailure{Cause: err})
			}
		}

		out = &ConsumedSetup{
			Email:       intent.Email,
			AppPassword: plainPassword,
			Platform:    platform,
			Setup:       setup,
			ProfileData: profileData,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	switch state {
	case qrcode.ValidationStateValid:
		return out, nil
	case qrcode.ValidationStateInvalid:
		return nil, ErrInvalidToken{}
	case qrcode.ValidationStateExpired:
		return nil, ErrExpiredToken{}
	case qrcode.ValidationStateUsed:
		return nil, ErrUsedToken{}
	default:
		return nil, ErrInvalidToken{}
	}
}

func (m *DefaultModule) GenerateMobileConfigFromToken(token string) ([]byte, string, error) {
	setup, err := m.ConsumeSetupIntent(ConsumeSetupInput{Token: token})
	if err != nil {
		return nil, "", err
	}
	if len(setup.ProfileData) == 0 {
		return nil, "", ErrGenerationFailure{Cause: fmt.Errorf("mobileconfig artifact not generated")}
	}
	return setup.ProfileData, setup.Email, nil
}

func (m *DefaultModule) GenerateQRURL(email string) (string, time.Time, error) {
	expiresAt := m.now().Add(qrcode.DefaultTokenExpiry)
	token, err := m.tokens.Create(email, expiresAt)
	if err != nil {
		return "", time.Time{}, ErrGenerationFailure{Cause: err}
	}
	url := fmt.Sprintf("https://%s/profile/download?token=%s", m.config.Hostname, token)
	return url, expiresAt, nil
}

func (m *DefaultModule) GenerateQRPNG(email string, size int) ([]byte, error) {
	url, _, err := m.GenerateQRURL(email)
	if err != nil {
		return nil, err
	}
	png, err := qrcode.GenerateQRCodePNG(url, size)
	if err != nil {
		return nil, ErrGenerationFailure{Cause: err}
	}
	return png, nil
}

func (m *DefaultModule) GeneratePlatformSetup(platform Platform, email, appPassword string) (*PlatformSetup, error) {
	return m.generatePlatformSetup(platform, email, appPassword, true)
}

func (m *DefaultModule) generatePlatformSetup(platform Platform, email, appPassword string, issueAppleToken bool) (*PlatformSetup, error) {
	host := m.config.Hostname
	setup := &PlatformSetup{Platform: platform}

	switch platform {
	case PlatformIOS, PlatformMacOS:
		setup.Title = "iOS/macOS Setup"
		setup.Steps = []string{
			"Follow the prompts to install the configuration profile",
			"Your email, calendar, and contacts will sync automatically",
		}
		setup.DownloadLabel = "Download Profile (.mobileconfig)"
		if issueAppleToken {
			url, expiry, err := m.GenerateQRURL(email)
			if err != nil {
				return nil, err
			}
			png, err := qrcode.GenerateQRCodePNG(url, 256)
			if err != nil {
				return nil, ErrGenerationFailure{Cause: err}
			}
			setup.Steps = append([]string{
				"Scan the QR code below with your device camera, OR",
				"Click the \"Download Profile\" button below",
			}, setup.Steps...)
			setup.QRCodeData = base64.StdEncoding.EncodeToString(png)
			setup.ProfileURL = url
			setup.TokenExpiry = expiry.Format(time.RFC3339)
			setup.TokenTTL = time.Until(expiry).Round(time.Second).String()
		}

	case PlatformAndroid:
		autoconfigURL := fmt.Sprintf("https://%s/.well-known/autoconfig/mail/config-v1.1.xml?emailaddress=%s", host, email)
		png, err := qrcode.GenerateQRCodePNG(autoconfigURL, 256)
		if err != nil {
			return nil, ErrGenerationFailure{Cause: err}
		}
		setup.Title = "Android Setup"
		setup.Steps = []string{
			"Open your email app (Gmail, K-9 Mail, etc.)",
			"Add a new account",
			fmt.Sprintf("Enter your email: %s", email),
			fmt.Sprintf("Enter this app password: %s", appPassword),
			"Follow the app's configuration wizard",
		}
		setup.Fields = []InstructionField{
			{Key: "IMAP Server", Value: host + " (Port 993, SSL)"},
			{Key: "SMTP Server", Value: host + " (Port 587, STARTTLS)"},
			{Key: "Username", Value: email},
		}
		setup.QRCodeData = base64.StdEncoding.EncodeToString(png)
		setup.ProfileURL = autoconfigURL

	case PlatformThunderbird, PlatformOutlook:
		platformTitle := strings.ToUpper(string(platform[:1])) + string(platform[1:])
		setup.Title = platformTitle + " Setup"
		setup.Summary = fmt.Sprintf("Just enter your email address and this app password - %s will auto-discover the settings:", platformTitle)
		setup.Steps = []string{
			fmt.Sprintf("Open %s", platformTitle),
			"Add a new email account",
			fmt.Sprintf("Email: %s", email),
			fmt.Sprintf("Password: %s", appPassword),
			"Click \"Continue\" - settings will be detected automatically",
		}

	case PlatformOther:
		setup.Title = "Manual Setup"
		setup.Summary = "Configure your email client with these settings:"
		setup.Fields = []InstructionField{
			{Key: "Email", Value: email},
			{Key: "Password", Value: appPassword},
			{Key: "IMAP Server", Value: host},
			{Key: "IMAP Port", Value: "993 (SSL/TLS)"},
			{Key: "SMTP Server", Value: host},
			{Key: "SMTP Port", Value: "587 (STARTTLS)"},
			{Key: "Username", Value: email + " (full email address)"},
		}
	default:
		return nil, ErrUnknownPlatform{Platform: string(platform)}
	}

	return setup, nil
}
