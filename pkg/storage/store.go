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
