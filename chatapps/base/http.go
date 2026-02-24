package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient provides common HTTP request patterns for chat adapters.
// This eliminates the duplicate HTTP request building code across adapters.
type HTTPClient struct {
	client     *http.Client
	maxRetries int
	retryDelay time.Duration
}

// NewHTTPClient creates a new HTTPClient with the default http.Client.
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client:     http.DefaultClient,
		maxRetries: 0,
		retryDelay: 100 * time.Millisecond,
	}
}

// NewHTTPClientWithConfig creates a new HTTPClient with custom configuration.
func NewHTTPClientWithConfig(timeout time.Duration, maxRetries int) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		retryDelay: 100 * time.Millisecond,
	}
}

// PostJSON sends a POST request with a JSON body and returns the response body.
// headers is a map of header key-value pairs to add to the request.
func (c *HTTPClient) PostJSON(ctx context.Context, url string, payload any, headers map[string]string) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.doRequest(req)
}

// PostJSONWithResponse sends a POST request and decodes the JSON response into dest.
func (c *HTTPClient) PostJSONWithResponse(ctx context.Context, url string, payload any, headers map[string]string, dest any) error {
	respBody, err := c.PostJSON(ctx, url, payload, headers)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(respBody, dest); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

// Get sends a GET request and returns the response body.
func (c *HTTPClient) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.doRequest(req)
}

// doRequest executes the request with optional retry logic.
func (c *HTTPClient) doRequest(req *http.Request) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.retryDelay * time.Duration(attempt))
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}
		defer func() { _ = resp.Body.Close() }()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("read response body: %w", err)
			continue
		}

		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
			continue
		}

		return respBody, nil
	}

	return nil, lastErr
}

// ExtractStringFromMetadata extracts a string value from message metadata.
// Returns empty string if the key doesn't exist or the value is not a string.
func ExtractStringFromMetadata(msg *ChatMessage, key string) string {
	if msg == nil || msg.Metadata == nil {
		return ""
	}
	if val, ok := msg.Metadata[key].(string); ok {
		return val
	}
	return ""
}

// ExtractInt64FromMetadata extracts an int64 value from message metadata.
// Returns 0 if the key doesn't exist or the value is not a numeric type.
func ExtractInt64FromMetadata(msg *ChatMessage, key string) int64 {
	if msg == nil || msg.Metadata == nil {
		return 0
	}
	switch v := msg.Metadata[key].(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case int:
		return int64(v)
	}
	return 0
}
