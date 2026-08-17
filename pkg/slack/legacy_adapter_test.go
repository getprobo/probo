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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slack"
)

type legacyLookupFixture struct {
	client         *pg.Client
	scope          *coredata.Scope
	organizationID gid.GID
}

func newLegacyLookupFixture(t *testing.T) legacyLookupFixture {
	t.Helper()

	ctx := context.Background()
	client := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	now := time.Now()
	organization := &coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Legacy Slack Lookup Test",
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
					return organization.Delete(ctx, tx, organization.ID)
				},
			)
		},
	)

	return legacyLookupFixture{
		client:         client,
		scope:          scope,
		organizationID: organization.ID,
	}
}

func insertDeliveredLegacyMessage(
	t *testing.T,
	ctx context.Context,
	fixture legacyLookupFixture,
	initialMessageID gid.GID,
	channelID string,
	messageTS string,
	body string,
) *coredata.SlackMessage {
	t.Helper()

	message := coredata.NewSlackMessage(
		fixture.scope,
		fixture.organizationID,
		coredata.SlackMessageTypeWelcome,
		map[string]any{"text": body},
	)
	if initialMessageID != gid.Nil {
		message.InitialSlackMessageID = initialMessageID
	}

	message.ChannelID = new(channelID)
	message.MessageTS = new(messageTS)

	require.NoError(
		t,
		fixture.client.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				if err := message.Insert(ctx, tx, fixture.scope); err != nil {
					return err
				}

				now := time.Now()
				message.SentAt = &now
				message.UpdatedAt = now

				return message.Update(ctx, tx, fixture.scope)
			},
		),
	)

	return message
}

func TestLegacyLookup_OrganizationIsolatesChannelTimestampCollision(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	first := newLegacyLookupFixture(t)
	second := newLegacyLookupFixture(t)

	const (
		channelID = "C-LEGACY-COLLISION"
		messageTS = "1234567890.888888"
	)

	firstInitial := insertDeliveredLegacyMessage(
		t,
		ctx,
		first,
		gid.Nil,
		channelID,
		messageTS,
		"first initial",
	)
	firstRevision := insertDeliveredLegacyMessage(
		t,
		ctx,
		first,
		firstInitial.ID,
		channelID,
		messageTS,
		"first revision",
	)
	_ = insertDeliveredLegacyMessage(
		t,
		ctx,
		second,
		gid.Nil,
		channelID,
		messageTS,
		"second tenant",
	)

	service := slack.NewService(first.client, "", "", log.NewLogger())
	loaded, err := service.GetInitialByOrganizationIDChannelAndTS(
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

func TestService_QueueIsIdempotentBySourceEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newLegacyLookupFixture(t)
	service := slack.NewService(fixture.client, "", "", log.NewLogger())
	sourceEventID := gid.New(
		fixture.scope.GetTenantID(),
		coredata.AgentExecutionEntityType,
	).String()
	req := slack.QueueRequest{
		OrganizationID: fixture.organizationID,
		MessageType:    coredata.SlackMessageTypeWelcome,
		Body:           map[string]any{"text": "first"},
		Metadata:       map[string]any{},
		SourceEventID:  &sourceEventID,
	}

	first, err := service.Queue(ctx, fixture.scope, req)
	require.NoError(t, err)

	req.Body = map[string]any{"text": "retry"}
	second, err := service.Queue(ctx, fixture.scope, req)
	require.NoError(t, err)

	assert.Equal(t, first.ID, second.ID)
	assert.Equal(t, "first", second.Body["text"])
}

func TestService_QueueReloadsOnSourceEventRace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixture := newLegacyLookupFixture(t)
	service := slack.NewService(fixture.client, "", "", log.NewLogger())
	sourceEventID := gid.New(
		fixture.scope.GetTenantID(),
		coredata.BotMessageEntityType,
	).String()

	var (
		wg            sync.WaitGroup
		first, second *coredata.SlackMessage
		err1, err2    error
	)

	queue := func(message **coredata.SlackMessage, queueErr *error) {
		defer wg.Done()

		queued, err := service.Queue(
			ctx,
			fixture.scope,
			slack.QueueRequest{
				OrganizationID: fixture.organizationID,
				MessageType:    coredata.SlackMessageTypeWelcome,
				Body:           map[string]any{"text": "race"},
				Metadata:       map[string]any{},
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
