package cf

import (
	cloudflarev6 "github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

// createCloudflareV6ClientFromConfig creates a v6 SDK client using the same
// credential precedence as the legacy client factory.
func createCloudflareV6ClientFromConfig(apiToken, apiKey, email string) (*cloudflarev6.Client, error) {
	var opts []option.RequestOption
	if baseURL := GetAPIBaseURL(); baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	opts = append(opts, option.WithHTTPClient(NewAPIHTTPClient()))

	switch {
	case apiToken != "":
		opts = append(opts, option.WithAPIToken(apiToken))
	case apiKey != "" && email != "":
		opts = append(opts, option.WithAPIKey(apiKey), option.WithAPIEmail(email))
	default:
		return nil, ErrNoCredentials
	}

	return cloudflarev6.NewClient(opts...), nil
}
