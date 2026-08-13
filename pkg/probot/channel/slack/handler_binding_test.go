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
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type (
	stubBindingGate struct {
		binding       *identitybinding.Binding
		bindURL       string
		lookupSubject identitybinding.Subject
		bindSubject   identitybinding.Subject
	}

	recordingSlackClient struct {
		mu sync.Mutex

		openIMUser   string
		openIMChan   string
		posts        []recordedPost
		statuses     []recordedStatus
		clientIDs    []string
		createErrors []error
	}

	recordedPost struct {
		channel  string
		text     string
		threadTS string
		blocks   []any
	}

	recordedStatus struct {
		channelID string
		threadTS  string
		status    string
		loading   []string
	}

	stubInstallationClients struct {
		client       *Client
		installation *coredata.SlackbotInstallation
	}
)

func (s *stubBindingGate) Lookup(
	_ context.Context,
	subject identitybinding.Subject,
) (*identitybinding.Binding, error) {
	s.lookupSubject = subject
	if s.binding == nil {
		return nil, coredata.ErrResourceNotFound
	}

	return s.binding, nil
}

func (s *stubBindingGate) BindURL(
	_ context.Context,
	subject identitybinding.Subject,
) (string, error) {
	s.bindSubject = subject

	return s.bindURL, nil
}

func (c *recordingSlackClient) OpenIM(_ context.Context, userID string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.openIMUser = userID
	if c.openIMChan == "" {
		c.openIMChan = "D-test-dm"
	}

	return c.openIMChan, nil
}

func (c *recordingSlackClient) CreateMessage(
	_ context.Context,
	channel string,
	text string,
	threadTS string,
	clientMsgID string,
) (*MessageRef, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.createErrors) > 0 {
		err := c.createErrors[0]

		c.createErrors = c.createErrors[1:]
		if err != nil {
			c.clientIDs = append(c.clientIDs, clientMsgID)
			return nil, err
		}
	}

	c.posts = append(
		c.posts,
		recordedPost{
			channel:  channel,
			text:     text,
			threadTS: threadTS,
		},
	)
	c.clientIDs = append(c.clientIDs, clientMsgID)

	return &MessageRef{Channel: channel, TS: "reply-ts"}, nil
}

func (c *recordingSlackClient) CreateMessageWithBlocks(
	_ context.Context,
	channel string,
	text string,
	blocks []any,
	clientMsgID string,
) (*MessageRef, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.createErrors) > 0 {
		err := c.createErrors[0]

		c.createErrors = c.createErrors[1:]
		if err != nil {
			c.clientIDs = append(c.clientIDs, clientMsgID)
			return nil, err
		}
	}

	c.posts = append(
		c.posts,
		recordedPost{
			channel: channel,
			text:    text,
			blocks:  blocks,
		},
	)
	c.clientIDs = append(c.clientIDs, clientMsgID)

	return &MessageRef{Channel: channel, TS: "reply-ts"}, nil
}

func (c *recordingSlackClient) postsSnapshot() []recordedPost {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]recordedPost, len(c.posts))
	copy(out, c.posts)

	return out
}

func (c *recordingSlackClient) SetStatus(
	_ context.Context,
	channelID, threadTS, status string,
	loadingMessages []string,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.statuses = append(
		c.statuses,
		recordedStatus{
			channelID: channelID,
			threadTS:  threadTS,
			status:    status,
			loading:   loadingMessages,
		},
	)

	return nil
}

func (c *recordingSlackClient) statusesSnapshot() []recordedStatus {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]recordedStatus, len(c.statuses))
	copy(out, c.statuses)

	return out
}

func (s *stubInstallationClients) ClientByTeamID(
	context.Context,
	string,
) (*Client, *coredata.SlackbotInstallation, error) {
	return s.client, s.installation, nil
}

func (s *stubInstallationClients) DisableByTeamID(context.Context, string) error {
	return nil
}

func TestHandler_HandleInteraction_UnboundPostsCTA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	slackClient := &recordingSlackClient{}
	eventID := "E-unbound-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()
	teamID := "T-unbound-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()
	bindings := &stubBindingGate{
		bindURL: "https://console.example.com/me/probot/bind?token=test",
	}

	handler := &Handler{
		client:   slackClient,
		bindings: bindings,
		pg:       client,
		logger:   log.NewLogger(),
	}

	require.NoError(t, handler.handleInteraction(
		ctx,
		eventID,
		teamID,
		&EventBody{
			Type:        EventTypeAppMention,
			User:        "U456",
			Text:        "hello",
			Channel:     "C789",
			TS:          "111.222",
			ChannelType: ChannelTypeChannel,
		},
	))

	posts := slackClient.postsSnapshot()
	require.Len(t, posts, 1)
	assert.Equal(t, "C789", posts[0].channel)
	assert.Equal(t, bindRequiredText, posts[0].text)
	assert.Nil(t, posts[0].blocks)
	assert.Equal(t, "111.222", posts[0].threadTS)
	assert.Empty(t, slackClient.openIMUser)
	assert.Equal(t, IdentitySubject(teamID, "U456"), bindings.lookupSubject)
	assert.Empty(t, bindings.bindSubject.ExternalUserID)
	assert.Empty(t, slackClient.statusesSnapshot())

	require.NoError(t, handler.handleInteraction(
		ctx,
		eventID,
		teamID,
		&EventBody{
			Type:        EventTypeAppMention,
			User:        "U456",
			Text:        "hello",
			Channel:     "C789",
			TS:          "111.222",
			ChannelType: ChannelTypeChannel,
		},
	))

	assert.Len(t, slackClient.postsSnapshot(), 1)
	assert.Empty(t, slackClient.statusesSnapshot())
}

