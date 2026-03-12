// Copyright (C) 2026 The Artificer of Ciphers, LLC. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-or-later


package health

import (
	"encoding/json"
	"net/http"
)

// jsonError writes a structured JSON error response with the given message and HTTP status code.
func jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"error": message, "code": code})
}

// LivenessHandler returns an HTTP handler for liveness checks.
// GET /health/live always returns 200 OK if the process is running.
func LivenessHandler(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := checker.Liveness()

		w.Header().Set("Content-Type", "application/health+json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(status)
	}
}

// ReadinessHandler returns an HTTP handler for readiness checks.
// GET /health/ready returns 200 OK if all checks pass, 503 Service Unavailable if any fail.
func ReadinessHandler(checker *Checker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			jsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		status := checker.Readiness(ctx)

		w.Header().Set("Content-Type", "application/health+json")

		if status.Status == "up" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(status)
	}
}
