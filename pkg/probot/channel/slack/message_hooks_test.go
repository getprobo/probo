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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestMessageService_OnSlackbotDeliverySuccessBindsThreadAndAnchor(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := newQueueTestOrganization(t)
	service := NewMessageService(pgClient, nil, nil, nil, log.NewLogger())
	target := coredata.NewBotDeliveryDestination(
		scope,
		organizationID,
		ProviderName,
		"compliance_portal",
		organizationID.String(),
	)
	target.ExternalDestinationID = "C-hook"
	channelID := "C-hook"
	messageTS := "123.456"
	source := "provider"
	sessionKey := "hook-session-" + organizationID.String()
	now := time.Now().UTC()
	execution := coredata.AgentExecution{
		ID:                gid.New(scope.GetTenantID(), coredata.AgentExecutionEntityType),
		OrganizationID:    organizationID,
		StartAgentName:    "assistant",
		Source:            &source,
		SessionKey:        &sessionKey,
		SourceCoordinates: []byte(`{"workspace":"opaque"}`),
		SessionMessages:   []byte(`[]`),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if _, err := target.Upsert(ctx, tx, scope); err != nil {
					return err
				}

				_, err := execution.UpsertBySourceSession(ctx, tx, scope)

				return err
			},
		),
	)

	message := &coredata.SlackbotMessage{
		ID:             gid.New(scope.GetTenantID(), coredata.SlackbotMessageEntityType),
		OrganizationID: organizationID,
		MessageType:    "ACCESS_REQUEST",
		ChannelID:      &channelID,
		MessageTS:      &messageTS,
		Metadata: map[string]any{
			deliverySourceEventMetadata:      gid.New(scope.GetTenantID(), coredata.BotMessageEntityType).String(),
			deliveryAgentExecutionMetadata:   execution.ID.String(),
			deliveryTargetNamespaceMetadata:  target.TargetNamespace,
			deliveryTargetKeyMetadata:        target.TargetKey,
			deliverySubjectNamespaceMetadata: "compliance_portal_access",
			deliverySubjectKeyMetadata:       "access-1",
			deliveryCapabilityMetadata:       "compliance_access",
			deliveryMessageTypeMetadata:      "ACCESS_REQUEST",
		},
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return service.OnSlackbotDeliverySuccess(ctx, tx, scope, message)
			},
		),
	)

	var subject coredata.BotThreadSubject

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return subject.LoadBySubject(
					ctx,
					conn,
					scope,
					organizationID,
					"compliance_portal_access",
					"access-1",
				)
			},
		),
	)
	assert.Equal(t, channelID, subject.ExternalConversationID)
	assert.Equal(t, messageTS, subject.ExternalMessageID)
	assert.Equal(t, "compliance_access", subject.Capability)

	var anchor coredata.AgentExecutionAnchor

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return anchor.LoadByProviderCoordinates(
					ctx,
					conn,
					scope,
					organizationID,
					ProviderName,
					channelID,
					messageTS,
				)
			},
		),
	)
	assert.Equal(t, execution.ID, anchor.AgentExecutionID)

	var destination coredata.BotDeliveryDestination

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return destination.LoadByTarget(
					ctx,
					conn,
					scope,
					organizationID,
					ProviderName,
					target.TargetNamespace,
					target.TargetKey,
				)
			},
		),
	)
	require.NotNil(t, destination.VerifiedAt)
}

func TestMessageService_OnSlackbotDeliverySuccessMissingDestinationIsNonFatal(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := newQueueTestOrganization(t)
	service := NewMessageService(pgClient, nil, nil, nil, log.NewLogger())
	channelID := "C-missing-dest"
	messageTS := "123.789"
	message := &coredata.SlackbotMessage{
		ID:             gid.New(scope.GetTenantID(), coredata.SlackbotMessageEntityType),
		OrganizationID: organizationID,
		MessageType:    "ACCESS_REQUEST",
		ChannelID:      &channelID,
		MessageTS:      &messageTS,
		Metadata: map[string]any{
			deliveryTargetNamespaceMetadata: "compliance_portal",
			deliveryTargetKeyMetadata:       organizationID.String(),
		},
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return service.OnSlackbotDeliverySuccess(ctx, tx, scope, message)
			},
		),
	)
}

func TestMessageService_OnSlackbotDeliverySuccessIgnoresSourceEventIDForAnchor(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := newQueueTestOrganization(t)
	service := NewMessageService(pgClient, nil, nil, nil, log.NewLogger())
	source := "provider"
	sessionKey := "source-event-session-" + organizationID.String()
	now := time.Now().UTC()
	execution := coredata.AgentExecution{
		ID:                gid.New(scope.GetTenantID(), coredata.AgentExecutionEntityType),
		OrganizationID:    organizationID,
		StartAgentName:    "assistant",
		Source:            &source,
		SessionKey:        &sessionKey,
		SourceCoordinates: []byte(`{"workspace":"opaque"}`),
		SessionMessages:   []byte(`[]`),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				_, err := execution.UpsertBySourceSession(ctx, tx, scope)

				return err
			},
		),
	)

	channelID := "C-source-only"
	messageTS := "999.001"
	message := &coredata.SlackbotMessage{
		ID:             gid.New(scope.GetTenantID(), coredata.SlackbotMessageEntityType),
		OrganizationID: organizationID,
		MessageType:    "ACCESS_REQUEST",
		ChannelID:      &channelID,
		MessageTS:      &messageTS,
		Metadata: map[string]any{
			deliverySourceEventMetadata: execution.ID.String(),
		},
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				return service.OnSlackbotDeliverySuccess(ctx, tx, scope, message)
			},
		),
	)

	var anchor coredata.AgentExecutionAnchor

	err := pgClient.WithConn(
		t.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			return anchor.LoadByProviderCoordinates(
				ctx,
				conn,
				scope,
				organizationID,
				ProviderName,
				channelID,
				messageTS,
			)
		},
	)
	require.ErrorIs(t, err, coredata.ErrResourceNotFound)
}
