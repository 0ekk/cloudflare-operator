// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 The Cloudflare Operator Authors

package gateway

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	networkingv1alpha2 "github.com/0ekk/cloudflare-operator/api/v1alpha2"
	tunnelpkg "github.com/0ekk/cloudflare-operator/internal/controller/tunnel"
)

func setupGatewayTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(s))
	require.NoError(t, networkingv1alpha2.AddToScheme(s))
	return s
}

func TestReconcileDNS_DNSRecordModeUsesTunnelStatusZoneIDWhenSpecMissing(t *testing.T) {
	scheme := setupGatewayTestScheme(t)
	ctx := context.Background()

	tunnel := &networkingv1alpha2.Tunnel{
		ObjectMeta: metav1.ObjectMeta{Name: "k8s-tunnel", Namespace: "default"},
		Spec: networkingv1alpha2.TunnelSpec{
			Cloudflare: networkingv1alpha2.CloudflareDetails{
				Domain:         "nixai.de",
				CredentialsRef: &networkingv1alpha2.CloudflareCredentialsRef{Name: "default"},
			},
		},
		Status: networkingv1alpha2.TunnelStatus{
			TunnelId: "073ccf1c-238c-4ac8-9249-cc290b4aaade",
			ZoneId:   "c7bfa5480177d069702856e32003a55c",
		},
	}

	r := &GatewayReconciler{Client: fake.NewClientBuilder().WithScheme(scheme).Build(), Scheme: scheme}
	gw := &gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "gw", Namespace: "default", UID: types.UID("gw-uid")}}
	cfg := &networkingv1alpha2.TunnelGatewayClassConfig{
		Spec: networkingv1alpha2.TunnelGatewayClassConfigSpec{
			DNSManagement: networkingv1alpha2.DNSManagementDNSRecord,
		},
	}

	err := r.reconcileDNS(ctx, gw, &tunnelpkg.TunnelWrapper{Tunnel: tunnel}, cfg, []string{"open-webui.nixai.de"}, "073ccf1c-238c-4ac8-9249-cc290b4aaade.cfargotunnel.com")
	require.NoError(t, err)

	var created networkingv1alpha2.DNSRecord
	err = r.Get(ctx, client.ObjectKey{Name: "gw-open-webui-nixai-de", Namespace: "default"}, &created)
	require.NoError(t, err)

	assert.Equal(t, "c7bfa5480177d069702856e32003a55c", created.Spec.Cloudflare.ZoneId)
	assert.Equal(t, "nixai.de", created.Spec.Cloudflare.Domain)
	require.NotNil(t, created.Spec.Cloudflare.CredentialsRef)
	assert.Equal(t, "default", created.Spec.Cloudflare.CredentialsRef.Name)
}
