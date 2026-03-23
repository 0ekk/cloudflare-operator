// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 The Cloudflare Operator Authors

package tunnelconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewConfigMap_SkipsCrossNamespaceOwnerRef(t *testing.T) {
	owner := &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k8s-tunnel",
			Namespace: "default",
			UID:       "12345",
		},
	}

	cm := NewConfigMap(
		"cloudflare-operator-system",
		"18005f22-90cd-4150-98de-201977a3cb04",
		owner,
		metav1.GroupVersionKind{
			Group:   "networking.cloudflare-operator.io",
			Version: "v1alpha2",
			Kind:    "Tunnel",
		},
	)

	assert.Empty(t, cm.OwnerReferences, "cross-namespace namespaced ownerRef must be omitted")
}

func TestNewConfigMap_SetsOwnerRefForClusterScopedOwner(t *testing.T) {
	owner := &metav1.PartialObjectMetadata{
		ObjectMeta: metav1.ObjectMeta{
			Name: "shared-tunnel",
			UID:  "67890",
		},
	}

	cm := NewConfigMap(
		"cloudflare-operator-system",
		"18005f22-90cd-4150-98de-201977a3cb04",
		owner,
		metav1.GroupVersionKind{
			Group:   "networking.cloudflare-operator.io",
			Version: "v1alpha2",
			Kind:    "ClusterTunnel",
		},
	)

	if assert.Len(t, cm.OwnerReferences, 1) {
		assert.Equal(t, "shared-tunnel", cm.OwnerReferences[0].Name)
		assert.Equal(t, "ClusterTunnel", cm.OwnerReferences[0].Kind)
	}
}
