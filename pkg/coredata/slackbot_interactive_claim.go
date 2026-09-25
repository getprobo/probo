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
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/gid"
)

type SlackbotInteractiveClaim struct {
	InteractionDigest   []byte     `db:"interaction_digest"`
	OrganizationID      gid.GID    `db:"organization_id"`
	ProcessingToken     *string    `db:"processing_token"`
	ProcessingStartedAt *time.Time `db:"processing_started_at"`
	CompletedAt         *time.Time `db:"completed_at"`
	CreatedAt           time.Time  `db:"created_at"`
}

func NewSlackbotInteractiveClaim(
	organizationID gid.GID,
	interactionKey string,
) *SlackbotInteractiveClaim {
	return &SlackbotInteractiveClaim{
		InteractionDigest: hash.SHA256String(interactionKey),
		OrganizationID:    organizationID,
		CreatedAt:         time.Now(),
	}
}

// Claim atomically acquires an interaction digest. Completed claims remain
// deduplicated while abandoned processing claims can be reclaimed after the
// stale interval.
func (c *SlackbotInteractiveClaim) Claim(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
	now time.Time,
	staleAfter time.Duration,
) (bool, error) {
	q := `
INSERT INTO slackbot_interactive_claims (
	interaction_digest,
	tenant_id,
	organization_id,
	processing_token,
	processing_started_at,
	created_at
) VALUES (
	@interaction_digest,
	@tenant_id,
	@organization_id,
	@processing_token,
	@processing_started_at,
	@created_at
)
ON CONFLICT (interaction_digest) DO UPDATE
SET
	processing_token = EXCLUDED.processing_token,
	processing_started_at = EXCLUDED.processing_started_at
WHERE slackbot_interactive_claims.completed_at IS NULL
	AND slackbot_interactive_claims.tenant_id = EXCLUDED.tenant_id
	AND slackbot_interactive_claims.organization_id = EXCLUDED.organization_id
	AND (
		slackbot_interactive_claims.processing_started_at IS NULL
		OR slackbot_interactive_claims.processing_started_at <= @stale_before
	)
`

	args := pgx.StrictNamedArgs{
		"interaction_digest":    c.InteractionDigest,
		"tenant_id":             scope.GetTenantID(),
		"organization_id":       c.OrganizationID,
		"processing_token":      processingToken,
		"processing_started_at": now,
		"stale_before":          now.Add(-staleAfter),
		"created_at":            c.CreatedAt,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot claim Slackbot interactive interaction: %w", err)
	}

	claimed := result.RowsAffected() > 0
	if claimed {
		c.ProcessingToken = &processingToken
		c.ProcessingStartedAt = &now
	}

	return claimed, nil
}

func (c *SlackbotInteractiveClaim) Complete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
	now time.Time,
) error {
	q := `
UPDATE slackbot_interactive_claims
SET
	processing_token = NULL,
	processing_started_at = NULL,
	completed_at = @completed_at
WHERE interaction_digest = @interaction_digest
	AND tenant_id = @tenant_id
	AND organization_id = @organization_id
	AND processing_token = @processing_token
	AND completed_at IS NULL
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"interaction_digest": c.InteractionDigest,
			"tenant_id":          scope.GetTenantID(),
			"organization_id":    c.OrganizationID,
			"processing_token":   processingToken,
			"completed_at":       now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot complete Slackbot interactive interaction: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (c *SlackbotInteractiveClaim) Release(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
) error {
	q := `
DELETE FROM slackbot_interactive_claims
WHERE interaction_digest = @interaction_digest
	AND tenant_id = @tenant_id
	AND organization_id = @organization_id
	AND processing_token = @processing_token
	AND completed_at IS NULL
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"interaction_digest": c.InteractionDigest,
			"tenant_id":          scope.GetTenantID(),
			"organization_id":    c.OrganizationID,
			"processing_token":   processingToken,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot release Slackbot interactive interaction: %w", err)
	}

	return nil
}
