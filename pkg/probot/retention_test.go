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
	"io"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestRetentionHandlerDeletesExpiredCompletedRecords(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	now := time.Now().UTC().Round(time.Microsecond)
	event := coredata.NewSlackbotEvent("retention-event-"+now.Format(time.RFC3339Nano), []byte(`{}`))
	command := coredata.NewSlackbotInteractiveCommand([]byte(event.EventID), []byte("encrypted"))
	processed := coredata.SlackbotProcessedEvent{
		EventID:   event.EventID,
		CreatedAt: now.Add(-2 * time.Hour),
	}

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if _, err := event.Insert(ctx, tx); err != nil {
					return err
				}

				event.ProcessedAt = new(now.Add(-2 * time.Hour))

				event.UpdatedAt = now.Add(-2 * time.Hour)
				if err := event.UpdateProcessingState(ctx, tx); err != nil {
					return err
				}

				if _, err := command.Insert(ctx, tx); err != nil {
					return err
				}

				command.ProcessedAt = new(now.Add(-2 * time.Hour))

				command.UpdatedAt = now.Add(-2 * time.Hour)
				if err := command.UpdateProcessingState(ctx, tx); err != nil {
					return err
				}

				_, err := processed.Claim(ctx, tx)

				return err
			},
		),
	)

	h := &retentionHandler{
		pg:                        pgClient,
		logger:                    log.NewLogger(log.WithOutput(io.Discard)),
		retention:                 time.Hour,
		deadLetterRetention:       3 * time.Hour,
		operationReceiptRetention: 4 * time.Hour,
		batchSize:                 10,
		now:                       func() time.Time { return now },
	}
	require.NoError(t, h.Run(t.Context()))

	var count int

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT
						(SELECT count(*) FROM slackbot_events WHERE event_id = @event_id) +
						(SELECT count(*) FROM slackbot_interactive_commands WHERE id = @command_id) +
						(SELECT count(*) FROM slackbot_processed_events WHERE event_id = @event_id)`,
					pgx.StrictNamedArgs{
						"event_id":   event.EventID,
						"command_id": command.ID,
					},
				).Scan(&count)
			},
		),
	)
	require.Zero(t, count)
}

func TestRetentionHandlerPreservesReceiptsBeyondReplayState(t *testing.T) {
	t.Parallel()

	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	now := time.Now().UTC().Round(time.Microsecond)
	old := now.Add(-5 * time.Hour)
	recent := now.Add(-2 * time.Hour)
	source := "retention-test"
	sessionKey := "retention-" + organizationID.String()
	execution := &coredata.AgentExecution{
		ID:              gid.New(tenantID, coredata.AgentRunEntityType),
		OrganizationID:  organizationID,
		StartAgentName:  "probot",
		Source:          &source,
		SessionKey:      &sessionKey,
		SessionMessages: []byte("[]"),
		MaxAttempts:     coredata.AgentExecutionDefaultMaxAttempts,
		CreatedAt:       old,
		UpdatedAt:       old,
	}
	input := &coredata.AgentInput{
		ID:             gid.New(tenantID, coredata.AgentInputEntityType),
		OrganizationID: organizationID,
		Source:         source,
		SourceEventID:  new("old-event-" + organizationID.String()),
		Message:        []byte(`{"role":"user","parts":[]}`),
		ProcessedAt:    new(old),
		MaxAttempts:    coredata.AgentInputDefaultMaxAttempts,
		CreatedAt:      old,
		UpdatedAt:      old,
	}
	anchor := &coredata.AgentExecutionAnchor{
		ID:                     gid.New(tenantID, coredata.AgentExecutionAnchorEntityType),
		OrganizationID:         organizationID,
		Provider:               "slack",
		ExternalConversationID: "C-retention",
		ExternalMessageID:      "1.0",
		CreatedAt:              old,
		UpdatedAt:              old,
	}
	oldReceipt := coredata.NewOperationReceipt(scope, organizationID, "old-receipt")
	oldReceipt.CreatedAt = old
	recentReceipt := coredata.NewOperationReceipt(scope, organizationID, "recent-receipt")
	recentReceipt.CreatedAt = recent

	require.NoError(
		t,
		pgClient.WithTx(
			t.Context(),
			func(ctx context.Context, tx pg.Tx) error {
				if _, err := tx.Exec(
					ctx,
					`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at)
					 VALUES ($1, $2, $3, $4, $4)`,
					organizationID,
					tenantID,
					"retention-"+organizationID.String(),
					now,
				); err != nil {
					return err
				}

				if _, err := execution.UpsertConversationalBySourceSession(ctx, tx, scope); err != nil {
					return err
				}

				input.AgentRunID = execution.ID
				if _, err := input.EnqueueIdempotently(ctx, tx, scope); err != nil {
					return err
				}

				anchor.AgentRunID = execution.ID
				if _, err := anchor.Upsert(ctx, tx, scope); err != nil {
					return err
				}

				if _, err := oldReceipt.Claim(ctx, tx, scope); err != nil {
					return err
				}

				_, err := recentReceipt.Claim(ctx, tx, scope)

				return err
			},
		),
	)
	t.Cleanup(
		func() {
			_ = pgClient.WithConn(
				context.Background(),
				func(ctx context.Context, conn pg.Querier) error {
					_, err := conn.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID)
					return err
				},
			)
		},
	)

	h := &retentionHandler{
		pg:                        pgClient,
		logger:                    log.NewLogger(log.WithOutput(io.Discard)),
		retention:                 time.Hour,
		deadLetterRetention:       3 * time.Hour,
		operationReceiptRetention: 4 * time.Hour,
		batchSize:                 100,
		now:                       func() time.Time { return now },
	}
	require.NoError(t, h.Run(t.Context()))

	var executionCount, inputCount, anchorCount, oldReceiptCount, recentReceiptCount int

	require.NoError(
		t,
		pgClient.WithConn(
			t.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				return conn.QueryRow(
					ctx,
					`SELECT
						(SELECT count(*) FROM agent_runs WHERE id = @execution_id),
						(SELECT count(*) FROM agent_inputs WHERE id = @input_id),
						(SELECT count(*) FROM agent_execution_anchors WHERE id = @anchor_id),
						(SELECT count(*) FROM operation_receipts WHERE id = @old_receipt_id),
						(SELECT count(*) FROM operation_receipts WHERE id = @recent_receipt_id)`,
					pgx.StrictNamedArgs{
						"execution_id":      execution.ID,
						"input_id":          input.ID,
						"anchor_id":         anchor.ID,
						"old_receipt_id":    oldReceipt.ID,
						"recent_receipt_id": recentReceipt.ID,
					},
				).Scan(
					&executionCount,
					&inputCount,
					&anchorCount,
					&oldReceiptCount,
					&recentReceiptCount,
				)
			},
		),
	)
	require.Zero(t, executionCount)
	require.Zero(t, inputCount)
	require.Zero(t, anchorCount)
	require.Zero(t, oldReceiptCount)
	require.Equal(t, 1, recentReceiptCount)
}
