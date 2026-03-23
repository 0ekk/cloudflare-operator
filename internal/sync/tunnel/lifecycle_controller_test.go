// SPDX-License-Identifier: Apache-2.0
// Copyright 2025-2026 The Cloudflare Operator Authors

package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/0ekk/cloudflare-operator/api/v1alpha2"
	"github.com/0ekk/cloudflare-operator/internal/clients/cf"
	tunnelsvc "github.com/0ekk/cloudflare-operator/internal/service/tunnel"
)

func init() {
	_ = v1alpha2.AddToScheme(scheme.Scheme)
}

// createLifecycleConfig creates a test lifecycle config with the given action
func createLifecycleConfig(action tunnelsvc.LifecycleAction, tunnelName, tunnelID string) *tunnelsvc.LifecycleConfig {
	return &tunnelsvc.LifecycleConfig{
		Action:     action,
		TunnelName: tunnelName,
		TunnelID:   tunnelID,
		ConfigSrc:  "local",
	}
}

// createLifecycleSyncState creates a test SyncState for tunnel lifecycle
func createLifecycleSyncState(
	name string,
	config *tunnelsvc.LifecycleConfig,
	status v1alpha2.SyncStatus,
	withFinalizer bool,
) *v1alpha2.CloudflareSyncState {
	configBytes, _ := json.Marshal(config)

	syncState := &v1alpha2.CloudflareSyncState{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha2.CloudflareSyncStateSpec{
			ResourceType: v1alpha2.SyncResourceTunnelLifecycle,
			CloudflareID: config.TunnelID,
			AccountID:    "test-account-123",
			CredentialsRef: v1alpha2.CredentialsReference{
				Name: "test-credentials",
			},
			Sources: []v1alpha2.ConfigSource{
				{
					Ref: v1alpha2.SourceReference{
						Kind:      "Tunnel",
						Namespace: "test-ns",
						Name:      "test-tunnel",
					},
					Config:   runtime.RawExtension{Raw: configBytes},
					Priority: 10,
				},
			},
		},
		Status: v1alpha2.CloudflareSyncStateStatus{
			SyncStatus: status,
		},
	}

	if withFinalizer {
		syncState.Finalizers = []string{TunnelLifecycleFinalizerName}
	}

	return syncState
}

func TestNewLifecycleController(t *testing.T) {
	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	c := NewLifecycleController(client)

	require.NotNil(t, c)
	assert.NotNil(t, c.BaseSyncController)
	assert.NotNil(t, c.Client)
	assert.NotNil(t, c.Debouncer)
}

func TestLifecycleController_GetLifecycleConfig_Success(t *testing.T) {
	config := createLifecycleConfig(tunnelsvc.LifecycleActionCreate, "my-tunnel", "")
	syncState := createLifecycleSyncState("test-sync", config, v1alpha2.SyncStatusPending, true)

	c := &LifecycleController{}
	result, err := c.getLifecycleConfig(syncState)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, tunnelsvc.LifecycleActionCreate, result.Action)
	assert.Equal(t, "my-tunnel", result.TunnelName)
	assert.Equal(t, "local", result.ConfigSrc)
}

func TestLifecycleController_GetLifecycleConfig_NoSources(t *testing.T) {
	syncState := &v1alpha2.CloudflareSyncState{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: v1alpha2.CloudflareSyncStateSpec{
			ResourceType: v1alpha2.SyncResourceTunnelLifecycle,
			Sources:      []v1alpha2.ConfigSource{},
		},
	}

	c := &LifecycleController{}
	result, err := c.getLifecycleConfig(syncState)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no sources")
}

