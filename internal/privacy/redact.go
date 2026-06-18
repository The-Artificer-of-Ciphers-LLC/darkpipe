// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package privacy centralizes privacy-preserving formatting for logs and
// generated diagnostics.
package privacy

import (
	"net/url"
	"strings"
)

// RedactEmail masks the local-part of an email address, preserving the domain.
func RedactEmail(addr string) string {
	if addr == "" {
		return ""
	}

	at := strings.LastIndex(addr, "@")
	if at < 0 {
		return addr
	}

	local := addr[:at]
	domain := addr[at:] // includes "@"

	switch len(local) {
	case 0:
		return domain
	case 1:
		return "*" + domain
	case 2:
		return string(local[0]) + "*" + domain
	default:
		return string(local[0]) + "***" + string(local[len(local)-1]) + domain
	}
}

// RedactEmails applies RedactEmail to each element of a string slice.
func RedactEmails(addrs []string) []string {
	if addrs == nil {
		return nil
	}
	out := make([]string, len(addrs))
	for i, a := range addrs {
		out[i] = RedactEmail(a)
	}
	return out
}

var sensitiveQueryKeys = map[string]bool{
	"emailaddress": true,
	"email":        true,
	"token":        true,
}

// RedactQueryParams parses a raw query string and replaces values for
// sensitive keys with "[REDACTED]". Non-sensitive parameters are preserved.
func RedactQueryParams(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}

	params, err := url.ParseQuery(rawQuery)
	if err != nil {
		return "[REDACTED]"
	}

	for key := range params {
		if sensitiveQueryKeys[strings.ToLower(key)] {
			params.Set(key, "[REDACTED]")
		}
	}

	return params.Encode()
}
