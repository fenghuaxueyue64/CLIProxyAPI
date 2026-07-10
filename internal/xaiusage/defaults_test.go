package xaiusage

import (
	"strings"
	"testing"
)

func TestApplyOAuthDefaults(t *testing.T) {
	metadata := map[string]any{"type": "xai", "auth_kind": "oauth"}

	ApplyOAuthDefaults(metadata)

	if got := metadata["base_url"]; got != FreeBaseURL {
		t.Fatalf("base_url = %#v, want %q", got, FreeBaseURL)
	}
	headers, ok := metadata["headers"].(map[string]any)
	if !ok {
		t.Fatalf("headers = %#v, want map", metadata["headers"])
	}
	if got := headers[TokenAuthHeader]; got != TokenAuthValue {
		t.Fatalf("%s = %#v, want %q", TokenAuthHeader, got, TokenAuthValue)
	}
	if got := headers[ClientVersionHeader]; got != DefaultClientVersion {
		t.Fatalf("%s = %#v, want %q", ClientVersionHeader, got, DefaultClientVersion)
	}
	if got := metadata["free_mode"]; got != true {
		t.Fatalf("free_mode = %#v, want true", got)
	}
}

func TestApplyOAuthDefaultsPreservesExplicitConfiguration(t *testing.T) {
	metadata := map[string]any{
		"type":      "xai",
		"auth_kind": "oauth",
		"base_url":  "https://api.x.ai/v1",
		"headers": map[string]any{
			TokenAuthHeader: "custom",
		},
	}

	ApplyOAuthDefaults(metadata)

	if got := metadata["base_url"]; got != "https://api.x.ai/v1" {
		t.Fatalf("base_url = %#v, want explicit paid URL", got)
	}
	headers := metadata["headers"].(map[string]any)
	if got := headers[TokenAuthHeader]; got != "custom" {
		t.Fatalf("explicit token auth header = %#v, want custom", got)
	}
	if _, ok := metadata["free_mode"]; ok {
		t.Fatalf("free_mode unexpectedly added: %#v", metadata["free_mode"])
	}
}

func TestApplyOAuthDefaultsHonorsOptOut(t *testing.T) {
	metadata := map[string]any{"type": "xai", "auth_kind": "oauth", "free_mode": false}

	ApplyOAuthDefaults(metadata)

	if _, ok := metadata["base_url"]; ok {
		t.Fatalf("base_url unexpectedly added: %#v", metadata["base_url"])
	}
}

func TestApplyOAuthDefaultsPreservesCaseInsensitiveHeaderOverride(t *testing.T) {
	metadata := map[string]any{
		"type":      "xai",
		"auth_kind": "oauth",
		"headers": map[string]any{
			"x-xai-token-auth": "custom",
		},
	}

	ApplyOAuthDefaults(metadata)

	headers := metadata["headers"].(map[string]any)
	var matches int
	for name, value := range headers {
		if strings.EqualFold(name, TokenAuthHeader) {
			matches++
			if value != "custom" {
				t.Fatalf("%s = %#v, want custom", name, value)
			}
		}
	}
	if matches != 1 {
		t.Fatalf("case-insensitive token auth header count = %d, want 1; headers=%#v", matches, headers)
	}
}
