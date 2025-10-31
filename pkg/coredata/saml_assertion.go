// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package coredata

import (
	"context"
	"fmt"
	"time"

	"go.probo.inc/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type SAMLAssertion struct {
	ID             string    `db:"id"`
	OrganizationID gid.GID   `db:"organization_id"`
	UsedAt         time.Time `db:"used_at"`
	ExpiresAt      time.Time `db:"expires_at"`
}

type ErrAssertionAlreadyUsed struct {
	AssertionID string
}

func (e ErrAssertionAlreadyUsed) Error() string {
	return fmt.Sprintf("assertion ID %q has already been used (replay attack)", e.AssertionID)
}

func (s *SAMLAssertion) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	query := `
INSERT INTO auth_saml_assertions (id, tenant_id, organization_id, used_at, expires_at)
VALUES (@id, @tenant_id, @organization_id, @used_at, @expires_at)
`

	args := pgx.NamedArgs{
		"id":              s.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"used_at":         s.UsedAt,
		"expires_at":      s.ExpiresAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert saml_assertion: %w", err)
	}

	return nil
}

func DeleteExpiredSAMLAssertions(ctx context.Context, conn pg.Conn, now time.Time) (int64, error) {
	query := `
DELETE FROM auth_saml_assertions
WHERE expires_at < @now
`

	result, err := conn.Exec(ctx, query, pgx.NamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired saml_assertions: %w", err)
	}

	return result.RowsAffected(), nil
}
