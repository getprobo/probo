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
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type stubBindingGate struct {
	binding       *identitybinding.Binding
	bindURL       string
	lookupSubject identitybinding.Subject
	bindSubject   identitybinding.Subject
	lookupErr     error
}

func (s *stubBindingGate) Lookup(
	_ context.Context,
	subject identitybinding.Subject,
) (*identitybinding.Binding, error) {
	s.lookupSubject = subject
	if s.lookupErr != nil {
		return nil, s.lookupErr
	}

	if s.binding == nil {
		return nil, coredata.ErrResourceNotFound
	}

	return s.binding, nil
}

func (s *stubBindingGate) BindURL(
	_ context.Context,
	subject identitybinding.Subject,
	_ gid.GID,
) (string, error) {
	s.bindSubject = subject

	return s.bindURL, nil
}

type recordingSlackAPI struct {
	URL string

	mu              sync.Mutex
	posts           []map[string]any
	successfulPosts []map[string]any
	statuses        []map[string]any
	clientIDs       []string
}

func TestService_HandleInteraction_UnboundPostsCTA(t *testing.T) {
	t.Parallel()

	slackAPI := newRecordingSlackAPI(t, nil)
	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	bindings := &stubBindingGate{
		bindURL: "https://console.example.com/organizations/org/employee/bind?token=test",
	}
	handler := newInteractionService(t, pgClient, slackAPI.URL, bindings)
	eventID := "E-unbound-" + organizationID.String()

	require.NoError(
		t,
		handler.handleInteraction(
			t.Context(),
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
		),
	)

	posts := slackAPI.postsSnapshot()
	require.Len(t, posts, 1)
	assert.Equal(t, "C789", posts[0]["channel"])
	assert.Equal(t, bindRequiredText, posts[0]["text"])
	assert.Equal(t, "111.222", posts[0]["thread_ts"])
	assert.Nil(t, posts[0]["blocks"])
	assert.Equal(t, IdentitySubject(teamID, "U456"), bindings.lookupSubject)
	assert.Empty(t, bindings.bindSubject.ExternalUserID)
	assert.Empty(t, slackAPI.statusesSnapshot())

	require.NoError(
		t,
		handler.handleInteraction(
			t.Context(),
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
		),
	)

	assert.Len(t, slackAPI.postsSnapshot(), 1)
	assert.Empty(t, slackAPI.statusesSnapshot())
}

func TestService_HandleInteraction_BindCTARetriesWithStableClientID(t *testing.T) {
	t.Parallel()

	var attempts int

	slackAPI := newRecordingSlackAPI(
		t,
		func(w http.ResponseWriter, _ map[string]any) bool {
			if attempts == 0 {
				attempts++
				_, err := w.Write([]byte(`{"ok":false,"error":"fatal_error"}`))
				require.NoError(t, err)

				return true
			}

			return false
		},
	)
	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	handler := newInteractionService(
		t,
		pgClient,
		slackAPI.URL,
		&stubBindingGate{bindURL: "https://console.example.com/organizations/org/employee/bind?token=test"},
	)
	event := &EventBody{
		Type:        EventTypeAppMention,
		User:        "U456",
		Text:        "hello",
		Channel:     "C789",
		TS:          "111.222",
		ChannelType: ChannelTypeChannel,
	}
	eventID := "E-bind-retry-" + organizationID.String()

	require.Error(t, handler.handleInteraction(t.Context(), eventID, teamID, event))
	require.NoError(t, handler.handleInteraction(t.Context(), eventID, teamID, event))

	clientIDs := slackAPI.clientIDsSnapshot()
	require.Len(t, clientIDs, 2)
	assert.Equal(t, clientIDs[0], clientIDs[1])
	assert.Len(t, clientIDs[0], 36)
	assert.Len(t, slackAPI.successfulPostsSnapshot(), 1)
	assert.Empty(t, slackAPI.statusesSnapshot())
}

