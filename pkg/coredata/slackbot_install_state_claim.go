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

type SlackbotInstallStateClaim struct {
	StateDigest         []byte     `db:"state_digest"`
	OrganizationID      gid.GID    `db:"organization_id"`
	ProcessingToken     *string    `db:"processing_token"`
	ProcessingStartedAt *time.Time `db:"processing_started_at"`
	CompletedAt         *time.Time `db:"completed_at"`
	CreatedAt           time.Time  `db:"created_at"`
}

func NewSlackbotInstallStateClaim(
	organizationID gid.GID,
	state string,
) *SlackbotInstallStateClaim {
	return &SlackbotInstallStateClaim{
		StateDigest:    hash.SHA256String(state),
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
	}
}

func (c *SlackbotInstallStateClaim) Claim(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
	now time.Time,
	staleAfter time.Duration,
) (bool, error) {
	q := `
INSERT INTO slackbot_install_state_claims (
	state_digest,
	tenant_id,
	organization_id,
	processing_token,
	processing_started_at,
	created_at
) VALUES (
	@state_digest,
	@tenant_id,
	@organization_id,
	@processing_token,
	@processing_started_at,
	@created_at
)
ON CONFLICT (state_digest) DO UPDATE
SET
	processing_token = EXCLUDED.processing_token,
	processing_started_at = EXCLUDED.processing_started_at
WHERE slackbot_install_state_claims.completed_at IS NULL
	AND slackbot_install_state_claims.tenant_id = EXCLUDED.tenant_id
	AND slackbot_install_state_claims.organization_id = EXCLUDED.organization_id
	AND (
		slackbot_install_state_claims.processing_started_at IS NULL
		OR slackbot_install_state_claims.processing_started_at <= @stale_before
	)
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"state_digest":          c.StateDigest,
			"tenant_id":             scope.GetTenantID(),
			"organization_id":       c.OrganizationID,
			"processing_token":      processingToken,
			"processing_started_at": now,
			"stale_before":          now.Add(-staleAfter),
			"created_at":            c.CreatedAt,
		},
	)
	if err != nil {
		return false, fmt.Errorf("cannot claim Slackbot install state: %w", err)
	}

	claimed := result.RowsAffected() > 0
	if claimed {
		c.ProcessingToken = &processingToken
		c.ProcessingStartedAt = &now
	}

	return claimed, nil
}

func (c *SlackbotInstallStateClaim) Complete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
	now time.Time,
) error {
	q := `
UPDATE slackbot_install_state_claims
SET
	processing_token = NULL,
	processing_started_at = NULL,
	completed_at = @completed_at
WHERE state_digest = @state_digest
	AND tenant_id = @tenant_id
	AND organization_id = @organization_id
	AND processing_token = @processing_token
	AND completed_at IS NULL
`

	result, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"state_digest":     c.StateDigest,
			"tenant_id":        scope.GetTenantID(),
			"organization_id":  c.OrganizationID,
			"processing_token": processingToken,
			"completed_at":     now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot complete Slackbot install state: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	c.ProcessingToken = nil
	c.ProcessingStartedAt = nil
	c.CompletedAt = &now

	return nil
}

func (c *SlackbotInstallStateClaim) Release(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
) error {
	q := `
DELETE FROM slackbot_install_state_claims
WHERE state_digest = @state_digest
	AND tenant_id = @tenant_id
	AND organization_id = @organization_id
	AND processing_token = @processing_token
	AND completed_at IS NULL
`

	_, err := conn.Exec(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"state_digest":     c.StateDigest,
			"tenant_id":        scope.GetTenantID(),
			"organization_id":  c.OrganizationID,
			"processing_token": processingToken,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot release Slackbot install state: %w", err)
	}

	return nil
}
