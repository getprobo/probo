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

package complianceportal_v1

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
	"go.probo.inc/probo/pkg/server/api/complianceportal/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

func newNDADirective(
	logger *log.Logger,
	visitorSvc *visitor.Service,
	esignSvc *esign.Service,
) func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
		identity := authn.IdentityFromContext(ctx)
		if identity == nil {
			return next(ctx)
		}

		compliancePage := complianceportal.CompliancePortalFromContext(ctx)
		if compliancePage == nil {
			logger.ErrorCtx(ctx, "cannot get compliance page from context")
			return nil, gqlutils.Internal(ctx)
		}

		scope := coredata.NewScopeFromObjectID(compliancePage.OrganizationID)

		skip, err := shouldSkipNDAForPublic(
			ctx,
			visitorSvc,
			scope,
			compliancePage.OrganizationID,
		)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot check target visibility for NDA gate", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}
		if skip {
			return next(ctx)
		}

		membership, err := visitorSvc.GetPortalMembership(ctx, compliancePage.ID, identity.ID)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot get compliance page membership", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		if membership.ElectronicSignatureID == nil {
			return next(ctx)
		}

		sig, err := esignSvc.GetSignatureByID(ctx, scope, *membership.ElectronicSignatureID)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot get NDA signature", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		// We need full name before user signs NDA
		if identity.FullName == "" {
			return nil, gqlutils.FullNameRequiredf(ctx, "full name is required")
		}

		if sig.Status != coredata.ElectronicSignatureStatusCompleted {
			return nil, gqlutils.NDASignatureRequiredf(ctx, "NDA signature required")
		}

		return next(ctx)
	}
}

// shouldSkipNDAForPublic reports whether the @nda field targets a single PUBLIC
// document / report / file and can therefore skip the signature gate. Bulk
// requestAccesses always returns false. Not-found / not-visible targets also
// skip so the resolver can return its usual NotFound/Invalid.
func shouldSkipNDAForPublic(
	ctx context.Context,
	visitorSvc *visitor.Service,
	scope coredata.Scoper,
	organizationID gid.GID,
) (bool, error) {
	fc := graphql.GetFieldContext(ctx)
	if fc == nil {
		return false, nil
	}

	switch fc.Field.Name {
	case "requestAccesses":
		return false, nil

	case "exportDocumentPDF":
		input, ok := fc.Args["input"].(types.ExportDocumentPDFInput)
		if !ok {
			return false, nil
		}
		return isPublicDocument(ctx, visitorSvc, scope, organizationID, input.DocumentID)

	case "requestDocumentAccess":
		input, ok := fc.Args["input"].(types.RequestDocumentAccessInput)
		if !ok {
			return false, nil
		}
		return isPublicDocument(ctx, visitorSvc, scope, organizationID, input.DocumentID)

	case "exportReportPDF":
		input, ok := fc.Args["input"].(types.ExportReportPDFInput)
		if !ok {
			return false, nil
		}
		return isPublicReport(ctx, visitorSvc, scope, input.ReportID)

	case "requestReportAccess":
		input, ok := fc.Args["input"].(types.RequestReportAccessInput)
		if !ok {
			return false, nil
		}
		return isPublicReport(ctx, visitorSvc, scope, input.ReportID)

	case "exportCompliancePortalFile":
		input, ok := fc.Args["input"].(types.ExportCompliancePortalFileInput)
		if !ok {
			return false, nil
		}
		return isPublicPortalFile(ctx, visitorSvc, scope, organizationID, input.CompliancePortalFileID)

	case "requestCompliancePortalFileAccess":
		input, ok := fc.Args["input"].(types.RequestCompliancePortalFileAccessInput)
		if !ok {
			return false, nil
		}
		return isPublicPortalFile(ctx, visitorSvc, scope, organizationID, input.CompliancePortalFileID)

	default:
		return false, nil
	}
}

func isPublicDocument(
	ctx context.Context,
	visitorSvc *visitor.Service,
	scope coredata.Scoper,
	organizationID gid.GID,
	documentID gid.GID,
) (bool, error) {
	document, err := visitorSvc.GetDocument(ctx, scope, organizationID, documentID)
	if err != nil {
		if isMissingDocument(err) {
			return true, nil
		}
		return false, fmt.Errorf("cannot load document: %w", err)
	}

	return document.CompliancePortalVisibility == coredata.CompliancePortalVisibilityPublic, nil
}

func isPublicReport(
	ctx context.Context,
	visitorSvc *visitor.Service,
	scope coredata.Scoper,
	reportID gid.GID,
) (bool, error) {
	audit, err := visitorSvc.GetAuditByReportFileID(ctx, scope, reportID)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return true, nil
		}
		return false, fmt.Errorf("cannot load audit: %w", err)
	}

	return audit.CompliancePortalVisibility == coredata.CompliancePortalVisibilityPublic, nil
}

func isPublicPortalFile(
	ctx context.Context,
	visitorSvc *visitor.Service,
	scope coredata.Scoper,
	organizationID gid.GID,
	fileID gid.GID,
) (bool, error) {
	file, err := visitorSvc.GetPortalFile(ctx, scope, organizationID, fileID)
	if err != nil {
		if errors.Is(err, visitor.ErrPortalFileNotFound) ||
			errors.Is(err, visitor.ErrPortalFileNotVisible) ||
			errors.Is(err, coredata.ErrResourceNotFound) {
			return true, nil
		}
		return false, fmt.Errorf("cannot load portal file: %w", err)
	}

	return file.CompliancePortalVisibility == coredata.CompliancePortalVisibilityPublic, nil
}

func isMissingDocument(err error) bool {
	if errors.Is(err, visitor.ErrDocumentNotFound) ||
		errors.Is(err, visitor.ErrDocumentNotVisible) ||
		errors.Is(err, coredata.ErrResourceNotFound) {
		return true
	}
	_, ok := errors.AsType[*visitor.ErrDocumentArchived](err)
	return ok
}