func TestService_HandleInteraction_UnboundInDMPostsSingleMessage(t *testing.T) {
	t.Parallel()

	slackAPI := newRecordingSlackAPI(t, nil)
	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	handler := newInteractionService(
		t,
		pgClient,
		slackAPI.URL,
		&stubBindingGate{bindURL: "https://console.example.com/organizations/org/employee/bind?token=test"},
	)

	require.NoError(
		t,
		handler.handleInteraction(
			t.Context(),
			"E-dm-"+organizationID.String(),
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

	posts := slackAPI.postsSnapshot()
	require.Len(t, posts, 1)
	assert.Equal(t, "D789", posts[0]["channel"])
	assert.Equal(t, bindRequiredText, posts[0]["text"])
	assert.Nil(t, posts[0]["thread_ts"])
	assert.Empty(t, slackAPI.statusesSnapshot())
}

func TestService_HandleInteraction_BoundMentionSetsAssistantStatus(t *testing.T) {
	t.Parallel()

	slackAPI := newRecordingSlackAPI(t, nil)
	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	handler := newInteractionService(
		t,
		pgClient,
		slackAPI.URL,
		&stubBindingGate{
			binding: &identitybinding.Binding{
				IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
			},
		},
	)

	require.NoError(
		t,
		handler.handleInteraction(
			t.Context(),
			"E-bound-mention-"+organizationID.String(),
			teamID,
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

	statuses := slackAPI.statusesSnapshot()
	require.Len(t, statuses, 1)
	assert.Equal(t, "C789", statuses[0]["channel_id"])
	assert.Equal(t, "100.000", statuses[0]["thread_ts"])
	assert.Equal(t, assistantWorkingStatus, statuses[0]["status"])
	assert.Equal(
		t,
		[]any{
			"Looking that up…",
			"Checking your workspace…",
			"Drafting a reply…",
		},
		statuses[0]["loading_messages"],
	)
}

func TestService_InteractiveActorBound(t *testing.T) {
	t.Parallel()

	payload := InteractivePayload{}
	payload.Team.ID = "T789"
	payload.User.ID = "U456"

	t.Run(
		"returns false when binding is missing",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{}
			handler := NewService(bindings, nil, nil, log.NewLogger())

			bound, err := handler.InteractiveActorBound(t.Context(), payload)
			require.NoError(t, err)
			assert.False(t, bound)
			assert.Equal(t, IdentitySubject("T789", "U456"), bindings.lookupSubject)
		},
	)

	t.Run(
		"returns true when binding exists",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{
				binding: &identitybinding.Binding{
					IdentityID: gid.New(gid.NilTenant, coredata.IdentityEntityType),
				},
			}
			handler := NewService(bindings, nil, nil, log.NewLogger())

			bound, err := handler.InteractiveActorBound(t.Context(), payload)
			require.NoError(t, err)
			assert.True(t, bound)
		},
	)

	t.Run(
		"prefers the user team id",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{
				binding: &identitybinding.Binding{
					IdentityID: gid.New(gid.NilTenant, coredata.IdentityEntityType),
				},
			}
			handler := NewService(bindings, nil, nil, log.NewLogger())
			actorPayload := payload
			actorPayload.User.TeamID = "T-actor"

			bound, err := handler.InteractiveActorBound(t.Context(), actorPayload)
			require.NoError(t, err)
			assert.True(t, bound)
			assert.Equal(t, IdentitySubject("T-actor", "U456"), bindings.lookupSubject)
		},
	)

	t.Run(
		"skips the gate when bindings are unavailable",
		func(t *testing.T) {
			t.Parallel()

			handler := NewService(nil, nil, nil, log.NewLogger())

			bound, err := handler.InteractiveActorBound(t.Context(), payload)
			require.NoError(t, err)
			assert.True(t, bound)
		},
	)

	t.Run(
		"returns lookup errors",
		func(t *testing.T) {
			t.Parallel()

			bindings := &stubBindingGate{lookupErr: errors.New("lookup failed")}
			handler := NewService(bindings, nil, nil, log.NewLogger())

			bound, err := handler.InteractiveActorBound(t.Context(), payload)
			require.Error(t, err)
			assert.False(t, bound)
			assert.ErrorContains(t, err, "cannot lookup Slack identity binding")
		},
	)
}

func TestService_HandleInteraction_BoundDMSetsAssistantStatus(t *testing.T) {
	t.Parallel()

	slackAPI := newRecordingSlackAPI(t, nil)
	pgClient, organizationID, teamID := executionAdapterDatabase(t)
	handler := newInteractionService(
		t,
		pgClient,
		slackAPI.URL,
		&stubBindingGate{
			binding: &identitybinding.Binding{
				IdentityID: gid.New(organizationID.TenantID(), coredata.IdentityEntityType),
			},
		},
	)

	require.NoError(
		t,
		handler.handleInteraction(
			t.Context(),
			"E-bound-dm-"+organizationID.String(),
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

	statuses := slackAPI.statusesSnapshot()
	require.Len(t, statuses, 1)
	assert.Equal(t, "D789", statuses[0]["channel_id"])
	assert.Equal(t, "333.444", statuses[0]["thread_ts"])
	assert.Equal(t, assistantWorkingStatus, statuses[0]["status"])
}

func newInteractionService(
	t *testing.T,
	pgClient *pg.Client,
	apiBaseURL string,
	bindings identitybinding.Gate,
) *Service {
	t.Helper()

	return &Service{
		installations: newTestInstallationService(t, pgClient, apiBaseURL),
		bindings:      bindings,
		pg:            pgClient,
		logger:        log.NewLogger(),
	}
}

func newRecordingSlackAPI(
	t *testing.T,
	failPost func(http.ResponseWriter, map[string]any) bool,
) *recordingSlackAPI {
	t.Helper()

	api := &recordingSlackAPI{}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var decoded map[string]any
				if len(body) > 0 {
					require.NoError(t, json.Unmarshal(body, &decoded))
				}

				switch r.URL.Path {
				case "/api/chat.postMessage":
					api.mu.Lock()

					api.posts = append(api.posts, decoded)
					if clientID, ok := decoded["client_msg_id"].(string); ok {
						api.clientIDs = append(api.clientIDs, clientID)
					}

					api.mu.Unlock()

					if failPost != nil && failPost(w, decoded) {
						return
					}

					api.mu.Lock()
					api.successfulPosts = append(api.successfulPosts, decoded)
					api.mu.Unlock()

					channel, _ := decoded["channel"].(string)
					_, err := w.Write(
						[]byte(`{"ok":true,"channel":"` + channel + `","ts":"reply-ts"}`),
					)
					require.NoError(t, err)
				case "/api/assistant.threads.setStatus":
					api.mu.Lock()
					api.statuses = append(api.statuses, decoded)
					api.mu.Unlock()

					_, err := w.Write([]byte(`{"ok":true}`))
					require.NoError(t, err)
				case "/api/auth.test":
					_, err := w.Write([]byte(`{"ok":true,"bot_id":"BPROBOT"}`))
					require.NoError(t, err)
				case "/api/conversations.replies":
					_, err := w.Write([]byte(`{"ok":true,"messages":[]}`))
					require.NoError(t, err)
				default:
					t.Errorf("unexpected Slack method %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			},
		),
	)
	t.Cleanup(server.Close)
	api.URL = server.URL + "/api"

	return api
}

func (a *recordingSlackAPI) postsSnapshot() []map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make([]map[string]any, len(a.successfulPosts))
	copy(out, a.successfulPosts)

	return out
}

func (a *recordingSlackAPI) successfulPostsSnapshot() []map[string]any {
	return a.postsSnapshot()
}

func (a *recordingSlackAPI) statusesSnapshot() []map[string]any {
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make([]map[string]any, len(a.statuses))
	copy(out, a.statuses)

	return out
}

func (a *recordingSlackAPI) clientIDsSnapshot() []string {
	a.mu.Lock()
	defer a.mu.Unlock()

	out := make([]string, len(a.clientIDs))
	copy(out, a.clientIDs)

	return out
}
