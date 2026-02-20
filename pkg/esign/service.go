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
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/mail"
	"golang.org/x/sync/errgroup"
)

// Service manages the electronic signature lifecycle.
type (
	Service struct {
		pg             *pg.Client
		fileManager    *filemanager.Service
		tsaClient      *TSAClient
		certificateGen *CertificateGenerator
		bucket         string
		logger         *log.Logger
	}

	CreateSignatureRequest struct {
		OrganizationID gid.GID
		DocumentType   coredata.ElectronicSignatureDocumentType
		FileID         gid.GID
		SignerEmail    mail.Addr
		ConsentText    string // optional; required when DocumentType == OTHER
	}

	AcceptSignatureRequest struct {
		SignatureID    gid.GID
		SignerFullName string
		SignerEmail    mail.Addr
		SignerIPAddr   string
		SignerUA       string
	}

	RecordEventRequest struct {
		SignatureID gid.GID
		EventType   coredata.ElectronicSignatureEventType
		EventSource coredata.ElectronicSignatureEventSource
		ActorEmail  mail.Addr
		ActorIPAddr string
		ActorUA     string
	}
)

func NewService(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	html2pdfConverter *html2pdf.Converter,
	tsaURL string,
	bucket string,
	logger *log.Logger,
) *Service {
	httpClient := httpclient.DefaultPooledClient(
		httpclient.WithLogger(logger),
	)

	return &Service{
		pg:          pgClient,
		fileManager: fileManager,
		tsaClient:   &TSAClient{URL: tsaURL, HTTPClient: httpClient},
		certificateGen: &CertificateGenerator{
			HTML2PDFConverter: html2pdfConverter,
		},
		bucket: bucket,
		logger: logger,
	}
}

func (s *Service) Run(ctx context.Context, presenterConfigFunc EmailPresenterConfigFunc) error {
	g := errgroup.Group{}

	sealingWorkerCtx, stopSealingWorker := context.WithCancel(context.Background())
	sealingWorker := NewSealingWorker(
		s.pg,
		s.fileManager,
		s.tsaClient,
		s.logger.Named("sealing-worker"),
	)
	g.Go(func() error { return sealingWorker.Run(sealingWorkerCtx) })

	certWorkerCtx, stopCertWorker := context.WithCancel(context.Background())
	certWorker := NewCompletionCertificateWorker(
		s.pg,
		s.fileManager,
		s.certificateGen,
		presenterConfigFunc,
		s.bucket,
		s.logger.Named("completion-certificate-worker"),
	)
	g.Go(func() error { return certWorker.Run(certWorkerCtx) })

	<-ctx.Done()

	stopSealingWorker()
	stopCertWorker()

	return g.Wait()
}

func (s *Service) CreateSignature(
	ctx context.Context,
	conn pg.Conn,
	req *CreateSignatureRequest,
) (*coredata.ElectronicSignature, error) {
	consentText := req.ConsentText
	if consentText == "" {
		var err error
		consentText, err = req.DocumentType.ConsentText()
		if err != nil {
			return nil, fmt.Errorf("cannot derive consent text: %w", err)
		}
	} else {
		if !strings.HasSuffix(consentText, coredata.ESignProcessConsentText) {
			consentText = consentText + " " + coredata.ESignProcessConsentText
		}
	}

	now := time.Now()
	scope := coredata.NewScopeFromObjectID(req.OrganizationID)

	signatureID := gid.New(scope.GetTenantID(), coredata.ElectronicSignatureEntityType)

	stampedFileID, err := s.createStampedDocument(ctx, conn, scope, req.OrganizationID, req.FileID, signatureID)
	if err != nil {
		return nil, fmt.Errorf("cannot create stamped document: %w", err)
	}

	sig := &coredata.ElectronicSignature{
		ID:             signatureID,
		OrganizationID: req.OrganizationID,
		Status:         coredata.ElectronicSignatureStatusPending,
		DocumentType:   req.DocumentType,
		FileID:         stampedFileID,
		SignerEmail:    req.SignerEmail.String(),
		ConsentText:    consentText,
		SealVersion:    1,
		AttemptCount:   0,
		MaxAttempts:    10,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := sig.Insert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot insert electronic signature: %w", err)
	}

	return sig, nil
}

