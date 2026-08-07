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

// JSONLWriter appends sessions as JSON Lines to disk.
type JSONLWriter struct {
	mu   sync.Mutex
	file *os.File
}

func NewJSONLWriter(path string) (*JSONLWriter, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &JSONLWriter{file: f}, nil
}

func (w *JSONLWriter) WriteSession(s *SessionRecord) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	b = append(b, byte(10))
	_, err = w.file.Write(b)
	return err
}

func (w *JSONLWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}
