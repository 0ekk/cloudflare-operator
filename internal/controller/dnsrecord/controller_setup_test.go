// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 The Cloudflare Operator Authors

package dnsrecord

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestSourceWatchRegistrations_GatewayDisabled(t *testing.T) {
	r := &DNSRecordReconciler{GatewayAPIAvailable: false}
	watches := r.sourceWatchRegistrations()

	assert.Len(t, watches, 3)
	assert.IsType(t, &corev1.Service{}, watches[0].object)
	assert.IsType(t, &networkingv1.Ingress{}, watches[1].object)
	assert.IsType(t, &corev1.Node{}, watches[2].object)
}

func TestSourceWatchRegistrations_GatewayEnabled(t *testing.T) {
	r := &DNSRecordReconciler{GatewayAPIAvailable: true}
	watches := r.sourceWatchRegistrations()

	assert.Len(t, watches, 5)
	assert.IsType(t, &corev1.Service{}, watches[0].object)
	assert.IsType(t, &networkingv1.Ingress{}, watches[1].object)
	assert.IsType(t, &gatewayv1.Gateway{}, watches[2].object)
	assert.IsType(t, &gatewayv1.HTTPRoute{}, watches[3].object)
	assert.IsType(t, &corev1.Node{}, watches[4].object)
}
