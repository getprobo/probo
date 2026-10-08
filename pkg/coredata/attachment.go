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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	Attachment struct {
		FileID         gid.GID   `db:"file_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		ParentID       gid.GID   `db:"parent_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	Attachments []*Attachment
)

func (l Attachment) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO
    attachments (
        file_id,
        tenant_id,
        organization_id,
        document_version_id,
        task_id,
        task_comment_id,
        risk_analysis_id,
        created_at
    )
VALUES (
    @file_id,
    @tenant_id,
    @organization_id,
    @document_version_id,
    @task_id,
    @task_comment_id,
    @risk_analysis_id,
    @created_at
)
`

	var documentVersionID, taskID, taskCommentID, riskAnalysisID *gid.GID

	id := l.ParentID
	switch id.EntityType() {
	case DocumentVersionEntityType:
		documentVersionID = &id
	case TaskEntityType:
		taskID = &id
	case TaskCommentEntityType:
		taskCommentID = &id
	case RiskAnalysisEntityType:
		riskAnalysisID = &id
	}

	args := pgx.StrictNamedArgs{
		"file_id":             l.FileID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     l.OrganizationID,
		"document_version_id": documentVersionID,
		"task_id":             taskID,
		"task_comment_id":     taskCommentID,
		"risk_analysis_id":    riskAnalysisID,
		"created_at":          l.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert attachment: %w", err)
	}

	return nil
}

func (l *Attachments) LoadByFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	fileIDs []gid.GID,
) error {
	q := `
SELECT
    file_id,
    organization_id,
    parent_id,
    created_at
FROM
    attachments
WHERE
    %s
    AND file_id = ANY(@file_ids::text[])
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"file_ids": fileIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query attachments: %w", err)
	}

	attachments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Attachment])
	if err != nil {
		return fmt.Errorf("cannot collect attachments: %w", err)
	}

	*l = attachments

	return nil
}

func (l *Attachments) LoadByParentFileIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	parentID gid.GID,
	fileIDs []gid.GID,
) error {
	q := `
SELECT
    file_id,
    organization_id,
    parent_id,
    created_at
FROM
    attachments
WHERE
    %s
    AND organization_id = @organization_id
    AND parent_id = @parent_id
    AND file_id = ANY(@file_ids::text[])
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"parent_id":       parentID,
		"file_ids":        fileIDs,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query attachments: %w", err)
	}

	attachments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Attachment])
	if err != nil {
		return fmt.Errorf("cannot collect attachments: %w", err)
	}

	*l = attachments

	return nil
}

func (l Attachments) DeleteExcept(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	parentID gid.GID,
	keep []gid.GID,
) error {
	if keep == nil {
		keep = []gid.GID{}
	}

	q := `
DELETE FROM
    attachments
WHERE
    %s
    AND organization_id = @organization_id
    AND parent_id = @parent_id
    AND NOT (file_id = ANY(@file_ids::text[]))
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"parent_id":       parentID,
		"file_ids":        keep,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete attachments: %w", err)
	}

	return nil
}
