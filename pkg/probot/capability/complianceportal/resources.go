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

package complianceportal

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func loadResources(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	accessID gid.GID,
) ([]messageDocument, []messageReport, []messageFile, error) {
	documents := []messageDocument{}
	reports := []messageReport{}
	files := []messageFile{}

	accesses, err := page.LoadAll(
		ctx,
		page.OrderBy[coredata.CompliancePortalDocumentAccessOrderField]{
			Field:     coredata.CompliancePortalDocumentAccessOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
		func(ctx context.Context, cursor *page.Cursor[coredata.CompliancePortalDocumentAccessOrderField]) ([]*coredata.CompliancePortalDocumentAccess, error) {
			var batch coredata.CompliancePortalDocumentAccesses
			if err := batch.LoadByCompliancePortalAccessID(ctx, conn, scope, accessID, cursor); err != nil {
				return nil, fmt.Errorf("cannot load compliance portal resource accesses: %w", err)
			}

			return batch, nil
		},
	)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, access := range accesses {
		if access.DocumentID != nil {
			var document coredata.Document
			if err := document.LoadByID(ctx, conn, scope, *access.DocumentID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load document: %w", err)
			}

			if document.CurrentPublishedMajor != nil {
				var link coredata.CompliancePortalDocument

				err := link.LoadByCompliancePortalIDAndDocumentID(
					ctx,
					conn,
					scope,
					compliancePortalID,
					*access.DocumentID,
				)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return nil, nil, nil, fmt.Errorf("cannot load compliance portal document: %w", err)
				}

				if err == nil && link.Visibility != coredata.CompliancePortalVisibilityNone {
					documents = append(
						documents,
						messageDocument{
							ID:     access.DocumentID.String(),
							Title:  document.Title,
							Status: access.Status.String(),
						},
					)
				}
			}
		}

		if access.ReportFileID != nil {
			var portalAudit coredata.CompliancePortalAudit

			err := portalAudit.LoadByCompliancePortalIDAndReportFileID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				*access.ReportFileID,
			)
			if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, nil, nil, fmt.Errorf("cannot load compliance portal audit: %w", err)
			}

			if err == nil {
				var audit coredata.Audit
				if err := audit.LoadByID(ctx, conn, scope, portalAudit.AuditID); err != nil {
					return nil, nil, nil, fmt.Errorf("cannot load audit: %w", err)
				}

				var framework coredata.Framework
				if err := framework.LoadByID(ctx, conn, scope, audit.FrameworkID); err != nil {
					return nil, nil, nil, fmt.Errorf("cannot load framework: %w", err)
				}

				title := framework.Name
				if audit.Name != nil && *audit.Name != "" {
					title += " - " + *audit.Name
				}

				reports = append(
					reports,
					messageReport{
						ID:      access.ReportFileID.String(),
						Title:   title,
						AuditID: audit.ID.String(),
						Status:  access.Status.String(),
					},
				)
			}
		}

		if access.CompliancePortalFileID != nil {
			var file coredata.CompliancePortalFile

			err := file.LoadByCompliancePortalIDAndID(
				ctx,
				conn,
				scope,
				compliancePortalID,
				*access.CompliancePortalFileID,
			)
			if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, nil, nil, fmt.Errorf("cannot load compliance portal file: %w", err)
			}

			if err == nil {
				files = append(
					files,
					messageFile{
						ID:       access.CompliancePortalFileID.String(),
						Name:     file.Name,
						Category: file.Category,
						Status:   access.Status.String(),
					},
				)
			}
		}
	}

	return documents, reports, files, nil
}
