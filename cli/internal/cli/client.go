package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pablogtrz/tablon/internal/version"
)

const maxResponseBytes = 10 << 20

type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	switch {
	case e.Status != 0 && e.Code != "" && e.Message != "":
		return fmt.Sprintf("%d %s: %s", e.Status, e.Code, e.Message)
	case e.Status != 0 && e.Code != "":
		return fmt.Sprintf("%d %s", e.Status, e.Code)
	case e.Status != 0:
		return fmt.Sprintf("request failed with status %d", e.Status)
	case e.Code != "":
		return e.Code
	default:
		return "request failed"
	}
}

type Client struct {
	baseURL   *url.URL
	token     string
	userAgent string
	http      *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8448"
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("URL must start with http:// or https://")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("URL must include a host")
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &Client{
		baseURL:   parsed,
		token:     strings.TrimSpace(token),
		userAgent: "tablon-cli/" + version.Version,
		http:      &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) Do(
	ctx context.Context,
	method, path string,
	body any,
	headers map[string]string,
) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(raw)
	}

	relative, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	target := *c.baseURL
	target.Path = strings.TrimRight(c.baseURL.Path, "/") + relative.Path
	target.RawQuery = relative.RawQuery
	req, err := http.NewRequestWithContext(ctx, method, target.String(), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		var envelope struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(raw, &envelope)
		return nil, &APIError{
			Status:  resp.StatusCode,
			Code:    envelope.Error.Code,
			Message: envelope.Error.Message,
		}
	}
	return raw, nil
}
