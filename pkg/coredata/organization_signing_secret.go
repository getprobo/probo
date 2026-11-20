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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type OrganizationSigningSecret struct {
	ID             gid.GID                   `db:"id"`
	OrganizationID gid.GID                   `db:"organization_id"`
	SecretKeyHash  string                    `db:"secret_key_hash"`
	Mode           DocumentSigningConfigMode `db:"mode"`
	CreatedAt      time.Time                 `db:"created_at"`
	UpdatedAt      time.Time                 `db:"updated_at"`
}

func (s *OrganizationSigningSecret) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO organization_signing_secrets (
	id, tenant_id, organization_id,
	secret_key_hash, mode,
	created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id,
	@secret_key_hash, @mode,
	@created_at, @updated_at
)
ON CONFLICT (organization_id) DO UPDATE SET
	secret_key_hash = EXCLUDED.secret_key_hash,
	mode            = EXCLUDED.mode,
	updated_at      = EXCLUDED.updated_at
RETURNING id, organization_id, secret_key_hash, mode, created_at, updated_at
`
	originalID := s.ID

	args := pgx.StrictNamedArgs{
		"id":              s.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"secret_key_hash": s.SecretKeyHash,
		"mode":            s.Mode,
		"created_at":      s.CreatedAt,
		"updated_at":      s.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert organization signing secret: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OrganizationSigningSecret])
	if err != nil {
		return false, fmt.Errorf("cannot collect upsert result: %w", err)
	}

	*s = row

	return originalID == s.ID, nil
}

func (s *OrganizationSigningSecret) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT id, organization_id, secret_key_hash, mode, created_at, updated_at
FROM organization_signing_secrets
WHERE %s AND organization_id = @organization_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query organization signing secret: %w", err)
	}

	secret, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OrganizationSigningSecret])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect organization signing secret: %w", err)
	}

	*s = secret

	return nil
}
