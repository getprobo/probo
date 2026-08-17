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

// ProbotIdentityBindingChallenge is an opaque bind token. The table is
// unscoped because confirmation happens while the user is signed in, before
// the external workspace is mapped to a tenant.
type ProbotIdentityBindingChallenge struct {
	HashedToken        []byte     `db:"hashed_token"`
	Provider           string     `db:"provider"`
	ExternalTenantID   string     `db:"external_tenant_id"`
	ExternalUserID     string     `db:"external_user_id"`
	ExternalTenantName string     `db:"external_tenant_name"`
	ExternalUserName   string     `db:"external_user_name"`
	ExpiresAt          time.Time  `db:"expires_at"`
	ConfirmedByID      *gid.GID   `db:"confirmed_by_identity_id"`
	ConfirmedAt        *time.Time `db:"confirmed_at"`
	CreatedAt          time.Time  `db:"created_at"`
}

func (c *ProbotIdentityBindingChallenge) Insert(
	ctx context.Context,
	conn pg.Querier,
) error {
	q := `
INSERT INTO probot_identity_binding_challenges (
    hashed_token,
    provider,
    external_tenant_id,
    external_user_id,
    external_tenant_name,
    external_user_name,
    expires_at,
    confirmed_by_identity_id,
    confirmed_at,
    created_at
) VALUES (
    @hashed_token,
    @provider,
    @external_tenant_id,
    @external_user_id,
    @external_tenant_name,
    @external_user_name,
    @expires_at,
    @confirmed_by_identity_id,
    @confirmed_at,
    @created_at
);
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{
		"hashed_token":             c.HashedToken,
		"provider":                 c.Provider,
		"external_tenant_id":       c.ExternalTenantID,
		"external_user_id":         c.ExternalUserID,
		"external_tenant_name":     c.ExternalTenantName,
		"external_user_name":       c.ExternalUserName,
		"expires_at":               c.ExpiresAt,
		"confirmed_by_identity_id": c.ConfirmedByID,
		"confirmed_at":             c.ConfirmedAt,
		"created_at":               c.CreatedAt,
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "probot_identity_binding_challenges_pkey" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert probot identity binding challenge: %w", err)
	}

	return nil
}

func (c *ProbotIdentityBindingChallenge) LoadByHashedToken(
	ctx context.Context,
	conn pg.Querier,
	hashedToken []byte,
) error {
	return c.loadByHashedToken(ctx, conn, hashedToken, false)
}

func (c *ProbotIdentityBindingChallenge) LoadByHashedTokenForUpdate(
	ctx context.Context,
	conn pg.Querier,
	hashedToken []byte,
) error {
	return c.loadByHashedToken(ctx, conn, hashedToken, true)
}

func (c *ProbotIdentityBindingChallenge) loadByHashedToken(
	ctx context.Context,
	conn pg.Querier,
	hashedToken []byte,
	forUpdate bool,
) error {
	lockClause := ""
	if forUpdate {
		lockClause = "FOR UPDATE"
	}

	q := fmt.Sprintf(`
SELECT
    hashed_token,
    provider,
    external_tenant_id,
    external_user_id,
    external_tenant_name,
    external_user_name,
    expires_at,
    confirmed_by_identity_id,
    confirmed_at,
    created_at
FROM
    probot_identity_binding_challenges
WHERE
    hashed_token = @hashed_token
LIMIT 1
%s;
`, lockClause)

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{"hashed_token": hashedToken},
	)
	if err != nil {
		return fmt.Errorf("cannot query probot identity binding challenge: %w", err)
	}

	challenge, err := pgx.CollectExactlyOneRow(
		rows,
		pgx.RowToStructByName[ProbotIdentityBindingChallenge],
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect probot identity binding challenge: %w", err)
	}

	*c = challenge

	return nil
}

func (c *ProbotIdentityBindingChallenge) MarkConfirmed(
	ctx context.Context,
	conn pg.Querier,
	identityID gid.GID,
	confirmedAt time.Time,
) error {
	result, err := conn.Exec(
		ctx,
		`
UPDATE probot_identity_binding_challenges
SET
    confirmed_by_identity_id = @identity_id,
    confirmed_at = @confirmed_at
WHERE
    hashed_token = @hashed_token;
`,
		pgx.StrictNamedArgs{
			"hashed_token": c.HashedToken,
			"identity_id":  identityID,
			"confirmed_at": confirmedAt,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot confirm probot identity binding challenge: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	c.ConfirmedByID = &identityID
	c.ConfirmedAt = &confirmedAt

	return nil
}

func DeleteExpiredProbotIdentityBindingChallenges(
	ctx context.Context,
	conn pg.Querier,
	expiredBefore time.Time,
) error {
	_, err := conn.Exec(
		ctx,
		`
DELETE FROM probot_identity_binding_challenges
WHERE expires_at < @expired_before;
`,
		pgx.StrictNamedArgs{"expired_before": expiredBefore},
	)
	if err != nil {
		return fmt.Errorf("cannot delete expired probot identity binding challenges: %w", err)
	}

	return nil
}

func DeleteProbotIdentityBindingChallengesByProviderAndExternalTenant(
	ctx context.Context,
	conn pg.Querier,
	provider string,
	externalTenantID string,
) error {
	_, err := conn.Exec(
		ctx,
		`
DELETE FROM probot_identity_binding_challenges
WHERE provider = @provider
	AND external_tenant_id = @external_tenant_id;
`,
		pgx.StrictNamedArgs{
			"provider":           provider,
			"external_tenant_id": externalTenantID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot delete probot identity binding challenges: %w", err)
	}

	return nil
}

func DeleteUnconfirmedProbotIdentityBindingChallengesBySubject(
	ctx context.Context,
	conn pg.Querier,
	provider string,
	externalTenantID string,
	externalUserID string,
	exceptHashedToken []byte,
) error {
	_, err := conn.Exec(
		ctx,
		`
WITH doomed AS (
	SELECT hashed_token
	FROM probot_identity_binding_challenges
	WHERE provider = @provider
		AND external_tenant_id = @external_tenant_id
		AND external_user_id = @external_user_id
		AND confirmed_at IS NULL
		AND hashed_token <> @hashed_token
	FOR UPDATE SKIP LOCKED
)
DELETE FROM probot_identity_binding_challenges
WHERE hashed_token IN (SELECT hashed_token FROM doomed);
`,
		pgx.StrictNamedArgs{
			"provider":           provider,
			"external_tenant_id": externalTenantID,
			"external_user_id":   externalUserID,
			"hashed_token":       exceptHashedToken,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot delete sibling probot identity binding challenges: %w", err)
	}

	return nil
}
