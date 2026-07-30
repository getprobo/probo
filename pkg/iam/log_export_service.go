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

package iam

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/safecsv"
)

type LogExportService struct {
	pg      *pg.Client
	fm      *filemanager.Service
	bucket  string
	baseURL string
}

func NewLogExportService(
	pgClient *pg.Client,
	fm *filemanager.Service,
	bucket string,
	baseURL string,
) *LogExportService {
	return &LogExportService{
		pg:      pgClient,
		fm:      fm,
		bucket:  bucket,
		baseURL: baseURL,
	}
}

func (s *LogExportService) BuildAndUploadExport(
	ctx context.Context,
	scope coredata.Scoper,
	exportJobID gid.GID,
) (*coredata.ExportJob, error) {
	exportJob := &coredata.ExportJob{}

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return exportJob.LoadByID(ctx, conn, scope, exportJobID)
		},
	); err != nil {
		return nil, fmt.Errorf("cannot load export job: %w", err)
	}

	args, err := exportJob.GetLogExportArguments()
	if err != nil {
		return nil, err
	}

	typeName := "audit-log"
	if exportJob.Type == coredata.ExportJobTypeSCIMEvent {
		typeName = "scim-event"
	}

	now := time.Now()
	fileKey := uuid.MustNewV4().String()
	fileName := fmt.Sprintf(
		"%s-export-%s-to-%s.csv",
		typeName,
		args.FromTime.Format("2006-01-02"),
		args.ToTime.Format("2006-01-02"),
	)

	file := coredata.File{
		ID:             gid.New(exportJob.ID.TenantID(), coredata.FileEntityType),
		OrganizationID: exportJob.OrganizationID,
		BucketName:     s.bucket,
		MimeType:       "text/csv",
		FileName:       fileName,
		FileKey:        fileKey,
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	pr, pw := io.Pipe()

	var uploadErr error

	var fileSize int64

	uploadDone := make(chan struct{})

	go func() {
		defer close(uploadDone)

		fileSize, uploadErr = s.fm.PutFile(
			ctx,
			&file,
			pr,
			map[string]string{
				"type":            typeName + "-export",
				"export-job-id":   exportJob.ID.String(),
				"organization-id": exportJob.OrganizationID.String(),
			},
			filemanager.WithAttachmentContentDisposition(),
		)
		_ = pr.CloseWithError(uploadErr)
	}()

	writeErr := s.streamCSV(ctx, exportJob, args, scope, pw)
	if writeErr != nil {
		_ = pw.CloseWithError(writeErr)
	} else {
		_ = pw.Close()
	}

	<-uploadDone

	if writeErr != nil {
		return nil, fmt.Errorf("cannot write CSV: %w", writeErr)
	}

	if uploadErr != nil {
		return nil, fmt.Errorf("cannot upload file to S3: %w", uploadErr)
	}

	file.FileSize = fileSize

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			exportJob.FileID = &file.ID
			if err := exportJob.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update export job: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return exportJob, nil
}

func (s *LogExportService) SendExportEmail(
	ctx context.Context,
	scope coredata.Scoper,
	fileID gid.GID,
	recipientName string,
	recipientEmail mail.Addr,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			file := &coredata.File{}
			if err := file.LoadByID(ctx, tx, scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			downloadURL := s.fm.GenerateFileURL(file)

			emailPresenter := emails.NewPresenter(s.baseURL, recipientName)

			isSCIMEvent := strings.HasPrefix(file.FileName, "scim-event-export")

			subject, textBody, htmlBody, err := emailPresenter.RenderLogExport(ctx, downloadURL, isSCIMEvent)
			if err != nil {
				return fmt.Errorf("cannot render log export email: %w", err)
			}

			email := coredata.NewEmail(
				recipientName,
				recipientEmail,
				subject,
				textBody,
				htmlBody,
				nil,
			)
			if err := email.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)
}

func (s *LogExportService) streamCSV(
	ctx context.Context,
	exportJob *coredata.ExportJob,
	args *coredata.LogExportArguments,
	scope coredata.Scoper,
	pw io.Writer,
) error {
	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, exportJob.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization for log export: %w", err)
			}

			w := safecsv.NewWriter(pw)

			switch exportJob.Type {
			case coredata.ExportJobTypeAuditLog:
				filter := coredata.NewAuditLogEntryFilter().
					WithCreatedAtGte(args.FromTime).
					WithCreatedAtLt(args.ToTime)

				return s.streamAuditLogCSV(
					ctx,
					conn,
					scope,
					exportJob.OrganizationID,
					organization.Name,
					filter,
					w,
				)
			case coredata.ExportJobTypeSCIMEvent:
				filter := coredata.NewSCIMEventFilter().
					WithCreatedAtGte(args.FromTime).
					WithCreatedAtLt(args.ToTime)

				return s.streamSCIMEventCSV(
					ctx,
					conn,
					scope,
					exportJob.OrganizationID,
					organization.Name,
					filter,
					w,
				)
			default:
				return fmt.Errorf("unsupported log export type: %q", exportJob.Type)
			}
		},
	)
}
