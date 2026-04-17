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

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/kit/worker"
	"go.probo.inc/probo/pkg/coredata"
)

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
			return version.ClaimNextPublishedWithoutFileForUpdate(ctx, tx)
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
	tenantService := h.service.WithTenant(version.ID.TenantID())

	if err := tenantService.Documents.generateAndUploadPublicationPDF(ctx, &version); err != nil {
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
