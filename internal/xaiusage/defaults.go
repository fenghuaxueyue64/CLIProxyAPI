package xaiusage

import "strings"

const (
	FreeBaseURL          = "https://cli-chat-proxy.grok.com/v1"
	TokenAuthHeader      = "X-XAI-Token-Auth"
	TokenAuthValue       = "xai-grok-cli"
	ClientVersionHeader  = "x-grok-client-version"
	DefaultClientVersion = "0.2.93"
	ModelOverrideHeader  = "x-grok-model-override"
)

func ApplyOAuthDefaults(metadata map[string]any) {
	if !isXAI(metadata) || !isOAuth(metadata) || explicitlyDisabled(metadata) {
		return
	}

	baseURL := metadataString(metadata, "base_url")
	if baseURL != "" && !strings.EqualFold(strings.TrimRight(baseURL, "/"), strings.TrimRight(FreeBaseURL, "/")) {
		return
	}
	if baseURL == "" {
		metadata["base_url"] = FreeBaseURL
	}
	metadata["free_mode"] = true
	setHeaderDefault(metadata, TokenAuthHeader, TokenAuthValue)
	setHeaderDefault(metadata, ClientVersionHeader, DefaultClientVersion)
}

func IsFreeMetadata(metadata map[string]any) bool {
	if metadata == nil {
		return false
	}
	if enabled, ok := metadata["free_mode"].(bool); ok {
		return enabled
	}
	return strings.EqualFold(strings.TrimRight(metadataString(metadata, "base_url"), "/"), strings.TrimRight(FreeBaseURL, "/"))
}

func isXAI(metadata map[string]any) bool {
	return strings.EqualFold(metadataString(metadata, "type"), "xai")
}

func isOAuth(metadata map[string]any) bool {
	kind := strings.ToLower(metadataString(metadata, "auth_kind"))
	if kind != "" {
		return kind == "oauth"
	}
	return metadataString(metadata, "access_token") != "" && metadataString(metadata, "api_key") == ""
}

func explicitlyDisabled(metadata map[string]any) bool {
	enabled, ok := metadata["free_mode"].(bool)
	return ok && !enabled
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, _ := metadata[key].(string)
	return strings.TrimSpace(value)
}

func setHeaderDefault(metadata map[string]any, name, value string) {
	switch headers := metadata["headers"].(type) {
	case map[string]any:
		for existingName, existingValue := range headers {
			if !strings.EqualFold(existingName, name) {
				continue
			}
			if existing, _ := existingValue.(string); strings.TrimSpace(existing) == "" {
				headers[existingName] = value
			}
			return
		}
		headers[name] = value
	case map[string]string:
		for existingName, existingValue := range headers {
			if !strings.EqualFold(existingName, name) {
				continue
			}
			if strings.TrimSpace(existingValue) == "" {
				headers[existingName] = value
			}
			return
		}
		headers[name] = value
	default:
		metadata["headers"] = map[string]any{name: value}
	}
}
