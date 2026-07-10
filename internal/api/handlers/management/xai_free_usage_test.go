package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/xaiusage"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

func TestGetXAIFreeUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authDir := t.TempDir()
	manager := coreauth.NewManager(nil, nil, nil)
	_, errRegister := manager.Register(context.Background(), &coreauth.Auth{
		ID:       "xai-user.json",
		Provider: "xai",
		Metadata: map[string]any{
			"auth_kind": "oauth",
			"email":     "user@example.com",
			"free_mode": true,
			"base_url":  xaiusage.FreeBaseURL,
		},
	})
	if errRegister != nil {
		t.Fatalf("Register() error = %v", errRegister)
	}
	xaiusage.NewStore(authDir).Record("xai-user.json", "user@example.com", "grok-4.5", http.Header{
		"X-Ratelimit-Limit-Requests":     {"21"},
		"X-Ratelimit-Remaining-Requests": {"20"},
		"X-Ratelimit-Limit-Tokens":       {"1000000"},
		"X-Ratelimit-Remaining-Tokens":   {"999775"},
	}, xaiusage.Usage{Total: 225})

	h := NewHandlerWithoutConfigFilePath(&config.Config{AuthDir: authDir}, manager)
	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/v0/management/xai-free-usage", nil)
	h.GetXAIFreeUsage(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Accounts []xaiFreeUsageAccount `json:"accounts"`
	}
	if errJSON := json.Unmarshal(recorder.Body.Bytes(), &response); errJSON != nil {
		t.Fatalf("Unmarshal() error = %v", errJSON)
	}
	if len(response.Accounts) != 1 || !response.Accounts[0].Known {
		t.Fatalf("accounts = %#v", response.Accounts)
	}
	if response.Accounts[0].Tokens.Used != 225 || response.Accounts[0].ObservedTokens.Total != 225 || response.Accounts[0].ObservedRequests != 1 {
		t.Fatalf("account usage = %#v", response.Accounts[0])
	}
}
