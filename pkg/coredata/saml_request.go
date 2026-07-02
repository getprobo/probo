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
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type SAMLRequest struct {
	ID             string    `db:"id"`
	OrganizationID gid.GID   `db:"organization_id"`
	CreatedAt      time.Time `db:"created_at"`
	ExpiresAt      time.Time `db:"expires_at"`
}

func (s *SAMLRequest) Insert(
	ctx context.Context,
	conn pg.Tx,
) error {
	query := `
INSERT INTO iam_saml_requests (id, organization_id, created_at, expires_at)
VALUES (@id, @organization_id, @created_at, @expires_at)
`

	args := pgx.NamedArgs{
		"id":              s.ID,
		"organization_id": s.OrganizationID,
		"created_at":      s.CreatedAt,
		"expires_at":      s.ExpiresAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert saml_request: %w", err)
	}

	return nil
}

func LoadValidRequestIDsForOrganization(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
	now time.Time,
) ([]string, error) {
	query := `
SELECT id
FROM iam_saml_requests
WHERE organization_id = @organization_id AND expires_at > @now
`

	args := pgx.NamedArgs{
		"organization_id": organizationID,
		"now":             now,
	}

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query saml_requests: %w", err)
	}

	requestIDs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		var id string

		err := row.Scan(&id)

		return id, err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot collect request IDs: %w", err)
	}

	return requestIDs, nil
}

func DeleteExpiredSAMLRequests(ctx context.Context, conn pg.Tx, now time.Time) (int64, error) {
	query := `
DELETE FROM iam_saml_requests
WHERE expires_at < @now
`

	result, err := conn.Exec(ctx, query, pgx.NamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired saml_requests: %w", err)
	}

	return result.RowsAffected(), nil
}
