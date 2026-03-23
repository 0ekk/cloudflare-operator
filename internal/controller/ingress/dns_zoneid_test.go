// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 The Cloudflare Operator Authors

package ingress

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	networkingv1alpha2 "github.com/0ekk/cloudflare-operator/api/v1alpha2"
)

func TestReconcileDNSRecords_UsesTunnelStatusZoneIDWhenSpecMissing(t *testing.T) {
	scheme := setupTestScheme(t)
	ctx := context.Background()

	tunnel := &networkingv1alpha2.Tunnel{
		ObjectMeta: metav1.ObjectMeta{Name: "k8s-tunnel", Namespace: "default"},
		Spec: networkingv1alpha2.TunnelSpec{
			Cloudflare: networkingv1alpha2.CloudflareDetails{
				Domain: "nixai.de",
				CredentialsRef: &networkingv1alpha2.CloudflareCredentialsRef{
					Name: "default",
				},
			},
		},
		Status: networkingv1alpha2.TunnelStatus{
			TunnelId: "073ccf1c-238c-4ac8-9249-cc290b4aaade",
			ZoneId:   "c7bfa5480177d069702856e32003a55c",
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tunnel).Build()
	r := &Reconciler{Client: fakeClient, Scheme: scheme, OperatorNamespace: "cloudflare-system"}

	ing := &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "open-webui", Namespace: "default"}}
	cfg := &networkingv1alpha2.TunnelIngressClassConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "cf-tunnel", Namespace: "default"},
		Spec: networkingv1alpha2.TunnelIngressClassConfigSpec{
			DNSManagement: networkingv1alpha2.DNSManagementAutomatic,
			TunnelRef:     networkingv1alpha2.TunnelReference{Kind: "Tunnel", Name: "k8s-tunnel"},
		},
	}

	err := r.reconcileDNSRecords(ctx, ing, []string{"open-webui.nixai.de"}, cfg)
	require.NoError(t, err)

	var created networkingv1alpha2.DNSRecord
	err = fakeClient.Get(ctx, client.ObjectKey{Name: r.sanitizeDNSRecordName("open-webui.nixai.de", ing), Namespace: "default"}, &created)
	require.NoError(t, err)

	assert.Equal(t, "c7bfa5480177d069702856e32003a55c", created.Spec.Cloudflare.ZoneId)
	assert.Equal(t, "nixai.de", created.Spec.Cloudflare.Domain)
	require.NotNil(t, created.Spec.Cloudflare.CredentialsRef)
	assert.Equal(t, "default", created.Spec.Cloudflare.CredentialsRef.Name)
}
