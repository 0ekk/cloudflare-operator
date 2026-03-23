package cf

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudflare/cloudflare-go"
)

func TestGetTunnelConfiguration_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel/tun-1/configurations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"tunnel_id":"tun-1","version":7,"config":{"ingress":[{"hostname":"app.example.com","service":"http://svc:8080"},{"service":"http_status:404"}]}}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	cfg, err := api.GetTunnelConfiguration(context.Background(), "tun-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.TunnelID != "tun-1" || cfg.Version != 7 {
		t.Fatalf("unexpected config header: %+v", cfg)
	}
	if len(cfg.Config.Ingress) != 2 || cfg.Config.Ingress[0].Hostname != "app.example.com" {
		t.Fatalf("unexpected ingress: %+v", cfg.Config.Ingress)
	}
}

func TestUpdateTunnelConfiguration_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel/tun-1/configurations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		configObj, ok := body["config"].(map[string]any)
		if !ok {
			t.Fatalf("missing config in request: %#v", body)
		}
		ingress, ok := configObj["ingress"].([]any)
		if !ok || len(ingress) != 2 {
			t.Fatalf("unexpected ingress in request: %#v", configObj["ingress"])
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"tunnel_id":"tun-1","version":8,"config":{"ingress":[{"hostname":"app.example.com","service":"http://svc:8080"},{"service":"http_status:404"}]}}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	input := cloudflare.TunnelConfiguration{
		Ingress: []cloudflare.UnvalidatedIngressRule{
			{Hostname: "app.example.com", Service: "http://svc:8080"},
			{Service: "http_status:404"},
		},
	}

	result, err := api.UpdateTunnelConfiguration(context.Background(), "tun-1", input)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Version != 8 || result.TunnelID != "tun-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
