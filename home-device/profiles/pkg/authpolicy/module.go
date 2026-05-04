// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package authpolicy

import (
	"fmt"
	"strings"
)

// Module verifies adapter-supplied credentials against auth policy.
type Module interface {
	Verify(username, password string, requireEmailUsername bool) (string, error)
}

type Config struct {
	AdminUser string
	AdminPass string
}

type ErrUnauthorized struct{}

func (e ErrUnauthorized) Error() string { return "unauthorized" }

type ErrMisconfigured struct{ Reason string }

func (e ErrMisconfigured) Error() string {
	return fmt.Sprintf("misconfigured auth policy: %s", e.Reason)
}

type ErrPolicyViolation struct{ Reason string }

func (e ErrPolicyViolation) Error() string { return fmt.Sprintf("policy violation: %s", e.Reason) }

type DefaultModule struct {
	cfg Config
}

func New(cfg Config) Module { return &DefaultModule{cfg: cfg} }

func (m *DefaultModule) Verify(username, password string, requireEmailUsername bool) (string, error) {
	if m.cfg.AdminUser == "" || m.cfg.AdminPass == "" {
		return "", ErrMisconfigured{Reason: "admin credentials not configured"}
	}
	if requireEmailUsername && !strings.Contains(username, "@") {
		return "", ErrPolicyViolation{Reason: "username must be an email address"}
	}
	if username != m.cfg.AdminUser || password != m.cfg.AdminPass {
		return "", ErrUnauthorized{}
	}
	return username, nil
}
