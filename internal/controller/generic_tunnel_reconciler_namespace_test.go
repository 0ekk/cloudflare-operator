package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/0ekk/cloudflare-operator/internal/controller/common"

	networkingv1alpha2 "github.com/0ekk/cloudflare-operator/api/v1alpha2"
)

func TestGetTunnelConfigNamespace_UsesOperatorNamespace(t *testing.T) {
	common.SetOperatorNamespace("operator-ns")
	t.Cleanup(func() {
		common.SetOperatorNamespace("cloudflare-operator-system")
	})

	tunnel := TunnelAdapter{&networkingv1alpha2.Tunnel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
	}}

	got := getTunnelConfigNamespace(tunnel)
	if got != "operator-ns" {
		t.Fatalf("getTunnelConfigNamespace() = %q, want %q", got, "operator-ns")
	}
}
