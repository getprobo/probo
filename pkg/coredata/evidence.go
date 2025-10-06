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
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Evidence struct {
		ID             gid.GID       `db:"id"`
		MeasureID      gid.GID       `db:"measure_id"`
		TaskID         *gid.GID      `db:"task_id"`
		State          EvidenceState `db:"state"`
		ReferenceID    string        `db:"reference_id"`
		Type           EvidenceType  `db:"type"`
		URL            string        `db:"url"`
		EvidenceFileId *gid.GID      `db:"evidence_file_id"`
		Description    string        `db:"description"`
		CreatedAt      time.Time     `db:"created_at"`
		UpdatedAt      time.Time     `db:"updated_at"`
	}

	Evidences []*Evidence
)

func (e Evidence) CursorKey(orderBy EvidenceOrderField) page.CursorKey {
	switch orderBy {
	case EvidenceOrderFieldCreatedAt:
		return page.NewCursorKey(e.ID, e.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (e Evidence) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    evidences (
        tenant_id,
        id,
        measure_id,
        task_id,
        reference_id,
        state,
        type,
        url,
        evidence_file_id,
        description,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @evidence_id,
    @measure_id,
    @task_id,
    @reference_id,
    @state,
    @type,
    @url,
    @evidence_file_id,
    @description,
    @created_at,
    @updated_at
)
ON CONFLICT (task_id, reference_id) DO UPDATE SET
	description = @description,
	type = @type,
	updated_at = @updated_at
WHERE evidences.state = 'REQUESTED';
`

	args := pgx.StrictNamedArgs{
		"tenant_id":        scope.GetTenantID(),
		"evidence_id":      e.ID,
		"measure_id":       e.MeasureID,
		"task_id":          e.TaskID,
		"reference_id":     e.ReferenceID,
		"evidence_file_id": e.EvidenceFileId,
		"created_at":       e.CreatedAt,
		"updated_at":       e.UpdatedAt,
		"state":            e.State,
		"type":             e.Type,
		"url":              e.URL,
		"description":      e.Description,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (e Evidence) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    evidences (
        tenant_id,
        id,
        measure_id,
        task_id,
        reference_id,
        state,
        type,
        url,
        evidence_file_id,
        description,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @evidence_id,
    @measure_id,
    @task_id,
    @reference_id,
    @state,
    @type,
    @url,
    @evidence_file_id,
    @description,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":        scope.GetTenantID(),
		"evidence_id":      e.ID,
		"measure_id":       e.MeasureID,
		"task_id":          e.TaskID,
		"reference_id":     e.ReferenceID,
		"evidence_file_id": e.EvidenceFileId,
		"created_at":       e.CreatedAt,
		"updated_at":       e.UpdatedAt,
		"state":            e.State,
		"type":             e.Type,
		"url":              e.URL,
		"description":      e.Description,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (e *Evidence) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	evidenceID gid.GID,
) error {
	q := `
SELECT
    id,
    task_id,
    measure_id,
    reference_id,
    state,
    type,
    url,
    evidence_file_id,
    description,
    created_at,
    updated_at
FROM
    evidences
WHERE
    %s
    AND id = @evidence_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"evidence_id": evidenceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query evidence: %w", err)
	}

	evidence, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Evidence])
	if err != nil {
		return fmt.Errorf("cannot collect evidence: %w", err)
	}

	*e = evidence

	return nil
}

func (e *Evidences) CountByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	evidences
WHERE
	%s
	AND measure_id = @measure_id
	`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot collect evidence: %w", err)
	}

	return count, nil
}

func (e *Evidences) LoadByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[EvidenceOrderField],
) error {
	q := `
SELECT
	id,
	measure_id,
	task_id,
	reference_id,
	state,
	type,
	url,
	evidence_file_id,
	description,
	created_at,
	updated_at
FROM
	evidences
WHERE
	%s
	AND measure_id = @measure_id
	AND %s
	`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query evidence: %w", err)
	}

	evidences, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Evidence])
	if err != nil {
		return fmt.Errorf("cannot collect evidence: %w", err)
	}

	*e = evidences

	return nil
}

func (e *Evidences) CountByTaskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	taskID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    evidences
WHERE
    %s
    AND task_id = @task_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot collect evidence: %w", err)
	}

	return count, nil
}

func (e *Evidences) LoadByTaskID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	taskID gid.GID,
	cursor *page.Cursor[EvidenceOrderField],
) error {
	q := `
SELECT
    id,
    measure_id,
    task_id,
    reference_id,
    state,
    type,
    url,
    evidence_file_id,
    description,
    created_at,
    updated_at
FROM
    evidences
WHERE
    %s
    AND task_id = @task_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"task_id": taskID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query evidence: %w", err)
	}

	evidences, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Evidence])
	if err != nil {
		return fmt.Errorf("cannot collect evidence: %w", err)
	}

	*e = evidences

	return nil
}

func (e Evidence) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
    evidences
SET
	type = @type,
	state = @state,
	evidence_file_id = @evidence_file_id,
	url = @url,
	description = @description,
	updated_at = @updated_at
WHERE
    %s
	AND id = @evidence_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"evidence_id":      e.ID,
		"type":             e.Type,
		"state":            e.State,
		"evidence_file_id": e.EvidenceFileId,
		"url":              e.URL,
		"description":      e.Description,
		"updated_at":       e.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (e Evidence) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) (*string, error) {
	q := `
DELETE FROM
    evidences
WHERE
    %s
    AND id = @evidence_id
RETURNING evidence_file_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"evidence_id": e.ID}
	maps.Copy(args, scope.SQLArguments())

	var evidenceFileId *gid.GID
	err := conn.QueryRow(ctx, q, args).Scan(&evidenceFileId)

	if err != nil {
		return nil, fmt.Errorf("failed to delete evidence: %w", err)
	}

	if evidenceFileId != nil {
		var err error
		var fileKey *string
		file := &File{ID: *evidenceFileId}
		if fileKey, err = file.HardDelete(ctx, conn, scope); err != nil {
			return nil, fmt.Errorf("failed to soft delete evidence file: %w", err)
		}
		return fileKey, nil

	}

	return nil, err
}