func TestLifecycleController_GetLifecycleConfig_InvalidJSON(t *testing.T) {
	syncState := &v1alpha2.CloudflareSyncState{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
		Spec: v1alpha2.CloudflareSyncStateSpec{
			ResourceType: v1alpha2.SyncResourceTunnelLifecycle,
			Sources: []v1alpha2.ConfigSource{
				{
					Ref: v1alpha2.SourceReference{
						Kind: "Tunnel",
						Name: "invalid",
					},
					Config: runtime.RawExtension{Raw: []byte("not valid json")},
				},
			},
		},
	}

	c := &LifecycleController{}
	result, err := c.getLifecycleConfig(syncState)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLifecycleController_Reconcile_NotFound(t *testing.T) {
	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	c := NewLifecycleController(client)
	ctx := context.Background()

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: "non-existent",
		},
	}

	result, err := c.Reconcile(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestLifecycleController_Reconcile_WrongResourceType(t *testing.T) {
	syncState := &v1alpha2.CloudflareSyncState{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-sync",
		},
		Spec: v1alpha2.CloudflareSyncStateSpec{
			ResourceType: v1alpha2.SyncResourceDNSRecord, // Wrong type
			CloudflareID: "test-id",
			AccountID:    "test-account",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(syncState).
		Build()

	c := NewLifecycleController(client)
	ctx := context.Background()

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-sync",
		},
	}

	result, err := c.Reconcile(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestLifecycleController_Reconcile_AlreadySynced(t *testing.T) {
	config := createLifecycleConfig(tunnelsvc.LifecycleActionCreate, "my-tunnel", "tunnel-123")
	syncState := createLifecycleSyncState("test-sync", config, v1alpha2.SyncStatusSynced, true)

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(syncState).
		Build()

	c := NewLifecycleController(client)
	ctx := context.Background()

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-sync",
		},
	}

	result, err := c.Reconcile(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestLifecycleController_Reconcile_NoSources_NoFinalizer(t *testing.T) {
	syncState := &v1alpha2.CloudflareSyncState{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-sync",
		},
		Spec: v1alpha2.CloudflareSyncStateSpec{
			ResourceType: v1alpha2.SyncResourceTunnelLifecycle,
			CloudflareID: "pending-abc123",
			AccountID:    "test-account",
			CredentialsRef: v1alpha2.CredentialsReference{
				Name: "test-credentials",
			},
			Sources: []v1alpha2.ConfigSource{},
		},
		Status: v1alpha2.CloudflareSyncStateStatus{
			SyncStatus: v1alpha2.SyncStatusPending,
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(syncState).
		WithStatusSubresource(syncState).
		Build()

	c := NewLifecycleController(client)
	ctx := context.Background()

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-sync",
		},
	}

	result, err := c.Reconcile(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Without finalizer, handleDeletion returns early
	var updatedState v1alpha2.CloudflareSyncState
	err = client.Get(ctx, types.NamespacedName{Name: "test-sync"}, &updatedState)
	require.NoError(t, err)
	assert.Equal(t, v1alpha2.SyncStatusPending, updatedState.Status.SyncStatus)
}

func TestLifecycleController_Reconcile_AddsFinalizer(t *testing.T) {
	config := createLifecycleConfig(tunnelsvc.LifecycleActionCreate, "my-tunnel", "")
	syncState := createLifecycleSyncState("test-sync", config, v1alpha2.SyncStatusPending, false)

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(syncState).
		Build()

	c := NewLifecycleController(client)
	ctx := context.Background()

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-sync",
		},
	}

	result, err := c.Reconcile(ctx, req)

	require.NoError(t, err)
	assert.True(t, result.Requeue)

	// Verify finalizer was added
	var updatedState v1alpha2.CloudflareSyncState
	err = client.Get(ctx, types.NamespacedName{Name: "test-sync"}, &updatedState)
	require.NoError(t, err)
	assert.Contains(t, updatedState.Finalizers, TunnelLifecycleFinalizerName)
}

func TestLifecycleController_Reconcile_DebouncePending(t *testing.T) {
	config := createLifecycleConfig(tunnelsvc.LifecycleActionCreate, "my-tunnel", "")
	syncState := createLifecycleSyncState("test-sync", config, v1alpha2.SyncStatusPending, true)

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(syncState).
		Build()

	c := NewLifecycleController(client)
	ctx := context.Background()

	// Add a pending debounce
	c.Debouncer.Debounce("test-sync", func() {})

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: "test-sync",
		},
	}

	result, err := c.Reconcile(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Cancel the debounce
	c.Debouncer.Cancel("test-sync")
}

