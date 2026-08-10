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

package slackbot

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	stubBindingGate struct {
		binding *coredata.SlackIdentityBinding
		bindURL string
	}

	recordingSlackClient struct {
		mu sync.Mutex

		openIMUser string
		openIMChan string
		posts      []recordedPost
	}

	recordedPost struct {
		channel  string
		text     string
		threadTS string
		blocks   []any
	}
)

func (s *stubBindingGate) Lookup(
	_ context.Context,
	_, _ string,
) (*coredata.SlackIdentityBinding, error) {
	if s.binding == nil {
		return nil, coredata.ErrResourceNotFound
	}

	return s.binding, nil
}

func (s *stubBindingGate) BindURL(_, _ string) (string, error) {
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

func (c *recordingSlackClient) PostMessage(
	ctx context.Context,
	channel, text, threadTS string,
) error {
	return c.PostMessageWithBlocks(ctx, channel, text, threadTS, nil)
}

func (c *recordingSlackClient) PostMessageWithBlocks(
	_ context.Context,
	channel, text, threadTS string,
	blocks []any,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.posts = append(
		c.posts,
		recordedPost{
			channel:  channel,
			text:     text,
			threadTS: threadTS,
			blocks:   blocks,
		},
	)

	return nil
}

func (c *recordingSlackClient) SetStatus(context.Context, string, string, string) error {
	return nil
}

func (c *recordingSlackClient) AddReaction(context.Context, string, string, string) error {
	return nil
}

func (c *recordingSlackClient) postsSnapshot() []recordedPost {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := make([]recordedPost, len(c.posts))
	copy(out, c.posts)

	return out
}

func countSlackbotAgents(t *testing.T, ctx context.Context, client *pg.Client, sessionPrefix string) int {
	t.Helper()

	var count int

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var err error
		count, err = coredata.CountSlackbotAgentsBySessionIDPrefix(ctx, conn, sessionPrefix)

		return err
	}))

	return count
}

func cleanupSlackbotAgents(t *testing.T, ctx context.Context, client *pg.Client, sessionPrefix string) {
	t.Helper()

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var agents coredata.SlackbotAgents
		if err := agents.LoadBySessionIDPrefix(ctx, conn, sessionPrefix, 100); err != nil {
			return err
		}

		agentIDs := make([]string, 0, len(agents))
		for _, agent := range agents {
			agentIDs = append(agentIDs, agent.AgentID)
		}

		if err := coredata.DeleteSlackbotInteractionsByAgentIDs(ctx, conn, agentIDs); err != nil {
			return err
		}

		for _, agent := range agents {
			if err := agent.Delete(ctx, conn); err != nil {
				return err
			}
		}

		return nil
	}))
}

func TestHandler_HandleInteraction_UnboundPostsCTA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	slackClient := &recordingSlackClient{}
	eventID := "E-unbound-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()
	teamID := "T-unbound-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()

	handler := &Handler{
		client:     slackClient,
		bindings:   &stubBindingGate{bindURL: "https://console.example.com/me/slack/bind?token=test"},
		pg:         client,
		logger:     log.NewLogger(),
		serviceCtx: ctx,
	}

	handler.handleInteraction(
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
	)

	posts := slackClient.postsSnapshot()
	require.Len(t, posts, 2)
	assert.Equal(t, "C789", posts[0].channel)
	assert.Equal(t, bindRequiredPublicText, posts[0].text)
	assert.Nil(t, posts[0].blocks)
	assert.Equal(t, "111.222", posts[0].threadTS)
	assert.Equal(t, "D-test-dm", posts[1].channel)
	assert.Equal(t, bindRequiredDMText, posts[1].text)
	assert.NotEmpty(t, posts[1].blocks)
	assert.Equal(t, "U456", slackClient.openIMUser)

	assert.Equal(t, 0, countSlackbotAgents(t, ctx, client, teamID+":"))

	handler.handleInteraction(
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
	)

	assert.Len(t, slackClient.postsSnapshot(), 2)
}

func TestHandler_HandleInteraction_UnboundInDMPostsSingleMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	slackClient := &recordingSlackClient{}
	eventID := "E-dm-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()
	teamID := "T-dm-" + gid.New(gid.NilTenant, coredata.IdentityEntityType).String()

	handler := &Handler{
		client:     slackClient,
		bindings:   &stubBindingGate{bindURL: "https://console.example.com/me/slack/bind?token=test"},
		pg:         client,
		logger:     log.NewLogger(),
		serviceCtx: ctx,
	}

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
	)

	posts := slackClient.postsSnapshot()
	require.Len(t, posts, 1)
	assert.Equal(t, "D789", posts[0].channel)
	assert.Equal(t, bindRequiredDMText, posts[0].text)
	assert.NotEmpty(t, posts[0].blocks)
	assert.Empty(t, slackClient.openIMUser)
}

func TestHandler_HandleInteraction_BoundCreatesAgent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := test.PGClient(t)
	slackClient := &recordingSlackClient{}
	identityID := gid.New(gid.NilTenant, coredata.IdentityEntityType)
	eventID := "E-bound-" + identityID.String()
	teamID := "T-bound-" + identityID.String()
	channel := "C-bound-" + identityID.String()

	handler := &Handler{
		client: slackClient,
		bindings: &stubBindingGate{
			binding: &coredata.SlackIdentityBinding{
				IdentityID:  identityID,
				TeamID:      teamID,
				SlackUserID: "U456",
			},
		},
		pg:         client,
		logger:     log.NewLogger(),
		serviceCtx: ctx,
	}

	handler.handleInteraction(
		ctx,
		eventID,
		teamID,
		&EventBody{
			Type:        EventTypeAppMention,
			User:        "U456",
			Text:        "hello",
			Channel:     channel,
			TS:          "222.333",
			ChannelType: ChannelTypeChannel,
		},
	)

	sessionPrefix := teamID + ":" + channel + ":"
	assert.Equal(t, 0, len(slackClient.postsSnapshot()))
	assert.Equal(t, 1, countSlackbotAgents(t, ctx, client, sessionPrefix))

	t.Cleanup(func() {
		cleanupSlackbotAgents(t, ctx, client, sessionPrefix)
	})
}