func TestHandler_HandleInteraction_BindCTARetriesWithStableClientID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	slackClient := &recordingSlackClient{
		createErrors: []error{errors.New("temporary Slack failure")},
	}
	eventID := "E-bind-retry-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()
	teamID := "T-bind-retry-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()
	handler := &Handler{
		client:   slackClient,
		bindings: &stubBindingGate{bindURL: "https://console.example.com/me/probot/bind?token=test"},
		pg:       client,
		logger:   log.NewLogger(),
	}
	event := &EventBody{
		Type:        EventTypeAppMention,
		User:        "U456",
		Text:        "hello",
		Channel:     "C789",
		TS:          "111.222",
		ChannelType: ChannelTypeChannel,
	}

	require.Error(t, handler.handleInteraction(ctx, eventID, teamID, event))
	require.NoError(t, handler.handleInteraction(ctx, eventID, teamID, event))

	require.Len(t, slackClient.clientIDs, 2)
	assert.Equal(t, slackClient.clientIDs[0], slackClient.clientIDs[1])
	assert.Len(t, slackClient.clientIDs[0], 36)
	assert.Len(t, slackClient.postsSnapshot(), 1)
	assert.Empty(t, slackClient.statusesSnapshot())
}

func TestHandler_HandleInteraction_UnboundInDMPostsSingleMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	slackClient := &recordingSlackClient{}
	eventID := "E-dm-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()
	teamID := "T-dm-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()

	handler := &Handler{
		client:   slackClient,
		bindings: &stubBindingGate{bindURL: "https://console.example.com/me/probot/bind?token=test"},
		pg:       client,
		logger:   log.NewLogger(),
	}

	require.NoError(
		t,
		handler.handleInteraction(
			ctx,
			eventID,
			teamID,
			&EventBody{
				Type:        EventTypeMessage,
				User:        "U456",
				Text:        "hello",
				Channel:     "D789",
				TS:          "333.444",
				ChannelType: ChannelTypeIM,
			},
		),
	)

	posts := slackClient.postsSnapshot()
	require.Len(t, posts, 1)
	assert.Equal(t, "D789", posts[0].channel)
	assert.Equal(t, bindRequiredText, posts[0].text)
	assert.Nil(t, posts[0].blocks)
	assert.Empty(t, slackClient.openIMUser)
	assert.Empty(t, slackClient.statusesSnapshot())
}

func TestHandler_HandleInteraction_BoundMentionSetsAssistantStatus(t *testing.T) {
	t.Parallel()

	pgClient, _, organizationID := executionIngressDatabase(t)
	var got []map[string]any
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/conversations.replies":
					_, err := w.Write([]byte(`{"ok":true,"messages":[]}`))
					require.NoError(t, err)
				case "/api/assistant.threads.setStatus":
					var body map[string]any
					require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
					got = append(got, body)
					_, err := w.Write([]byte(`{"ok":true}`))
					require.NoError(t, err)
				default:
					t.Errorf("unexpected Slack method %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			},
		),
	)
	t.Cleanup(server.Close)

	handler := &Handler{
		installations: &stubInstallationClients{
			client: newTestClient(server.URL + "/api"),
			installation: &coredata.SlackbotInstallation{
				OrganizationID: organizationID,
				BotUserID:      "UBOT",
			},
		},
		bindings: &stubBindingGate{
			binding: &identitybinding.Binding{
				IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
			},
		},
		pg:     pgClient,
		logger: log.NewLogger(),
	}

	require.NoError(
		t,
		handler.handleInteraction(
			t.Context(),
			"E-bound-mention-"+organizationID.String(),
			"T-bound",
			&EventBody{
				Type:        EventTypeAppMention,
				User:        "U456",
				Text:        "hello",
				Channel:     "C789",
				TS:          "111.222",
				ThreadTS:    "100.000",
				ChannelType: ChannelTypeChannel,
			},
		),
	)

	require.Len(t, got, 1)
	assert.Equal(t, "C789", got[0]["channel_id"])
	assert.Equal(t, "100.000", got[0]["thread_ts"])
	assert.Equal(t, assistantWorkingStatus, got[0]["status"])
	assert.Equal(
		t,
		[]any{
			"Looking that up…",
			"Checking your workspace…",
			"Drafting a reply…",
		},
		got[0]["loading_messages"],
	)
}

func TestHandler_HandleInteraction_BoundDMSetsAssistantStatus(t *testing.T) {
	t.Parallel()

	pgClient, _, organizationID := executionIngressDatabase(t)
	var got []map[string]any
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/assistant.threads.setStatus", r.URL.Path)
				var body map[string]any
				require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
				got = append(got, body)
				_, err := w.Write([]byte(`{"ok":true}`))
				require.NoError(t, err)
			},
		),
	)
	t.Cleanup(server.Close)

	handler := &Handler{
		installations: &stubInstallationClients{
			client: newTestClient(server.URL + "/api"),
			installation: &coredata.SlackbotInstallation{
				OrganizationID: organizationID,
				BotUserID:      "UBOT",
			},
		},
		bindings: &stubBindingGate{
			binding: &identitybinding.Binding{
				IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
			},
		},
		pg:     pgClient,
		logger: log.NewLogger(),
	}

	require.NoError(
		t,
		handler.handleInteraction(
			t.Context(),
			"E-bound-dm-"+organizationID.String(),
			"T-bound",
			&EventBody{
				Type:        EventTypeMessage,
				User:        "U456",
				Text:        "hello",
				Channel:     "D789",
				TS:          "333.444",
				ChannelType: ChannelTypeIM,
			},
		),
	)

	require.Len(t, got, 1)
	assert.Equal(t, "D789", got[0]["channel_id"])
	assert.Equal(t, "333.444", got[0]["thread_ts"])
	assert.Equal(t, assistantWorkingStatus, got[0]["status"])
}
