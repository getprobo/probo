// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package iam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	emails "go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
)

const (
	logExportEmailExpiresIn = 24 * time.Hour
)

type (
	LogExportWorker struct {
		pg             *pg.Client
		fm             *filemanager.Service
		bucket         string
		baseURL        string
		logger         *log.Logger
		interval       time.Duration
		staleAfter     time.Duration
		maxConcurrency int
	}

	LogExportWorkerOption func(*LogExportWorker)
)

func WithLogExportWorkerInterval(d time.Duration) LogExportWorkerOption {
	return func(w *LogExportWorker) { w.interval = d }
}

func WithLogExportWorkerStaleAfter(d time.Duration) LogExportWorkerOption {
	return func(w *LogExportWorker) { w.staleAfter = d }
}

func WithLogExportWorkerMaxConcurrency(n int) LogExportWorkerOption {
	return func(w *LogExportWorker) {
		if n > 0 {
			w.maxConcurrency = n
		}
	}
}

func NewLogExportWorker(
	pgClient *pg.Client,
	fm *filemanager.Service,
	bucket string,
	baseURL string,
	logger *log.Logger,
	opts ...LogExportWorkerOption,
) *LogExportWorker {
	w := &LogExportWorker{
		pg:             pgClient,
		fm:             fm,
		bucket:         bucket,
		baseURL:        baseURL,
		logger:         logger,
		interval:       10 * time.Second,
		staleAfter:     5 * time.Minute,
		maxConcurrency: 3,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func (w *LogExportWorker) Run(ctx context.Context) error {
	var (
		wg     sync.WaitGroup
		sem    = make(chan struct{}, w.maxConcurrency)
		ticker = time.NewTicker(w.interval)
	)
	defer ticker.Stop()
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			nonCancelableCtx := context.WithoutCancel(ctx)
			w.recoverStaleRows(nonCancelableCtx)
			for {
				if err := w.processNext(ctx, sem, &wg); err != nil {
					if !errors.Is(err, coredata.ErrResourceNotFound) {
						w.logger.ErrorCtx(nonCancelableCtx, "cannot process log export", log.Error(err))
					}
					break
				}
			}
		}
	}
}

func (w *LogExportWorker) processNext(ctx context.Context, sem chan struct{}, wg *sync.WaitGroup) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	var (
		export           coredata.LogExport
		now              = time.Now()
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := w.pg.WithTx(
		nonCancelableCtx,
		func(tx pg.Conn) error {
			if err := export.LoadNextPendingForUpdateSkipLocked(nonCancelableCtx, tx); err != nil {
				return err
			}

			scope := coredata.NewScopeFromObjectID(export.ID)
			export.Status = coredata.LogExportStatusProcessing
			export.StartedAt = new(now)

			if err := export.Update(nonCancelableCtx, tx, scope); err != nil {
				return fmt.Errorf("cannot update log export: %w", err)
			}

			return nil
		},
	); err != nil {
		<-sem
		return err
	}

	wg.Add(1)
	go func(export coredata.LogExport) {
		defer wg.Done()
		defer func() { <-sem }()

		if err := w.handle(nonCancelableCtx, &export); err != nil {
			w.logger.ErrorCtx(
				nonCancelableCtx,
				"cannot handle log export",
				log.Error(err),
				log.String("log_export_id", export.ID.String()),
			)
			if err := w.handleFailure(nonCancelableCtx, &export, err); err != nil {
				w.logger.ErrorCtx(nonCancelableCtx, "cannot mark log export as failed", log.Error(err))
			}
		}
	}(export)

	return nil
}

