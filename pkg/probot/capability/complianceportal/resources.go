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
		if access.CompliancePortalDocumentID != nil {
			var link coredata.CompliancePortalDocument
			if err := link.LoadByID(ctx, conn, scope, *access.CompliancePortalDocumentID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					continue
				}

				return nil, nil, nil, fmt.Errorf("cannot load compliance portal document: %w", err)
			}

			if link.CompliancePortalID != compliancePortalID {
				continue
			}

			var document coredata.Document
			if err := document.LoadByID(ctx, conn, scope, link.DocumentID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load document: %w", err)
			}

			if document.CurrentPublishedMajor != nil {
				documents = append(
					documents,
					messageDocument{
						ID:     document.ID.String(),
						Title:  document.Title,
						Status: access.Status.String(),
					},
				)
			}
		}

		if access.CompliancePortalAuditID != nil {
			var portalAudit coredata.CompliancePortalAudit
			if err := portalAudit.LoadByID(ctx, conn, scope, *access.CompliancePortalAuditID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					continue
				}

				return nil, nil, nil, fmt.Errorf("cannot load compliance portal audit: %w", err)
			}

			if portalAudit.CompliancePortalID != compliancePortalID {
				continue
			}

			var audit coredata.Audit
			if err := audit.LoadByID(ctx, conn, scope, portalAudit.AuditID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load audit: %w", err)
			}

			var framework coredata.Framework
			if err := framework.LoadByID(ctx, conn, scope, audit.FrameworkID); err != nil {
				return nil, nil, nil, fmt.Errorf("cannot load framework: %w", err)
			}

			if audit.ReportFileID == nil {
				continue
			}

			title := framework.Name
			if audit.Name != nil && *audit.Name != "" {
				title += " - " + *audit.Name
			}

			reports = append(
				reports,
				messageReport{
					ID:      audit.ReportFileID.String(),
					Title:   title,
					AuditID: audit.ID.String(),
					Status:  access.Status.String(),
				},
			)
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

func resolveAccessResourceIDs(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	resourceIDs []gid.GID,
) ([]gid.GID, []gid.GID, []gid.GID, error) {
	var (
		documentIDs             []gid.GID
		publicDocumentIDs       []gid.GID
		auditIDs                []gid.GID
		reportFileIDs           []gid.GID
		compliancePortalFileIDs []gid.GID
	)

	for _, resourceID := range resourceIDs {
		switch resourceID.EntityType() {
		case coredata.CompliancePortalDocumentEntityType:
			documentIDs = append(documentIDs, resourceID)
		case coredata.DocumentEntityType:
			publicDocumentIDs = append(publicDocumentIDs, resourceID)
		case coredata.CompliancePortalAuditEntityType:
			auditIDs = append(auditIDs, resourceID)
		case coredata.FileEntityType:
			reportFileIDs = append(reportFileIDs, resourceID)
		case coredata.CompliancePortalFileEntityType:
			compliancePortalFileIDs = append(compliancePortalFileIDs, resourceID)
		}
	}

	if len(publicDocumentIDs) > 0 {
		var links coredata.CompliancePortalDocuments
		if err := links.LoadByCompliancePortalIDAndDocumentIDs(
			ctx,
			conn,
			scope,
			compliancePortalID,
			publicDocumentIDs,
		); err != nil {
			return nil, nil, nil, fmt.Errorf("cannot load portal document links: %w", err)
		}

		for _, link := range links {
			documentIDs = append(documentIDs, link.ID)
		}
	}

	if len(reportFileIDs) > 0 {
		var audits coredata.Audits
		if err := audits.LoadByReportFileIDs(ctx, conn, scope, reportFileIDs); err != nil {
			return nil, nil, nil, fmt.Errorf("cannot load audits by report file IDs: %w", err)
		}

		legacyAuditIDs := make([]gid.GID, 0, len(audits))
		for _, audit := range audits {
			legacyAuditIDs = append(legacyAuditIDs, audit.ID)
		}

		linksByAuditID, err := coredata.LoadCompliancePortalAuditsByCompliancePortalIDAndAuditIDs(
			ctx,
			conn,
			scope,
			compliancePortalID,
			legacyAuditIDs,
		)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("cannot load portal audit links: %w", err)
		}

		for _, link := range linksByAuditID {
			auditIDs = append(auditIDs, link.ID)
		}
	}

	documentIDs, err := filterDocumentLinksOnPortal(ctx, conn, scope, compliancePortalID, documentIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	auditIDs, err = filterAuditLinksOnPortal(ctx, conn, scope, compliancePortalID, auditIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	compliancePortalFileIDs, err = filterFilesOnPortal(
		ctx,
		conn,
		scope,
		compliancePortalID,
		compliancePortalFileIDs,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	return documentIDs, auditIDs, compliancePortalFileIDs, nil
}

func filterDocumentLinksOnPortal(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	documentLinkIDs []gid.GID,
) ([]gid.GID, error) {
	if len(documentLinkIDs) == 0 {
		return nil, nil
	}

	var links coredata.CompliancePortalDocuments
	if err := links.LoadByIDs(ctx, conn, scope, documentLinkIDs); err != nil &&
		!errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load portal document links: %w", err)
	}

	filtered := make([]gid.GID, 0, len(links))
	for _, link := range links {
		if link.CompliancePortalID == compliancePortalID {
			filtered = append(filtered, link.ID)
		}
	}

	return filtered, nil
}

func filterAuditLinksOnPortal(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	auditLinkIDs []gid.GID,
) ([]gid.GID, error) {
	if len(auditLinkIDs) == 0 {
		return nil, nil
	}

	var links coredata.CompliancePortalAudits
	if err := links.LoadByIDs(ctx, conn, scope, auditLinkIDs); err != nil &&
		!errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load portal audit links: %w", err)
	}

	filtered := make([]gid.GID, 0, len(links))
	for _, link := range links {
		if link.CompliancePortalID == compliancePortalID {
			filtered = append(filtered, link.ID)
		}
	}

	return filtered, nil
}

func filterFilesOnPortal(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	compliancePortalID gid.GID,
	fileIDs []gid.GID,
) ([]gid.GID, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}

	var files coredata.CompliancePortalFiles
	if err := files.LoadByIDs(ctx, conn, scope, fileIDs); err != nil &&
		!errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load portal files: %w", err)
	}

	filtered := make([]gid.GID, 0, len(files))
	for _, file := range files {
		if file.CompliancePortalID == compliancePortalID {
			filtered = append(filtered, file.ID)
		}
	}

	return filtered, nil
}
