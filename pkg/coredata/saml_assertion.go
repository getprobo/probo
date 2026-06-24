// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type SAMLAssertion struct {
	ID             string    `db:"id"`
	OrganizationID gid.GID   `db:"organization_id"`
	UsedAt         time.Time `db:"used_at"`
	ExpiresAt      time.Time `db:"expires_at"`
}

func (s *SAMLAssertion) Insert(
	ctx context.Context,
	conn pg.Tx,
) error {
	query := `
INSERT INTO iam_saml_assertions (id, organization_id, used_at, expires_at)
VALUES (@id, @organization_id, @used_at, @expires_at)
`

	args := pgx.NamedArgs{
		"id":              s.ID,
		"organization_id": s.OrganizationID,
		"used_at":         s.UsedAt,
		"expires_at":      s.ExpiresAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "iam_saml_assertions_pkey" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert saml_assertion: %w", err)
	}

	return nil
}

func DeleteExpiredSAMLAssertions(ctx context.Context, conn pg.Tx, now time.Time) (int64, error) {
	query := `
DELETE FROM iam_saml_assertions
WHERE expires_at < @now
`

	result, err := conn.Exec(ctx, query, pgx.NamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired saml_assertions: %w", err)
	}

	return result.RowsAffected(), nil
}
