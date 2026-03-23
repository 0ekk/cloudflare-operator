package cf

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	testRouteTunnelID = "tun-1"
)

func TestCreateVirtualNetwork_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/teamnet/virtual_networks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":{"id":"vnet-1","name":"vnet-a","comment":"demo","is_default_network":false,"created_at":"2026-01-01T00:00:00Z"}}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	result, err := api.CreateVirtualNetwork(context.Background(), VirtualNetworkParams{
		Name:             "vnet-a",
		Comment:          "demo",
		IsDefaultNetwork: false,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != "vnet-1" || result.Name != "vnet-a" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestCreateTunnelRoute_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/teamnet/routes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"errors":[],"messages":[],"result":{"id":"route-1","network":"10.0.0.0/8","tunnel_id":"%s","virtual_network_id":"vnet-1","comment":"demo"}}`, testRouteTunnelID)))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	result, err := api.CreateTunnelRoute(context.Background(), TunnelRouteParams{
		Network:          "10.0.0.0/8",
		TunnelID:         testRouteTunnelID,
		VirtualNetworkID: "vnet-1",
		Comment:          "demo",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Network != "10.0.0.0/8" || result.TunnelID != testRouteTunnelID {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDeleteTunnelRoute_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/teamnet/routes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"errors":[],"messages":[],"result":[{"id":"route-1","network":"10.0.0.0/8","tunnel_id":"%s","tunnel_name":"my-tunnel","virtual_network_id":"vnet-1"}],"result_info":{"count":1,"page":1,"per_page":20,"total_count":1}}`, testRouteTunnelID)))
	})
	mux.HandleFunc("/accounts/acc-1/teamnet/routes/route-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"errors":[],"messages":[],"result":{"id":"route-1","network":"10.0.0.0/8","tunnel_id":"%s","virtual_network_id":"vnet-1"}}`, testRouteTunnelID)))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	if err := api.DeleteTunnelRoute(context.Background(), "10.0.0.0/8", "vnet-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListTunnelRoutesByTunnelID_UsesV6Client(t *testing.T) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/accounts/acc-1/teamnet/routes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if got := r.URL.Query().Get("tunnel_id"); got != testRouteTunnelID {
			t.Fatalf("expected tunnel_id=%s, got %q", testRouteTunnelID, got)
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page") == "2" {
			_, _ = w.Write([]byte(`{"success":true,"errors":[],"messages":[],"result":[],"result_info":{"count":0,"page":2,"per_page":20,"total_count":1}}`))
			return
		}
		_, _ = w.Write([]byte(fmt.Sprintf(`{"success":true,"errors":[],"messages":[],"result":[{"id":"route-1","network":"10.0.0.0/8","tunnel_id":"%s","tunnel_name":"my-tunnel","virtual_network_id":"vnet-1","comment":"demo"}],"result_info":{"count":1,"page":1,"per_page":20,"total_count":1}}`, testRouteTunnelID)))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfV6 := newTestCloudflareV6Client(t, srv.URL)
	api := &API{
		CloudflareV6:   cfV6,
		ValidAccountId: "acc-1",
	}

	routes, err := api.ListTunnelRoutesByTunnelID(context.Background(), testRouteTunnelID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(routes) != 1 || routes[0].TunnelName != "my-tunnel" {
		t.Fatalf("unexpected routes: %+v", routes)
	}
}
