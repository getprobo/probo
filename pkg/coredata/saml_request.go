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

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type SAMLRequest struct {
	ID             string    `db:"id"`
	OrganizationID gid.GID   `db:"organization_id"`
	CreatedAt      time.Time `db:"created_at"`
	ExpiresAt      time.Time `db:"expires_at"`
}

type ErrSAMLRequestNotFound struct {
	RequestID string
}

func (e ErrSAMLRequestNotFound) Error() string {
	return fmt.Sprintf("SAML request ID %q not found", e.RequestID)
}

type ErrSAMLRequestExpired struct {
	RequestID string
	ExpiresAt time.Time
}

func (e ErrSAMLRequestExpired) Error() string {
	return fmt.Sprintf("SAML request ID %q expired at %v", e.RequestID, e.ExpiresAt)
}

func (s *SAMLRequest) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	query := `
INSERT INTO auth_saml_requests (id, organization_id, tenant_id, created_at, expires_at)
VALUES (@id, @organization_id, @tenant_id, @created_at, @expires_at)
`

	args := pgx.NamedArgs{
		"id":              s.ID,
		"organization_id": s.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      s.CreatedAt,
		"expires_at":      s.ExpiresAt,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert saml_request: %w", err)
	}

	return nil
}

func (s *SAMLRequest) Load(
	ctx context.Context,
	conn pg.Conn,
	requestID string,
	organizationID gid.GID,
) error {
	query := `
SELECT id, organization_id, created_at, expires_at
FROM auth_saml_requests
WHERE id = @id AND organization_id = @organization_id
LIMIT 1
`

	args := pgx.NamedArgs{
		"id":              requestID,
		"organization_id": organizationID,
	}

	rows, err := conn.Query(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot query saml_requests: %w", err)
	}

	req, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[SAMLRequest])
	if err == pgx.ErrNoRows {
		return ErrSAMLRequestNotFound{RequestID: requestID}
	}
	if err != nil {
		return fmt.Errorf("cannot collect saml_request: %w", err)
	}

	*s = req
	return nil
}

func (s *SAMLRequest) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt) || now.Equal(s.ExpiresAt)
}

func (s *SAMLRequest) Delete(
	ctx context.Context,
	conn pg.Conn,
) error {
	query := `
DELETE FROM auth_saml_requests
WHERE id = @id
`

	_, err := conn.Exec(ctx, query, pgx.NamedArgs{"id": s.ID})
	if err != nil {
		return fmt.Errorf("cannot delete saml_request: %w", err)
	}

	return nil
}

func DeleteExpiredSAMLRequests(ctx context.Context, conn pg.Conn, now time.Time) (int64, error) {
	query := `
DELETE FROM auth_saml_requests
WHERE expires_at < @now
`

	result, err := conn.Exec(ctx, query, pgx.NamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired saml_requests: %w", err)
	}

	return result.RowsAffected(), nil
}
