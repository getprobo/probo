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

package mcp_v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
	"go.probo.inc/probo/pkg/validator"
)

const feedbackTestBaseURL = "https://probod.example"

var feedbackIdentity = &coredata.Identity{
	FullName:     "Ada Lovelace",
	EmailAddress: mail.Addr("ada@example.com"),
}

func TestNewFeedbackConfig_Enabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want bool
		url  string
	}{
		{name: "hidden when unset"},
		{name: "hidden when blank", raw: "   "},
		{
			name: "available when a url is set",
			raw:  "  https://example.com/feedback  ",
			want: true,
			url:  "https://example.com/feedback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewFeedbackConfig(tt.raw)
			assert.Equal(t, tt.want, cfg.Enabled())
			assert.Equal(t, tt.url, cfg.URL)

			server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
			mcp.AddTool(
				server,
				&mcp.Tool{Name: submitFeedbackToolName, Description: "Send product feedback"},
				func(context.Context, *mcp.CallToolRequest, map[string]any) (*mcp.CallToolResult, map[string]any, error) {
					return nil, map[string]any{}, nil
				},
			)
			configureFeedbackTool(server, cfg)

			names := listToolNames(t, server)
			if tt.want {
				assert.Contains(t, names, submitFeedbackToolName)
			} else {
				assert.NotContains(t, names, submitFeedbackToolName)
			}
		})
	}
}

func TestSubmitFeedbackTool_postsWebhookPayload(t *testing.T) {
	t.Parallel()

	var (
		got     postedFeedback
		headers http.Header
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		headers = r.Header.Clone()
		require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	_, out, err := feedbackResolver(t, srv.URL).SubmitFeedbackTool(
		feedbackContext(),
		nil,
		&types.SubmitFeedbackInput{Message: "listRisks drops closed risks"},
	)
	require.NoError(t, err)
	assert.True(t, out.Submitted)

	_, err = uuid.Parse(got.EventID)
	require.NoError(t, err)
	assert.Nil(t, got.SubscriptionID)
	assert.Nil(t, got.OrganizationID)
	assert.Equal(t, feedbackEventType, got.EventType)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, 5*time.Second)
	assert.Nil(t, got.UpdatedFrom)
	assert.Equal(t, "Ada Lovelace", got.Actor.FullName)
	assert.Equal(t, "ada@example.com", got.Actor.EmailAddress)
	assert.Empty(t, got.Actor.Role)
	assert.Equal(t, "listRisks drops closed risks", got.Data.Message)

	assert.Equal(t, feedbackEventType, headers.Get("X-Probo-Webhook-Event"))
	assert.Equal(t, strconv.FormatInt(got.CreatedAt.Unix(), 10), headers.Get("X-Probo-Webhook-Timestamp"))
	assert.Equal(t, feedbackTestBaseURL, headers.Get("X-Probo-Webhook-Host"))
	assert.Empty(t, headers.Get("X-Probo-Webhook-Signature"))
}

func TestSubmitFeedbackTool_unavailableWithoutURL(t *testing.T) {
	t.Parallel()

	_, _, err := (&Resolver{
		feedback: NewFeedbackConfig(""),
	}).SubmitFeedbackTool(
		feedbackContext(),
		nil,
		&types.SubmitFeedbackInput{Message: "the list is missing a filter"},
	)
	require.EqualError(t, err, "feedback is not available")
}

func TestSubmitFeedbackTool_requiresIdentity(t *testing.T) {
	t.Parallel()

	_, _, err := feedbackResolver(t, "https://example.com/feedback").SubmitFeedbackTool(
		context.Background(),
		nil,
		&types.SubmitFeedbackInput{Message: "the list is missing a filter"},
	)
	require.EqualError(t, err, "missing identity")
}

func TestSubmitFeedbackTool_rejectsEmptyMessage(t *testing.T) {
	t.Parallel()

	_, _, err := feedbackResolver(t, "https://example.com/feedback").SubmitFeedbackTool(
		feedbackContext(),
		nil,
		&types.SubmitFeedbackInput{Message: "   "},
	)

	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	assert.NotEmpty(t, validationErrors)
}

func TestSubmitFeedbackTool_acceptsMessageAtRuneLimit(t *testing.T) {
	t.Parallel()

	message := strings.Repeat("é", feedbackMessageMaxLen)
	require.Greater(t, len(message), feedbackMessageMaxLen)

	var (
		got       postedFeedback
		decodeErr error
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decodeErr = json.NewDecoder(r.Body).Decode(&got)

		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	_, out, err := feedbackResolver(t, srv.URL).SubmitFeedbackTool(
		feedbackContext(),
		nil,
		&types.SubmitFeedbackInput{Message: message},
	)
	require.NoError(t, err)
	require.NoError(t, decodeErr)
	assert.True(t, out.Submitted)
	assert.Equal(t, message, got.Data.Message)
}

func TestSubmitFeedbackTool_rejectsMessageOverRuneLimit(t *testing.T) {
	t.Parallel()

	_, _, err := feedbackResolver(t, "https://example.com/feedback").SubmitFeedbackTool(
		feedbackContext(),
		nil,
		&types.SubmitFeedbackInput{Message: strings.Repeat("é", feedbackMessageMaxLen+1)},
	)

	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	require.NotEmpty(t, validationErrors)
	assert.Equal(t, validator.ErrorCodeTooLong, validationErrors[0].Code)
}

func TestFeedbackDeliveryError_dropsRequestURL(t *testing.T) {
	t.Parallel()

	err := feedbackDeliveryError(&url.Error{
		Op:  http.MethodPost,
		URL: "https://hooks.example/secret-token",
		Err: errors.New("connection refused"),
	})

	assert.EqualError(t, err, "connection refused")
	assert.NotContains(t, err.Error(), "secret-token")
}

func TestSubmitFeedbackTool_rejectsNonSuccessStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	_, _, err := feedbackResolver(t, srv.URL).SubmitFeedbackTool(
		feedbackContext(),
		nil,
		&types.SubmitFeedbackInput{Message: "the list is missing a filter"},
	)
	require.EqualError(t, err, "cannot deliver feedback")
}

type postedFeedback struct {
	EventID        string          `json:"eventId"`
	SubscriptionID *string         `json:"subscriptionId"`
	OrganizationID *string         `json:"organizationId"`
	EventType      string          `json:"eventType"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedFrom    json.RawMessage `json:"updatedFrom"`
	Actor          struct {
		FullName     string `json:"fullName"`
		EmailAddress string `json:"emailAddress"`
		Role         string `json:"role"`
	} `json:"actor"`
	Data struct {
		Message string `json:"message"`
	} `json:"data"`
}

func feedbackContext() context.Context {
	return authn.ContextWithIdentity(context.Background(), feedbackIdentity)
}

func feedbackResolver(t *testing.T, rawURL string) *Resolver {
	t.Helper()

	cfg := NewFeedbackConfig(rawURL)

	return &Resolver{
		logger:   log.NewLogger(),
		baseURL:  baseurl.MustParse(feedbackTestBaseURL),
		feedback: cfg,
		feedbackClient: httpclient.DefaultPooledClient(
			httpclient.WithSSRFProtection(),
			httpclient.WithSSRFAllowLoopback(),
		),
	}
}

func listToolNames(t *testing.T, server *mcp.Server) []string {
	t.Helper()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil)
	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })

	clientSession, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	result, err := clientSession.ListTools(ctx, nil)
	require.NoError(t, err)

	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}

	return names
}
