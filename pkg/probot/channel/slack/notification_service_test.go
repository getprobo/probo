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

package slack_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
)

type notificationFixture struct {
	client         *pg.Client
	scope          *coredata.Scope
	organizationID gid.GID
	service        *slackchannel.NotificationService
}

func newNotificationFixture(t *testing.T) notificationFixture {
	t.Helper()

	ctx := context.Background()
	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)

	err := client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var message coredata.SlackbotMessage

			return message.LoadByID(
				ctx,
				conn,
				scope,
				gid.New(tenantID, coredata.SlackbotMessageEntityType),
			)
		},
	)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		t.Skipf("slackbot_messages is unavailable in the test database: %v", err)
	}

	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()
	organization := &coredata.Organization{
		ID:        organizationID,
		TenantID:  tenantID,
		Name:      "Slackbot Notification Service Test",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(
		t,
		client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Insert(ctx, tx)
			},
		),
	)

	t.Cleanup(
		func() {
			_ = client.WithTx(
				context.Background(),
				func(ctx context.Context, tx pg.Tx) error {
					return organization.Delete(ctx, tx, organizationID)
				},
			)
		},
	)

	return notificationFixture{
		client:         client,
		scope:          scope,
		organizationID: organizationID,
		service:        slackchannel.NewNotificationService(client),
	}
}

func deliverNotification(
	t *testing.T,
	ctx context.Context,
	fixture notificationFixture,
	message *coredata.SlackbotMessage,
	channelID string,
	messageTS string,
) {
	t.Helper()

	now := time.Now()
	message.ChannelID = new(channelID)
	message.MessageTS = new(messageTS)
	message.SentAt = &now
	message.UpdatedAt = now

	require.NoError(
		t,
		fixture.client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := message.UpdateDeliveryState(ctx, tx, fixture.scope); err != nil {
					return err
				}

				return message.PropagateDeliveryReferenceToRevisions(ctx, tx, fixture.scope)
			},
		),
	)
}

func TestNotificationService_QueueRejectsIncompleteDedup(t *testing.T) {
	t.Parallel()

	dedupKey := "dedup-key"
	service := slackchannel.NewNotificationService(nil)
	_, err := service.Queue(
		context.Background(),
		coredata.NewScope(gid.NewTenantID()),
		slackchannel.QueueNotificationRequest{
			DedupKey: &dedupKey,
		},
	)

	require.EqualError(t, err, "dedup key and window must be provided together")
}

func TestNotificationService_QueueAndLookup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newNotificationFixture(t)
	dedupKey := "generic-resource-123"
	dedupWindow := time.Hour
	metadata := map[string]any{"opaque": "value"}

	initial, err := fixture.service.Queue(
		ctx,
		fixture.scope,
		slackchannel.QueueNotificationRequest{
			OrganizationID: fixture.organizationID,
			ChannelID:      "C123",
			MessageType:    "GENERIC_NOTIFICATION",
			Body:           map[string]any{"text": "first"},
			Metadata:       metadata,
			DedupKey:       &dedupKey,
			DedupWindow:    &dedupWindow,
		},
	)
	require.NoError(t, err)
	assert.Equal(t, initial.ID, initial.InitialSlackbotMessageID)
	assert.NotContains(t, metadata, coredata.SlackbotMessageMetadataDedupKey)

	deduplicated, err := fixture.service.Queue(
		ctx,
		fixture.scope,
		slackchannel.QueueNotificationRequest{
			OrganizationID: fixture.organizationID,
			ChannelID:      "C123",
			MessageType:    "GENERIC_NOTIFICATION",
			Body:           map[string]any{"text": "second"},
			Metadata:       map[string]any{"opaque": "updated"},
			DedupKey:       &dedupKey,
			DedupWindow:    &dedupWindow,
		},
	)
	require.NoError(t, err)
	assert.NotEqual(t, initial.ID, deduplicated.ID)
	assert.Equal(t, initial.ID, deduplicated.InitialSlackbotMessageID)

	revision, err := fixture.service.QueueRevision(
		ctx,
		fixture.scope,
		deduplicated.ID,
		map[string]any{"text": "third"},
		map[string]any{"revision": true},
	)
	require.NoError(t, err)
	assert.Equal(t, initial.ID, revision.InitialSlackbotMessageID)
	assert.Equal(t, initial.MessageType, revision.MessageType)
	assert.Equal(t, initial.OrganizationID, revision.OrganizationID)

	loaded, err := fixture.service.GetByID(ctx, fixture.scope, revision.ID)
	require.NoError(t, err)
	assert.Equal(t, revision.ID, loaded.ID)
	assert.Equal(t, map[string]any{"text": "third"}, loaded.Body)
}

