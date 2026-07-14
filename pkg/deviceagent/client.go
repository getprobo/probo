// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package deviceagent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
)

type (
	// Client calls the /api/agent/v1 REST API.
	Client struct {
		ServerURL string
		APIKey    string
		UserAgent string
		HTTP      *http.Client
	}
)

// NewClient creates an API client.
func NewClient(serverURL, apiKey, userAgent string) *Client {
	httpClient := httpclient.DefaultPooledClient()
	httpClient.Timeout = 30 * time.Second

	return &Client{
		ServerURL: strings.TrimRight(serverURL, "/"),
		APIKey:    apiKey,
		UserAgent: userAgent,
		HTTP:      httpClient,
	}
}

type (
	HeartbeatRequest struct {
		HardwareUUID string  `json:"hardware_uuid"`
		SerialNumber *string `json:"serial_number,omitempty"`
		Hostname     string  `json:"hostname"`
		Platform     string  `json:"platform"`
		OSVersion    string  `json:"os_version"`
		AgentVersion string  `json:"agent_version"`
	}

	HeartbeatResponse struct {
		DeviceID         string `json:"device_id"`
		HeartbeatSeconds int    `json:"heartbeat_interval_seconds"`
		PostureSeconds   int    `json:"posture_interval_seconds"`
		ServerTime       string `json:"server_time"`
	}

	PostureResultPayload struct {
		CheckKey   string          `json:"check_key"`
		Status     string          `json:"status"`
		Evidence   json.RawMessage `json:"evidence,omitempty"`
		ObservedAt time.Time       `json:"observed_at"`
	}

	PosturesRequest struct {
		Results []PostureResultPayload `json:"results"`
	}
)

// Heartbeat sends a periodic device heartbeat.
func (c *Client) Heartbeat(ctx context.Context, req HeartbeatRequest) (*HeartbeatResponse, error) {
	var resp HeartbeatResponse
	if err := c.do(
		ctx,
		http.MethodPost,
		"/api/agent/v1/heartbeat",
		true,
		req,
		&resp,
	); err != nil {
		return nil, err
	}

	return &resp, nil
}

// PushPostures sends posture check results.
func (c *Client) PushPostures(ctx context.Context, results []PostureResultPayload) error {
	if len(results) == 0 {
		return nil
	}

	return c.do(
		ctx,
		http.MethodPost,
		"/api/agent/v1/postures",
		true,
		PosturesRequest{Results: results},
		nil,
	)
}

// Unenroll asks the server to revoke the device.
func (c *Client) Unenroll(ctx context.Context) error {
	return c.do(
		ctx,
		http.MethodPost,
		"/api/agent/v1/unenroll",
		true,
		nil,
		nil,
	)
}

type enrollResponse struct {
	APIKey string `json:"api_key"`
}

// ExchangeEnrollmentToken redeems a one-shot enrollment token for the
// permanent device API key.
func (c *Client) ExchangeEnrollmentToken(
	ctx context.Context,
	token string,
) (string, error) {
	var resp enrollResponse
	if err := c.do(
		ctx,
		http.MethodPost,
		"/api/agent/v1/enroll",
		false,
		map[string]string{"token": token},
		&resp,
	); err != nil {
		return "", err
	}

	if resp.APIKey == "" {
		return "", errors.New("agent api: empty api key in enroll response")
	}

	return resp.APIKey, nil
}

// HTTPError captures a non-2xx API response.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("agent api: %d %s", e.StatusCode, e.Body)
}

// IsUnauthorized reports whether err is an API 401.
func IsUnauthorized(err error) bool {
	var herr *HTTPError
	if !errors.As(err, &herr) {
		return false
	}

	return herr.StatusCode == http.StatusUnauthorized
}

func (c *Client) do(
	ctx context.Context,
	method, path string,
	authed bool,
	in any,
	out any,
) error {
	url := c.ServerURL + path

	var body io.Reader

	if in != nil {
		buf, err := json.Marshal(in)
		if err != nil {
			return fmt.Errorf("cannot marshal request: %w", err)
		}

		body = bytes.NewReader(buf)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("cannot build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	if authed {
		if c.APIKey == "" {
			return errors.New("agent client: no api key set")
		}

		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("cannot perform request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		buf, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return &HTTPError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(buf))}
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("cannot decode response: %w", err)
	}

	return nil
}
