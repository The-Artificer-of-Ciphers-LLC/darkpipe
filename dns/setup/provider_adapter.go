// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package setup

import (
	"context"
	"fmt"

	"github.com/darkpipe/darkpipe/dns/provider"
)

type ProviderContext struct {
	Domain string
	ZoneID string
}

type ErrUnsupportedRecordType struct{ RecordType string }

func (e ErrUnsupportedRecordType) Error() string {
	return fmt.Sprintf("unsupported record type: %s", e.RecordType)
}

type RecordApplyResult struct {
	RecordType string
	Name       string
	Action     string // create|update|skip|fail
	Reason     string
}

type ProviderCapabilities struct {
	CanCreate bool
	CanUpdate bool
	CanList   bool
	CanDelete bool
	Types     map[string]bool
}

type ProviderAdapter struct {
	inner provider.DNSProvider
	ctx   ProviderContext
	caps  ProviderCapabilities
}

func NewProviderAdapter(ctx context.Context, domain string) (*ProviderAdapter, error) {
	p, err := provider.NewProviderFromDetection(ctx, domain, nil)
	if err != nil || p == nil {
		return nil, ErrManualGuideRequired{Provider: "unknown", Domain: domain}
	}

	zoneID, _ := p.GetZoneID(ctx, domain)
	caps := ProviderCapabilities{
		CanCreate: true,
		CanUpdate: true,
		CanList:   true,
		CanDelete: false,
		Types: map[string]bool{
			"TXT":   true,
			"MX":    true,
			"CNAME": true,
			"A":     true,
			"SRV":   false,
		},
	}

	return &ProviderAdapter{inner: p, ctx: ProviderContext{Domain: domain, ZoneID: zoneID}, caps: caps}, nil
}

func (a *ProviderAdapter) Capabilities() ProviderCapabilities { return a.caps }

func (a *ProviderAdapter) ApplyRecord(ctx context.Context, rec provider.Record) RecordApplyResult {
	out := RecordApplyResult{RecordType: rec.Type, Name: rec.Name}
	if !a.caps.Types[rec.Type] {
		out.Action = "skip"
		out.Reason = ErrUnsupportedRecordType{RecordType: rec.Type}.Error()
		return out
	}

	existing, err := a.inner.ListRecords(ctx, provider.RecordFilter{Type: rec.Type, Name: rec.Name})
	if err != nil {
		out.Action = "fail"
		out.Reason = err.Error()
		return out
	}
	if len(existing) == 0 {
		if err := a.inner.CreateRecord(ctx, rec); err != nil {
			out.Action = "fail"
			out.Reason = err.Error()
			return out
		}
		out.Action = "create"
		return out
	}

	if err := a.inner.UpdateRecord(ctx, existing[0].ID, rec); err != nil {
		out.Action = "fail"
		out.Reason = err.Error()
		return out
	}
	out.Action = "update"
	return out
}
