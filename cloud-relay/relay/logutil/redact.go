// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logutil adapts shared privacy redaction policy for relay logging.
package logutil

import "github.com/darkpipe/darkpipe/internal/privacy"

// RedactEmail masks the local-part of an email address, preserving the domain.
func RedactEmail(addr string) string {
	return privacy.RedactEmail(addr)
}

// RedactEmails applies RedactEmail to each element of a string slice.
func RedactEmails(addrs []string) []string {
	return privacy.RedactEmails(addrs)
}