func TestLifecycleController_HandleDeletion_NoFinalizer(t *testing.T) {
	config := createLifecycleConfig(tunnelsvc.LifecycleActionCreate, "my-tunnel", "tunnel-123")
	syncState := createLifecycleSyncState("test-sync", config, v1alpha2.SyncStatusSynced, false)

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(syncState).
		Build()

	c := NewLifecycleController(client)
	ctx := context.Background()

	result, err := c.handleDeletion(ctx, syncState)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestLifecycleController_HandleDeletion_PendingID(t *testing.T) {
	config := createLifecycleConfig(tunnelsvc.LifecycleActionCreate, "my-tunnel", "pending-123")
	syncState := createLifecycleSyncState("test-sync", config, v1alpha2.SyncStatusPending, true)
	syncState.Spec.CloudflareID = "pending-123"
	syncState.Spec.Sources = []v1alpha2.ConfigSource{} // Empty sources to trigger cleanup

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(syncState).
		Build()

	c := NewLifecycleController(client)
	ctx := context.Background()

	result, err := c.handleDeletion(ctx, syncState)

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Verify SyncState was deleted (orphan cleanup)
	var updatedState v1alpha2.CloudflareSyncState
	err = client.Get(ctx, types.NamespacedName{Name: "test-sync"}, &updatedState)
	assert.True(t, err != nil, "SyncState should be deleted")
}

func TestLifecycleController_HandleError_PermanentErrorDoesNotRequeueOrReturnError(t *testing.T) {
	config := createLifecycleConfig(tunnelsvc.LifecycleActionCreate, "my-tunnel", "")
	syncState := createLifecycleSyncState("test-sync", config, v1alpha2.SyncStatusPending, true)

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(syncState).
		WithStatusSubresource(syncState).
		Build()

	c := NewLifecycleController(client)
	ctx := context.Background()

	result, err := c.handleError(ctx, syncState, fmt.Errorf("create API client: %w", cf.ErrResourceNotFound))

	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	var updatedState v1alpha2.CloudflareSyncState
	err = client.Get(ctx, types.NamespacedName{Name: "test-sync"}, &updatedState)
	require.NoError(t, err)
	assert.Equal(t, v1alpha2.SyncStatusError, updatedState.Status.SyncStatus)
	assert.Contains(t, updatedState.Status.Error, "resource not found")
}

func TestLifecycleController_SetupWithManager(t *testing.T) {
	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
	c := NewLifecycleController(client)

	// Verify the controller is properly configured
	assert.NotNil(t, c.Client)
	assert.NotNil(t, c.Debouncer)
}

func TestLifecycleConfig_AllActions(t *testing.T) {
	tests := []struct {
		name       string
		action     tunnelsvc.LifecycleAction
		tunnelName string
		tunnelID   string
	}{
		{
			name:       "create action",
			action:     tunnelsvc.LifecycleActionCreate,
			tunnelName: "new-tunnel",
			tunnelID:   "",
		},
		{
			name:       "delete action",
			action:     tunnelsvc.LifecycleActionDelete,
			tunnelName: "existing-tunnel",
			tunnelID:   "tunnel-abc-123",
		},
		{
			name:       "adopt action",
			action:     tunnelsvc.LifecycleActionAdopt,
			tunnelName: "adopted-tunnel",
			tunnelID:   "tunnel-xyz-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &tunnelsvc.LifecycleConfig{
				Action:     tt.action,
				TunnelName: tt.tunnelName,
				TunnelID:   tt.tunnelID,
			}

			configBytes, err := json.Marshal(config)
			require.NoError(t, err)

			var parsed tunnelsvc.LifecycleConfig
			err = json.Unmarshal(configBytes, &parsed)
			require.NoError(t, err)

			assert.Equal(t, tt.action, parsed.Action)
			assert.Equal(t, tt.tunnelName, parsed.TunnelName)
			assert.Equal(t, tt.tunnelID, parsed.TunnelID)
		})
	}
}

func TestLifecycleResult_Serialization(t *testing.T) {
	result := &tunnelsvc.LifecycleResult{
		TunnelID:    "tunnel-123",
		TunnelName:  "my-tunnel",
		TunnelToken: "secret-token",
		Credentials: "base64-creds",
		AccountTag:  "account-abc",
	}

	resultBytes, err := json.Marshal(result)
	require.NoError(t, err)

	var parsed tunnelsvc.LifecycleResult
	err = json.Unmarshal(resultBytes, &parsed)
	require.NoError(t, err)

	assert.Equal(t, result.TunnelID, parsed.TunnelID)
	assert.Equal(t, result.TunnelName, parsed.TunnelName)
	assert.Equal(t, result.TunnelToken, parsed.TunnelToken)
	assert.Equal(t, result.Credentials, parsed.Credentials)
	assert.Equal(t, result.AccountTag, parsed.AccountTag)
}

