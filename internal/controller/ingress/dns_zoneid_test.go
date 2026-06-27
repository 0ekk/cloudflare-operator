// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 The Cloudflare Operator Authors

package ingress

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
			Cloudflare: networkingv1alpha2.TunnelCloudflareDetails{
				CloudflareDetails: networkingv1alpha2.CloudflareDetails{
					Domain: "nixai.de",
					CredentialsRef: &networkingv1alpha2.CloudflareCredentialsRef{
						Name: "default",
					},
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

func TestReconcileDNSRecords_UsesMatchedDomainOrDefaultSuffixAndDeletesStaleRecords(t *testing.T) {
	scheme := setupTestScheme(t)
	ctx := context.Background()

	tunnel := &networkingv1alpha2.Tunnel{
		ObjectMeta: metav1.ObjectMeta{Name: "k8s-tunnel", Namespace: "default"},
		Spec: networkingv1alpha2.TunnelSpec{
			Cloudflare: networkingv1alpha2.TunnelCloudflareDetails{
				CloudflareDetails: networkingv1alpha2.CloudflareDetails{
					Domain: "nixai.de",
					CredentialsRef: &networkingv1alpha2.CloudflareCredentialsRef{
						Name: "default",
					},
				},
				Domains: []string{"linuxpods.com"},
			},
		},
		Status: networkingv1alpha2.TunnelStatus{
			TunnelId: "073ccf1c-238c-4ac8-9249-cc290b4aaade",
			ZoneId:   "zone-nixai",
			Domains: []networkingv1alpha2.TunnelDomainStatus{
				{
					Domain: "linuxpods.com",
					ZoneId: "zone-linuxpods",
					State:  networkingv1alpha2.TunnelDomainStateReady,
				},
			},
		},
	}
	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: "code-server", Namespace: "default"},
	}
	stale := &networkingv1alpha2.DNSRecord{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "code-server-code-evil-com",
			Namespace: "default",
			Labels: map[string]string{
				ManagedByAnnotation:           ManagedByValue,
				"cloudflare.com/ingress-name": "code-server",
			},
		},
		Spec: networkingv1alpha2.DNSRecordSpec{
			Name:    "code.evil.com",
			Type:    "CNAME",
			Content: "old.cfargotunnel.com",
			Cloudflare: networkingv1alpha2.CloudflareDetails{
				ZoneId: "zone-nixai",
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tunnel, ing, stale).Build()
	r := &Reconciler{Client: fakeClient, Scheme: scheme, OperatorNamespace: "cloudflare-system"}
	cfg := &networkingv1alpha2.TunnelIngressClassConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "cf-tunnel", Namespace: "default"},
		Spec: networkingv1alpha2.TunnelIngressClassConfigSpec{
			DNSManagement: networkingv1alpha2.DNSManagementAutomatic,
			TunnelRef:     networkingv1alpha2.TunnelReference{Kind: "Tunnel", Name: "k8s-tunnel"},
		},
	}

	err := r.reconcileDNSRecords(ctx, ing, []string{"code.linuxpods.com", "code-server"}, cfg)
	require.NoError(t, err)

	var created networkingv1alpha2.DNSRecord
	err = fakeClient.Get(ctx, client.ObjectKey{Name: r.sanitizeDNSRecordName("code.linuxpods.com", ing), Namespace: "default"}, &created)
	require.NoError(t, err)
	assert.Equal(t, "code.linuxpods.com", created.Spec.Name)
	assert.Equal(t, "zone-linuxpods", created.Spec.Cloudflare.ZoneId)
	assert.Equal(t, "linuxpods.com", created.Spec.Cloudflare.Domain)
	require.NotNil(t, created.Spec.Cloudflare.CredentialsRef)
	assert.Equal(t, "default", created.Spec.Cloudflare.CredentialsRef.Name)

	var defaulted networkingv1alpha2.DNSRecord
	err = fakeClient.Get(ctx, client.ObjectKey{Name: r.sanitizeDNSRecordName("code-server.nixai.de", ing), Namespace: "default"}, &defaulted)
	require.NoError(t, err)
	assert.Equal(t, "code-server.nixai.de", defaulted.Spec.Name)
	assert.Equal(t, "zone-nixai", defaulted.Spec.Cloudflare.ZoneId)
	assert.Equal(t, "nixai.de", defaulted.Spec.Cloudflare.Domain)

	var invalid networkingv1alpha2.DNSRecord
	err = fakeClient.Get(ctx, client.ObjectKey{Name: stale.Name, Namespace: stale.Namespace}, &invalid)
	assert.True(t, apierrors.IsNotFound(err))
}

func TestReconcileDNSRecords_SkipsUnresolvedConfiguredDomainAndDefaultsOtherHosts(t *testing.T) {
	scheme := setupTestScheme(t)
	ctx := context.Background()

	tunnel := &networkingv1alpha2.Tunnel{
		ObjectMeta: metav1.ObjectMeta{Name: "k8s-tunnel", Namespace: "default"},
		Spec: networkingv1alpha2.TunnelSpec{
			Cloudflare: networkingv1alpha2.TunnelCloudflareDetails{
				CloudflareDetails: networkingv1alpha2.CloudflareDetails{
					Domain: "nixai.de",
					CredentialsRef: &networkingv1alpha2.CloudflareCredentialsRef{
						Name: "default",
					},
				},
				Domains: []string{"linuxpods.com"},
			},
		},
		Status: networkingv1alpha2.TunnelStatus{
			TunnelId: "073ccf1c-238c-4ac8-9249-cc290b4aaade",
			ZoneId:   "zone-nixai",
			Domains: []networkingv1alpha2.TunnelDomainStatus{
				{
					Domain: "linuxpods.com",
					State:  networkingv1alpha2.TunnelDomainStateError,
				},
			},
		},
	}

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(tunnel).Build()
	r := &Reconciler{Client: fakeClient, Scheme: scheme, OperatorNamespace: "cloudflare-system"}
	ing := &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "code-server", Namespace: "default"}}
	cfg := &networkingv1alpha2.TunnelIngressClassConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "cf-tunnel", Namespace: "default"},
		Spec: networkingv1alpha2.TunnelIngressClassConfigSpec{
			DNSManagement: networkingv1alpha2.DNSManagementAutomatic,
			TunnelRef:     networkingv1alpha2.TunnelReference{Kind: "Tunnel", Name: "k8s-tunnel"},
		},
	}

	err := r.reconcileDNSRecords(ctx, ing, []string{"code-server", "code.linuxpods.com"}, cfg)
	require.NoError(t, err)

	var record networkingv1alpha2.DNSRecord
	err = fakeClient.Get(ctx, client.ObjectKey{Name: r.sanitizeDNSRecordName("code-server.nixai.de", ing), Namespace: "default"}, &record)
	require.NoError(t, err)
	assert.Equal(t, "code-server.nixai.de", record.Spec.Name)
	assert.Equal(t, "zone-nixai", record.Spec.Cloudflare.ZoneId)

	err = fakeClient.Get(ctx, client.ObjectKey{Name: r.sanitizeDNSRecordName("code.linuxpods.com", ing), Namespace: "default"}, &record)
	assert.True(t, apierrors.IsNotFound(err))
}
