package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/0ekk/cloudflare-operator/internal/clients/cf"
)

func TestApplyResolvedAPIContext_SetsTunnelID(t *testing.T) {
	api := &cf.API{}

	applyResolvedAPIContext(api, "acc-1", "example.com", "tunnel-123")

	assert.Equal(t, "acc-1", api.ValidAccountId)
	assert.Equal(t, "example.com", api.Domain)
	assert.Equal(t, "tunnel-123", api.ValidTunnelId)
}
