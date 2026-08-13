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

package probot

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/bot"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

var messageWorkerTestMu sync.Mutex

type (
	outboundTestCapability struct {
		name        string
		messageType string
	}

	recordingDelivery struct {
		deliveries []OutboundDelivery
		err        error
	}
)

func (c outboundTestCapability) Name() string {
	return c.name
}

func (c outboundTestCapability) MessageTypes() []string {
	return []string{c.messageType}
}

func (c outboundTestCapability) RenderMessage(
	context.Context,
	Message,
) (MessageIntent, error) {
	return MessageIntent{FallbackText: c.name}, nil
}

func (c outboundTestCapability) BuildOutboundMessage(
	_ context.Context,
	organizationID gid.GID,
	messageType string,
	_ map[string]any,
) (OutboundMessage, error) {
	return OutboundMessage{
		Message: Message{
			OrganizationID: organizationID,
			Type:           messageType,
		},
		Intent: MessageIntent{FallbackText: "queued"},
		DeliveryTarget: DeliveryTarget{
			Namespace: "test-ns",
			Key:       "test-key",
		},
	}, nil
}

func (d *recordingDelivery) DeliverOutbound(
	_ context.Context,
	delivery OutboundDelivery,
) error {
	d.deliveries = append(d.deliveries, delivery)

	return d.err
}

func TestMessageWorkerHandler_ProcessMarksMessageProcessed(t *testing.T) {
	t.Parallel()

	fixture := newMessageWorkerFixture(t)
	message := fixture.enqueue(t, fixture.capability.Name(), fixture.capability.messageType)
	now := time.Now().UTC()
	handler := fixture.handler(now)

	claimed, err := handler.Claim(t.Context())
	require.NoError(t, err)
	assert.Equal(t, message.ID, claimed.ID)
	require.NoError(t, handler.Process(t.Context(), claimed))

	loaded := fixture.load(t, message.ID)
	require.NotNil(t, loaded.ProcessedAt)
	assert.Nil(t, loaded.DeadLetteredAt)
	assert.Nil(t, loaded.NextAttemptAt)
	require.Len(t, fixture.delivery.deliveries, 1)
	assert.Equal(t, message.ID.String(), fixture.delivery.deliveries[0].SourceEventID)
	assert.Equal(t, fixture.capability.Name(), fixture.delivery.deliveries[0].Capability)
}

func TestMessageWorkerHandler_TransientFailureSchedulesRetry(t *testing.T) {
	t.Parallel()

	fixture := newMessageWorkerFixture(t)
	fixture.delivery.err = errors.New("slack unavailable")
	message := fixture.enqueue(t, fixture.capability.Name(), fixture.capability.messageType)
	now := time.Now().UTC()
	handler := fixture.handler(now)

	claimed, err := handler.Claim(t.Context())
	require.NoError(t, err)
	require.Error(t, handler.Process(t.Context(), claimed))

	failed := fixture.load(t, message.ID)
	assert.Nil(t, failed.ProcessedAt)
	assert.Nil(t, failed.DeadLetteredAt)
	require.NotNil(t, failed.LastError)
	require.NotNil(t, failed.NextAttemptAt)
	assert.WithinDuration(t, now.Add(time.Second), *failed.NextAttemptAt, time.Microsecond)

	_, err = handler.Claim(t.Context())
	require.ErrorIs(t, err, worker.ErrNoTask)

	now = now.Add(time.Second)
	handler.now = func() time.Time { return now }
	claimed, err = handler.Claim(t.Context())
	require.NoError(t, err)
	assert.Equal(t, message.ID, claimed.ID)
	assert.Equal(t, 2, claimed.AttemptCount)
}

func TestMessageWorkerHandler_UnknownCapabilityDeadLetters(t *testing.T) {
	t.Parallel()

	fixture := newMessageWorkerFixture(t)
	message := fixture.enqueue(t, "missing-capability", fixture.capability.messageType)
	now := time.Now().UTC()
	handler := fixture.handler(now)

	claimed, err := handler.Claim(t.Context())
	require.NoError(t, err)
	require.Error(t, handler.Process(t.Context(), claimed))

	failed := fixture.load(t, message.ID)
	assert.Nil(t, failed.ProcessedAt)
	require.NotNil(t, failed.DeadLetteredAt)
	assert.Nil(t, failed.NextAttemptAt)
	assert.Empty(t, fixture.delivery.deliveries)
}

type messageWorkerFixture struct {
	pg           *pg.Client
	scope        coredata.Scoper
	organization coredata.Organization
	capability   outboundTestCapability
	delivery     *recordingDelivery
}

func newMessageWorkerFixture(t *testing.T) messageWorkerFixture {
	t.Helper()

	messageWorkerTestMu.Lock()
	t.Cleanup(messageWorkerTestMu.Unlock)

	ctx := t.Context()
	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "probot-message-worker-" + tenantID.String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Insert(ctx, tx)
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					return organization.Delete(ctx, tx, organization.ID)
				},
			)
		},
	)

	return messageWorkerFixture{
		pg:           pgClient,
		scope:        scope,
		organization: organization,
		capability: outboundTestCapability{
			name:        "test-capability",
			messageType: "TEST_MESSAGE",
		},
		delivery: &recordingDelivery{},
	}
}

func (f messageWorkerFixture) handler(now time.Time) *messageWorkerHandler {
	registry := NewCapabilityRegistry()
	if err := registry.Register(f.capability); err != nil {
		panic(err)
	}

	return &messageWorkerHandler{
		pg:           f.pg,
		capabilities: registry,
		deliveries:   f.delivery,
		logger:       log.NewLogger(log.WithName("test")),
		staleAfter:   5 * time.Minute,
		retryBase:    time.Second,
		retryMax:     time.Minute,
		now:          func() time.Time { return now },
	}
}

func (f messageWorkerFixture) enqueue(
	t *testing.T,
	capability string,
	messageType string,
) *coredata.BotMessage {
	t.Helper()

	var message *coredata.BotMessage

	require.NoError(
		t,
		f.pg.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				var err error

				message, err = bot.NewService(bot.ServiceConfig{}).EnqueueMessage(
					ctx,
					tx,
					f.scope,
					bot.MessageParams{
						OrganizationID:   f.organization.ID,
						Capability:       capability,
						MessageType:      messageType,
						SubjectNamespace: "test-subject",
						SubjectKey:       gid.New(f.scope.GetTenantID(), coredata.BotMessageEntityType).String(),
						EventKey:         gid.New(f.scope.GetTenantID(), coredata.BotMessageEntityType).String(),
						Purpose:          coredata.BotMessagePurposePost,
					},
				)

				return err
			},
		),
	)
	require.NotNil(t, message)

	return message
}

func (f messageWorkerFixture) load(
	t *testing.T,
	messageID gid.GID,
) coredata.BotMessage {
	t.Helper()

	var message coredata.BotMessage

	require.NoError(
		t,
		f.pg.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return message.LoadByID(ctx, conn, f.scope, messageID)
			},
		),
	)

	return message
}
