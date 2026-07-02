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

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

const maxPDFAttempts = 3

type documentPDFHandler struct {
	service *Service
	logger  *log.Logger
}

func NewDocumentPDFWorker(
	service *Service,
	logger *log.Logger,
	opts ...worker.Option,
) *worker.Worker[coredata.DocumentVersion] {
	h := &documentPDFHandler{
		service: service,
		logger:  logger,
	}

	return worker.New(
		"document-pdf-worker",
		h,
		logger,
		opts...,
	)
}

func (h *documentPDFHandler) Claim(ctx context.Context) (coredata.DocumentVersion, error) {
	var version coredata.DocumentVersion

	if err := h.service.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return version.ClaimNextPublishedWithoutFileForUpdate(ctx, tx, maxPDFAttempts)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrNoDocumentPDFJobAvailable) {
			return coredata.DocumentVersion{}, worker.ErrNoTask
		}

		return coredata.DocumentVersion{}, err
	}

	return version, nil
}

func (h *documentPDFHandler) Process(ctx context.Context, version coredata.DocumentVersion) error {
	scope := coredata.NewScope(version.ID.TenantID())

	if err := h.service.Documents.generateAndUploadPublicationPDF(ctx, scope, &version); err != nil {
		h.logger.ErrorCtx(
			ctx,
			"document pdf worker failure",
			log.Error(err),
			log.String("document_version_id", version.ID.String()),
			log.Int("attempt", version.PdfAttemptCount),
		)

		return err
	}

	return nil
}
