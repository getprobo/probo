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

package esign

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
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

type completionCertificateHandler struct {
	pg                  *pg.Client
	fileManager         *filemanager.Service
	certificateGen      *CertificateGenerator
	presenterConfigFunc EmailPresenterConfigFunc
	bucket              string
	logger              *log.Logger
	staleAfter          time.Duration
}

const (
	certificateFilename = "certificate-of-completion.pdf"
)

func NewCompletionCertificateWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	certificateGen *CertificateGenerator,
	presenterConfigFunc EmailPresenterConfigFunc,
	bucket string,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.ElectronicSignature] {
	h := &completionCertificateHandler{
		pg:                  pgClient,
		fileManager:         fileManager,
		certificateGen:      certificateGen,
		presenterConfigFunc: presenterConfigFunc,
		bucket:              bucket,
		logger:              logger,
		staleAfter:          10 * time.Minute,
	}

	return worker.New(
		"completion-certificate-worker",
		h,
		logger,
		opts...,
	)
}

func (h *completionCertificateHandler) Claim(ctx context.Context) (coredata.ElectronicSignature, error) {
	var signature coredata.ElectronicSignature

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := signature.LoadNextCompletedWithoutCertificateForUpdate(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			scope := coredata.NewScopeFromObjectID(signature.ID)
			signature.CertificateProcessingStartedAt = &now
			signature.AttemptCount++
			signature.LastAttemptedAt = &now
			signature.UpdatedAt = now

			if err := signature.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.ElectronicSignature{}, worker.ErrNoTask
		}
		return coredata.ElectronicSignature{}, err
	}

	return signature, nil
}

func (h *completionCertificateHandler) Process(ctx context.Context, signature coredata.ElectronicSignature) error {
	scope := coredata.NewScopeFromObjectID(signature.ID)

	if err := h.generateAndCommit(ctx, &signature); err != nil {
		if err := h.handleCertFailure(ctx, &signature, scope, err); err != nil {
			h.logger.ErrorCtx(ctx, "cannot handle certificate failure", log.Error(err))
		}
		return err
	}

	return nil
}

func (h *completionCertificateHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleCertificateProcessing(ctx, conn, h.staleAfter)
		},
	)
}

func (h *completionCertificateHandler) generateAndCommit(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
) error {
	scope := coredata.NewScopeFromObjectID(signature.ID)

	email, attachments, err := h.generateCertificate(ctx, signature, scope)
	if err != nil {
		return err
	}

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
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

func (h *completionCertificateHandler) generateCertificate(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
	scope coredata.Scoper,
) (*coredata.Email, coredata.EmailAttachments, error) {
	var (
		events       = coredata.ElectronicSignatureEvents{}
		signedFile   = coredata.File{}
		organization = coredata.Organization{}
	)

	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := events.LoadBySignatureID(ctx, conn, scope, signature.ID); err != nil {
				return fmt.Errorf("cannot load events: %w", err)
			}

			if err := signedFile.LoadByID(ctx, conn, scope, signature.FileID); err != nil {
				return fmt.Errorf("cannot load signed file: %w", err)
			}

			if err := organization.LoadByID(ctx, conn, scope, signature.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	certificatePDFReader, err := h.certificateGen.Generate(ctx, signature, events)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate certificate: %w", err)
	}

	certificateOfCompletionFile := coredata.File{
		ID:             gid.New(scope.GetTenantID(), coredata.FileEntityType),
		OrganizationID: signature.OrganizationID,
		BucketName:     h.bucket,
		MimeType:       "application/pdf",
		FileName:       certificateFilename,
		FileKey:        uuid.MustNewV4().String(),
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	certificateOfCompletionFileSize, err := h.fileManager.PutFile(
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

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := certificateOfCompletionFile.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert certificate of completion file: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	presenterCfg, err := h.presenterConfigFunc(ctx, signature.OrganizationID)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot resolve presenter config: %w", err)
	}
	emailPresenter := emails.NewPresenterFromConfig(h.fileManager, presenterCfg, ref.UnrefOrZero(signature.SignerFullName))

	docName := ref.UnrefOrZero(signature.DocumentName)
	if docName == "" {
		docName = signature.DocumentType.DisplayName()
	}
	subject, textBody, htmlBody, err := emailPresenter.RenderElectronicSignatureCertificate(ctx, ref.UnrefOrZero(signature.SignerFullName), docName)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot render email: %w", err)
	}

	email := coredata.NewEmail(
		ref.UnrefOrZero(signature.SignerFullName),
		mail.Addr(signature.SignerEmail),
		subject,
		textBody,
		htmlBody,
		&coredata.EmailOptions{
			SenderName: new(organization.Name),
		},
	)

	attachments := coredata.EmailAttachments{
		coredata.NewEmailAttachment(
			email.ID,
			signedFile.ID,
			signedFile.FileName,
		),
		coredata.NewEmailAttachment(
			email.ID,
			certificateOfCompletionFile.ID,
			certificateFilename,
		),
	}

	return email, attachments, nil
}

func (h *completionCertificateHandler) handleCertFailure(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
	scope coredata.Scoper,
	processingError error,
) error {
	h.logger.ErrorCtx(
		ctx,
		"certificate worker failure",
		log.Error(processingError),
		log.String("signature_id", signature.ID.String()),
	)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			errStr := processingError.Error()
			signature.LastError = &errStr
			signature.CertificateProcessingStartedAt = nil
			signature.UpdatedAt = time.Now()

			if signature.AttemptCount >= signature.MaxAttempts {
				signature.Status = coredata.ElectronicSignatureStatusFailed
			}

			if err := signature.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature: %w", err)
			}

			return nil
		},
	)
}
