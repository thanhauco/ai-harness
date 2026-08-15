package telemetry

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type traceContextKey struct{}

type Span struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Name       string            `json:"name"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Attributes map[string]string `json:"attributes"`
	mu         sync.Mutex
}

func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func StartSpan(ctx context.Context, name string) (context.Context, *Span) {
	parent, _ := ctx.Value(traceContextKey{}).(*Span)

	traceID := ""
	parentID := ""
	if parent != nil {
		traceID = parent.TraceID
		parentID = parent.SpanID
	} else {
		traceID = randomHex(16)
	}

	span := &Span{
		TraceID:    traceID,
		SpanID:     randomHex(8),
		ParentID:   parentID,
		Name:       name,
		StartTime:  time.Now(),
		Attributes: make(map[string]string),
	}

	return context.WithValue(ctx, traceContextKey{}, span), span
}

func (s *Span) SetAttribute(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Attributes[key] = value
}

func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.EndTime = time.Now()
}

func (s *Span) Duration() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.EndTime.Sub(s.StartTime)
}
