// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package oauthflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
	"golang.org/x/oauth2"
)

// UINotifier is a UI adapter seam for device flow UX.
type UINotifier interface {
	ShowVerification(verificationURL, userCode string)
	WaitingStart(message string) (stop func())
	ShowSuccess()
	ShowError(err error)
}

// Config controls module behavior.
type Config struct {
	PollInterval time.Duration
	MaxWait      time.Duration
}

// Module is the OAuth device-flow seam.
type Module interface {
	RunDeviceFlow(ctx context.Context, cfg providers.OAuthConfig, ui UINotifier, opts Config) (*oauth2.Token, error)
}

type DefaultModule struct{}

func New() Module { return &DefaultModule{} }

type ErrDenied struct{}

func (e ErrDenied) Error() string { return "oauth authorization denied" }

type ErrExpiredCode struct{}

func (e ErrExpiredCode) Error() string { return "oauth device code expired" }

type ErrRateLimited struct{ RetryAfter time.Duration }

func (e ErrRateLimited) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("oauth rate limited; retry after %s", e.RetryAfter)
	}
	return "oauth rate limited"
}

type ErrTemporaryNetwork struct{ Cause error }

func (e ErrTemporaryNetwork) Error() string {
	return fmt.Sprintf("oauth temporary network error: %v", e.Cause)
}

type ErrTimeout struct{}

func (e ErrTimeout) Error() string { return "oauth flow timed out" }

func (m *DefaultModule) RunDeviceFlow(ctx context.Context, cfg providers.OAuthConfig, ui UINotifier, opts Config) (*oauth2.Token, error) {
	if opts.MaxWait > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.MaxWait)
		defer cancel()
	}

	var stop func()
	token, err := providers.RunDeviceFlow(ctx, &cfg, func(verificationURL, userCode string) {
		if ui == nil {
			return
		}
		ui.ShowVerification(verificationURL, userCode)
		stop = ui.WaitingStart("Waiting for authorization... (complete in your browser)")
	})

	if stop != nil {
		stop()
	}

	if err != nil {
		typed := classifyError(ctx, err)
		if ui != nil {
			ui.ShowError(typed)
		}
		return nil, typed
	}

	if ui != nil {
		ui.ShowSuccess()
	}
	return token, nil
}

func classifyError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "deadline") || strings.Contains(strings.ToLower(err.Error()), "timeout") {
		return ErrTimeout{}
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "access_denied"), strings.Contains(msg, "authorization denied"), strings.Contains(msg, "denied"):
		return ErrDenied{}
	case strings.Contains(msg, "expired_token"), strings.Contains(msg, "expired"):
		return ErrExpiredCode{}
	case strings.Contains(msg, "slow_down"), strings.Contains(msg, "too many"), strings.Contains(msg, "rate"):
		return ErrRateLimited{}
	case strings.Contains(msg, "temporar"), strings.Contains(msg, "connection"), strings.Contains(msg, "dial"), strings.Contains(msg, "network"):
		return ErrTemporaryNetwork{Cause: err}
	default:
		return err
	}
}
