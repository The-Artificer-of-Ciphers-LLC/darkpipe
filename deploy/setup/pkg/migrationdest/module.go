// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package migrationdest

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-webdav/caldav"
	"github.com/emersion/go-webdav/carddav"
)

type TLSPolicy string

const (
	RequireTLS              TLSPolicy = "require_tls"
	AllowInsecureForLocalDev TLSPolicy = "allow_insecure_for_local_dev"
)

type Config struct {
	DestIMAP    string
	DestCalDAV  string
	DestCardDAV string
	DestUser    string
	DestPass    string
	TLSPolicy   TLSPolicy
}

type Capability string

const (
	CapabilityIMAP    Capability = "imap"
	CapabilityCalDAV  Capability = "caldav"
	CapabilityCardDAV Capability = "carddav"
)

type DestinationAdapters struct {
	IMAP    *imapclient.Client
	CalDAV  *caldav.Client
	CardDAV *carddav.Client
	Warnings []error
}

type Module interface {
	Validate(cfg Config) error
	Connect(ctx context.Context, cfg Config) (*DestinationAdapters, error)
	Capabilities(cfg Config) map[Capability]bool
}

type ErrInvalidConfig struct{ Msg string }
func (e ErrInvalidConfig) Error() string { return e.Msg }

type ErrOptionalAdapterUnavailable struct {
	Adapter string
	Cause error
}
func (e ErrOptionalAdapterUnavailable) Error() string { return fmt.Sprintf("optional %s unavailable: %v", e.Adapter, e.Cause) }

type DefaultModule struct{}

func New() Module { return &DefaultModule{} }

func (m *DefaultModule) Validate(cfg Config) error {
	if strings.TrimSpace(cfg.DestUser) == "" || strings.TrimSpace(cfg.DestPass) == "" {
		return ErrInvalidConfig{Msg: "destination username/password required"}
	}
	if strings.TrimSpace(cfg.DestIMAP) == "" {
		return ErrInvalidConfig{Msg: "destination IMAP endpoint required"}
	}
	if cfg.TLSPolicy == "" {
		cfg.TLSPolicy = RequireTLS
	}
	if cfg.TLSPolicy != RequireTLS && cfg.TLSPolicy != AllowInsecureForLocalDev {
		return ErrInvalidConfig{Msg: "invalid TLS policy"}
	}
	return nil
}

func (m *DefaultModule) Capabilities(cfg Config) map[Capability]bool {
	return map[Capability]bool{
		CapabilityIMAP: true,
		CapabilityCalDAV: strings.TrimSpace(cfg.DestCalDAV) != "",
		CapabilityCardDAV: strings.TrimSpace(cfg.DestCardDAV) != "",
	}
}

func (m *DefaultModule) Connect(ctx context.Context, cfg Config) (*DestinationAdapters, error) {
	if err := m.Validate(cfg); err != nil { return nil, err }

	host, port, err := parseHostPort(cfg.DestIMAP, 993)
	if err != nil { return nil, err }

	p := &providers.GenericProvider{
		IMAPHost: host,
		IMAPPort: port,
		Username: cfg.DestUser,
		Password: cfg.DestPass,
		CalDAVURL: cfg.DestCalDAV,
		CardDAVURL: cfg.DestCardDAV,
		UseTLS: cfg.TLSPolicy != AllowInsecureForLocalDev,
	}

	imapClient, err := p.ConnectIMAP(ctx)
	if err != nil { return nil, err }

	out := &DestinationAdapters{IMAP: imapClient}
	if cfg.DestCalDAV != "" {
		c, err := p.ConnectCalDAV(ctx)
		if err != nil { out.Warnings = append(out.Warnings, ErrOptionalAdapterUnavailable{Adapter:"caldav", Cause: err}) } else { out.CalDAV = c }
	}
	if cfg.DestCardDAV != "" {
		c, err := p.ConnectCardDAV(ctx)
		if err != nil { out.Warnings = append(out.Warnings, ErrOptionalAdapterUnavailable{Adapter:"carddav", Cause: err}) } else { out.CardDAV = c }
	}
	return out, nil
}

func parseHostPort(endpoint string, defaultPort int) (string, int, error) {
	e := strings.TrimSpace(endpoint)
	if e == "" { return "", 0, ErrInvalidConfig{Msg: "empty endpoint"} }
	parts := strings.Split(e, ":")
	if len(parts) == 1 { return parts[0], defaultPort, nil }
	if len(parts) == 2 {
		p, err := strconv.Atoi(parts[1]); if err != nil { return "",0, ErrInvalidConfig{Msg:"invalid IMAP port"} }
		return parts[0], p, nil
	}
	return "", 0, ErrInvalidConfig{Msg: "invalid endpoint format"}
}