func (w *LogExportWorker) handle(ctx context.Context, export *coredata.LogExport) error {
	scope := coredata.NewScopeFromObjectID(export.ID)

	fileKey := uuid.MustNewV4().String()
	now := time.Now()
	typeName := "audit-log"
	if export.Type == coredata.LogExportTypeSCIMEvent {
		typeName = "scim-event"
	}
	fileName := fmt.Sprintf(
		"%s-export-%s-to-%s.jsonl",
		typeName,
		export.FromTime.Format("2006-01-02"),
		export.ToTime.Format("2006-01-02"),
	)

	file := coredata.File{
		ID:             gid.New(export.ID.TenantID(), coredata.FileEntityType),
		OrganizationID: export.OrganizationID,
		BucketName:     w.bucket,
		MimeType:       "application/x-ndjson",
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
		fileSize, uploadErr = w.fm.PutFile(
			ctx,
			&file,
			pr,
			map[string]string{
				"type":            typeName + "-export",
				"log-export-id":   export.ID.String(),
				"organization-id": export.OrganizationID.String(),
			},
		)
		pr.CloseWithError(uploadErr)
	}()

	writeErr := w.streamJSONL(ctx, export, scope, pw)
	if writeErr != nil {
		_ = pw.CloseWithError(writeErr)
	} else {
		_ = pw.Close()
	}

	<-uploadDone

	if writeErr != nil {
		return fmt.Errorf("cannot write JSONL: %w", writeErr)
	}
	if uploadErr != nil {
		return fmt.Errorf("cannot upload file to S3: %w", uploadErr)
	}
	file.FileSize = fileSize

	downloadURL, err := w.fm.GenerateFileUrl(ctx, &file, logExportEmailExpiresIn)
	if err != nil {
		return fmt.Errorf("cannot generate download URL: %w", err)
	}

	emailPresenter := emails.NewPresenter(w.fm, w.bucket, w.baseURL, export.RecipientName)
	subject, textBody, htmlBody, err := emailPresenter.RenderLogExport(ctx, downloadURL)
	if err != nil {
		return fmt.Errorf("cannot render log export email: %w", err)
	}

	return w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			email := coredata.NewEmail(
				export.RecipientName,
				export.RecipientEmail,
				subject,
				textBody,
				htmlBody,
				nil,
			)
			if err := email.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			export.FileID = &file.ID
			export.Status = coredata.LogExportStatusCompleted
			export.CompletedAt = new(time.Now())
			if err := export.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update log export: %w", err)
			}

			return nil
		},
	)
}

func (w *LogExportWorker) streamJSONL(
	ctx context.Context,
	export *coredata.LogExport,
	scope coredata.Scoper,
	pw io.Writer,
) error {
	return w.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			enc := json.NewEncoder(pw)

			switch export.Type {
			case coredata.LogExportTypeAuditLog:
				for entry, err := range coredata.AuditLogEntriesByOrganizationIDAndTimeRange(
					ctx,
					conn,
					scope,
					export.OrganizationID,
					export.FromTime,
					export.ToTime,
				) {
					if err != nil {
						return err
					}
					if err := enc.Encode(entry); err != nil {
						return fmt.Errorf("cannot encode audit log entry: %w", err)
					}
				}
				return nil

			case coredata.LogExportTypeSCIMEvent:
				for event, err := range coredata.SCIMEventsByOrganizationIDAndTimeRange(
					ctx,
					conn,
					scope,
					export.OrganizationID,
					export.FromTime,
					export.ToTime,
				) {
					if err != nil {
						return err
					}
					if err := enc.Encode(event); err != nil {
						return fmt.Errorf("cannot encode SCIM event: %w", err)
					}
				}
				return nil

			default:
				return fmt.Errorf("unsupported log export type: %q", export.Type)
			}
		},
	)
}

func (w *LogExportWorker) handleFailure(ctx context.Context, export *coredata.LogExport, failureErr error) error {
	return w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			scope := coredata.NewScopeFromObjectID(export.ID)
			errorMsg := failureErr.Error()
			export.Error = &errorMsg
			export.Status = coredata.LogExportStatusFailed
			export.CompletedAt = new(time.Now())

			if err := export.Update(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update log export: %w", err)
			}

			return nil
		},
	)
}

func (w *LogExportWorker) recoverStaleRows(ctx context.Context) {
	if err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return coredata.ResetStaleLogExports(ctx, conn, w.staleAfter)
		},
	); err != nil {
		w.logger.ErrorCtx(ctx, "cannot recover stale log exports", log.Error(err))
	}
}