func TestNotificationService_OrganizationLookupIsolatesChannelTimestampCollision(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := newNotificationFixture(t)
	second := newNotificationFixture(t)

	const (
		channelID = "C-COLLISION"
		messageTS = "1234567890.999999"
	)

	firstInitial, err := first.service.Queue(
		ctx,
		first.scope,
		slackchannel.QueueNotificationRequest{
			OrganizationID: first.organizationID,
			ChannelID:      channelID,
			MessageType:    "GENERIC_NOTIFICATION",
			Body:           map[string]any{"text": "first initial"},
			Metadata:       map[string]any{},
		},
	)
	require.NoError(t, err)
	deliverNotification(t, ctx, first, firstInitial, channelID, messageTS)

	firstRevision, err := first.service.QueueRevision(
		ctx,
		first.scope,
		firstInitial.ID,
		map[string]any{"text": "first revision"},
		map[string]any{},
	)
	require.NoError(t, err)
	deliverNotification(t, ctx, first, firstRevision, channelID, messageTS)

	secondMessage, err := second.service.Queue(
		ctx,
		second.scope,
		slackchannel.QueueNotificationRequest{
			OrganizationID: second.organizationID,
			ChannelID:      channelID,
			MessageType:    "GENERIC_NOTIFICATION",
			Body:           map[string]any{"text": "second tenant"},
			Metadata:       map[string]any{},
		},
	)
	require.NoError(t, err)
	deliverNotification(t, ctx, second, secondMessage, channelID, messageTS)

	loaded, err := first.service.GetInitialByOrganizationIDChannelAndTS(
		ctx,
		first.scope,
		first.organizationID,
		channelID,
		messageTS,
	)
	require.NoError(t, err)
	assert.Equal(t, firstRevision.ID, loaded.ID)
	assert.Equal(t, first.organizationID, loaded.OrganizationID)
}

func TestNotificationService_QueueWithoutDedup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newNotificationFixture(t)
	req := slackchannel.QueueNotificationRequest{
		OrganizationID: fixture.organizationID,
		ChannelID:      "C456",
		MessageType:    "GENERIC_NOTIFICATION",
		Body:           map[string]any{"text": "body"},
		Metadata:       map[string]any{},
	}

	first, err := fixture.service.Queue(ctx, fixture.scope, req)
	require.NoError(t, err)
	second, err := fixture.service.Queue(ctx, fixture.scope, req)
	require.NoError(t, err)

	assert.Equal(t, first.ID, first.InitialSlackbotMessageID)
	assert.Equal(t, second.ID, second.InitialSlackbotMessageID)
	assert.NotEqual(t, first.ID, second.ID)
}

func TestNotificationService_QueueIsIdempotentBySourceEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newNotificationFixture(t)
	sourceEventID := gid.New(
		fixture.scope.GetTenantID(),
		coredata.AgentExecutionEntityType,
	).String()
	req := slackchannel.QueueNotificationRequest{
		OrganizationID: fixture.organizationID,
		ChannelID:      "C-source-event",
		MessageType:    "GENERIC_NOTIFICATION",
		Body:           map[string]any{"text": "first"},
		Metadata:       map[string]any{},
		SourceEventID:  &sourceEventID,
	}

	first, err := fixture.service.Queue(ctx, fixture.scope, req)
	require.NoError(t, err)

	req.Body = map[string]any{"text": "retry"}
	second, err := fixture.service.Queue(ctx, fixture.scope, req)
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "first", second.Body["text"])
}

