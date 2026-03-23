package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"

	networkingv1alpha2 "github.com/0ekk/cloudflare-operator/api/v1alpha2"
)

func TestDeploymentForTunnel_UsesStartupProbeAndConservativeLiveness(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = networkingv1alpha2.AddToScheme(scheme)

	tunnel := &networkingv1alpha2.Tunnel{
		Spec: networkingv1alpha2.TunnelSpec{
			Protocol: "auto",
		},
	}

	r := &TunnelReconciler{
		Scheme: scheme,
		tunnel: TunnelAdapter{tunnel},
	}

	dep := deploymentForTunnel(r)
	cloudflared := dep.Spec.Template.Spec.Containers[0]

	assert.NotNil(t, cloudflared.StartupProbe, "startup probe should protect slow startup")
	assert.NotNil(t, cloudflared.LivenessProbe, "liveness probe must be configured")
	assert.GreaterOrEqual(t, cloudflared.LivenessProbe.FailureThreshold, int32(3), "liveness should not restart after a single transient failure")
}
