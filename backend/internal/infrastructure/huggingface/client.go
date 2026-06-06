// Package huggingface implements perception adapters backed by the Hugging Face
// serverless Inference Providers API (hf-inference). All outbound calls use a
// context deadline (backend.mdc: no timeoutless HTTP) and handle the cold-start
// 503 + estimated_time pattern with bounded retry (design doc 7.1).
package huggingface

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultBaseURL is the hf-inference router base. Override with HF_INFERENCE_BASE_URL.
const DefaultBaseURL = "https://router.huggingface.co/hf-inference/models"

// Client is a thin HF inference HTTP client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	maxWait    time.Duration // total budget incl. cold-start retries
}

// Option configures the Client.
type Option func(*Client)

// WithHTTPClient overrides the underlying http.Client (testing).
func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.httpClient = h } }

// WithBaseURL overrides the inference base URL.
func WithBaseURL(u string) Option {
	return func(c *Client) {
		if u != "" {
			c.baseURL = u
		}
	}
}

// WithMaxWait sets the total time budget for a detection incl. retries.
func WithMaxWait(d time.Duration) Option { return func(c *Client) { c.maxWait = d } }

// NewClient builds a client. token may be empty, but live calls then fail with
// an auth error; callers should gate registration on a non-empty token.
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 20 * time.Second},
		maxWait:    15 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// rawBox mirrors the HF object-detection response box.
type rawBox struct {
	XMin float64 `json:"xmin"`
	YMin float64 `json:"ymin"`
	XMax float64 `json:"xmax"`
	YMax float64 `json:"ymax"`
}

// detectionResponse is one element of the HF object-detection response array.
type detectionResponse struct {
	Label string  `json:"label"`
	Score float64 `json:"score"`
	Box   rawBox  `json:"box"`
}

// hfError is the JSON body HF returns for transient/cold-start conditions.
type hfError struct {
	Error         string  `json:"error"`
	EstimatedTime float64 `json:"estimated_time"`
}

// detect posts image bytes to a model and returns parsed detections. It retries
// on HTTP 503 honoring estimated_time, bounded by the client's maxWait budget
// and the caller's context deadline.
func (c *Client) detect(ctx context.Context, model string, image []byte) ([]detectionResponse, error) {
	deadline := time.Now().Add(c.maxWait)
	url := c.baseURL + "/" + model

	for attempt := 0; ; attempt++ {
		body, status, err := c.post(ctx, url, image)
		if err != nil {
			return nil, err
		}

		switch {
		case status == http.StatusOK:
			var out []detectionResponse
			if err := json.Unmarshal(body, &out); err != nil {
				return nil, fmt.Errorf("huggingface: decode response: %w", err)
			}
			return out, nil

		case status == http.StatusServiceUnavailable:
			wait := coldStartWait(body)
			if time.Now().Add(wait).After(deadline) {
				return nil, fmt.Errorf("huggingface: model %s cold (503), exceeds %s budget", model, c.maxWait)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}

		default:
			return nil, fmt.Errorf("huggingface: model %s status %d: %s", model, status, truncate(body, 200))
		}
	}
}

func (c *Client) post(ctx context.Context, url string, image []byte) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(image))
	if err != nil {
		return nil, 0, fmt.Errorf("huggingface: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("huggingface: request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("huggingface: read body: %w", err)
	}
	return body, resp.StatusCode, nil
}

// coldStartWait extracts estimated_time (capped) or falls back to a short wait.
func coldStartWait(body []byte) time.Duration {
	var e hfError
	if json.Unmarshal(body, &e) == nil && e.EstimatedTime > 0 {
		d := time.Duration(e.EstimatedTime*float64(time.Second)) + 500*time.Millisecond
		if d > 10*time.Second {
			d = 10 * time.Second
		}
		return d
	}
	return 2 * time.Second
}

func truncate(b []byte, n int) string {
	if len(b) <= n {
		return string(b)
	}
	return string(b[:n]) + "..."
}
