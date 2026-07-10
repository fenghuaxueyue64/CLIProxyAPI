package xaiusage

import (
	"net/http"
	"path/filepath"
	"testing"
)

func TestStoreRecordAndReload(t *testing.T) {
	authDir := t.TempDir()
	store := NewStore(authDir)
	headers := http.Header{
		"X-Ratelimit-Limit-Requests":     {"21"},
		"X-Ratelimit-Remaining-Requests": {"20"},
		"X-Ratelimit-Limit-Tokens":       {"1000000"},
		"X-Ratelimit-Remaining-Tokens":   {"999775"},
		"X-Request-Id":                   {"request-1"},
	}

	store.Record("auth-1", "user@example.com", "grok-4.5", headers, Usage{Input: 201, Output: 24, Total: 225})

	snapshot, ok := store.Get("auth-1")
	if !ok {
		t.Fatal("Get() ok = false, want true")
	}
	if snapshot.Requests.Used != 1 || snapshot.Tokens.Used != 225 {
		t.Fatalf("derived usage = requests %d tokens %d", snapshot.Requests.Used, snapshot.Tokens.Used)
	}
	if snapshot.Observed.Total != 225 || snapshot.LastRequestID != "request-1" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	if snapshot.ObservedRequests != 1 {
		t.Fatalf("observed requests = %d, want 1", snapshot.ObservedRequests)
	}

	reloaded := NewStore(authDir)
	reloadedSnapshot, ok := reloaded.Get("auth-1")
	if !ok || reloadedSnapshot.Tokens.Remaining != 999775 {
		t.Fatalf("reloaded snapshot = %#v, ok=%v", reloadedSnapshot, ok)
	}
	if got := reloaded.Path(); got != filepath.Join(authDir, ".state", "xai-free-usage.json") {
		t.Fatalf("Path() = %q", got)
	}
}

func TestStoreCountsOnlyCompletedObservedRequests(t *testing.T) {
	store := NewStore(t.TempDir())
	headers := http.Header{
		"X-Ratelimit-Limit-Requests":     {"21"},
		"X-Ratelimit-Remaining-Requests": {"21"},
	}
	store.Record("auth-1", "", "grok-4.5", headers, Usage{})
	store.Record("auth-1", "", "grok-4.5", headers, Usage{Input: 10, Output: 5, Total: 15})

	snapshot, _ := store.Get("auth-1")
	if snapshot.ObservedRequests != 1 {
		t.Fatalf("observed requests = %d, want 1", snapshot.ObservedRequests)
	}
}

func TestStorePreservesPreviousQuotaOnMalformedHeaders(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Record("auth-1", "", "grok-4.5", http.Header{
		"X-Ratelimit-Limit-Requests":     {"21"},
		"X-Ratelimit-Remaining-Requests": {"20"},
	}, Usage{})
	store.Record("auth-1", "", "grok-4.5", http.Header{
		"X-Ratelimit-Limit-Requests":     {"bad"},
		"X-Ratelimit-Remaining-Requests": {"19"},
	}, Usage{Total: 3})

	snapshot, _ := store.Get("auth-1")
	if snapshot.Requests.Limit != 21 || snapshot.Requests.Remaining != 20 {
		t.Fatalf("requests overwritten by malformed headers: %#v", snapshot.Requests)
	}
	if snapshot.Observed.Total != 3 {
		t.Fatalf("observed total = %d, want 3", snapshot.Observed.Total)
	}
}

func TestStorePreservesObservedHistoryWhenQuotaRemainingIncreases(t *testing.T) {
	store := NewStore(t.TempDir())
	store.Record("auth-1", "", "grok-4.5", http.Header{
		"X-Ratelimit-Limit-Requests":     {"21"},
		"X-Ratelimit-Remaining-Requests": {"10"},
		"X-Ratelimit-Limit-Tokens":       {"1000"},
		"X-Ratelimit-Remaining-Tokens":   {"400"},
	}, Usage{Input: 400, Output: 200, Total: 600})

	store.Record("auth-1", "", "grok-4.5", http.Header{
		"X-Ratelimit-Limit-Requests":     {"21"},
		"X-Ratelimit-Remaining-Requests": {"21"},
		"X-Ratelimit-Limit-Tokens":       {"1000"},
		"X-Ratelimit-Remaining-Tokens":   {"1000"},
	}, Usage{})

	snapshot, _ := store.Get("auth-1")
	if snapshot.ObservedRequests != 1 {
		t.Fatalf("observed requests = %d, want lifetime count 1", snapshot.ObservedRequests)
	}
	if snapshot.Observed != (Usage{Input: 400, Output: 200, Total: 600}) {
		t.Fatalf("observed tokens = %#v, want lifetime history preserved", snapshot.Observed)
	}
}
