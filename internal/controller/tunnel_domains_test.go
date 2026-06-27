package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkingv1alpha2 "github.com/0ekk/cloudflare-operator/api/v1alpha2"
)

func TestApplyTunnelDomainStatuses_PartialFailureSetsDomainsReadyFalse(t *testing.T) {
	status := networkingv1alpha2.TunnelStatus{}

	applyTunnelDomainStatuses(&status, []tunnelDomainResolution{
		{
			Domain: "linuxpods.com",
			ZoneID: "zone-linuxpods",
		},
		{
			Domain: "missing.example",
			Err:    assert.AnError,
		},
	}, 7)

	require.Len(t, status.Domains, 2)
	assert.Equal(t, "linuxpods.com", status.Domains[0].Domain)
	assert.Equal(t, "zone-linuxpods", status.Domains[0].ZoneId)
	assert.Equal(t, networkingv1alpha2.TunnelDomainStateReady, status.Domains[0].State)
	assert.Equal(t, "missing.example", status.Domains[1].Domain)
	assert.Equal(t, networkingv1alpha2.TunnelDomainStateError, status.Domains[1].State)

	condition := meta.FindStatusCondition(status.Conditions, "DomainsReady")
	require.NotNil(t, condition)
	assert.Equal(t, metav1.ConditionFalse, condition.Status)
	assert.Equal(t, int64(7), condition.ObservedGeneration)
}
