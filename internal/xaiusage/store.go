package xaiusage

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Limit struct {
	Limit     int64 `json:"limit"`
	Remaining int64 `json:"remaining"`
	Used      int64 `json:"used"`
}

type Usage struct {
	Input  int64 `json:"input"`
	Output int64 `json:"output"`
	Total  int64 `json:"total"`
}

type Snapshot struct {
	AuthID           string    `json:"auth_id"`
	Email            string    `json:"email,omitempty"`
	Requests         Limit     `json:"requests"`
	Tokens           Limit     `json:"tokens"`
	Observed         Usage     `json:"observed_tokens"`
	ObservedRequests int64     `json:"observed_requests"`
	LastModel        string    `json:"last_model,omitempty"`
	LastRequestID    string    `json:"last_request_id,omitempty"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Store struct {
	mu        sync.RWMutex
	path      string
	snapshots map[string]Snapshot
}

var sharedStores sync.Map

func NewStore(authDir string) *Store {
	store := &Store{
		path:      filepath.Join(strings.TrimSpace(authDir), ".state", "xai-free-usage.json"),
		snapshots: make(map[string]Snapshot),
	}
	store.load()
	return store
}

func SharedStore(authDir string) *Store {
	path := filepath.Join(strings.TrimSpace(authDir), ".state", "xai-free-usage.json")
	if existing, ok := sharedStores.Load(path); ok {
		return existing.(*Store)
	}
	created := NewStore(authDir)
	actual, _ := sharedStores.LoadOrStore(path, created)
	return actual.(*Store)
}

func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}

func (s *Store) Record(authID, email, model string, headers http.Header, observed Usage) {
	if s == nil || strings.TrimSpace(authID) == "" {
		return
	}

	s.mu.Lock()
	snapshot := s.snapshots[authID]
	snapshot.AuthID = authID
	if email = strings.TrimSpace(email); email != "" {
		snapshot.Email = email
	}
	if model = strings.TrimSpace(model); model != "" {
		snapshot.LastModel = model
	}
	if requestID := strings.TrimSpace(headers.Get("X-Request-ID")); requestID != "" {
		snapshot.LastRequestID = requestID
	}
	if limit, ok := parseLimit(headers, "X-Ratelimit-Limit-Requests", "X-Ratelimit-Remaining-Requests"); ok {
		snapshot.Requests = limit
	}
	if limit, ok := parseLimit(headers, "X-Ratelimit-Limit-Tokens", "X-Ratelimit-Remaining-Tokens"); ok {
		snapshot.Tokens = limit
	}
	snapshot.Observed.Input += maxZero(observed.Input)
	snapshot.Observed.Output += maxZero(observed.Output)
	snapshot.Observed.Total += maxZero(observed.Total)
	if observed.Input > 0 || observed.Output > 0 || observed.Total > 0 {
		snapshot.ObservedRequests++
	}
	snapshot.UpdatedAt = time.Now().UTC()
	s.snapshots[authID] = snapshot
	errSave := s.saveLocked()
	s.mu.Unlock()
	if errSave != nil {
		log.WithError(errSave).Warn("failed to persist xAI Free usage state")
	}
}

func (s *Store) Get(authID string) (Snapshot, bool) {
	if s == nil {
		return Snapshot{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	snapshot, ok := s.snapshots[authID]
	return snapshot, ok
}

func (s *Store) List() []Snapshot {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Snapshot, 0, len(s.snapshots))
	for _, snapshot := range s.snapshots {
		out = append(out, snapshot)
	}
	return out
}

func parseLimit(headers http.Header, limitName, remainingName string) (Limit, bool) {
	limit, errLimit := strconv.ParseInt(strings.TrimSpace(headers.Get(limitName)), 10, 64)
	remaining, errRemaining := strconv.ParseInt(strings.TrimSpace(headers.Get(remainingName)), 10, 64)
	if errLimit != nil || errRemaining != nil || limit < 0 || remaining < 0 {
		return Limit{}, false
	}
	if remaining > limit {
		remaining = limit
	}
	return Limit{Limit: limit, Remaining: remaining, Used: limit - remaining}, true
}

func (s *Store) load() {
	data, errRead := os.ReadFile(s.path)
	if errRead != nil {
		return
	}
	var snapshots map[string]Snapshot
	if errJSON := json.Unmarshal(data, &snapshots); errJSON == nil && snapshots != nil {
		s.snapshots = snapshots
	}
}

func (s *Store) saveLocked() error {
	if errMkdir := os.MkdirAll(filepath.Dir(s.path), 0o700); errMkdir != nil {
		return errMkdir
	}
	data, errJSON := json.MarshalIndent(s.snapshots, "", "  ")
	if errJSON != nil {
		return errJSON
	}
	tempPath := s.path + ".tmp"
	if errWrite := os.WriteFile(tempPath, data, 0o600); errWrite != nil {
		return errWrite
	}
	return os.Rename(tempPath, s.path)
}

func maxZero(value int64) int64 {
	if value < 0 {
		return 0
	}
	return value
}
