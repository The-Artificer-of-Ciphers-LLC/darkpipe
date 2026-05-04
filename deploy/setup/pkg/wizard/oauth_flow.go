// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

package wizard

import (
	"context"

	"github.com/darkpipe/darkpipe/deploy/setup/pkg/oauthflow"
	"github.com/darkpipe/darkpipe/deploy/setup/pkg/providers"
	"golang.org/x/oauth2"
)

// RunOAuthDeviceFlow executes OAuth2 device flow via oauthflow module.
func RunOAuthDeviceFlow(ctx context.Context, config *providers.OAuthConfig) (*oauth2.Token, error) {
	return oauthflow.New().RunDeviceFlow(ctx, *config, oauthflow.NewPTermUI(), oauthflow.Config{})
}
