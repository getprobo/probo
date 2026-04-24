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
	"go.probo.inc/probo/pkg/evidenceassessor"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
)

type (
	evidenceAssessmentHandler struct {
		pg          *pg.Client
		fileManager *filemanager.Service
		assessor    *evidenceassessor.Assessor
		logger      *log.Logger
		staleAfter  time.Duration
	}

	EvidenceAssessmentWorkerConfig struct {
		StaleAfter time.Duration
	}
)

func NewEvidenceAssessmentWorker(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	assessor *evidenceassessor.Assessor,
	logger *log.Logger,
	cfg EvidenceAssessmentWorkerConfig,
	opts ...worker.Option,
) *worker.Worker[coredata.Evidence] {
	staleAfter := cfg.StaleAfter
	if staleAfter == 0 {
		staleAfter = 5 * time.Minute
	}

	h := &evidenceAssessmentHandler{
		pg:          pgClient,
		fileManager: fileManager,
		assessor:    assessor,
		logger:      logger,
		staleAfter:  staleAfter,
	}

	return worker.New(
		"evidence-assessment-worker",
		h,
		logger,
		opts...,
	)
}

func (h *evidenceAssessmentHandler) Claim(ctx context.Context) (coredata.Evidence, error) {
	var evidence coredata.Evidence

	if err := h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := evidence.LoadNextPendingAssessmentForUpdateSkipLocked(ctx, tx); err != nil {
				return err
			}

			now := time.Now()
			evidence.AssessmentStatus = coredata.EvidenceAssessmentStatusProcessing
			evidence.AssessmentProcessingStartedAt = &now
			evidence.UpdatedAt = now
			if err := evidence.Update(ctx, tx, coredata.NewScopeFromObjectID(evidence.ID)); err != nil {
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

func (h *evidenceAssessmentHandler) Process(ctx context.Context, evidence coredata.Evidence) error {
	if err := h.assessAndCommit(ctx, evidence); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"evidence assessment worker failure",
			log.Error(err),
			log.String("evidence_id", evidence.ID.String()),
		)

		if err := h.failEvidence(ctx, evidence.ID); err != nil {
			h.logger.ErrorCtx(ctx, "cannot mark evidence assessment as failed", log.Error(err))
		}

		return err
	}

	return nil
}

func (h *evidenceAssessmentHandler) RecoverStale(ctx context.Context) error {
	return h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := coredata.ResetStaleAssessmentProcessing(ctx, conn, h.staleAfter); err != nil {
				return fmt.Errorf("cannot reset stale assessment processing: %w", err)
			}
			return nil
		},
	)
}

// assessAndCommit deliberately takes evidence by value; mutations made
// inside the transaction stay local, so a failed commit cannot leak
// partial state to the subsequent failEvidence call.
func (h *evidenceAssessmentHandler) assessAndCommit(
	ctx context.Context,
	evidence coredata.Evidence,
) error {
	if evidence.EvidenceFileID == nil {
		return fmt.Errorf("cannot assess evidence %s: no file attached", evidence.ID)
	}

	scope := coredata.NewScopeFromObjectID(evidence.ID)

	var file coredata.File
	if err := h.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadByID(ctx, conn, scope, *evidence.EvidenceFileID); err != nil {
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

	assessment, err := h.assessor.Assess(ctx, file.FileName, mimeType, base64Data)
	if err != nil {
		return fmt.Errorf("cannot assess evidence: %w", err)
	}

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := evidence.SetAssessment(assessment); err != nil {
				return err
			}
			summary := assessment.Summary
			evidence.Description = &summary
			evidence.AssessmentStatus = coredata.EvidenceAssessmentStatusCompleted
			evidence.AssessmentProcessingStartedAt = nil
			evidence.UpdatedAt = time.Now()
			if err := evidence.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update evidence: %w", err)
			}

			return nil
		},
	)
}

func (h *evidenceAssessmentHandler) failEvidence(ctx context.Context, evidenceID gid.GID) error {
	scope := coredata.NewScopeFromObjectID(evidenceID)

	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return coredata.MarkEvidenceAssessmentFailed(ctx, tx, scope, evidenceID)
		},
	)
}
