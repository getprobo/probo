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

package esign

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"

	emails "go.probo.inc/probo/packages/emails"
)

// EmailPresenterConfigFunc resolves the emails.PresenterConfig for the
// organization that owns the given trust center.
type EmailPresenterConfigFunc func(ctx context.Context, organizationID gid.GID) (emails.PresenterConfig, error)

type (
	CompletionCertificateWorker struct {
		pg                  *pg.Client
		fileManager         *filemanager.Service
		certificateGen      *CertificateGenerator
		presenterConfigFunc EmailPresenterConfigFunc
		bucket              string
		logger              *log.Logger
		interval            time.Duration
		staleAfter          time.Duration
		maxConcurrency      int
	}

	CompletionCertificateWorkerOption func(*CompletionCertificateWorker)
)

const (
	certificateFilename = "certificate-of-completion.pdf"
)

func WithCompletionCertificateWorkerInterval(d time.Duration) CompletionCertificateWorkerOption {
	return func(w *CompletionCertificateWorker) { w.interval = d }
}

func WithCompletionCertificateWorkerStaleAfter(d time.Duration) CompletionCertificateWorkerOption {
	return func(w *CompletionCertificateWorker) { w.staleAfter = d }
}

func WithCompletionCertificateWorkerMaxConcurrency(n int) CompletionCertificateWorkerOption {
	return func(w *CompletionCertificateWorker) {
		if n > 0 {
			w.maxConcurrency = n
		}
	}
}

func NewCompletionCertificateWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	certificateGen *CertificateGenerator,
	presenterConfigFunc EmailPresenterConfigFunc,
	bucket string,
	logger *log.Logger,
	opts ...CompletionCertificateWorkerOption,
) *CompletionCertificateWorker {
	w := &CompletionCertificateWorker{
		pg:                  pgClient,
		fileManager:         fileManager,
		certificateGen:      certificateGen,
		presenterConfigFunc: presenterConfigFunc,
		bucket:              bucket,
		logger:              logger,
		interval:            10 * time.Second,
		staleAfter:          10 * time.Minute,
		maxConcurrency:      5,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *CompletionCertificateWorker) Run(ctx context.Context) error {
	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, w.maxConcurrency)
	)
	defer wg.Wait()

LOOP:
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(w.interval):
		// From there we should not accept cancelations anymore.
		nonCancelableCtx := context.WithoutCancel(ctx)
		w.recoverStaleCertificateRows(nonCancelableCtx)
		for {
			if err := w.processNext(ctx, sem, &wg); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					w.logger.ErrorCtx(ctx, "cannot process certificate", log.Error(err))
				}
				break
			}
		}
		goto LOOP
	}
}

func (w *CompletionCertificateWorker) processNext(ctx context.Context, sem chan struct{}, wg *sync.WaitGroup) error {
	select {
	case sem <- struct{}{}:
	case <-ctx.Done(): // FIXME: this will never be fired
		return ctx.Err()
	}

	var (
		signature coredata.ElectronicSignature
		now       = time.Now()

		// From there we should not accept cancelations anymore.
		nonCancelableCtx = context.WithoutCancel(ctx)
	)

	if err := w.pg.WithTx(
		nonCancelableCtx,
		func(tx pg.Conn) error {
			if err := signature.LoadNextCompletedWithoutCertificateForUpdate(nonCancelableCtx, tx); err != nil {
				return err
			}
			scope := coredata.NewScopeFromObjectID(signature.ID)
			signature.CertificateProcessingStartedAt = &now
			signature.UpdatedAt = now

			if err := signature.Update(nonCancelableCtx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature: %w", err)
			}

			return nil
		},
	); err != nil {
		<-sem
		return err
	}

	wg.Add(1)
	go func(signature coredata.ElectronicSignature) {
		defer wg.Done()
		defer func() { <-sem }()

		scope := coredata.NewScopeFromObjectID(signature.ID)

		if err := w.generateAndCommit(nonCancelableCtx, &signature); err != nil {
			if err := w.handleCertFailure(nonCancelableCtx, &signature, scope, err); err != nil {
				w.logger.ErrorCtx(nonCancelableCtx, "cannot handle certificate failure", log.Error(err))
			}
		}
	}(signature)

	return nil
}

