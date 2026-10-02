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

package management

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

type (
	UpdateCompliancePortalDocumentVisibilityRequest struct {
		CompliancePortalID         gid.GID
		DocumentID                 gid.GID
		CompliancePortalVisibility coredata.CompliancePortalVisibility
	}

	DeleteCompliancePortalDocumentRequest struct {
		ID gid.GID
	}

	UpdateCompliancePortalAuditVisibilityRequest struct {
		CompliancePortalID         gid.GID
		AuditID                    gid.GID
		CompliancePortalVisibility coredata.CompliancePortalVisibility
	}

	DeleteCompliancePortalAuditRequest struct {
		ID gid.GID
	}

	UpdateCompliancePortalThirdPartyPublishedRequest struct {
		CompliancePortalID gid.GID
		ThirdPartyID       gid.GID
		Published          bool
	}

	DeleteCompliancePortalThirdPartyRequest struct {
		ID gid.GID
	}
)

func (r *UpdateCompliancePortalDocumentVisibilityRequest) Validate() error {
	v := validator.New()

	v.Check(r.CompliancePortalID, "compliance_portal_id", validator.Required(), validator.GID(coredata.CompliancePortalEntityType))
	v.Check(r.DocumentID, "document_id", validator.Required(), validator.GID(coredata.DocumentEntityType))
	v.Check(
		r.CompliancePortalVisibility,
		"compliance_portal_visibility",
		validator.Required(),
		validator.OneOfSlice([]coredata.CompliancePortalVisibility{
			coredata.CompliancePortalVisibilityRestricted,
			coredata.CompliancePortalVisibilityPublic,
		}),
	)

	return v.Error()
}

func (r *DeleteCompliancePortalDocumentRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.CompliancePortalDocumentEntityType))

	return v.Error()
}

func (r *UpdateCompliancePortalAuditVisibilityRequest) Validate() error {
	v := validator.New()

	v.Check(r.CompliancePortalID, "compliance_portal_id", validator.Required(), validator.GID(coredata.CompliancePortalEntityType))
	v.Check(r.AuditID, "audit_id", validator.Required(), validator.GID(coredata.AuditEntityType))
	v.Check(
		r.CompliancePortalVisibility,
		"compliance_portal_visibility",
		validator.Required(),
		validator.OneOfSlice([]coredata.CompliancePortalVisibility{
			coredata.CompliancePortalVisibilityRestricted,
			coredata.CompliancePortalVisibilityPublic,
		}),
	)

	return v.Error()
}

func (r *DeleteCompliancePortalAuditRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.CompliancePortalAuditEntityType))

	return v.Error()
}

func (r *UpdateCompliancePortalThirdPartyPublishedRequest) Validate() error {
	v := validator.New()

	v.Check(r.CompliancePortalID, "compliance_portal_id", validator.Required(), validator.GID(coredata.CompliancePortalEntityType))
	v.Check(r.ThirdPartyID, "third_party_id", validator.Required(), validator.GID(coredata.ThirdPartyEntityType))
	v.Check(r.Published, "published", validator.OneOfSlice([]bool{true}))

	return v.Error()
}

func (r *DeleteCompliancePortalThirdPartyRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.CompliancePortalThirdPartyEntityType))

	return v.Error()
}

func (s *Service) UpdateDocumentVisibility(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateCompliancePortalDocumentVisibilityRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			portal := &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, tx, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			document := &coredata.Document{}
			if err := document.LoadByID(ctx, tx, scope, req.DocumentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			now := time.Now()
			row := &coredata.CompliancePortalDocument{
				ID:                 gid.New(scope.GetTenantID(), coredata.CompliancePortalDocumentEntityType),
				OrganizationID:     portal.OrganizationID,
				CompliancePortalID: req.CompliancePortalID,
				DocumentID:         req.DocumentID,
				Visibility:         req.CompliancePortalVisibility,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := row.Upsert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot upsert portal document: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) DeleteDocument(
	ctx context.Context,
	scope coredata.Scoper,
	req *DeleteCompliancePortalDocumentRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.DeleteCompliancePortalDocumentByID(
				ctx,
				tx,
				scope,
				req.ID,
			); err != nil {
				return fmt.Errorf("cannot delete portal document: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) UpdateAuditVisibility(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateCompliancePortalAuditVisibilityRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			portal := &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, tx, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			audit := &coredata.Audit{}
			if err := audit.LoadByID(ctx, tx, scope, req.AuditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			now := time.Now()
			row := &coredata.CompliancePortalAudit{
				ID:                 gid.New(scope.GetTenantID(), coredata.CompliancePortalAuditEntityType),
				OrganizationID:     portal.OrganizationID,
				CompliancePortalID: req.CompliancePortalID,
				AuditID:            req.AuditID,
				Visibility:         req.CompliancePortalVisibility,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := row.Upsert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot upsert portal audit: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) DeleteAudit(
	ctx context.Context,
	scope coredata.Scoper,
	req *DeleteCompliancePortalAuditRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.DeleteCompliancePortalAuditByID(
				ctx,
				tx,
				scope,
				req.ID,
			); err != nil {
				return fmt.Errorf("cannot delete portal audit: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) UpdateThirdPartyPublished(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateCompliancePortalThirdPartyPublishedRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			portal := &coredata.CompliancePortal{}
			if err := portal.LoadByID(ctx, tx, scope, req.CompliancePortalID); err != nil {
				return fmt.Errorf("cannot load compliance portal: %w", err)
			}

			thirdParty := &coredata.ThirdParty{}
			if err := thirdParty.LoadByID(ctx, tx, scope, req.ThirdPartyID); err != nil {
				return fmt.Errorf("cannot load third party: %w", err)
			}

			now := time.Now()
			row := &coredata.CompliancePortalThirdParty{
				ID:                 gid.New(scope.GetTenantID(), coredata.CompliancePortalThirdPartyEntityType),
				OrganizationID:     portal.OrganizationID,
				CompliancePortalID: req.CompliancePortalID,
				ThirdPartyID:       req.ThirdPartyID,
				CreatedAt:          now,
				UpdatedAt:          now,
			}

			if err := row.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot publish third party on portal: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) DeleteThirdParty(
	ctx context.Context,
	scope coredata.Scoper,
	req *DeleteCompliancePortalThirdPartyRequest,
) error {
	if err := req.Validate(); err != nil {
		return err
	}

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := coredata.DeleteCompliancePortalThirdPartyByID(
				ctx,
				tx,
				scope,
				req.ID,
			); err != nil {
				return fmt.Errorf("cannot delete portal third party: %w", err)
			}

			return nil
		},
	)
}
