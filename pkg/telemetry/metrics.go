package telemetry

import (
	"sync/atomic"
	"time"
)

// Metrics records atomic telemetry counters.
type Metrics struct {
	TotalRequests atomic.Int64
	TotalErrors   atomic.Int64
	CircuitTrips  atomic.Int64
	PromptTokens  atomic.Int64
	ComplTokens   atomic.Int64
	TotalDuration atomic.Int64 // in microseconds
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) RecordRequest(dur time.Duration, promptTok, complTok int, isErr bool) {
	m.TotalRequests.Add(1)
	m.TotalDuration.Add(dur.Microseconds())
	m.PromptTokens.Add(int64(promptTok))
	m.ComplTokens.Add(int64(complTok))
	if isErr {
		m.TotalErrors.Add(1)
	}
}

func (m *Metrics) RecordCircuitTrip() {
	m.CircuitTrips.Add(1)
}

type MetricsSnapshot struct {
	Requests    int64   `json:"requests"`
	Errors      int64   `json:"errors"`
	ErrorRate   float64 `json:"error_rate"`
	AvgLatencyMs float64`json:"avg_latency_ms"`
	TotalTokens int64   `json:"total_tokens"`
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	reqs := m.TotalRequests.Load()
	errs := m.TotalErrors.Load()
	durMicros := m.TotalDuration.Load()

	rate := 0.0
	avgMs := 0.0
	if reqs > 0 {
		rate = float64(errs) / float64(reqs)
		avgMs = float64(durMicros) / float64(reqs*1000)
	}

	return MetricsSnapshot{
		Requests:     reqs,
		Errors:       errs,
		ErrorRate:    rate,
		AvgLatencyMs: avgMs,
		TotalTokens:  m.PromptTokens.Load() + m.ComplTokens.Load(),
	}
}
