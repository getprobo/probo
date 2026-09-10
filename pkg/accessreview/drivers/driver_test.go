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

package drivers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveFromStatus(t *testing.T) {
	t.Parallel()

	// "active" → active, case-insensitive.
	for _, s := range []string{"active", "ACTIVE", " Active "} {
		got := activeFromStatus(s)
		require.NotNilf(t, got, "status %q", s)
		assert.Truef(t, *got, "status %q", s)
	}

	// Empty/whitespace status → no signal (nil).
	assert.Nil(t, activeFromStatus(""))
	assert.Nil(t, activeFromStatus("   "))

	// Any other non-empty status is treated as inactive (Brevo "pending",
	// Pylon "deactivated", and any unrecognised future value).
	for _, s := range []string{"pending", "deactivated", "disabled", "whatever"} {
		got := activeFromStatus(s)
		require.NotNilf(t, got, "status %q", s)
		assert.Falsef(t, *got, "status %q", s)
	}
}

// TestRetryRoundTripperRewindsBody covers the retry of a POST. The first
// attempt reads the body to the end, so an unrewound retry sends an empty one
// and the provider answers 400 — turning a recoverable 5xx into a failed sync.
func TestRetryRoundTripperRewindsBody(t *testing.T) {
	t.Parallel()

	var bodies []string

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		bodies = append(bodies, string(body))

		status := http.StatusServiceUnavailable
		if len(bodies) > 1 {
			status = http.StatusOK
		}

		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(`{}`)),
			Header:     http.Header{},
		}, nil
	})

	payload := `{"query":"{ viewer { id } }"}`

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"https://example.com/graphql",
		strings.NewReader(payload),
	)
	require.NoError(t, err)

	client := &http.Client{Transport: &retryRoundTripper{next: transport, maxRetries: 3}}

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, bodies, 2)
	// Both attempts carry the whole query, not just the first.
	assert.Equal(t, payload, bodies[0])
	assert.Equal(t, payload, bodies[1])
}
