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
		DocumentIDs    []gid.GID  `json:"document_ids"`
		WithWatermark  bool       `json:"with_watermark"`
		WatermarkEmail *mail.Addr `json:"watermark_email"`
		WithSignatures bool       `json:"with_signatures"`
	}

	FrameworkExportArguments struct {
		FrameworkID gid.GID `json:"framework_id"`
	}
)

var (
	ErrNoExportJobAvailable = errors.New("no export job available")
)

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (ej *ExportJob) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM export_jobs WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, ej.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query export job authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (ej *ExportJob) Insert(
	ctx context.Context,
	conn pg.Conn,
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
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
	export_jobs
SET
	status = @status,
	error = @error,
	file_id = @file_id,
	started_at = @started_at,
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
		"started_at":   ej.StartedAt,
		"completed_at": ej.CompletedAt,
		"id":           ej.ID,
	}
	maps.Copy(args, scope.SQLArguments())
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (ej *ExportJob) LoadByID(
	ctx context.Context,
	conn pg.Conn,
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
	conn pg.Conn,
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
