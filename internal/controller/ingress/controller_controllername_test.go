package ingress

import "testing"

func TestIsManagedIngressClassController(t *testing.T) {
	tests := []struct {
		name       string
		controller string
		want       bool
	}{
		{
			name:       "current controller name",
			controller: ControllerName,
			want:       true,
		},
		{
			name:       "legacy migration doc controller name",
			controller: "cloudflare-operator.io/tunnel-ingress-controller",
			want:       true,
		},
		{
			name:       "unrelated controller",
			controller: "k8s.io/ingress-nginx",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isManagedIngressClassController(tt.controller); got != tt.want {
				t.Fatalf("isManagedIngressClassController(%q) = %v, want %v", tt.controller, got, tt.want)
			}
		})
	}
}
