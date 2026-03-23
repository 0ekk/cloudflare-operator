package cf

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateWARPConnector_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/warp_connector", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"id":"wc-1","name":"connector-a","account_tag":"acc-1"}}`))
	})
	mux.HandleFunc("/accounts/acc-1/warp_connector/wc-1/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":"token-xyz"}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	result, err := api.CreateWARPConnector(context.Background(), "connector-a")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != "wc-1" || result.TunnelToken != "token-xyz" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGetWARPConnectorToken_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/warp_connector/wc-1/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":"token-xyz"}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	result, err := api.GetWARPConnectorToken(context.Background(), "wc-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Token != "token-xyz" {
		t.Fatalf("expected token-xyz, got %s", result.Token)
	}
}

func TestDeleteWARPConnector_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/warp_connector/wc-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"id":"wc-1","name":"connector-a"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	if err := api.DeleteWARPConnector(context.Background(), "wc-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