func (s *Service) createStampedDocument(
	ctx context.Context,
	conn pg.Conn,
	scope coredata.Scoper,
	organizationID gid.GID,
	originalFileID gid.GID,
	signatureID gid.GID,
) (gid.GID, error) {
	var originalFile coredata.File
	if err := originalFile.LoadByID(ctx, conn, scope, originalFileID); err != nil {
		return gid.GID{}, fmt.Errorf("cannot load original file: %w", err)
	}

	pdfData, err := s.fileManager.GetFileBytes(ctx, &originalFile)
	if err != nil {
		return gid.GID{}, fmt.Errorf("cannot download original file: %w", err)
	}

	stampedData, err := StampSignatureID(pdfData, signatureID.String())
	if err != nil {
		return gid.GID{}, fmt.Errorf("cannot stamp signature ID: %w", err)
	}

	now := time.Now()
	stampedFile := coredata.File{
		ID:             gid.New(scope.GetTenantID(), coredata.FileEntityType),
		OrganizationID: organizationID,
		BucketName:     s.bucket,
		MimeType:       "application/pdf",
		FileName:       originalFile.FileName,
		FileKey:        uuid.MustNewV4().String(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	stampedSize, err := s.fileManager.PutFile(
		ctx,
		&stampedFile,
		bytes.NewReader(stampedData),
		map[string]string{
			"type":         "stamped-document",
			"signature-id": signatureID.String(),
		},
	)
	if err != nil {
		return gid.GID{}, fmt.Errorf("cannot upload stamped file: %w", err)
	}

	stampedFile.FileSize = stampedSize

	if err := stampedFile.Insert(ctx, conn, scope); err != nil {
		return gid.GID{}, fmt.Errorf("cannot insert stamped file record: %w", err)
	}

	return stampedFile.ID, nil
}

func (s *Service) AcceptSignature(ctx context.Context, req *AcceptSignatureRequest) (*coredata.ElectronicSignature, error) {
	var (
		scope     = coredata.NewScopeFromObjectID(req.SignatureID)
		now       = time.Now()
		signature = coredata.ElectronicSignature{}
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := signature.LoadByID(ctx, tx, scope, req.SignatureID); err != nil {
				return fmt.Errorf("cannot load electronic signature: %w", err)
			}

			if signature.Status != coredata.ElectronicSignatureStatusPending &&
				signature.Status != coredata.ElectronicSignatureStatusFailed {
				return fmt.Errorf("cannot accept electronic signature in status %s", signature.Status)
			}

			if signature.Status == coredata.ElectronicSignatureStatusFailed {
				signature.AttemptCount = 0
				signature.LastError = nil
			}

			signature.SignerFullName = &req.SignerFullName
			signature.SignerIPAddress = &req.SignerIPAddr
			signature.SignerUserAgent = &req.SignerUA
			signature.SignedAt = &now
			signature.Status = coredata.ElectronicSignatureStatusAccepted
			signature.UpdatedAt = now

			if err := signature.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature: %w", err)
			}

			if err := s.recordEvent(
				ctx,
				tx,
				&RecordEventRequest{
					SignatureID: signature.ID,
					EventType:   coredata.ElectronicSignatureEventTypeSignatureAccepted,
					EventSource: coredata.ElectronicSignatureEventSourceServer,
					ActorEmail:  req.SignerEmail,
					ActorIPAddr: req.SignerIPAddr,
					ActorUA:     req.SignerUA,
				},
			); err != nil {
				return fmt.Errorf("cannot record event: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &signature, nil
}

func (s *Service) RecordEvent(ctx context.Context, req *RecordEventRequest) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			return s.recordEvent(ctx, tx, req)
		},
	)
}

func (s *Service) recordEvent(ctx context.Context, tx pg.Conn, req *RecordEventRequest) error {
	var (
		now   = time.Now()
		scope = coredata.NewScopeFromObjectID(req.SignatureID)
	)

	event := coredata.ElectronicSignatureEvent{
		ID:                    gid.New(scope.GetTenantID(), coredata.ElectronicSignatureEventEntityType),
		ElectronicSignatureID: req.SignatureID,
		EventType:             req.EventType,
		EventSource:           req.EventSource,
		ActorEmail:            req.ActorEmail.String(),
		ActorIPAddress:        req.ActorIPAddr,
		ActorUserAgent:        req.ActorUA,
		OccurredAt:            now,
		CreatedAt:             now,
	}

	if err := event.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert signing event: %w", err)
	}

	return nil
}

func (s *Service) GetSignatureByID(ctx context.Context, id gid.GID) (*coredata.ElectronicSignature, error) {
	var (
		scope     = coredata.NewScopeFromObjectID(id)
		signature = coredata.ElectronicSignature{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := signature.LoadByID(ctx, conn, scope, id); err != nil {
				return fmt.Errorf("cannot load electronic signature: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &signature, nil
}

func (s *Service) GenerateCertificateFileURL(
	ctx context.Context,
	certificateFileID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	var (
		scope = coredata.NewScopeFromObjectID(certificateFileID)
		file  = coredata.File{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := file.LoadByID(ctx, conn, scope, certificateFileID); err != nil {
				return fmt.Errorf("cannot load certificate file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	url, err := s.fileManager.GenerateFileUrl(ctx, &file, expiresIn)
	if err != nil {
		return "", fmt.Errorf("cannot generate certificate file URL: %w", err)
	}

	return url, nil
}

func (s *Service) GenerateSignatureFileURL(
	ctx context.Context,
	signatureID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	var (
		scope     = coredata.NewScopeFromObjectID(signatureID)
		signature coredata.ElectronicSignature
		file      coredata.File
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := signature.LoadByID(ctx, conn, scope, signatureID); err != nil {
				return fmt.Errorf("cannot load electronic signature: %w", err)
			}

			if err := file.LoadByID(ctx, conn, scope, signature.FileID); err != nil {
				return fmt.Errorf("cannot load signature file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	url, err := s.fileManager.GenerateFileUrl(ctx, &file, expiresIn)
	if err != nil {
		return "", fmt.Errorf("cannot generate signature file URL: %w", err)
	}

	return url, nil
}

func (s *Service) GetEventsBySignatureID(
	ctx context.Context,
	signatureID gid.GID,
) (coredata.ElectronicSignatureEvents, error) {
	var (
		scope  = coredata.NewScopeFromObjectID(signatureID)
		events = coredata.ElectronicSignatureEvents{}
	)
	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := events.LoadBySignatureID(ctx, conn, scope, signatureID); err != nil {
				return fmt.Errorf("cannot load events: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return events, nil
}
