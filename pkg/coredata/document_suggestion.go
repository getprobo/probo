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
	"encoding"
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
		AnchorStart         []byte                   `db:"anchor_start"`
		AnchorEnd           []byte                   `db:"anchor_end"`
		AnchorQuote         *string                  `db:"anchor_quote"`
	}

	DocumentSuggestions []*DocumentSuggestion
)

const (
	DocumentSuggestionStatusOpen     DocumentSuggestionStatus = "OPEN"
	DocumentSuggestionStatusAccepted DocumentSuggestionStatus = "ACCEPTED"
	DocumentSuggestionStatusRejected DocumentSuggestionStatus = "REJECTED"
)

var (
	_ fmt.Stringer             = DocumentSuggestionStatus("")
	_ encoding.TextMarshaler   = DocumentSuggestionStatus("")
	_ encoding.TextUnmarshaler = (*DocumentSuggestionStatus)(nil)
)

func DocumentSuggestionStatuses() []DocumentSuggestionStatus {
	return []DocumentSuggestionStatus{
		DocumentSuggestionStatusOpen,
		DocumentSuggestionStatusAccepted,
		DocumentSuggestionStatusRejected,
	}
}

func (v DocumentSuggestionStatus) IsValid() bool {
	switch v {
	case
		DocumentSuggestionStatusOpen,
		DocumentSuggestionStatusAccepted,
		DocumentSuggestionStatusRejected:
		return true
	}

	return false
}

func (v DocumentSuggestionStatus) String() string {
	return string(v)
}

func (v DocumentSuggestionStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *DocumentSuggestionStatus) UnmarshalText(text []byte) error {
	val := DocumentSuggestionStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid DocumentSuggestionStatus value: %q", string(text))
	}

	*v = val

	return nil
}

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
    resolved_by_profile_id,
    anchor_start,
    anchor_end,
    anchor_quote
FROM
    document_suggestions
WHERE
    %s
    AND id = @suggestion_id
LIMIT 1
`

	return s.collectByID(ctx, conn, scope, suggestionID, q)
}

// LoadByIDForUpdate is LoadByID under FOR UPDATE so concurrent accept/reject
// decisions serialize on the same row.
func (s *DocumentSuggestion) LoadByIDForUpdate(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	suggestionID gid.GID,
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
    resolved_by_profile_id,
    anchor_start,
    anchor_end,
    anchor_quote
FROM
    document_suggestions
WHERE
    %s
    AND id = @suggestion_id
LIMIT 1
FOR UPDATE
`

	return s.collectByID(ctx, tx, scope, suggestionID, q)
}

func (s *DocumentSuggestion) collectByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	suggestionID gid.GID,
	q string,
) error {
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

// LoadMarkableByIDs loads the suggestions whose marks may be in the text: open
// ones, and automatically rejected ones, which reopen when their marks return.
func (s *DocumentSuggestions) LoadMarkableByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	documentID gid.GID,
	suggestionIDs []gid.GID,
) error {
	if len(suggestionIDs) == 0 {
		*s = nil
		return nil
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
    resolved_by_profile_id,
    anchor_start,
    anchor_end,
    anchor_quote
FROM
    document_suggestions
WHERE
    %s
    AND document_id = @document_id
    AND (status = @open_status OR auto_rejected_at IS NOT NULL)
    AND id = ANY(@suggestion_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":    documentID,
		"open_status":    DocumentSuggestionStatusOpen,
		"suggestion_ids": suggestionIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query markable document suggestions: %w", err)
	}

	suggestions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DocumentSuggestion])
	if err != nil {
		return fmt.Errorf("cannot collect markable document suggestions: %w", err)
	}

	*s = suggestions

	return nil
}

