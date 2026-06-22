// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package trust_v1

import (
	"context"
	"errors"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/compliancepage"
	"go.probo.inc/probo/pkg/server/api/trust/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/trust"
)

func (r *queryResolver) nodeByGID(
	ctx context.Context,
	id gid.GID,
	notFoundLabel string,
) (types.Node, error) {
	scope := coredata.NewScopeFromObjectID(id)
	trustService := r.trust

	switch id.EntityType() {
	case coredata.OrganizationEntityType:
		organization, err := trustService.Organizations.Get(ctx, scope, id)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot get organization", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		return types.NewOrganization(organization), nil

	case coredata.DocumentEntityType:
		trustCenter := compliancepage.CompliancePageFromContext(ctx)

		document, err := trustService.Documents.Get(ctx, scope, trustCenter.OrganizationID, id)
		if err != nil {
			if errors.Is(err, trust.ErrDocumentNotFound) || errors.Is(err, trust.ErrDocumentNotVisible) || errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, gqlutils.NotFoundf(ctx, "node %q not found", notFoundLabel)
			}

			if _, ok := errors.AsType[*trust.ErrDocumentArchived](err); ok {
				return nil, gqlutils.NotFoundf(ctx, "node %q not found", notFoundLabel)
			}

			r.logger.ErrorCtx(ctx, "cannot get document", log.Error(err))

			return nil, gqlutils.Internal(ctx)
		}

		return types.NewDocument(document), nil

	case coredata.FrameworkEntityType:
		framework, err := trustService.Frameworks.Get(ctx, scope, id)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot get framework", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		return types.NewFramework(framework), nil

	case coredata.FileEntityType:
		trustCenter := compliancepage.CompliancePageFromContext(ctx)

		file, err := trustService.Reports.Get(ctx, scope, trustCenter.OrganizationID, id)
		if err != nil {
			if errors.Is(err, trust.ErrReportNotFound) || errors.Is(err, coredata.ErrResourceNotFound) {
				return nil, gqlutils.NotFoundf(ctx, "node %q not found", notFoundLabel)
			}

			r.logger.ErrorCtx(ctx, "cannot get audit report file", log.Error(err))

			return nil, gqlutils.Internal(ctx)
		}

		return types.NewAuditReport(file), nil

	case coredata.AuditEntityType:
		audit, err := trustService.Audits.Get(ctx, scope, id)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot get audit", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		return types.NewAudit(audit), nil

	case coredata.ThirdPartyEntityType:
		thirdParty, err := trustService.ThirdParties.Get(ctx, scope, id)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot get thirdParty", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		return types.NewSubprocessor(thirdParty), nil

	case coredata.TrustCenterEntityType:
		trustCenter, err := trustService.TrustCenters.Get(ctx, scope, id)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot get trust center", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		return types.NewTrustCenter(trustCenter), nil

	case coredata.TrustCenterReferenceEntityType:
		reference, err := trustService.TrustCenterReferences.Get(ctx, scope, id)
		if err != nil {
			r.logger.ErrorCtx(ctx, "cannot get trust center reference", log.Error(err))
			return nil, gqlutils.Internal(ctx)
		}

		return types.NewTrustCenterReference(reference), nil

	case coredata.TrustCenterFileEntityType:
		trustCenter := compliancepage.CompliancePageFromContext(ctx)

		trustCenterFile, err := trustService.TrustCenterFiles.Get(ctx, scope, trustCenter.OrganizationID, id)
		if err != nil {
			if errors.Is(err, trust.ErrTrustCenterFileNotFound) || errors.Is(err, trust.ErrTrustCenterFileNotVisible) {
				return nil, gqlutils.NotFoundf(ctx, "node %q not found", notFoundLabel)
			}

			r.logger.ErrorCtx(ctx, "cannot get trust center file", log.Error(err))

			return nil, gqlutils.Internal(ctx)
		}

		return types.NewTrustCenterFile(trustCenterFile), nil

	default:
		return nil, gqlutils.NotFoundf(ctx, "node %q not found", notFoundLabel)
	}
}
