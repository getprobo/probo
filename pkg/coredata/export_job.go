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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/mail"
)

type (
	ExportJob struct {
		ID             gid.GID         `db:"id"`
		OrganizationID gid.GID         `db:"organization_id"`
		Type           ExportJobType   `db:"type"`
		Arguments      json.RawMessage `db:"arguments"`
		Error          *string         `db:"error"`
		Status         ExportJobStatus `db:"status"`
		FileID         *gid.GID        `db:"file_id"`
		RecipientEmail mail.Addr       `db:"recipient_email"`
		RecipientName  string          `db:"recipient_name"`
		CreatedAt      time.Time       `db:"created_at"`
		StartedAt      *time.Time      `db:"started_at"`
		CompletedAt    *time.Time      `db:"completed_at"`
	}

	ExportJobs []*ExportJob

	DocumentExportArguments struct {
		DocumentIDs    []gid.GID `json:"document_ids"`
		WithWatermark  bool      `json:"with_watermark"`
		WatermarkText  *string   `json:"watermark_text"`
		WithSignatures bool      `json:"with_signatures"`
	}

	FrameworkExportArguments struct {
		FrameworkID gid.GID `json:"framework_id"`
	}

	LogExportArguments struct {
		FromTime time.Time `json:"from_time"`
		ToTime   time.Time `json:"to_time"`
	}
)

var (
	ErrNoExportJobAvailable = errors.New("no export job available")
)

func (a *DocumentExportArguments) UnmarshalJSON(data []byte) error {
	var arguments struct {
		DocumentIDs    []gid.GID `json:"document_ids"`
		WithWatermark  bool      `json:"with_watermark"`
		WatermarkText  *string   `json:"watermark_text"`
		WatermarkEmail *string   `json:"watermark_email"`
		WithSignatures bool      `json:"with_signatures"`
	}

	if err := json.Unmarshal(data, &arguments); err != nil {
		return fmt.Errorf("cannot unmarshal document export arguments: %w", err)
	}

	a.DocumentIDs = arguments.DocumentIDs
	a.WithWatermark = arguments.WithWatermark
	a.WatermarkText = arguments.WatermarkText
	if a.WatermarkText == nil {
		a.WatermarkText = arguments.WatermarkEmail
	}
	a.WithSignatures = arguments.WithSignatures

	return nil
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (ej *ExportJob) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM export_jobs WHERE id = ANY(@resource_ids::text[])`

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

func (ej *ExportJob) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO export_jobs (
	id,
	organization_id,
	tenant_id,
	type,
	arguments,
	status,
	recipient_email,
	recipient_name,
	created_at
) VALUES (
	@id,
	@organization_id,
	@tenant_id,
	@type,
	@arguments,
	@status,
	@recipient_email,
	@recipient_name,
	@created_at
)`
	args := pgx.StrictNamedArgs{
		"id":              ej.ID,
		"organization_id": ej.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"type":            ej.Type,
		"arguments":       ej.Arguments,
		"status":          ej.Status,
		"recipient_email": ej.RecipientEmail,
		"recipient_name":  ej.RecipientName,
		"created_at":      ej.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	return err
}

func (ej *ExportJob) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE
	export_jobs
SET
	status = @status,
	error = @error,
	file_id = @file_id,
	completed_at = @completed_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"status":       ej.Status,
		"error":        ej.Error,
		"file_id":      ej.FileID,
		"completed_at": ej.CompletedAt,
		"id":           ej.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update export job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// UpdateIfStatus updates the export job only when its current status matches
// expected. Used to finalize a claim without clobbering a job that was
// reclaimed and reassigned to another worker.
func (ej *ExportJob) UpdateIfStatus(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	expected ExportJobStatus,
) error {
	q := `
UPDATE
	export_jobs
SET
	status = @status,
	error = @error,
	file_id = @file_id,
	completed_at = @completed_at
WHERE
	%s
	AND id = @id
	AND status = @expected_status
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"status":          ej.Status,
		"error":           ej.Error,
		"file_id":         ej.FileID,
		"completed_at":    ej.CompletedAt,
		"id":              ej.ID,
		"expected_status": expected,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update export job: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// MarkProcessing claims the job for processing and starts its lease clock.
func (ej *ExportJob) MarkProcessing(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()
	ej.Status = ExportJobStatusProcessing
	ej.StartedAt = &now

	q := `
UPDATE
	export_jobs
SET
	status = @status,
	started_at = @started_at,
	error = NULL,
	completed_at = NULL
WHERE
	%s
	AND id = @id
	AND status = @pending_status
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"status":         ej.Status,
		"started_at":     ej.StartedAt,
		"id":             ej.ID,
		"pending_status": ExportJobStatusPending,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark export job as processing: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (ej *ExportJob) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	type,
	arguments,
	error,
	status,
	file_id,
	recipient_email,
	recipient_name,
	created_at,
	started_at,
	completed_at
FROM
	export_jobs
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return err
	}

	ej2, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ExportJob])
	if err != nil {
		return fmt.Errorf("cannot collect export job: %w", err)
	}

	*ej = ej2

	return nil
}

