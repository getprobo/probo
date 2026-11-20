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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	AgreementSubject struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		SiteID         gid.GID   `db:"site_id"`
		ExternalUserID string    `db:"external_user_id"`
		AccountID      *string   `db:"account_id"`
		Email          string    `db:"email"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	AgreementSubjects []*AgreementSubject
)

func (s *AgreementSubject) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM agreement_subjects WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query agreement subject authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan agreement subject authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate agreement subject authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (s *AgreementSubject) CursorKey(orderBy AgreementSubjectOrderField) page.CursorKey {
	switch orderBy {
	case AgreementSubjectOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *AgreementSubject) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO agreement_subjects (
	id, tenant_id, organization_id, site_id, external_user_id, account_id, email,
	created_at, updated_at
) VALUES (
	@id, @tenant_id, @organization_id, @site_id, @external_user_id, @account_id, @email,
	@created_at, @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               s.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  s.OrganizationID,
		"site_id":          s.SiteID,
		"external_user_id": s.ExternalUserID,
		"account_id":       s.AccountID,
		"email":            s.Email,
		"created_at":       s.CreatedAt,
		"updated_at":       s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "agreement_subjects_site_id_external_user_id_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert agreement subject: %w", err)
	}

	return nil
}

func (s *AgreementSubject) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE agreement_subjects SET
	email = @email,
	account_id = @account_id,
	updated_at = @updated_at
WHERE %s AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         s.ID,
		"email":      s.Email,
		"account_id": s.AccountID,
		"updated_at": s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update agreement subject: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *AgreementSubject) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id, organization_id, site_id, external_user_id, account_id, email,
	created_at, updated_at
FROM agreement_subjects
WHERE %s AND id = @id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement subject: %w", err)
	}

	sub, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgreementSubject])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agreement subject: %w", err)
	}

	*s = sub

	return nil
}

// LoadBySiteAndExternalUserID finds the subject for a given site + external_user_id pair.
func (s *AgreementSubject) LoadBySiteAndExternalUserID(
	ctx context.Context,
	conn pg.Querier,
	siteID gid.GID,
	externalUserID string,
) error {
	q := `
SELECT
	id, organization_id, site_id, external_user_id, account_id, email,
	created_at, updated_at
FROM agreement_subjects
WHERE site_id = @site_id AND external_user_id = @external_user_id
LIMIT 1
`

	args := pgx.StrictNamedArgs{
		"site_id":          siteID,
		"external_user_id": externalUserID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement subject: %w", err)
	}

	sub, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AgreementSubject])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect agreement subject: %w", err)
	}

	*s = sub

	return nil
}

func (ss *AgreementSubjects) LoadBySiteID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	siteID gid.GID,
	cursor *page.Cursor[AgreementSubjectOrderField],
) error {
	q := `
SELECT
	id, organization_id, site_id, external_user_id, account_id, email,
	created_at, updated_at
FROM agreement_subjects
WHERE
	%s
	AND site_id = @site_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"site_id": siteID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query agreement subjects: %w", err)
	}

	subjects, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AgreementSubject])
	if err != nil {
		return fmt.Errorf("cannot collect agreement subjects: %w", err)
	}

	*ss = subjects

	return nil
}
