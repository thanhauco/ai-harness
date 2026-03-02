package provider

import (
	"net/http"
	"time"
)

type ClientOptions struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxIdle    int
	HTTPClient *http.Client
}

func DefaultClientOptions() ClientOptions {
	return ClientOptions{
		Timeout: 30 * time.Second,
		MaxIdle: 100,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// HTTPProvider communicates with remote LLM HTTP API endpoints.
type HTTPProvider struct {
	name string
	opts ClientOptions
}

func NewHTTPProvider(name string, opts ClientOptions) *HTTPProvider {
	if opts.HTTPClient == nil {
		opts = DefaultClientOptions()
	}
	return &HTTPProvider{
		name: name,
		opts: opts,
	}
}

func (h *HTTPProvider) Name() string {
	return h.name
}
