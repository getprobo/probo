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

// InstallStateStaleAfter releases a claim abandoned mid-flight (a crashed
// process) so the customer's retry inside the state TTL still works.
//
// It must stay longer than the bounded network work either ceremony performs
// while holding a claim — Slack's token exchange and the connector install
// verification are both capped at 30s — or a live request would be reclaimed
// from under itself and run a second time.
const InstallStateStaleAfter = 2 * time.Minute

// The two ceremonies keep separate ledgers so one retention sweep never touches
// the other's rows.
const (
	connectorInstallStateClaimTable = "connector_install_state_claims"
	slackbotInstallStateClaimTable  = "slackbot_install_state_claims"
)

// InstallStateClaim is the single-use ledger entry for one app-install state
// token. Both install callbacks are public GETs whose vendor proof stays valid,
// so nothing but this row stops a refresh or a link-preview prefetch from
// binding a second installation.
//
// It keys on the state digest, not on the vendor tenant: every mint carries a
// fresh nonce, so two browser tabs are two claims and both succeed. Serializing
// two first installs of the same tenant is LockConnectorInstallResource's job.
type InstallStateClaim struct {
	// table and subject are all that separate the connector ledger from the
	// slackbot one; the mechanism is identical.
	table   string
	subject string

	StateDigest         []byte     `db:"state_digest"`
	OrganizationID      gid.GID    `db:"organization_id"`
	ProcessingToken     *string    `db:"processing_token"`
	ProcessingStartedAt *time.Time `db:"processing_started_at"`
	CompletedAt         *time.Time `db:"completed_at"`
	CreatedAt           time.Time  `db:"created_at"`
}

// NewConnectorInstallStateClaim derives the ledger entry for an access-review
// connector install state. Only the digest is kept: the raw state carries the
// organization and identity GIDs and has no business at rest.
func NewConnectorInstallStateClaim(
	organizationID gid.GID,
	state string,
) *InstallStateClaim {
	return newInstallStateClaim(
		connectorInstallStateClaimTable,
		"connector",
		organizationID,
		state,
	)
}

// NewSlackbotInstallStateClaim derives the ledger entry for a Slack app install
// state.
func NewSlackbotInstallStateClaim(
	organizationID gid.GID,
	state string,
) *InstallStateClaim {
	return newInstallStateClaim(
		slackbotInstallStateClaimTable,
		"slack",
		organizationID,
		state,
	)
}

func newInstallStateClaim(
	table string,
	subject string,
	organizationID gid.GID,
	state string,
) *InstallStateClaim {
	return &InstallStateClaim{
		table:          table,
		subject:        subject,
		StateDigest:    hash.SHA256String(state),
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
	}
}

// Claim takes exclusive ownership of the state for processingToken, reporting
// false when another request already holds or completed it. A claim abandoned
// mid-flight longer ago than staleAfter is reclaimable, so a crashed process
// does not cost the customer their window.
func (c *InstallStateClaim) Claim(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
	now time.Time,
	staleAfter time.Duration,
) (bool, error) {
	q := fmt.Sprintf(`
INSERT INTO %[1]s (
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
WHERE %[1]s.completed_at IS NULL
    AND %[1]s.tenant_id = EXCLUDED.tenant_id
    AND %[1]s.organization_id = EXCLUDED.organization_id
    AND (
        %[1]s.processing_started_at IS NULL
        OR %[1]s.processing_started_at <= @stale_before
    )
`, c.table)

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
		return false, fmt.Errorf("cannot claim %s install state: %w", c.subject, err)
	}

	claimed := result.RowsAffected() > 0
	if claimed {
		c.ProcessingToken = &processingToken
		c.ProcessingStartedAt = &now
	}

	return claimed, nil
}

// Complete burns the state for good. It must run in the same transaction that
// persists the installation: a crash between the two would otherwise leave a
// spent state and nothing to show for it.
func (c *InstallStateClaim) Complete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
	now time.Time,
) error {
	q := fmt.Sprintf(`
UPDATE %s
SET
    processing_token = NULL,
    processing_started_at = NULL,
    completed_at = @completed_at
WHERE state_digest = @state_digest
    AND tenant_id = @tenant_id
    AND organization_id = @organization_id
    AND processing_token = @processing_token
    AND completed_at IS NULL
`, c.table)

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
		return fmt.Errorf("cannot complete %s install state: %w", c.subject, err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	c.ProcessingToken = nil
	c.ProcessingStartedAt = nil
	c.CompletedAt = &now

	return nil
}

// Release hands the state back unspent so a retry inside its TTL still works.
// It is the right answer only to a failure a retry could fix.
func (c *InstallStateClaim) Release(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	processingToken string,
) error {
	q := fmt.Sprintf(`
DELETE FROM %s
WHERE state_digest = @state_digest
    AND tenant_id = @tenant_id
    AND organization_id = @organization_id
    AND processing_token = @processing_token
    AND completed_at IS NULL
`, c.table)

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
		return fmt.Errorf("cannot release %s install state: %w", c.subject, err)
	}

	return nil
}
