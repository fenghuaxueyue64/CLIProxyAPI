package management

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/xaiusage"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

type xaiFreeUsageAccount struct {
	ID               string         `json:"id"`
	Email            string         `json:"email,omitempty"`
	FreeMode         bool           `json:"free_mode"`
	Known            bool           `json:"known"`
	Requests         xaiusage.Limit `json:"requests"`
	Tokens           xaiusage.Limit `json:"tokens"`
	ObservedTokens   xaiusage.Usage `json:"observed_tokens"`
	ObservedRequests int64          `json:"observed_requests"`
	LastModel        string         `json:"last_model,omitempty"`
	LastRequestID    string         `json:"last_request_id,omitempty"`
	UpdatedAt        string         `json:"updated_at,omitempty"`
}

func (h *Handler) GetXAIFreeUsage(c *gin.Context) {
	accounts := make([]xaiFreeUsageAccount, 0)
	if h == nil || h.authManager == nil {
		c.JSON(http.StatusOK, gin.H{"accounts": accounts})
		return
	}

	store := xaiusage.SharedStore(h.cfg.AuthDir)
	for _, auth := range h.authManager.List() {
		if auth == nil || !strings.EqualFold(strings.TrimSpace(auth.Provider), "xai") || auth.AuthKind() != coreauth.AuthKindOAuth {
			continue
		}
		account := xaiFreeUsageAccount{
			ID:       auth.ID,
			Email:    authMetadataString(auth.Metadata, "email"),
			FreeMode: xaiusage.IsFreeMetadata(auth.Metadata),
		}
		if snapshot, ok := store.Get(auth.ID); ok {
			account.Known = snapshot.Requests.Limit > 0 || snapshot.Tokens.Limit > 0
			account.Requests = snapshot.Requests
			account.Tokens = snapshot.Tokens
			account.ObservedTokens = snapshot.Observed
			account.ObservedRequests = snapshot.ObservedRequests
			account.LastModel = snapshot.LastModel
			account.LastRequestID = snapshot.LastRequestID
			if !snapshot.UpdatedAt.IsZero() {
				account.UpdatedAt = snapshot.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")
			}
		}
		accounts = append(accounts, account)
	}
	sort.Slice(accounts, func(i, j int) bool {
		return strings.ToLower(accounts[i].Email+accounts[i].ID) < strings.ToLower(accounts[j].Email+accounts[j].ID)
	})
	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

func authMetadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, _ := metadata[key].(string)
	return strings.TrimSpace(value)
}
