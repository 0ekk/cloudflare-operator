package cf

import "testing"

func TestCreateCloudflareV6ClientFromConfig_WithAPIToken(t *testing.T) {
	t.Parallel()

	client, err := createCloudflareV6ClientFromConfig("token-123", "", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCreateCloudflareV6ClientFromConfig_WithAPIKey(t *testing.T) {
	t.Parallel()

	client, err := createCloudflareV6ClientFromConfig("", "key-123", "user@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestCreateCloudflareV6ClientFromConfig_NoCredentials(t *testing.T) {
	t.Parallel()

	client, err := createCloudflareV6ClientFromConfig("", "", "")
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
	if err != ErrNoCredentials {
		t.Fatalf("expected ErrNoCredentials, got %v", err)
	}
	if client != nil {
		t.Fatal("expected nil client on error")
	}
}
