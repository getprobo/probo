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
	DocumentSuggestionStatus string

	DocumentSuggestion struct {
		ID                  gid.GID                  `db:"id"`
		OrganizationID      gid.GID                  `db:"organization_id"`
		DocumentID          gid.GID                  `db:"document_id"`
		AuthorProfileID     gid.GID                  `db:"author_profile_id"`
		Status              DocumentSuggestionStatus `db:"status"`
		CreatedAt           time.Time                `db:"created_at"`
		UpdatedAt           time.Time                `db:"updated_at"`
		ResolvedAt          *time.Time               `db:"resolved_at"`
		ResolvedByProfileID *gid.GID                 `db:"resolved_by_profile_id"`
	}

	DocumentSuggestions []*DocumentSuggestion
)

const (
	DocumentSuggestionStatusOpen     DocumentSuggestionStatus = "OPEN"
	DocumentSuggestionStatusAccepted DocumentSuggestionStatus = "ACCEPTED"
	DocumentSuggestionStatusRejected DocumentSuggestionStatus = "REJECTED"
)

func (s DocumentSuggestion) CursorKey(orderBy DocumentSuggestionOrderField) page.CursorKey {
	switch orderBy {
	case DocumentSuggestionOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *DocumentSuggestion) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    organization_id
FROM
    document_suggestions
WHERE
    id = ANY(@resource_ids)
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query suggestion authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan suggestion authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate suggestion authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (s *DocumentSuggestion) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	suggestionID gid.GID,
) error {
	return s.loadByID(ctx, conn, scope, suggestionID, false)
}

// LoadByIDForUpdate is LoadByID under FOR UPDATE so concurrent accept/reject
// decisions serialize on the same row.
func (s *DocumentSuggestion) LoadByIDForUpdate(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	suggestionID gid.GID,
) error {
	return s.loadByID(ctx, tx, scope, suggestionID, true)
}

func (s *DocumentSuggestion) loadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	suggestionID gid.GID,
	forUpdate bool,
) error {
	q := `
SELECT
    id,
    organization_id,
    document_id,
    author_profile_id,
    status,
    created_at,
    updated_at,
    resolved_at,
    resolved_by_profile_id
FROM
    document_suggestions
WHERE
    %s
    AND id = @suggestion_id
LIMIT 1
`

	if forUpdate {
		q += "\nFOR UPDATE"
	}

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"suggestion_id": suggestionID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document suggestion: %w", err)
	}

	suggestion, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DocumentSuggestion])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect document suggestion: %w", err)
	}

	*s = suggestion

	return nil
}

func (s *DocumentSuggestions) LoadByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	cursor *page.Cursor[DocumentSuggestionOrderField],
	filter *DocumentSuggestionFilter,
) error {
	if filter == nil {
		filter = NewDocumentSuggestionFilter(nil)
	}

	q := `
SELECT
    id,
    organization_id,
    document_id,
    author_profile_id,
    status,
    created_at,
    updated_at,
    resolved_at,
    resolved_by_profile_id
FROM
    document_suggestions
WHERE
    %s
    AND document_id = @document_id
    AND %s
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query document suggestions: %w", err)
	}

	suggestions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentSuggestion])
	if err != nil {
		return fmt.Errorf("cannot collect document suggestions: %w", err)
	}

	*s = suggestions

	return nil
}

func (s *DocumentSuggestions) CountByDocumentID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	filter *DocumentSuggestionFilter,
) (int, error) {
	if filter == nil {
		filter = NewDocumentSuggestionFilter(nil)
	}

	q := `
SELECT
    COUNT(id)
FROM
    document_suggestions
WHERE
    %s
    AND document_id = @document_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int

	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count document suggestions: %w", err)
	}

	return count, nil
}

func (s DocumentSuggestion) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    document_suggestions (
        tenant_id,
        id,
        organization_id,
        document_id,
        author_profile_id,
        status,
        created_at,
        updated_at,
        resolved_at,
        resolved_by_profile_id
    )
VALUES (
    @tenant_id,
    @suggestion_id,
    @organization_id,
    @document_id,
    @author_profile_id,
    @status,
    @created_at,
    @updated_at,
    @resolved_at,
    @resolved_by_profile_id
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":              scope.GetTenantID(),
		"suggestion_id":          s.ID,
		"organization_id":        s.OrganizationID,
		"document_id":            s.DocumentID,
		"author_profile_id":      s.AuthorProfileID,
		"status":                 s.Status,
		"created_at":             s.CreatedAt,
		"updated_at":             s.UpdatedAt,
		"resolved_at":            s.ResolvedAt,
		"resolved_by_profile_id": s.ResolvedByProfileID,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert document suggestion: %w", err)
	}

	return nil
}

func (s DocumentSuggestion) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	return s.update(ctx, conn, scope, false)
}

// UpdateIfOpen resolves an open suggestion. Zero rows means another decision
// already closed it.
func (s DocumentSuggestion) UpdateIfOpen(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	return s.update(ctx, conn, scope, true)
}

func (s DocumentSuggestion) update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	onlyOpen bool,
) error {
	q := `
UPDATE
    document_suggestions
SET
    status = @status,
    updated_at = @updated_at,
    resolved_at = @resolved_at,
    resolved_by_profile_id = @resolved_by_profile_id
WHERE
    %s
    AND id = @suggestion_id
`

	if onlyOpen {
		q += "\n    AND status = @open_status"
	}

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"suggestion_id":          s.ID,
		"status":                 s.Status,
		"updated_at":             s.UpdatedAt,
		"resolved_at":            s.ResolvedAt,
		"resolved_by_profile_id": s.ResolvedByProfileID,
	}
	if onlyOpen {
		args["open_status"] = DocumentSuggestionStatusOpen
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update document suggestion: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