func (ej *ExportJob) LoadNextPendingForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
SELECT
	id,
	organization_id,
	type,
	arguments,
	error,
	status,
	file_id,
	recipient_email,
	recipient_name,
	created_at,
	started_at,
	completed_at
FROM
	export_jobs
WHERE
	status = @status
ORDER BY
	created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"status": ExportJobStatusPending,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return err
	}

	ej2, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ExportJob])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExportJobAvailable
		}

		return fmt.Errorf("cannot collect export job: %w", err)
	}

	*ej = ej2

	return nil
}

func (ej *ExportJob) GetDocumentExportArguments() (*DocumentExportArguments, error) {
	if ej.Type != ExportJobTypeDocument {
		return nil, fmt.Errorf("export job is not a document export")
	}

	var args DocumentExportArguments
	if err := json.Unmarshal(ej.Arguments, &args); err != nil {
		return nil, fmt.Errorf("cannot unmarshal document export arguments: %w", err)
	}

	return &args, nil
}

func (ej *ExportJob) GetFrameworkExportArguments() (*FrameworkExportArguments, error) {
	if ej.Type != ExportJobTypeFramework {
		return nil, fmt.Errorf("export job is not a framework export")
	}

	var args FrameworkExportArguments
	if err := json.Unmarshal(ej.Arguments, &args); err != nil {
		return nil, fmt.Errorf("cannot unmarshal framework export arguments: %w", err)
	}

	return &args, nil
}

func (ej *ExportJob) GetDocumentIDs() ([]gid.GID, error) {
	args, err := ej.GetDocumentExportArguments()
	if err != nil {
		return nil, err
	}

	return args.DocumentIDs, nil
}

func (ej *ExportJob) GetFrameworkID() (gid.GID, error) {
	args, err := ej.GetFrameworkExportArguments()
	if err != nil {
		return gid.GID{}, err
	}

	return args.FrameworkID, nil
}

func (ej *ExportJob) GetLogExportArguments() (*LogExportArguments, error) {
	switch ej.Type {
	case ExportJobTypeAuditLog, ExportJobTypeSCIMEvent:
	default:
		return nil, fmt.Errorf("export job is not a log export")
	}

	var args LogExportArguments
	if err := json.Unmarshal(ej.Arguments, &args); err != nil {
		return nil, fmt.Errorf("cannot unmarshal log export arguments: %w", err)
	}

	return &args, nil
}

func ResetStaleExportJobs(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	q := `
UPDATE export_jobs
SET
	status = @pending_status,
	started_at = NULL
WHERE
	status = @processing_status
	AND started_at < @stale_threshold
`

	args := pgx.StrictNamedArgs{
		"pending_status":    ExportJobStatusPending,
		"processing_status": ExportJobStatusProcessing,
		"stale_threshold":   time.Now().Add(-staleAfter),
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset stale export jobs: %w", err)
	}

	return nil
}

// TouchExportJobLease renews the processing lease so long-running exports
// are not requeued by ResetStaleExportJobs while still actively working.
func TouchExportJobLease(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
UPDATE export_jobs
SET
	started_at = @started_at
WHERE
	%s
	AND id = @id
	AND status = @processing_status
`

	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"id":                id,
		"started_at":        time.Now(),
		"processing_status": ExportJobStatusProcessing,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot touch export job lease: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