func (w *CompletionCertificateWorker) generateAndCommit(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
) error {
	var (
		scope = coredata.NewScopeFromObjectID(signature.ID)
	)

	email, attachments, err := w.generateCertificate(ctx, signature, scope)
	if err != nil {
		return err
	}

	if err := w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			signature.CertificateFileID = &attachments[1].FileID
			signature.UpdatedAt = time.Now()
			if err := signature.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature: %w", err)
			}

			event := signature.NewEvent(
				coredata.ElectronicSignatureEventTypeCertificateGenerated,
				coredata.ElectronicSignatureEventSourceServer,
			)
			if err := event.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert certificate event: %w", err)
			}

			if err := email.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert certificate email: %w", err)
			}

			for _, attachment := range attachments {
				if err := attachment.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert email attachment: %w", err)
				}
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
}

func (w *CompletionCertificateWorker) generateCertificate(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
	scope coredata.Scoper,
) (*coredata.Email, coredata.EmailAttachments, error) {
	var (
		events     = coredata.ElectronicSignatureEvents{}
		signedFile = coredata.File{}
	)

	if err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := events.LoadBySignatureID(ctx, conn, scope, signature.ID); err != nil {
				return fmt.Errorf("cannot load events: %w", err)
			}

			if err := signedFile.LoadByID(ctx, conn, scope, signature.FileID); err != nil {
				return fmt.Errorf("cannot load signed file: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	certificatePDFReader, err := w.certificateGen.Generate(ctx, signature, events)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate certificate: %w", err)
	}

	certificateOfCompletionFile := coredata.File{
		ID:             gid.New(scope.GetTenantID(), coredata.FileEntityType),
		OrganizationID: signature.OrganizationID,
		BucketName:     w.bucket,
		MimeType:       "application/pdf",
		FileName:       certificateFilename,
		FileKey:        uuid.MustNewV4().String(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	certificateOfCompletionFileSize, err := w.fileManager.PutFile(
		ctx,
		&certificateOfCompletionFile,
		certificatePDFReader,
		map[string]string{
			"type":         "certificate-of-completion",
			"signature-id": signature.ID.String(),
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot upload cert to S3: %w", err)
	}

	certificateOfCompletionFile.FileSize = certificateOfCompletionFileSize

	if err := w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := certificateOfCompletionFile.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert certificate of completion file: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	presenterCfg, err := w.presenterConfigFunc(ctx, signature.OrganizationID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot resolve presenter config: %w", err)
	}
	emailPresenter := emails.NewPresenterFromConfig(w.fileManager, presenterCfg, ref.UnrefOrZero(signature.SignerFullName))

	docTypeName := signature.DocumentType.DisplayName()
	subject, textBody, htmlBody, err := emailPresenter.RenderElectronicSignatureCertificate(ctx, ref.UnrefOrZero(signature.SignerFullName), docTypeName)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot render email: %w", err)
	}

	email := coredata.NewEmail(
		ref.UnrefOrZero(signature.SignerFullName),
		mail.Addr(signature.SignerEmail),
		subject,
		textBody,
		htmlBody,
	)

	attachments := coredata.EmailAttachments{
		coredata.NewEmailAttachment(
			email.ID,
			signedFile.ID,
			signedFile.FileName,
			signedFile.MimeType,
		),
		coredata.NewEmailAttachment(
			email.ID,
			certificateOfCompletionFile.ID,
			certificateFilename,
			"application/pdf",
		),
	}

	return email, attachments, nil
}

func (w *CompletionCertificateWorker) handleCertFailure(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
	scope coredata.Scoper,
	processingError error,
) error {
	w.logger.ErrorCtx(ctx, "certificate worker failure",
		log.Error(processingError),
		log.String("signature_id", signature.ID.String()),
	)

	return w.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			signature.CertificateProcessingStartedAt = nil
			signature.UpdatedAt = time.Now()

			if err := signature.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature: %w", err)
			}

			return nil
		},
	)
}

func (w *CompletionCertificateWorker) recoverStaleCertificateRows(ctx context.Context) {
	if err := w.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return coredata.ResetStaleCertificateProcessing(ctx, conn, w.staleAfter)
		},
	); err != nil {
		w.logger.ErrorCtx(ctx, "cannot recover stale certificates", log.Error(err))
	}
}
