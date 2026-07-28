// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	Evidence struct {
		ID                            gid.GID                  `db:"id"`
		OrganizationID                gid.GID                  `db:"organization_id"`
		MeasureID                     gid.GID                  `db:"measure_id"`
		TaskID                        *gid.GID                 `db:"task_id"`
		State                         EvidenceState            `db:"state"`
		ReferenceID                   string                   `db:"reference_id"`
		Type                          EvidenceType             `db:"type"`
		URL                           string                   `db:"url"`
		EvidenceFileID                *gid.GID                 `db:"evidence_file_id"`
		Description                   *string                  `db:"description"`
		Assessment                    jsonRawMessageOrNull     `db:"assessment"`
		AssessmentStatus              EvidenceAssessmentStatus `db:"assessment_status"`
		AssessmentProcessingStartedAt *time.Time               `db:"assessment_processing_started_at"`
		AssessmentAttempts            int                      `db:"assessment_attempts"`
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
func (e *Evidence) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM evidences WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
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
        assessment_attempts,
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
    @assessment_attempts,
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
		"evidence_file_id":                 e.EvidenceFileID,
		"created_at":                       e.CreatedAt,
		"updated_at":                       e.UpdatedAt,
		"state":                            e.State,
		"type":                             e.Type,
		"url":                              e.URL,
		"description":                      e.Description,
		"assessment":                       e.Assessment,
		"assessment_status":                e.AssessmentStatus,
		"assessment_processing_started_at": e.AssessmentProcessingStartedAt,
		"assessment_attempts":              e.AssessmentAttempts,
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
        assessment_attempts,
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
    @assessment_attempts,
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
		"evidence_file_id":                 e.EvidenceFileID,
		"created_at":                       e.CreatedAt,
		"updated_at":                       e.UpdatedAt,
		"state":                            e.State,
		"type":                             e.Type,
		"url":                              e.URL,
		"description":                      e.Description,
		"assessment":                       e.Assessment,
		"assessment_status":                e.AssessmentStatus,
		"assessment_processing_started_at": e.AssessmentProcessingStartedAt,
		"assessment_attempts":              e.AssessmentAttempts,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
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
    assessment_attempts,
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
	assessment_attempts,
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
    assessment_attempts,
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
	assessment_attempts = @assessment_attempts,
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
		"evidence_file_id":                 e.EvidenceFileID,
		"url":                              e.URL,
		"description":                      e.Description,
		"assessment":                       e.Assessment,
		"assessment_status":                e.AssessmentStatus,
		"assessment_processing_started_at": e.AssessmentProcessingStartedAt,
		"assessment_attempts":              e.AssessmentAttempts,
		"updated_at":                       e.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot update evidence: %w", err)
	}

	return nil
}

// MarkAssessmentProcessing claims a PENDING row for assessment: bumps the
// attempt counter and flips status to PROCESSING. Touches only the
// assessment status columns so concurrent edits to other fields are safe.
func (e *Evidence) MarkAssessmentProcessing(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()
	q := `
UPDATE
    evidences
SET
    assessment_status = @assessment_status,
    assessment_processing_started_at = @started_at,
    assessment_attempts = assessment_attempts + 1,
    updated_at = @updated_at
WHERE
    %s
    AND id = @evidence_id
RETURNING
    assessment_attempts
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"evidence_id":       e.ID,
		"assessment_status": EvidenceAssessmentStatusProcessing,
		"started_at":        now,
		"updated_at":        now,
	}
	maps.Copy(args, scope.SQLArguments())

	var attempts int
	if err := conn.QueryRow(ctx, q, args).Scan(&attempts); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot mark evidence assessment processing: %w", err)
	}

	e.AssessmentStatus = EvidenceAssessmentStatusProcessing
	e.AssessmentProcessingStartedAt = &now
	e.AssessmentAttempts = attempts
	e.UpdatedAt = now

	return nil
}

// SetAssessmentCompleted persists a successful assessment result, touching
// only assessment/description/status columns. Avoids clobbering concurrent
// edits with a snapshot held across the preceding LLM call.
func (e Evidence) SetAssessmentCompleted(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
    evidences
SET
    assessment = @assessment,
    description = @description,
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
		"assessment":        e.Assessment,
		"description":       e.Description,
		"assessment_status": EvidenceAssessmentStatusCompleted,
		"updated_at":        time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark evidence assessment completed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// SetAssessmentOutcome records a failed attempt. When attempts is below
// maxAttempts the row returns to PENDING for retry; otherwise it is marked
// FAILED. Touches only the assessment status columns.
func (e Evidence) SetAssessmentOutcome(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	maxAttempts int,
) error {
	status := EvidenceAssessmentStatusFailed
	if maxAttempts > 0 && e.AssessmentAttempts < maxAttempts {
		status = EvidenceAssessmentStatusPending
	}

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
		"assessment_status": status,
		"updated_at":        time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark evidence assessment outcome: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
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
    assessment_attempts,
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
