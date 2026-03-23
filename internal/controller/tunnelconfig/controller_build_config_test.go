// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 The Cloudflare Operator Authors

package tunnelconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCloudflareConfig_AlwaysIncludesCatchAllRule(t *testing.T) {
	r := &Reconciler{}

	cfg := &TunnelConfig{
		TunnelID:  "tid",
		AccountID: "aid",
		Sources:   map[string]*SourceConfig{},
	}

	out := r.buildCloudflareConfig(cfg)
	if assert.Len(t, out.Ingress, 1) {
		assert.Equal(t, "http_status:404", out.Ingress[0].Service)
	}
}