func TestIsTunnelLifecycleSyncState(t *testing.T) {
	tests := []struct {
		name         string
		resourceType v1alpha2.SyncResourceType
		expected     bool
	}{
		{
			name:         "TunnelLifecycle type",
			resourceType: v1alpha2.SyncResourceTunnelLifecycle,
			expected:     true,
		},
		{
			name:         "DNSRecord type",
			resourceType: v1alpha2.SyncResourceDNSRecord,
			expected:     false,
		},
		{
			name:         "TunnelConfiguration type",
			resourceType: v1alpha2.SyncResourceTunnelConfiguration,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncState := &v1alpha2.CloudflareSyncState{
				Spec: v1alpha2.CloudflareSyncStateSpec{
					ResourceType: tt.resourceType,
				},
			}

			result := isTunnelLifecycleSyncState(syncState)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldReconcileLifecycleUpdate_StatusOnlyChange(t *testing.T) {
	oldObj := &v1alpha2.CloudflareSyncState{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "sync-1",
			Generation: 1,
		},
		Spec: v1alpha2.CloudflareSyncStateSpec{
			ResourceType: v1alpha2.SyncResourceTunnelLifecycle,
		},
		Status: v1alpha2.CloudflareSyncStateStatus{
			SyncStatus: v1alpha2.SyncStatusPending,
		},
	}

	newObj := oldObj.DeepCopy()
	newObj.Status.SyncStatus = v1alpha2.SyncStatusError

	assert.False(t, shouldReconcileLifecycleUpdate(oldObj, newObj))
}

func TestShouldReconcileLifecycleUpdate_GenerationChanged(t *testing.T) {
	oldObj := &v1alpha2.CloudflareSyncState{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "sync-1",
			Generation: 1,
		},
		Spec: v1alpha2.CloudflareSyncStateSpec{
			ResourceType: v1alpha2.SyncResourceTunnelLifecycle,
		},
	}

	newObj := oldObj.DeepCopy()
	newObj.Generation = 2

	assert.True(t, shouldReconcileLifecycleUpdate(oldObj, newObj))
}

func TestLifecycleController_CreateAPIClient_FallbackToLegacyTunnelSecret(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cloudflare-credentials",
			Namespace: "default",
		},
		StringData: map[string]string{
			"CLOUDFLARE_API_TOKEN": "test-token",
		},
	}
	secretData := map[string][]byte{}
	for k, v := range secret.StringData {
		secretData[k] = []byte(v)
	}
	secret.Data = secretData

	tunnel := &v1alpha2.Tunnel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "k8s-tunnel",
			Namespace: "default",
		},
		Spec: v1alpha2.TunnelSpec{
			Cloudflare: v1alpha2.CloudflareDetails{
				Secret:    "cloudflare-credentials",
				AccountId: "acc-123",
				Domain:    "example.com",
			},
		},
	}

	syncState := &v1alpha2.CloudflareSyncState{
		ObjectMeta: metav1.ObjectMeta{Name: "tunnel-lifecycle-k8s-tunnel"},
		Spec: v1alpha2.CloudflareSyncStateSpec{
			ResourceType:   v1alpha2.SyncResourceTunnelLifecycle,
			CloudflareID:   "k8s-tunnel",
			CredentialsRef: v1alpha2.CredentialsReference{},
			Sources: []v1alpha2.ConfigSource{
				{
					Ref: v1alpha2.SourceReference{
						Kind:      "Tunnel",
						Namespace: "default",
						Name:      "k8s-tunnel",
					},
				},
			},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(secret, tunnel, syncState).
		Build()

	c := NewLifecycleController(client)
	api, err := c.createAPIClient(context.Background(), syncState)

	require.NoError(t, err)
	require.NotNil(t, api)
	assert.Equal(t, "test-token", api.APIToken)
	assert.Equal(t, "acc-123", api.AccountId)
	assert.Equal(t, "example.com", api.Domain)
}
