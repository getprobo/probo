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
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

const pdfClaimLease = 5 * time.Minute

type documentApprovalQuorumPDFHandler struct {
	service *Service
	logger  *log.Logger
}

func NewDocumentApprovalQuorumPDFWorker(
	service *Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.DocumentVersionApprovalQuorum] {
	h := &documentApprovalQuorumPDFHandler{
		service: service,
		logger:  logger,
	}

	return worker.New(
		"document-approval-quorum-pdf-worker",
		h,
		logger,
		opts...,
	)
}

func (h *documentApprovalQuorumPDFHandler) Claim(ctx context.Context) (coredata.DocumentVersionApprovalQuorum, error) {
	var quorum coredata.DocumentVersionApprovalQuorum

	if err := h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return quorum.ClaimNextWithoutFileForUpdate(
				ctx,
				tx,
				maxPDFAttempts,
				time.Now(),
				pdfClaimLease,
			)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrNoDocumentPDFJobAvailable) {
			return coredata.DocumentVersionApprovalQuorum{}, worker.ErrNoTask
		}

		return coredata.DocumentVersionApprovalQuorum{}, err
	}

	return quorum, nil
}

func (h *documentApprovalQuorumPDFHandler) Process(ctx context.Context, quorum coredata.DocumentVersionApprovalQuorum) error {
	scope := coredata.NewScope(quorum.ID.TenantID())

	if err := h.service.DocumentApprovals.generateAndUploadQuorumPDF(ctx, scope, &quorum); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"document approval quorum pdf worker failure",
			log.Error(err),
			log.String("approval_quorum_id", quorum.ID.String()),
			log.Int("attempt", quorum.PdfAttemptCount),
		)

		if releaseErr := h.releasePDFClaim(ctx, scope, &quorum); releaseErr != nil {
			h.logger.ErrorCtx(
				ctx,
				"cannot release document approval quorum pdf claim",
				log.Error(releaseErr),
				log.String("approval_quorum_id", quorum.ID.String()),
			)
		}

		return err
	}

	return nil
}

func (h *documentApprovalQuorumPDFHandler) releasePDFClaim(
	ctx context.Context,
	scope coredata.Scoper,
	quorum *coredata.DocumentVersionApprovalQuorum,
) error {
	return h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return quorum.ReleasePDFClaim(ctx, tx, scope, time.Now())
		},
	)
}
