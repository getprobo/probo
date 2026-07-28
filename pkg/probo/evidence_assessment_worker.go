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
)

type (
	evidenceAssessmentHandler struct {
		pg          *pg.Client
		fileManager *filemanager.Service
		assessor    *evidenceassessor.Assessor
		logger      *log.Logger
		staleAfter  time.Duration
		maxAttempts int
	}

	EvidenceAssessmentWorkerConfig struct {
		StaleAfter  time.Duration
		MaxAttempts int
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

	maxAttempts := cfg.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 3
	}

	h := &evidenceAssessmentHandler{
		pg:          pgClient,
		fileManager: fileManager,
		assessor:    assessor,
		logger:      logger,
		staleAfter:  staleAfter,
		maxAttempts: maxAttempts,
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

			if err := evidence.MarkAssessmentProcessing(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(evidence.ID),
			); err != nil {
				return fmt.Errorf("cannot mark evidence assessment processing: %w", err)
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
			log.Int("assessment_attempts", evidence.AssessmentAttempts),
		)

		if err := h.failEvidence(ctx, evidence); err != nil {
			h.logger.ErrorCtx(ctx, "cannot mark evidence assessment outcome", log.Error(err))
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

			if err := evidence.SetAssessmentCompleted(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update evidence: %w", err)
			}

			return nil
		},
	)
}

func (h *evidenceAssessmentHandler) failEvidence(ctx context.Context, evidence coredata.Evidence) error {
	return h.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return evidence.SetAssessmentOutcome(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(evidence.ID),
				h.maxAttempts,
			)
		},
	)
}
