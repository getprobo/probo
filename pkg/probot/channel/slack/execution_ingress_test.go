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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

func TestExecutionIngressCreatesIdleDirectConversationAndDeduplicatesEvent(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := executionIngressDatabase(t)
	handler := &Service{pg: pgClient}
	identityID := gid.New(scope.GetTenantID(), coredata.IdentityEntityType)
	event := EventBody{
		Type:        EventTypeAppMention,
		User:        "U123",
		Text:        "<@BOT> hello",
		Channel:     "C123",
		TS:          "100.200",
		ChannelType: ChannelTypeChannel,
	}
	require.NoError(
		t,
		handler.enqueueExecutionInput(
			t.Context(),
			"E123",
			"T123",
			"T123",
			organizationID,
			identityID,
			event,
			"hello",
		),
	)
	require.NoError(
		t,
		handler.enqueueExecutionInput(
			t.Context(),
			"E123",
			"T123",
			"T123",
			organizationID,
			gid.New(scope.GetTenantID(), coredata.IdentityEntityType),
			EventBody{
				Type:        EventTypeAppMention,
				User:        "U999",
				Text:        "changed replay payload",
				Channel:     "C999",
				TS:          "999.000",
				ChannelType: ChannelTypeChannel,
			},
			"changed replay payload",
		),
	)

	sessionID := sessionIDFor("T123", "C123", ChannelTypeChannel, "", "100.200")

	var (
		executionCount int
		inputCount     int
		status         coredata.AgentExecutionStatus
		messageRaw     []byte
		storedIdentity *string
	)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT count(*)
					 FROM agent_executions
					 WHERE tenant_id = $1 AND organization_id = $2
					   AND source = $3 AND session_key = $4`,
					scope.GetTenantID(),
					organizationID,
					ProviderName,
					sessionID,
				).Scan(&executionCount)
			},
		),
	)
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT status
					 FROM agent_executions
					 WHERE tenant_id = $1 AND organization_id = $2
					   AND source = $3 AND session_key = $4`,
					scope.GetTenantID(),
					organizationID,
					ProviderName,
					sessionID,
				).Scan(&status)
			},
		),
	)
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT count(*)
					 FROM agent_inputs
					 WHERE tenant_id = $1 AND organization_id = $2
					   AND source = $3 AND source_event_id = $4`,
					scope.GetTenantID(),
					organizationID,
					ProviderName,
					"E123",
				).Scan(&inputCount)
			},
		),
	)
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT message, identity_id
					 FROM agent_inputs
					 WHERE tenant_id = $1 AND organization_id = $2
					   AND source = $3 AND source_event_id = $4`,
					scope.GetTenantID(),
					organizationID,
					ProviderName,
					"E123",
				).Scan(&messageRaw, &storedIdentity)
			},
		),
	)
	assert.Equal(t, 1, executionCount)
	assert.Equal(t, coredata.AgentExecutionStatusIdle, status)
	assert.Equal(t, 1, inputCount)
	require.NotNil(t, storedIdentity)
	assert.Equal(t, identityID.String(), *storedIdentity)

	var message llm.Message
	require.NoError(t, json.Unmarshal(messageRaw, &message))
	assert.Equal(t, "hello", message.Text())
	assert.NotContains(t, message.Text(), "C123")
	assert.NotContains(t, message.Text(), "U123")
}

