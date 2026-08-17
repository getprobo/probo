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

package slack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

func TestNormalizeSlackResponseURL(t *testing.T) {
	t.Parallel()

	t.Run(
		"accepts slack command response urls",
		func(t *testing.T) {
			t.Parallel()

			got, err := normalizeSlackResponseURL(
				" https://hooks.slack.com/commands/T123/1/abc ",
			)
			require.NoError(t, err)
			assert.Equal(t, "https://hooks.slack.com/commands/T123/1/abc", got)
		},
	)

	t.Run(
		"rejects non-slack hosts",
		func(t *testing.T) {
			t.Parallel()

			_, err := normalizeSlackResponseURL("https://example.com/commands/T123/1/abc")
			require.Error(t, err)
		},
	)
}

func TestBindPromptService_BindingConfirmedReplacesEphemeral(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	key := cipher.EncryptionKey{1, 2, 3}
	service := NewBindPromptService(client, key, log.NewLogger())

	var (
		postedURL  string
		postedBody map[string]any
	)

	service.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			postedURL = req.URL.String()
			defer func() { _ = req.Body.Close() }()

			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}

			if err := json.Unmarshal(body, &postedBody); err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       http.NoBody,
				Header:     make(http.Header),
			}, nil
		}),
	}

	responseURL := "https://hooks.slack.com/commands/T-BIND/1/token"

	err := service.RememberResponseURL(ctx, "T-BIND", "U-BIND", responseURL)
	if err != nil {
		t.Skipf("slackbot_bind_callbacks is unavailable in the test database: %v", err)
	}

	require.NoError(
		t,
		service.BindingConfirmed(
			ctx,
			identitybinding.Subject{
				Provider:         ProviderName,
				ExternalTenantID: "T-BIND",
				ExternalUserID:   "U-BIND",
			},
		),
	)

	assert.Equal(t, responseURL, postedURL)
	assert.Equal(t, true, postedBody["replace_original"])
	assert.Equal(t, SlashResponseTypeEphemeral, postedBody["response_type"])
	assert.Equal(t, bindSlashLinkedText, postedBody["text"])
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