func TestNotificationService_ClaimInteractiveAction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newNotificationFixture(t)

	processingToken, claimed, err := fixture.service.ClaimInteractiveAction(
		ctx,
		fixture.scope,
		fixture.organizationID,
		"interaction-123",
	)
	require.NoError(t, err)
	assert.True(t, claimed)
	require.NoError(
		t,
		fixture.service.CompleteInteractiveAction(
			ctx,
			fixture.scope,
			fixture.organizationID,
			"interaction-123",
			processingToken,
		),
	)

	_, claimed, err = fixture.service.ClaimInteractiveAction(
		ctx,
		fixture.scope,
		fixture.organizationID,
		"interaction-123",
	)
	require.NoError(t, err)
	assert.False(t, claimed)
}

func TestNotificationService_ReleaseInteractiveActionAllowsRetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newNotificationFixture(t)
	processingToken, claimed, err := fixture.service.ClaimInteractiveAction(
		ctx,
		fixture.scope,
		fixture.organizationID,
		"interaction-retry",
	)
	require.NoError(t, err)
	assert.True(t, claimed)
	require.NoError(
		t,
		fixture.service.ReleaseInteractiveAction(
			ctx,
			fixture.scope,
			fixture.organizationID,
			"interaction-retry",
			processingToken,
		),
	)

	_, claimed, err = fixture.service.ClaimInteractiveAction(
		ctx,
		fixture.scope,
		fixture.organizationID,
		"interaction-retry",
	)
	require.NoError(t, err)
	assert.True(t, claimed)
}

func TestNotificationService_StaleInteractiveActionCanBeReclaimed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newNotificationFixture(t)
	claim := coredata.NewSlackbotInteractiveClaim(
		fixture.organizationID,
		"interaction-stale",
	)
	require.NoError(
		t,
		fixture.client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				claimed, err := claim.Claim(
					ctx,
					tx,
					fixture.scope,
					"00000000-0000-7000-8000-000000000008",
					time.Now().Add(-10*time.Minute),
					5*time.Minute,
				)
				if err != nil {
					return err
				}

				if !claimed {
					return errors.New("cannot create stale interactive claim")
				}

				return nil
			},
		),
	)

	_, claimed, err := fixture.service.ClaimInteractiveAction(
		ctx,
		fixture.scope,
		fixture.organizationID,
		"interaction-stale",
	)
	require.NoError(t, err)
	assert.True(t, claimed)
}

func TestNotificationService_QueueReloadsOnSourceEventRace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newNotificationFixture(t)
	sourceEventID := gid.New(
		fixture.scope.GetTenantID(),
		coredata.AgentExecutionEntityType,
	).String()

	var (
		wg            sync.WaitGroup
		first, second *coredata.SlackbotMessage
		err1, err2    error
	)

	queue := func(message **coredata.SlackbotMessage, queueErr *error) {
		defer wg.Done()

		queued, err := fixture.service.Queue(
			ctx,
			fixture.scope,
			slackchannel.QueueNotificationRequest{
				OrganizationID: fixture.organizationID,
				ChannelID:      "C-RACE",
				MessageType:    "GENERIC_NOTIFICATION",
				Body:           map[string]any{"text": "race"},
				SourceEventID:  &sourceEventID,
			},
		)
		*message = queued
		*queueErr = err
	}

	wg.Add(2)

	go queue(&first, &err1)
	go queue(&second, &err2)

	wg.Wait()

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, first)
	require.NotNil(t, second)
	assert.Equal(t, first.ID, second.ID)
}
