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

package coredata

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type AgentExecutionAnchor struct {
	ID                     gid.GID   `db:"id"`
	OrganizationID         gid.GID   `db:"organization_id"`
	AgentExecutionID       gid.GID   `db:"agent_execution_id"`
	Provider               string    `db:"provider"`
	ExternalConversationID string    `db:"external_conversation_id"`
	ExternalMessageID      string    `db:"external_message_id"`
	CreatedAt              time.Time `db:"created_at"`
	UpdatedAt              time.Time `db:"updated_at"`
}

func (a *AgentExecutionAnchor) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO agent_execution_anchors (
	id, tenant_id, organization_id, agent_execution_id, provider,
	external_conversation_id, external_message_id, created_at, updated_at
) SELECT
	@id, @tenant_id, @organization_id, @agent_execution_id, @provider,
	@external_conversation_id, @external_message_id, @created_at, @updated_at
FROM agent_executions
WHERE
	agent_executions.id = @agent_execution_id
	AND agent_executions.tenant_id = @tenant_id
	AND agent_executions.organization_id = @organization_id
ON CONFLICT (tenant_id, organization_id, agent_execution_id, provider) DO UPDATE SET
	external_conversation_id = EXCLUDED.external_conversation_id,
	external_message_id = EXCLUDED.external_message_id,
	updated_at = EXCLUDED.updated_at
RETURNING
	id, organization_id, agent_execution_id, provider, external_conversation_id,
	external_message_id, created_at, updated_at
`
	originalID := a.ID

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"id":                       a.ID,
			"tenant_id":                scope.GetTenantID(),
			"organization_id":          a.OrganizationID,
			"agent_execution_id":       a.AgentExecutionID,
			"provider":                 a.Provider,
			"external_conversation_id": a.ExternalConversationID,
			"external_message_id":      a.ExternalMessageID,
			"created_at":               a.CreatedAt,
			"updated_at":               a.UpdatedAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot upsert agent execution anchor: %w", err)
	}

	anchor, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[AgentExecutionAnchor],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrResourceNotFound
		}

		return false, fmt.Errorf("cannot collect agent execution anchor: %w", err)
	}

	*a = anchor

	return originalID == a.ID, nil
}

func (a *AgentExecutionAnchor) LoadByProviderCoordinates(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	provider string,
	externalConversationID string,
	externalMessageID string,
) error {
	q := `
SELECT
	id, organization_id, agent_execution_id, provider, external_conversation_id,
	external_message_id, created_at, updated_at
FROM agent_execution_anchors
WHERE
	tenant_id = @tenant_id
	AND organization_id = @organization_id
	AND provider = @provider
	AND external_conversation_id = @external_conversation_id
	AND external_message_id = @external_message_id
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"tenant_id":                scope.GetTenantID(),
			"organization_id":          organizationID,
			"provider":                 provider,
			"external_conversation_id": externalConversationID,
			"external_message_id":      externalMessageID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot load agent execution anchor: %w", err)
	}

	anchor, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[AgentExecutionAnchor],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agent execution anchor: %w", err)
	}

	*a = anchor

	return nil
}
