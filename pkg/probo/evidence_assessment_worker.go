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

package probo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/evidencedescriber"
	"go.probo.inc/probo/pkg/filemanager"
)

type (
	evidenceDescriptionHandler struct {
		pg          *pg.Client
		fileManager *filemanager.Service
		describer   *evidencedescriber.Describer
		logger      *log.Logger
		staleAfter  time.Duration
	}

	EvidenceDescriptionWorkerConfig struct {
		StaleAfter time.Duration
	}
)

func NewEvidenceDescriptionWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	describer *evidencedescriber.Describer,
	logger *log.Logger,
	cfg EvidenceDescriptionWorkerConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.Evidence] {
	staleAfter := cfg.StaleAfter
	if staleAfter == 0 {
		staleAfter = 5 * time.Minute
	}

	h := &evidenceDescriptionHandler{
		pg:          pgClient,
		fileManager: fileManager,
		describer:   describer,
		logger:      logger,
		staleAfter:  staleAfter,
	}

	return worker.New(
		"evidence-description-worker",
		h,
		logger,
		opts...,
	)
}

func (h *evidenceDescriptionHandler) Claim(ctx context.Context) (coredata.Evidence, error) {
	var evidence coredata.Evidence

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := evidence.LoadNextPendingDescriptionForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			evidence.DescriptionStatus = coredata.EvidenceDescriptionStatusProcessing
			evidence.DescriptionProcessingStartedAt = &now
			evidence.UpdatedAt = now
			if err := evidence.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update evidence: %w", err)
			}

			return nil
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return coredata.Evidence{}, worker.ErrNoTask
		}
		return coredata.Evidence{}, err
	}

	return evidence, nil
}

func (h *evidenceDescriptionHandler) Process(ctx context.Context, evidence coredata.Evidence) error {
	if err := h.describeAndCommit(ctx, &evidence); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"evidence description worker failure",
			log.Error(err),
			log.String("evidence_id", evidence.ID.String()),
		)

		if err := h.failEvidence(ctx, &evidence); err != nil {
			h.logger.ErrorCtx(ctx, "cannot mark evidence description as failed", log.Error(err))
		}

		return err
	}

	return nil
}

func (h *evidenceDescriptionHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleDescriptionProcessing(ctx, conn, h.staleAfter); err != nil {
				return fmt.Errorf("cannot reset stale description processing: %w", err)
			}
			return nil
		},
	)
}

func (h *evidenceDescriptionHandler) describeAndCommit(
	ctx context.Context,
	evidence *coredata.Evidence,
) error {
	if evidence.EvidenceFileId == nil {
		return fmt.Errorf("evidence %s has no file", evidence.ID)
	}

	scope := coredata.NewScopeFromObjectID(evidence.ID)

	var file coredata.File
	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadByID(ctx, conn, scope, *evidence.EvidenceFileId); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("cannot load file: %w", err)
	}

	base64Data, mimeType, err := h.fileManager.GetFileBase64(ctx, &file)
	if err != nil {
		return fmt.Errorf("cannot download file: %w", err)
	}

	description, err := h.describer.Describe(ctx, file.FileName, mimeType, base64Data)
	if err != nil {
		return fmt.Errorf("cannot describe evidence: %w", err)
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			evidence.Description = description
			evidence.DescriptionStatus = coredata.EvidenceDescriptionStatusCompleted
			evidence.DescriptionProcessingStartedAt = nil
			evidence.UpdatedAt = time.Now()
			if err := evidence.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update evidence: %w", err)
			}

			return nil
		},
	)
}

func (h *evidenceDescriptionHandler) failEvidence(
	ctx context.Context,
	evidence *coredata.Evidence,
) error {
	scope := coredata.NewScopeFromObjectID(evidence.ID)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			evidence.DescriptionStatus = coredata.EvidenceDescriptionStatusFailed
			evidence.DescriptionProcessingStartedAt = nil
			evidence.UpdatedAt = time.Now()
			if err := evidence.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update evidence: %w", err)
			}

			return nil
		},
	)
}
