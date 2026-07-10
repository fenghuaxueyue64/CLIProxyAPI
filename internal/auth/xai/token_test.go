package xai

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/xaiusage"
)

func TestCreateTokenStorageUsesFreeDefaults(t *testing.T) {
	auth := &XAIAuth{}
	storage := auth.CreateTokenStorage(&AuthBundle{TokenData: TokenData{AccessToken: "token"}})

	if storage.BaseURL != xaiusage.FreeBaseURL {
		t.Fatalf("BaseURL = %q, want %q", storage.BaseURL, xaiusage.FreeBaseURL)
	}
	if got := storage.Headers[xaiusage.TokenAuthHeader]; got != xaiusage.TokenAuthValue {
		t.Fatalf("token auth header = %q, want %q", got, xaiusage.TokenAuthValue)
	}
	if got := storage.Headers[xaiusage.ClientVersionHeader]; got != xaiusage.DefaultClientVersion {
		t.Fatalf("client version = %q, want %q", got, xaiusage.DefaultClientVersion)
	}
	if !storage.FreeMode {
		t.Fatal("FreeMode = false, want true")
	}
}

func TestCreateTokenStoragePreservesExplicitBaseURL(t *testing.T) {
	auth := &XAIAuth{}
	storage := auth.CreateTokenStorage(&AuthBundle{
		TokenData: TokenData{AccessToken: "token"},
		BaseURL:   "https://custom.example/v1",
	})

	if storage.BaseURL != "https://custom.example/v1" {
		t.Fatalf("BaseURL = %q, want explicit URL", storage.BaseURL)
	}
	if storage.FreeMode {
		t.Fatal("FreeMode = true, want false for explicit URL")
	}
}
