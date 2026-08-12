package telemetry

import (
	"context"
	"log/slog"
	"os"
	"regexp"
)

// RedactingHandler wraps an slog.Handler to scrub sensitive tokens.
type RedactingHandler struct {
	inner slog.Handler
	re    *regexp.Regexp
}

func NewRedactingHandler(inner slog.Handler) *RedactingHandler {
	// Redact OpenAI/Anthropic keys and Bearer tokens
	re := regexp.MustCompile(`(sk-[a-zA-Z0-9_-]{20,}|Bearer\s+[a-zA-Z0-9_\.\-]+|[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+)`)
	return &RedactingHandler{
		inner: inner,
		re:    re,
	}
}

func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *RedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	redactedMsg := h.re.ReplaceAllString(r.Message, "[REDACTED]")
	newRecord := slog.NewRecord(r.Time, r.Level, redactedMsg, r.PC)
	r.Attrs(func(a slog.Attr) bool {
		if a.Value.Kind() == slog.KindString {
			scrubbed := h.re.ReplaceAllString(a.Value.String(), "[REDACTED]")
			newRecord.AddAttrs(slog.String(a.Key, scrubbed))
		} else {
			newRecord.AddAttrs(a)
		}
		return true
	})
	return h.inner.Handle(ctx, newRecord)
}

func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &RedactingHandler{inner: h.inner.WithAttrs(attrs), re: h.re}
}

func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{inner: h.inner.WithGroup(name), re: h.re}
}

func NewProductionLogger() *slog.Logger {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(NewRedactingHandler(baseHandler))
}

func MaskString(input string) string {
	re := regexp.MustCompile(`(sk-[a-zA-Z0-9_-]{20,}|Bearer\s+[a-zA-Z0-9_\.\-]+|[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\.[a-zA-Z0-9-.]+)`)
	return re.ReplaceAllString(input, "[REDACTED]")
}
