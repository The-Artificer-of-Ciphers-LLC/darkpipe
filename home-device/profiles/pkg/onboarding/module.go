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

type Module interface {
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

func (m *DefaultModule) GenerateMobileConfigFromToken(token string) ([]byte, string, error) {
	email, state, err := m.tokens.Validate(token)
	if err != nil {
		return nil, "", ErrStoreFailure{Cause: err}
	}
	switch state {
	case qrcode.ValidationStateValid:
	case qrcode.ValidationStateInvalid:
		return nil, "", ErrInvalidToken{}
	case qrcode.ValidationStateExpired:
		return nil, "", ErrExpiredToken{}
	case qrcode.ValidationStateUsed:
		return nil, "", ErrUsedToken{}
	default:
		return nil, "", ErrInvalidToken{}
	}

	deviceName := fmt.Sprintf("QR-%d", m.now().Unix())
	plainPassword, err := apppassword.GenerateAppPassword()
	if err != nil {
		return nil, "", ErrGenerationFailure{Cause: err}
	}
	if _, err := m.appPass.Create(email, deviceName, plainPassword); err != nil {
		return nil, "", ErrStoreFailure{Cause: err}
	}

	profileCfg := mobileconfig.ProfileConfig{
		Domain:       m.config.Domain,
		MailHostname: m.config.Hostname,
		Email:        email,
		AppPassword:  plainPassword,
		CalDAVURL:    m.config.CalDAVURL,
		CardDAVURL:   m.config.CardDAVURL,
		CalDAVPort:   m.config.CalDAVPort,
		CardDAVPort:  m.config.CardDAVPort,
	}

	profileData, err := m.profileGen.GenerateProfile(profileCfg)
	if err != nil {
		return nil, "", ErrGenerationFailure{Cause: err}
	}

	return profileData, email, nil
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
	host := m.config.Hostname
	setup := &PlatformSetup{Platform: platform}

	switch platform {
	case PlatformIOS, PlatformMacOS:
		url, expiry, err := m.GenerateQRURL(email)
		if err != nil {
			return nil, err
		}
		png, err := qrcode.GenerateQRCodePNG(url, 256)
		if err != nil {
			return nil, ErrGenerationFailure{Cause: err}
		}
		setup.Title = "iOS/macOS Setup"
		setup.Steps = []string{
			"Scan the QR code below with your device camera, OR",
			"Click the \"Download Profile\" button below",
			"Follow the prompts to install the configuration profile",
			"Your email, calendar, and contacts will sync automatically",
		}
		setup.DownloadLabel = "Download Profile (.mobileconfig)"
		setup.QRCodeData = base64.StdEncoding.EncodeToString(png)
		setup.ProfileURL = url
		setup.TokenExpiry = expiry.Format(time.RFC3339)
		setup.TokenTTL = time.Until(expiry).Round(time.Second).String()

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
