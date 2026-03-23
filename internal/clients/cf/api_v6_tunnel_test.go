package cf

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cloudflarev6 "github.com/cloudflare/cloudflare-go/v6"
)

const (
	testTunnelName      = "my-tunnel"
	testTunnelID        = "tun-1"
	testTunnelConfigSrc = "cloudflare"
)

func TestGetTunnelToken_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel/tun-1/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":"token-abc"}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	token, err := api.GetTunnelToken(context.Background(), "tun-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token != "token-abc" {
		t.Fatalf("expected token-abc, got %s", token)
	}
}

func TestCreateTunnelWithParams_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}

		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		if body["name"] != testTunnelName {
			t.Fatalf("expected tunnel name %s, got %q", testTunnelName, body["name"])
		}
		if body["config_src"] != testTunnelConfigSrc {
			t.Fatalf("expected default config_src %s, got %q", testTunnelConfigSrc, body["config_src"])
		}
		if body["tunnel_secret"] == "" {
			t.Fatal("expected tunnel_secret to be set")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"id":"tun-1","name":"my-tunnel","account_tag":"acc-1","config_src":"cloudflare"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	result, err := api.CreateTunnelWithParams(context.Background(), testTunnelName, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != testTunnelID {
		t.Fatalf("expected tunnel id %s, got %s", testTunnelID, result.ID)
	}
	if result.Credentials == nil {
		t.Fatal("expected credentials")
	}
	if result.Credentials.TunnelSecret == "" {
		t.Fatal("expected tunnel secret in credentials")
	}
	if _, err := base64.StdEncoding.DecodeString(result.Credentials.TunnelSecret); err != nil {
		t.Fatalf("expected base64 tunnel secret, got error: %v", err)
	}
}

func TestGetTunnelIDByName_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.URL.Query().Get("name"); got != testTunnelName {
			t.Fatalf("expected query name=%s, got %q", testTunnelName, got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":[{"id":"tun-deleted","name":"my-tunnel","deleted_at":"2025-01-01T00:00:00Z"},{"id":"tun-active","name":"my-tunnel"}],"result_info":{"count":2,"page":1,"per_page":20,"total_count":2}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	tunnelID, err := api.GetTunnelIDByName(context.Background(), testTunnelName)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tunnelID != "tun-active" {
		t.Fatalf("expected tun-active, got %s", tunnelID)
	}
}

func TestCreateTunnel_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["name"] != testTunnelName {
			t.Fatalf("expected name=%s, got %q", testTunnelName, body["name"])
		}
		if body["config_src"] != testTunnelConfigSrc {
			t.Fatalf("expected config_src=%s, got %q", testTunnelConfigSrc, body["config_src"])
		}
		if body["tunnel_secret"] == "" {
			t.Fatal("expected tunnel_secret to be set")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"id":"tun-1","name":"my-tunnel","account_tag":"acc-1","config_src":"cloudflare"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
		TunnelName:     testTunnelName,
	}

	tunnelID, creds, err := api.CreateTunnel(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tunnelID != testTunnelID {
		t.Fatalf("expected tunnel id %s, got %s", testTunnelID, tunnelID)
	}
	var parsed TunnelCredentialsFile
	if err := json.Unmarshal([]byte(creds), &parsed); err != nil {
		t.Fatalf("unmarshal creds: %v", err)
	}
	if parsed.TunnelID != testTunnelID || parsed.TunnelName != testTunnelName {
		t.Fatalf("unexpected creds payload: %+v", parsed)
	}
}

func TestDeleteTunnelByID_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel/tun-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"id":"tun-1","name":"my-tunnel"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	if err := api.DeleteTunnelByID(context.Background(), "tun-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestGetTunnelId_ValidateByID_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel/tun-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"id":"tun-1","name":"my-tunnel"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
		TunnelId:       "tun-1",
	}

	id, err := api.GetTunnelId(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != testTunnelID {
		t.Fatalf("expected %s, got %s", testTunnelID, id)
	}
	if api.ValidTunnelName != "my-tunnel" {
		t.Fatalf("expected ValidTunnelName my-tunnel, got %s", api.ValidTunnelName)
	}
}

func TestGetTunnelCredsByID_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/cfd_tunnel/tun-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"id":"tun-1","name":"my-tunnel"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	creds, err := api.GetTunnelCredsByID(context.Background(), "tun-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if creds.TunnelID != testTunnelID || creds.TunnelName != testTunnelName {
		t.Fatalf("unexpected creds: %+v", creds)
	}
}

func newTestCloudflareV6Client(t *testing.T, baseURL string) *cloudflarev6.Client {
	t.Helper()

	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	t.Setenv(CloudflareAPIBaseURLEnv, baseURL)

	client, err := createCloudflareV6ClientFromConfig("test-token", "", "")
	if err != nil {
		t.Fatalf("create v6 client: %v", err)
	}
	return client
}
