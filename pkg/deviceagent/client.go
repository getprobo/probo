// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
	EnrollRequest struct {
		EnrollmentToken string  `json:"enrollment_token"`
		HardwareUUID    string  `json:"hardware_uuid"`
		SerialNumber    *string `json:"serial_number,omitempty"`
		Hostname        string  `json:"hostname"`
		Platform        string  `json:"platform"`
		OSVersion       string  `json:"os_version"`
		AgentVersion    string  `json:"agent_version"`
	}

	EnrollResponse struct {
		DeviceID         string `json:"device_id"`
		APIKey           string `json:"api_key"`
		HeartbeatSeconds int    `json:"heartbeat_interval_seconds"`
		PostureSeconds   int    `json:"posture_interval_seconds"`
		ServerTime       string `json:"server_time"`
	}

	HeartbeatRequest struct {
		AgentVersion string `json:"agent_version,omitempty"`
		Hostname     string `json:"hostname,omitempty"`
		OSVersion    string `json:"os_version,omitempty"`
		UptimeSec    int64  `json:"uptime_seconds,omitempty"`
	}

	HeartbeatResponse struct {
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

// Enroll exchanges an enrollment token for a device key.
func (c *Client) Enroll(ctx context.Context, req EnrollRequest) (*EnrollResponse, error) {
	var resp EnrollResponse
	if err := c.do(
		ctx,
		http.MethodPost,
		"/api/agent/v1/enroll",
		false,
		req,
		&resp,
	); err != nil {
		return nil, err
	}

	return &resp, nil
}

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
