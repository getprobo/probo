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

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/filemanager"
)

var (
	ErrLoadFile     = errors.New("esign: cannot load file")
	ErrDownloadPDF  = errors.New("esign: cannot download PDF")
	ErrComputeSeal  = errors.New("esign: cannot compute seal")
	ErrTSATimestamp = errors.New("esign: cannot get TSA timestamp")
)

const (
	currentSealVersion = 1
)

type (
	sealingHandler struct {
		pg          *pg.Client
		fileManager *filemanager.Service
		tsaClient   *TSAClient
		logger      *log.Logger
		tsaTimeout  time.Duration
		staleAfter  time.Duration
	}

	SealingWorkerOption func(*sealingHandler)
)

func WithSealingWorkerTSATimeout(d time.Duration) SealingWorkerOption {
	return func(h *sealingHandler) { h.tsaTimeout = d }
}

func WithSealingWorkerStaleAfter(d time.Duration) SealingWorkerOption {
	return func(h *sealingHandler) { h.staleAfter = d }
}

func NewSealingWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	tsaClient *TSAClient,
	logger *log.Logger,
	handlerOpts []SealingWorkerOption,
	workerOpts ...worker.Option,
) *worker.Worker[coredata.ElectronicSignature] {
	h := &sealingHandler{
		pg:          pgClient,
		fileManager: fileManager,
		tsaClient:   tsaClient,
		logger:      logger,
		tsaTimeout:  10 * time.Second,
		staleAfter:  5 * time.Minute,
	}

	for _, opt := range handlerOpts {
		opt(h)
	}

	return worker.New(
		"sealing-worker",
		h,
		logger,
		workerOpts...,
	)
}

func (h *sealingHandler) Claim(ctx context.Context) (coredata.ElectronicSignature, error) {
	var signature coredata.ElectronicSignature

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := signature.LoadNextAcceptedForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			signature.Status = coredata.ElectronicSignatureStatusProcessing
			signature.ProcessingStartedAt = &now
			signature.AttemptCount++
			signature.LastAttemptedAt = &now
			signature.UpdatedAt = now
			if err := signature.Update(ctx, tx, coredata.NewNoScope()); err != nil {
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

func (h *sealingHandler) Process(ctx context.Context, signature coredata.ElectronicSignature) error {
	if err := h.sealAndCommit(ctx, &signature); err != nil {
		if err := h.failSignature(ctx, &signature, err); err != nil {
			h.logger.ErrorCtx(ctx, "cannot fail signature", log.Error(err))
		}
		return err
	}
	return nil
}

func (h *sealingHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return coredata.ResetStaleProcessingSignatures(ctx, conn, h.staleAfter)
		},
	)
}

func (h *sealingHandler) sealAndCommit(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
) error {
	var (
		scope  = coredata.NewScopeFromObjectID(signature.ID)
		file   coredata.File
		events []coredata.ElectronicSignatureEvent
	)

	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadByID(ctx, conn, scope, signature.FileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			return nil
		},
	); err != nil {
		return fmt.Errorf("%w: %w", ErrLoadFile, err)
	}

	pdfBytes, err := h.fileManager.GetFileBytes(ctx, &file)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDownloadPDF, err)
	}

	fileHash := hash.SHA256Hex(pdfBytes)
	signature.FileHash = &fileHash

	seal, err := signature.ComputeSeal(currentSealVersion)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrComputeSeal, err)
	}
	signature.Seal = &seal
	signature.SealVersion = currentSealVersion
	events = append(
		events,
		signature.NewEvent(
			coredata.ElectronicSignatureEventTypeSealComputed,
			coredata.ElectronicSignatureEventSourceServer,
		),
	)

	tsaCtx, cancel := context.WithTimeout(ctx, h.tsaTimeout)
	defer cancel()
	tsaToken, err := h.tsaClient.Timestamp(tsaCtx, []byte(seal))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTSATimestamp, err)
	}
	signature.TSAToken = tsaToken
	events = append(
		events,
		signature.NewEvent(
			coredata.ElectronicSignatureEventTypeTimestampRequested,
			coredata.ElectronicSignatureEventSourceServer,
		),
	)

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var current coredata.ElectronicSignature
			if err := current.LoadByID(ctx, tx, scope, signature.ID); err != nil {
				return fmt.Errorf("cannot load signature: %w", err)
			}

			if current.Status != coredata.ElectronicSignatureStatusProcessing {
				return fmt.Errorf("esign: unexpected status %s, expected PROCESSING", current.Status)
			}

			signature.Status = coredata.ElectronicSignatureStatusCompleted
			signature.UpdatedAt = time.Now()
			if err := signature.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature: %w", err)
			}
			events = append(
				events,
				signature.NewEvent(
					coredata.ElectronicSignatureEventTypeSignatureCompleted,
					coredata.ElectronicSignatureEventSourceServer,
				),
			)

			for i := range events {
				if err := events[i].Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert %s event: %w", events[i].EventType, err)
				}
			}

			return nil
		},
	); err != nil {
		return fmt.Errorf("cannot commit signing results: %w", err)
	}

	return nil
}

func (h *sealingHandler) failSignature(
	ctx context.Context,
	signature *coredata.ElectronicSignature,
	processingError error,
) error {
	scope := coredata.NewScopeFromObjectID(signature.ID)

	h.logger.ErrorCtx(
		ctx,
		"sealing worker failure",
		log.Error(processingError),
		log.String("signature_id", signature.ID.String()),
	)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			errStr := userFacingError(processingError)
			signature.LastError = &errStr
			signature.ProcessingStartedAt = nil
			signature.UpdatedAt = time.Now()
			if signature.AttemptCount >= signature.MaxAttempts {
				signature.Status = coredata.ElectronicSignatureStatusFailed
			} else {
				signature.Status = coredata.ElectronicSignatureStatusAccepted
			}
			if err := signature.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature: %w", err)
			}

			event := signature.NewEvent(coredata.ElectronicSignatureEventTypeProcessingError, coredata.ElectronicSignatureEventSourceServer)
			if err := event.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert PROCESSING_ERROR event: %w", err)
			}

			return nil
		},
	)
}

func userFacingError(err error) string {
	switch {
	case errors.Is(err, ErrTSATimestamp):
		return "The timestamp authority is temporarily unavailable."
	case errors.Is(err, ErrLoadFile):
		return "Unable to load the document for signing."
	case errors.Is(err, ErrDownloadPDF):
		return "Unable to retrieve the document."
	case errors.Is(err, ErrComputeSeal):
		return "Unable to generate the cryptographic seal."
	default:
		return "An unexpected error occurred while processing your signature."
	}
}
