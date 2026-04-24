// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	Evidence struct {
		ID                            gid.GID                  `db:"id"`
		OrganizationID                gid.GID                  `db:"organization_id"`
		MeasureID                     gid.GID                  `db:"measure_id"`
		TaskID                        *gid.GID                 `db:"task_id"`
		State                         EvidenceState            `db:"state"`
		ReferenceID                   string                   `db:"reference_id"`
		Type                          EvidenceType             `db:"type"`
		URL                           string                   `db:"url"`
		EvidenceFileId                *gid.GID                 `db:"evidence_file_id"`
		Description                   *string                  `db:"description"`
		Assessment                    jsonRawMessageOrNull     `db:"assessment"`
		AssessmentStatus              EvidenceAssessmentStatus `db:"assessment_status"`
		AssessmentProcessingStartedAt *time.Time               `db:"assessment_processing_started_at"`
		CreatedAt                     time.Time                `db:"created_at"`
		UpdatedAt                     time.Time                `db:"updated_at"`
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

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (e *Evidence) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM evidences WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, e.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query evidence authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (e Evidence) Upsert(
	ctx context.Context,
	conn pg.Querier,
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
        assessment,
        assessment_status,
        assessment_processing_started_at,
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
    @assessment,
    @assessment_status,
    @assessment_processing_started_at,
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
		"tenant_id":                        scope.GetTenantID(),
		"evidence_id":                      e.ID,
		"measure_id":                       e.MeasureID,
		"task_id":                          e.TaskID,
		"reference_id":                     e.ReferenceID,
		"evidence_file_id":                 e.EvidenceFileId,
		"created_at":                       e.CreatedAt,
		"updated_at":                       e.UpdatedAt,
		"state":                            e.State,
		"type":                             e.Type,
		"url":                              e.URL,
		"description":                      e.Description,
		"assessment":                       e.Assessment.Arg(),
		"assessment_status":                e.AssessmentStatus,
		"assessment_processing_started_at": e.AssessmentProcessingStartedAt,
	}
	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert evidence: %w", err)
	}
	return nil
}

func (e Evidence) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    evidences (
        tenant_id,
        id,
        organization_id,
        measure_id,
        task_id,
        reference_id,
        state,
        type,
        url,
        evidence_file_id,
        description,
        assessment,
        assessment_status,
        assessment_processing_started_at,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @evidence_id,
    @organization_id,
    @measure_id,
    @task_id,
    @reference_id,
    @state,
    @type,
    @url,
    @evidence_file_id,
    @description,
    @assessment,
    @assessment_status,
    @assessment_processing_started_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                        scope.GetTenantID(),
		"evidence_id":                      e.ID,
		"organization_id":                  e.OrganizationID,
		"measure_id":                       e.MeasureID,
		"task_id":                          e.TaskID,
		"reference_id":                     e.ReferenceID,
		"evidence_file_id":                 e.EvidenceFileId,
		"created_at":                       e.CreatedAt,
		"updated_at":                       e.UpdatedAt,
		"state":                            e.State,
		"type":                             e.Type,
		"url":                              e.URL,
		"description":                      e.Description,
		"assessment":                       e.Assessment.Arg(),
		"assessment_status":                e.AssessmentStatus,
		"assessment_processing_started_at": e.AssessmentProcessingStartedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "evidences_reference_id_key" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert evidence: %w", err)
	}

	return nil
}

func (e *Evidence) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	evidenceID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    task_id,
    measure_id,
    reference_id,
    state,
    type,
    url,
    evidence_file_id,
    description,
    assessment,
    assessment_status,
    assessment_processing_started_at,
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
	conn pg.Querier,
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
	conn pg.Querier,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[EvidenceOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	measure_id,
	task_id,
	reference_id,
	state,
	type,
	url,
	evidence_file_id,
	description,
	assessment,
	assessment_status,
	assessment_processing_started_at,
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
	conn pg.Querier,
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
	conn pg.Querier,
	scope Scoper,
	taskID gid.GID,
	cursor *page.Cursor[EvidenceOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    measure_id,
    task_id,
    reference_id,
    state,
    type,
    url,
    evidence_file_id,
    description,
    assessment,
    assessment_status,
    assessment_processing_started_at,
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
	conn pg.Tx,
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
	assessment = @assessment,
	assessment_status = @assessment_status,
	assessment_processing_started_at = @assessment_processing_started_at,
	updated_at = @updated_at
WHERE
    %s
	AND id = @evidence_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"evidence_id":                      e.ID,
		"type":                             e.Type,
		"state":                            e.State,
		"evidence_file_id":                 e.EvidenceFileId,
		"url":                              e.URL,
		"description":                      e.Description,
		"assessment":                       e.Assessment.Arg(),
		"assessment_status":                e.AssessmentStatus,
		"assessment_processing_started_at": e.AssessmentProcessingStartedAt,
		"updated_at":                       e.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot update evidence: %w", err)
	}
	return nil
}

func (e Evidence) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
    evidences
WHERE
    %s
    AND id = @evidence_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"evidence_id": e.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete evidence: %w", err)
	}

	return nil
}

func (e *Evidence) LoadNextPendingAssessmentForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
SELECT
    id,
    organization_id,
    task_id,
    measure_id,
    reference_id,
    state,
    type,
    url,
    evidence_file_id,
    description,
    assessment,
    assessment_status,
    assessment_processing_started_at,
    created_at,
    updated_at
FROM
    evidences
WHERE
    assessment_status = 'PENDING'
    AND evidence_file_id IS NOT NULL
ORDER BY
    created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query evidence: %w", err)
	}

	evidence, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Evidence])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect evidence: %w", err)
	}

	*e = evidence

	return nil
}

// MarkAssessmentFailed transitions the evidence row to FAILED without
// touching description, assessment, or any other column. The worker's
// in-memory evidence pointer may carry mutations from a partially
// successful assess-and-commit (e.g. a new assessment that could not be
// persisted); using the full Update method would write those back with
// FAILED status. This targeted statement guarantees the failure path
// only changes state-transition columns.
func (e Evidence) MarkAssessmentFailed(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
    evidences
SET
    assessment_status = @assessment_status,
    assessment_processing_started_at = NULL,
    updated_at = @updated_at
WHERE
    %s
    AND id = @evidence_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"evidence_id":       e.ID,
		"assessment_status": EvidenceAssessmentStatusFailed,
		"updated_at":        time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot mark evidence assessment failed: %w", err)
	}
	return nil
}

func ResetStaleAssessmentProcessing(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	q := `
UPDATE evidences
SET
    assessment_status = 'PENDING',
    assessment_processing_started_at = NULL
WHERE
    assessment_status = 'PROCESSING'
    AND assessment_processing_started_at < $1;
`

	if _, err := conn.Exec(ctx, q, time.Now().Add(-staleAfter)); err != nil {
		return fmt.Errorf("cannot reset stale assessment processing: %w", err)
	}
	return nil
}
