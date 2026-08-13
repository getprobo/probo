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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

// ProbotIdentityBinding maps an external-provider actor to an identity.
// The table is unscoped because lookup happens before tenant resolution.
type ProbotIdentityBinding struct {
	ID               gid.GID   `db:"id"`
	Provider         string    `db:"provider"`
	ExternalTenantID string    `db:"external_tenant_id"`
	ExternalUserID   string    `db:"external_user_id"`
	IdentityID       gid.GID   `db:"identity_id"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

const probotIdentityBindingColumns = `
    id,
    provider,
    external_tenant_id,
    external_user_id,
    identity_id,
    created_at,
    updated_at
`

func (b *ProbotIdentityBinding) LoadByExternalSubject(
	ctx context.Context,
	conn pg.Querier,
	provider, externalTenantID, externalUserID string,
) error {
	q := `
SELECT
` + probotIdentityBindingColumns + `
FROM
    probot_identity_bindings
WHERE
    provider = @provider
    AND external_tenant_id = @external_tenant_id
    AND external_user_id = @external_user_id
LIMIT 1;
`

	return b.loadOne(ctx, conn, q, pgx.StrictNamedArgs{
		"provider":           provider,
		"external_tenant_id": externalTenantID,
		"external_user_id":   externalUserID,
	})
}

func (b *ProbotIdentityBinding) LoadByIdentityAndExternalTenant(
	ctx context.Context,
	conn pg.Querier,
	identityID gid.GID,
	provider, externalTenantID string,
) error {
	q := `
SELECT
` + probotIdentityBindingColumns + `
FROM
    probot_identity_bindings
WHERE
    identity_id = @identity_id
    AND provider = @provider
    AND external_tenant_id = @external_tenant_id
LIMIT 1;
`

	return b.loadOne(ctx, conn, q, pgx.StrictNamedArgs{
		"identity_id":        identityID,
		"provider":           provider,
		"external_tenant_id": externalTenantID,
	})
}

func (b *ProbotIdentityBinding) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	bindingID gid.GID,
) error {
	q := `
SELECT
` + probotIdentityBindingColumns + `
FROM
    probot_identity_bindings
WHERE
    id = @id
LIMIT 1;
`

	return b.loadOne(ctx, conn, q, pgx.StrictNamedArgs{"id": bindingID})
}

func (b *ProbotIdentityBinding) loadOne(
	ctx context.Context,
	conn pg.Querier,
	q string,
	args pgx.StrictNamedArgs,
) error {
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query probot identity binding: %w", err)
	}

	binding, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[ProbotIdentityBinding],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect probot identity binding: %w", err)
	}

	*b = binding

	return nil
}

func (b *ProbotIdentityBinding) Insert(ctx context.Context, conn pg.Querier) error {
	q := `
INSERT INTO probot_identity_bindings (
` + probotIdentityBindingColumns + `
) VALUES (
    @id,
    @provider,
    @external_tenant_id,
    @external_user_id,
    @identity_id,
    @created_at,
    @updated_at
);
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{
		"id":                 b.ID,
		"provider":           b.Provider,
		"external_tenant_id": b.ExternalTenantID,
		"external_user_id":   b.ExternalUserID,
		"identity_id":        b.IdentityID,
		"created_at":         b.CreatedAt,
		"updated_at":         b.UpdatedAt,
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "probot_identity_bindings_provider_external_tenant_id_external_user_id_key",
				"probot_identity_bindings_identity_id_provider_external_tenant_id_key":
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert probot identity binding: %w", err)
	}

	return nil
}

func (b *ProbotIdentityBinding) Delete(ctx context.Context, conn pg.Querier) error {
	_, err := conn.Exec(
		ctx,
		`DELETE FROM probot_identity_bindings WHERE id = @id;`,
		pgx.StrictNamedArgs{"id": b.ID},
	)
	if err != nil {
		return fmt.Errorf("cannot delete probot identity binding: %w", err)
	}

	return nil
}

func DeleteProbotIdentityBindingsByProviderAndExternalTenant(
	ctx context.Context,
	conn pg.Querier,
	provider string,
	externalTenantID string,
) error {
	_, err := conn.Exec(
		ctx,
		`
DELETE FROM probot_identity_bindings
WHERE provider = @provider
	AND external_tenant_id = @external_tenant_id;
`,
		pgx.StrictNamedArgs{
			"provider":           provider,
			"external_tenant_id": externalTenantID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot delete probot identity bindings: %w", err)
	}

	return nil
}