func TestExecutionIngressReusesThreadSessionAndStoresPerInputCoordinates(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := executionIngressDatabase(t)
	now := time.Now()
	accessID := gid.New(scope.GetTenantID(), coredata.CompliancePortalAccessEntityType)
	attributes, err := json.Marshal(map[string]any{"access_id": accessID.String()})
	require.NoError(t, err)

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				subject := &coredata.BotThreadSubject{
					ID:                     gid.New(scope.GetTenantID(), coredata.BotThreadSubjectEntityType),
					OrganizationID:         organizationID,
					Provider:               ProviderName,
					ExternalConversationID: "C999",
					ExternalMessageID:      "900.000",
					Capability:             "test",
					MessageType:            "TEST",
					Attributes:             attributes,
					SubjectNamespace:       "subject",
					SubjectKey:             "stable",
					CreatedAt:              now,
					UpdatedAt:              now,
				}
				_, err := subject.Upsert(ctx, tx, scope)

				return err
			},
		),
	)

	handler := &Service{pg: pgClient}
	aliceID := gid.New(scope.GetTenantID(), coredata.IdentityEntityType)
	bobID := gid.New(scope.GetTenantID(), coredata.IdentityEntityType)
	transcript := "Thread:\n<@U1>: we should grant it\n<@U2>: @probot grant it"

	require.NoError(
		t,
		handler.enqueueExecutionInput(
			t.Context(),
			"E-alice",
			"T999",
			"T999",
			organizationID,
			aliceID,
			EventBody{
				Type:        EventTypeAppMention,
				User:        "U1",
				Text:        "<@BOT> grant it",
				Channel:     "C999",
				TS:          "901.000",
				ThreadTS:    "900.000",
				ChannelType: ChannelTypeChannel,
			},
			transcript,
		),
	)
	require.NoError(
		t,
		handler.enqueueExecutionInput(
			t.Context(),
			"E-bob",
			"T999",
			"T999",
			organizationID,
			bobID,
			EventBody{
				Type:        EventTypeAppMention,
				User:        "U2",
				Text:        "<@BOT> don't",
				Channel:     "C999",
				TS:          "902.000",
				ThreadTS:    "900.000",
				ChannelType: ChannelTypeChannel,
			},
			transcript+"\n<@U2>: don't",
		),
	)

	sessionID := sessionIDFor("T999", "C999", ChannelTypeChannel, "900.000", "901.000")

	var (
		executionCount int
		inputCount     int
		aliceCoordsRaw []byte
		bobCoordsRaw   []byte
	)

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT count(*)
					 FROM agent_executions
					 WHERE tenant_id = $1 AND organization_id = $2
					   AND source = $3 AND session_key = $4`,
					scope.GetTenantID(),
					organizationID,
					ProviderName,
					sessionID,
				).Scan(&executionCount)
			},
		),
	)
	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				if err := conn.QueryRow(
					ctx,
					`SELECT count(*)
					 FROM agent_inputs
					 WHERE tenant_id = $1 AND organization_id = $2
					   AND source = $3`,
					scope.GetTenantID(),
					organizationID,
					ProviderName,
				).Scan(&inputCount); err != nil {
					return err
				}

				if err := conn.QueryRow(
					ctx,
					`SELECT source_coordinates
					 FROM agent_inputs
					 WHERE tenant_id = $1 AND organization_id = $2
					   AND source = $3 AND source_event_id = $4`,
					scope.GetTenantID(),
					organizationID,
					ProviderName,
					"E-alice",
				).Scan(&aliceCoordsRaw); err != nil {
					return err
				}

				return conn.QueryRow(
					ctx,
					`SELECT source_coordinates
					 FROM agent_inputs
					 WHERE tenant_id = $1 AND organization_id = $2
					   AND source = $3 AND source_event_id = $4`,
					scope.GetTenantID(),
					organizationID,
					ProviderName,
					"E-bob",
				).Scan(&bobCoordsRaw)
			},
		),
	)
	assert.Equal(t, 1, executionCount)
	assert.Equal(t, 2, inputCount)

	var aliceCoords ExecutionSourceCoordinates
	require.NoError(t, json.Unmarshal(aliceCoordsRaw, &aliceCoords))
	assert.Equal(t, "901.000", aliceCoords.MessageTS)

	var bobCoords ExecutionSourceCoordinates
	require.NoError(t, json.Unmarshal(bobCoordsRaw, &bobCoords))
	assert.Equal(t, "902.000", bobCoords.MessageTS)
}

func TestExecutionIngressRoutesAnchoredReplyWithoutChangingDomainSession(t *testing.T) {
	t.Parallel()

	pgClient, scope, organizationID := executionIngressDatabase(t)
	now := time.Now()
	source := ProviderName
	sessionKey := "domain-session-" + organizationID.String()
	execution := &coredata.AgentExecution{
		ID:              gid.New(scope.GetTenantID(), coredata.AgentExecutionEntityType),
		OrganizationID:  organizationID,
		StartAgentName:  defaultAgentProfile,
		Source:          &source,
		SessionKey:      &sessionKey,
		SessionMessages: json.RawMessage("[]"),
		MaxAttempts:     coredata.AgentExecutionDefaultMaxAttempts,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if _, err := execution.UpsertBySourceSession(ctx, tx, scope); err != nil {
					return err
				}

				anchor := &coredata.AgentExecutionAnchor{
					ID:                     gid.New(scope.GetTenantID(), coredata.AgentExecutionAnchorEntityType),
					OrganizationID:         organizationID,
					AgentExecutionID:       execution.ID,
					Provider:               ProviderName,
					ExternalConversationID: "C999",
					ExternalMessageID:      "900.000",
					CreatedAt:              now,
					UpdatedAt:              now,
				}
				_, err := anchor.Upsert(ctx, tx, scope)

				return err
			},
		),
	)

	domainSessionKey := *execution.SessionKey
	handler := &Service{pg: pgClient}
	require.NoError(
		t,
		handler.enqueueExecutionInput(
			t.Context(),
			"E-reply",
			"T999",
			"T999",
			organizationID,
			gid.New(scope.GetTenantID(), coredata.IdentityEntityType),
			EventBody{
				Type:        EventTypeMessage,
				User:        "U999",
				Text:        "approve it",
				Channel:     "C999",
				TS:          "901.000",
				ThreadTS:    "900.000",
				ChannelType: ChannelTypeChannel,
			},
			"approve it",
		),
	)

	var loaded coredata.AgentExecution

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return loaded.LoadByID(
					ctx,
					conn,
					scope,
					execution.ID,
				)
			},
		),
	)
	assert.Equal(t, domainSessionKey, *loaded.SessionKey)

	var coordinates ExecutionSourceCoordinates
	require.NoError(t, json.Unmarshal(loaded.SourceCoordinates, &coordinates))
	assert.Equal(t, "U999", coordinates.ExternalUserID)
	assert.Equal(t, "901.000", coordinates.MessageTS)
	assert.Equal(t, "900.000", coordinates.ThreadTS)
}

func executionIngressDatabase(
	t *testing.T,
) (*pg.Client, coredata.Scoper, gid.GID) {
	t.Helper()

	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now()

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				_, err := conn.Exec(
					ctx,
					`INSERT INTO organizations
					 (id, tenant_id, name, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $5)`,
					organizationID,
					tenantID,
					"slack-ingress-"+organizationID.String(),
					now,
					now,
				)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(
						ctx,
						`DELETE FROM organizations WHERE id = $1`,
						organizationID,
					)

					return err
				},
			)
		},
	)

	return pgClient, scope, organizationID
}
