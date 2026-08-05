package storage

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"
)

// SessionRecord aggregates conversational turns and metadata for a session.
type SessionRecord struct {
	SessionID string         `json:"session_id"`
	Messages  []any          `json:"messages"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// SessionStore manages persistent or in-memory agent sessions.
type SessionStore interface {
	Save(ctx context.Context, session *SessionRecord) error
	Load(ctx context.Context, sessionID string) (*SessionRecord, error)
}

// MemorySessionStore implements SessionStore in-memory.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*SessionRecord
}

func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]*SessionRecord),
	}
}

func (m *MemorySessionStore) Save(ctx context.Context, s *SessionRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s.UpdatedAt = time.Now()
	m.sessions[s.SessionID] = s
	return nil
}

func (m *MemorySessionStore) Load(ctx context.Context, sessionID string) (*SessionRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return s, nil
}
