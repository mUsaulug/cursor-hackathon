// Package streetview fetches Google Street View Static API frames. Street View
// frames are raw PII (faces, plates), so callers MUST anonymize before any
// inference, persistence, or transmission (docs/PRIVACY.md, design doc 5.3).
// This source is opt-in via GOOGLE_STREET_VIEW_API_KEY and disabled by default
// (MVP is PII-avoidance-by-design). All calls use a context deadline.
package streetview

import (
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"time"

	"cursor-hackathon/backend/internal/shared/imaging"
)

// DefaultBaseURL is the Street View Static API endpoint.
const DefaultBaseURL = "https://maps.googleapis.com/maps/api/streetview"

// Source fetches Street View frames.
type Source struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Option configures the Source.
type Option func(*Source)

// WithBaseURL overrides the API base (testing).
func WithBaseURL(u string) Option {
	return func(s *Source) {
		if u != "" {
			s.baseURL = u
		}
	}
}

// WithHTTPClient overrides the HTTP client (testing).
func WithHTTPClient(h *http.Client) Option { return func(s *Source) { s.client = h } }

// NewSource builds the Street View source.
func NewSource(apiKey string, opts ...Option) *Source {
	s := &Source{
		apiKey:  apiKey,
		baseURL: DefaultBaseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Fetch retrieves a JPEG frame for a coordinate. The returned bytes are RAW and
// MUST be passed through AnonymizeFrame before any further use.
func (s *Source) Fetch(ctx context.Context, lat, lng float64, size string) ([]byte, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("streetview: missing API key")
	}
	if size == "" {
		size = "640x640"
	}
	q := url.Values{}
	q.Set("size", size)
	q.Set("location", fmt.Sprintf("%f,%f", lat, lng))
	q.Set("key", s.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("streetview: build request: %w", err)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("streetview: request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("streetview: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("streetview: read body: %w", err)
	}
	return data, nil
}

// AnonymizeFrame irreversibly pixelates the given regions of a JPEG frame and
// returns anonymized JPEG bytes. This is the blur-before-inference step. The
// raw input must be discarded by the caller immediately after.
func AnonymizeFrame(jpegData []byte, regions []image.Rectangle, block int) ([]byte, error) {
	img, _, err := imaging.Decode(jpegData)
	if err != nil {
		return nil, err
	}
	anon := imaging.PixelateRegions(img, regions, block)
	return imaging.EncodeJPEG(anon, 88)
}
