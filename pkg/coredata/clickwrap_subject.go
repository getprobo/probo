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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	ClickwrapSubject struct {
		ID                gid.GID   `db:"id"`
		OrganizationID    gid.GID   `db:"organization_id"`
		ExternalUserID    string    `db:"external_user_id"`
		Email             string    `db:"email"`
		FullName          *string   `db:"full_name"`
		ExternalAccountID *string   `db:"external_account_id"`
		CreatedAt         time.Time `db:"created_at"`
		UpdatedAt         time.Time `db:"updated_at"`
	}

	ClickwrapSubjects []*ClickwrapSubject
)

func (s *ClickwrapSubject) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM clickwrap_subjects WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{"resource_ids": resourceIDs}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query clickwrap subject authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan clickwrap subject authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate clickwrap subject authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (s *ClickwrapSubject) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO clickwrap_subjects (
	id, tenant_id, organization_id,
	external_user_id, email, full_name, external_account_id,
	created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id,
	@external_user_id, @email, @full_name, @external_account_id,
	@created_at, @updated_at
)
ON CONFLICT (organization_id, external_user_id) DO UPDATE SET
	email               = EXCLUDED.email,
	full_name           = EXCLUDED.full_name,
	external_account_id = EXCLUDED.external_account_id,
	updated_at          = EXCLUDED.updated_at
RETURNING id, organization_id, external_user_id, email, full_name, external_account_id, created_at, updated_at
`
	originalID := s.ID

	args := pgx.StrictNamedArgs{
		"id":                  s.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     s.OrganizationID,
		"external_user_id":    s.ExternalUserID,
		"email":               s.Email,
		"full_name":           s.FullName,
		"external_account_id": s.ExternalAccountID,
		"created_at":          s.CreatedAt,
		"updated_at":          s.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert clickwrap subject: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ClickwrapSubject])
	if err != nil {
		return false, fmt.Errorf("cannot collect upsert result: %w", err)
	}

	*s = row

	return originalID == s.ID, nil
}

func (s *ClickwrapSubject) LoadByExternalUserID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	externalUserID string,
) error {
	q := `
SELECT id, organization_id, external_user_id, email, full_name, external_account_id, created_at, updated_at
FROM clickwrap_subjects
WHERE %s
	AND organization_id = @organization_id
	AND external_user_id = @external_user_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id":  organizationID,
		"external_user_id": externalUserID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query clickwrap subject: %w", err)
	}

	subject, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ClickwrapSubject])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect clickwrap subject: %w", err)
	}

	*s = subject

	return nil
}

func (s *ClickwrapSubject) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT id, organization_id, external_user_id, email, full_name, external_account_id, created_at, updated_at
FROM clickwrap_subjects
WHERE %s AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query clickwrap subject: %w", err)
	}

	subject, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ClickwrapSubject])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect clickwrap subject: %w", err)
	}

	*s = subject

	return nil
}

type ClickwrapSubjectOrderField string

const ClickwrapSubjectOrderFieldCreatedAt ClickwrapSubjectOrderField = "created_at"

func (f ClickwrapSubjectOrderField) String() string { return string(f) }
func (f ClickwrapSubjectOrderField) Column() string { return string(f) }

func (s ClickwrapSubject) CursorKey(orderBy ClickwrapSubjectOrderField) page.CursorKey {
	switch orderBy {
	case ClickwrapSubjectOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (ss *ClickwrapSubjects) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ClickwrapSubjectOrderField],
) error {
	q := `
SELECT id, organization_id, external_user_id, email, full_name, external_account_id, created_at, updated_at
FROM clickwrap_subjects
WHERE %s
	AND organization_id = @organization_id
	AND %s
ORDER BY created_at ASC, id ASC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query clickwrap subjects: %w", err)
	}

	subjects, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ClickwrapSubject])
	if err != nil {
		return fmt.Errorf("cannot collect clickwrap subjects: %w", err)
	}

	*ss = subjects

	return nil
}