func (s *DocumentSuggestions) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	suggestionIDs []gid.GID,
) error {
	if len(suggestionIDs) == 0 {
		*s = nil
		return nil
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
    resolved_by_profile_id,
    anchor_start,
    anchor_end,
    anchor_quote
FROM
    document_suggestions
WHERE
    %s
    AND id = ANY(@suggestion_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"suggestion_ids": suggestionIDs}
	maps.Copy(args, scope.SQLArguments())

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
    resolved_by_profile_id,
    anchor_start,
    anchor_end,
    anchor_quote
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
        resolved_by_profile_id,
        anchor_start,
        anchor_end,
        anchor_quote
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
    @resolved_by_profile_id,
    @anchor_start,
    @anchor_end,
    @anchor_quote
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
		"anchor_start":           s.AnchorStart,
		"anchor_end":             s.AnchorEnd,
		"anchor_quote":           s.AnchorQuote,
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
    resolved_by_profile_id = @resolved_by_profile_id,
    anchor_start = @anchor_start,
    anchor_end = @anchor_end,
    anchor_quote = @anchor_quote
WHERE
    %s
    AND id = @suggestion_id
    AND (@only_open::boolean = FALSE OR status = @open_status)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"suggestion_id":          s.ID,
		"status":                 s.Status,
		"updated_at":             s.UpdatedAt,
		"resolved_at":            s.ResolvedAt,
		"resolved_by_profile_id": s.ResolvedByProfileID,
		"anchor_start":           s.AnchorStart,
		"anchor_end":             s.AnchorEnd,
		"anchor_quote":           s.AnchorQuote,
		"only_open":              onlyOpen,
		"open_status":            DocumentSuggestionStatusOpen,
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

// MarkDocumentSuggestionsSeen records that the marks of the given suggestions
// reached the persisted text. Rows another transaction holds are skipped; a
// later persist marks them.
func MarkDocumentSuggestionsSeen(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentID gid.GID,
	suggestionIDs []gid.GID,
	now time.Time,
) error {
	if len(suggestionIDs) == 0 {
		return nil
	}

	q := `
UPDATE
    document_suggestions
SET
    marked_at = @now
WHERE
    id IN (
        SELECT
            id
        FROM
            document_suggestions
        WHERE
            %s
            AND document_id = @document_id
            AND id = ANY(@suggestion_ids)
            AND marked_at IS NULL
        FOR UPDATE SKIP LOCKED
    )
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":    documentID,
		"suggestion_ids": suggestionIDs,
		"now":            now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot mark document suggestions seen: %w", err)
	}

	return nil
}

// AutoRejectDocumentSuggestions rejects the open suggestions of a document
// whose marks were persisted once and are no longer among presentIDs: they
// amount to no change. Rows another transaction holds, such as a suggestion
// being decided, are skipped. It returns the rejected ids.
func AutoRejectDocumentSuggestions(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentID gid.GID,
	presentIDs []gid.GID,
	now time.Time,
) ([]gid.GID, error) {
	if presentIDs == nil {
		// A nil array is NULL in SQL, and NOT (id = ANY(NULL)) matches nothing.
		presentIDs = []gid.GID{}
	}

	q := `
UPDATE
    document_suggestions
SET
    status = @rejected_status,
    updated_at = @now,
    resolved_at = @now,
    resolved_by_profile_id = NULL,
    auto_rejected_at = @now,
    anchor_start = NULL,
    anchor_end = NULL,
    anchor_quote = ''
WHERE
    id IN (
        SELECT
            id
        FROM
            document_suggestions
        WHERE
            %s
            AND document_id = @document_id
            AND status = @open_status
            AND marked_at IS NOT NULL
            AND NOT (id = ANY(@present_ids))
        FOR UPDATE SKIP LOCKED
    )
RETURNING
    id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":     documentID,
		"present_ids":     presentIDs,
		"open_status":     DocumentSuggestionStatusOpen,
		"rejected_status": DocumentSuggestionStatusRejected,
		"now":             now,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot auto-reject document suggestions: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect auto-rejected document suggestions: %w", err)
	}

	return ids, nil
}

// ReopenAutoRejectedDocumentSuggestions reopens the automatically rejected
// suggestions whose marks are back among presentIDs, as after an undo is
// redone. Suggestions a reviewer rejected stay rejected. It returns the
// reopened ids.
func ReopenAutoRejectedDocumentSuggestions(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	documentID gid.GID,
	presentIDs []gid.GID,
	now time.Time,
) ([]gid.GID, error) {
	if len(presentIDs) == 0 {
		return nil, nil
	}

	q := `
UPDATE
    document_suggestions
SET
    status = @open_status,
    updated_at = @now,
    resolved_at = NULL,
    resolved_by_profile_id = NULL,
    auto_rejected_at = NULL,
    anchor_start = NULL,
    anchor_end = NULL,
    anchor_quote = NULL
WHERE
    id IN (
        SELECT
            id
        FROM
            document_suggestions
        WHERE
            %s
            AND document_id = @document_id
            AND status = @rejected_status
            AND auto_rejected_at IS NOT NULL
            AND id = ANY(@present_ids)
        FOR UPDATE SKIP LOCKED
    )
RETURNING
    id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"document_id":     documentID,
		"present_ids":     presentIDs,
		"open_status":     DocumentSuggestionStatusOpen,
		"rejected_status": DocumentSuggestionStatusRejected,
		"now":             now,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot reopen document suggestions: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[gid.GID])
	if err != nil {
		return nil, fmt.Errorf("cannot collect reopened document suggestions: %w", err)
	}

	return ids, nil
}
