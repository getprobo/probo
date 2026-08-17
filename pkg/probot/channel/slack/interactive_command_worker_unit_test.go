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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probot"
	"go.probo.inc/probo/pkg/probot/identitybinding"
)

type (
	fakeInteractiveInstallationResolver struct {
		installation *coredata.SlackbotInstallation
		err          error
	}

	fakeInteractiveBindingGate struct {
		binding *identitybinding.Binding
		err     error
	}

	fakeInteractiveMessageResolver struct {
		message *bot.DeliveredMessage
		err     error
	}

	recordingReplyPoster struct {
		url  string
		text string
	}

	forbiddenActionCapability struct{}
)

func (f fakeInteractiveInstallationResolver) ClientByTeamID(
	context.Context,
	string,
) (*Client, *coredata.SlackbotInstallation, error) {
	return nil, f.installation, f.err
}

func (f fakeInteractiveBindingGate) Lookup(
	context.Context,
	identitybinding.Subject,
) (*identitybinding.Binding, error) {
	return f.binding, f.err
}

func (fakeInteractiveBindingGate) BindURL(
	context.Context,
	identitybinding.Subject,
	gid.GID,
) (string, error) {
	return "", nil
}

func (f fakeInteractiveMessageResolver) GetInitialByChannelAndTS(
	context.Context,
	gid.GID,
	string,
	string,
) (*bot.DeliveredMessage, error) {
	return f.message, f.err
}

func (r *recordingReplyPoster) PostEphemeralReply(
	_ context.Context,
	responseURL string,
	text string,
) error {
	r.url = responseURL
	r.text = text

	return nil
}

func (forbiddenActionCapability) Name() string {
	return "forbidden"
}

func (forbiddenActionCapability) ActionPrefixes() []string {
	return []string{"test."}
}

func (forbiddenActionCapability) HandleAction(
	context.Context,
	probot.Action,
) (probot.ActionResult, error) {
	return probot.ActionResult{}, probot.ErrCapabilityForbidden
}

func TestInteractiveCommandOutcomeRetriesThenDeadLetters(t *testing.T) {
	t.Parallel()

	now := time.Now()
	command := coredata.SlackbotInteractiveCommand{
		AttemptCount: 1,
		MaxAttempts:  2,
	}
	applyInteractiveCommandOutcome(&command, errors.New("temporary"), now, time.Minute)
	require.NotNil(t, command.NextAttemptAt)
	assert.Equal(t, now.Add(time.Minute), *command.NextAttemptAt)
	assert.Nil(t, command.DeadLetteredAt)

	command.AttemptCount = 2
	applyInteractiveCommandOutcome(&command, errors.New("temporary"), now, time.Minute)
	assert.Nil(t, command.NextAttemptAt)
	assert.Equal(t, &now, command.DeadLetteredAt)
}

func TestInteractiveCommandDeadLettersRevokedBinding(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	command, key := encryptedInteractiveCommand(t)
	replies := &recordingReplyPoster{}
	h := &interactiveCommandHandler{
		encryptionKey: key,
		installations: fakeInteractiveInstallationResolver{
			installation: &coredata.SlackbotInstallation{
				OrganizationID: organizationID,
			},
		},
		bindings:     fakeInteractiveBindingGate{err: coredata.ErrResourceNotFound},
		messages:     fakeInteractiveMessageResolver{},
		capabilities: probot.NewCapabilityRegistry(),
		replies:      replies,
		logger:       log.NewLogger(),
	}

	err := h.dispatch(t.Context(), &command)
	require.Error(t, err)
	assert.True(t, isPermanent(err))
	assert.Equal(t, &organizationID, command.OrganizationID)
	assert.Equal(t, "https://hooks.slack.com/actions/T123/1/abc", replies.url)
	assert.Equal(t, bindRequiredText, replies.text)
}

func TestInteractiveCommandDeadLettersAuthorizationError(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	command, key := encryptedInteractiveCommand(t)
	registry := probot.NewCapabilityRegistry()
	require.NoError(t, registry.Register(forbiddenActionCapability{}))

	replies := &recordingReplyPoster{}
	h := &interactiveCommandHandler{
		encryptionKey: key,
		installations: fakeInteractiveInstallationResolver{
			installation: &coredata.SlackbotInstallation{
				OrganizationID: organizationID,
			},
		},
		bindings: fakeInteractiveBindingGate{
			binding: &identitybinding.Binding{
				IdentityID: gid.New(gid.NilTenant, coredata.IdentityEntityType),
			},
		},
		messages: fakeInteractiveMessageResolver{
			message: &bot.DeliveredMessage{
				Message: bot.Message{
					ID:             gid.New(tenantID, coredata.SlackbotMessageEntityType),
					OrganizationID: organizationID,
					Type:           "test-message",
				},
			},
		},
		capabilities: registry,
		replies:      replies,
		logger:       log.NewLogger(),
	}

	err := h.dispatch(t.Context(), &command)
	require.Error(t, err)
	assert.True(t, isPermanent(err))
	assert.ErrorIs(t, err, probot.ErrCapabilityForbidden)
	assert.Equal(t, "https://hooks.slack.com/actions/T123/1/abc", replies.url)
	assert.Equal(t, interactiveForbiddenText, replies.text)
}

func encryptedInteractiveCommand(
	t *testing.T,
) (coredata.SlackbotInteractiveCommand, cipher.EncryptionKey) {
	t.Helper()

	key := cipher.EncryptionKey{1, 2, 3}
	raw := []byte(`{"team":{"id":"T123"},"user":{"id":"U123"},"response_url":"https://hooks.slack.com/actions/T123/1/abc","container":{"channel_id":"C123","message_ts":"123.456"},"actions":[{"action_id":"test.approve","action_ts":"123.789","value":"resource-id"}]}`)
	encrypted, err := cipher.Encrypt(raw, key)
	require.NoError(t, err)

	return coredata.SlackbotInteractiveCommand{
		EncryptedPayload: encrypted,
		RequestDigest:    []byte("digest"),
	}, key
}
