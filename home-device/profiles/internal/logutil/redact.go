// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logutil adapts shared privacy redaction policy for profile logging.
package logutil

import "github.com/darkpipe/darkpipe/internal/privacy"

// RedactEmail masks the local-part of an email address, preserving the domain.
func RedactEmail(addr string) string {
	return privacy.RedactEmail(addr)
}

// RedactQueryParams parses a raw query string and replaces values for
// sensitive keys (emailaddress, email, token) with "[REDACTED]".
// Non-sensitive parameters are preserved as-is.
func RedactQueryParams(rawQuery string) string {
	return privacy.RedactQueryParams(rawQuery)
}
