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
	"errors"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnboundInteractiveResponse_KeepsOriginalMessage(t *testing.T) {
	t.Parallel()

	reply := UnboundInteractiveResponse()
	assert.Equal(t, SlashResponseTypeEphemeral, reply.ResponseType)
	assert.Equal(t, bindRequiredText, reply.Text)
	assert.False(t, reply.ReplaceOriginal)
}

func TestForbiddenInteractiveResponse_KeepsOriginalMessage(t *testing.T) {
	t.Parallel()

	reply := ForbiddenInteractiveResponse()
	assert.Equal(t, SlashResponseTypeEphemeral, reply.ResponseType)
	assert.Equal(t, interactiveForbiddenText, reply.Text)
	assert.False(t, reply.ReplaceOriginal)
}

func TestPostInteractiveEphemeral_PostsWithoutReplacing(t *testing.T) {
	t.Parallel()

	var (
		postedURL  string
		postedBody map[string]any
	)

	client := &http.Client{
		Transport: roundTripFunc(
			func(req *http.Request) (*http.Response, error) {
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
			},
		),
	}
	responseURL := "https://hooks.slack.com/actions/T123/1/abc"

	require.NoError(
		t,
		postInteractiveEphemeral(
			context.Background(),
			client,
			responseURL,
			interactiveForbiddenText,
		),
	)

	assert.Equal(t, responseURL, postedURL)
	assert.Equal(t, false, postedBody["replace_original"])
	assert.Equal(t, SlashResponseTypeEphemeral, postedBody["response_type"])
	assert.Equal(t, interactiveForbiddenText, postedBody["text"])
}

func TestPostInteractiveEphemeral_SkipsEmptyURL(t *testing.T) {
	t.Parallel()

	require.NoError(
		t,
		postInteractiveEphemeral(context.Background(), nil, "", bindRequiredText),
	)
}

func TestSanitizeHTTPError_OmitsURL(t *testing.T) {
	t.Parallel()

	sanitized := sanitizeHTTPError(
		&url.Error{
			Op:  "Post",
			URL: "https://hooks.slack.com/actions/T123/secret-token/abc",
			Err: errors.New("connection refused"),
		},
	)

	require.NotNil(t, sanitized)
	assert.NotContains(t, sanitized.Error(), "secret-token")
	assert.NotContains(t, sanitized.Error(), "hooks.slack.com")
	assert.Contains(t, sanitized.Error(), "Post")
	assert.Contains(t, sanitized.Error(), "connection refused")
}
