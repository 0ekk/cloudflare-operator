// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 The Cloudflare Operator Authors

package cf

import (
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	// CloudflareAPITimeoutEnv is the env var for Cloudflare API HTTP timeout.
	// Supports Go duration (e.g. "120s", "2m") or plain integer seconds (e.g. "120").
	CloudflareAPITimeoutEnv = "CLOUDFLARE_API_TIMEOUT"

	// DefaultCloudflareAPITimeout is used when env var is missing/invalid.
	DefaultCloudflareAPITimeout = 120 * time.Second

	// MaxCloudflareAPITimeout caps configured timeout to avoid runaway values.
	MaxCloudflareAPITimeout = 10 * time.Minute
)

// GetCloudflareAPITimeout returns Cloudflare API timeout from env with safe fallback.
func GetCloudflareAPITimeout() time.Duration {
	raw := os.Getenv(CloudflareAPITimeoutEnv)
	if raw == "" {
		return DefaultCloudflareAPITimeout
	}

	if d, err := time.ParseDuration(raw); err == nil {
		return clampTimeout(d)
	}

	if seconds, err := strconv.Atoi(raw); err == nil {
		return clampTimeout(time.Duration(seconds) * time.Second)
	}

	return DefaultCloudflareAPITimeout
}

func clampTimeout(d time.Duration) time.Duration {
	if d <= 0 {
		return DefaultCloudflareAPITimeout
	}
	if d > MaxCloudflareAPITimeout {
		return MaxCloudflareAPITimeout
	}
	return d
}

// NewAPIHTTPClient creates an HTTP client with operator-wide Cloudflare API timeout.
func NewAPIHTTPClient() *http.Client {
	return &http.Client{
		Timeout: GetCloudflareAPITimeout(),
	}
}
