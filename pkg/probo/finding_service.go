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
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type FindingService struct {
	svc *TenantService
}

type (
	CreateFindingRequest struct {
		OrganizationID     gid.GID
		Kind               coredata.FindingKind
		Description        *string
		Source             *string
		IdentifiedOn       *time.Time
		RootCause          *string
		CorrectiveAction   *string
		OwnerID            *gid.GID
		DueDate            *time.Time
		Status             *coredata.FindingStatus
		Priority           *coredata.FindingPriority
		RiskID             *gid.GID
		EffectivenessCheck *string
	}

	UpdateFindingRequest struct {
		ID                 gid.GID
		Description        **string
		Source             **string
		IdentifiedOn       **time.Time
		RootCause          **string
		CorrectiveAction   **string
		OwnerID            *gid.GID
		DueDate            **time.Time
		Status             *coredata.FindingStatus
		Priority           *coredata.FindingPriority
		RiskID             **gid.GID
		EffectivenessCheck **string
	}
)

func (r *CreateFindingRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(r.Kind, "kind", validator.Required(), validator.OneOfSlice(coredata.FindingKinds()))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(r.Source, "source", validator.SafeText(ContentMaxLength))
	v.Check(r.RootCause, "root_cause", validator.SafeText(ContentMaxLength))
	v.Check(r.CorrectiveAction, "corrective_action", validator.SafeText(ContentMaxLength))
	v.Check(r.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(r.Status, "status", validator.OneOfSlice(coredata.FindingStatuses()))
	v.Check(r.Priority, "priority", validator.OneOfSlice(coredata.FindingPriorities()))
	v.Check(r.RiskID, "risk_id", validator.GID(coredata.RiskEntityType))
	v.Check(r.EffectivenessCheck, "effectiveness_check", validator.SafeText(ContentMaxLength))

	if r.Status != nil && *r.Status == coredata.FindingStatusRiskAccepted {
		v.Check(r.RiskID, "risk_id", validator.Required())
	}

	return v.Error()
}

func (r *UpdateFindingRequest) Validate() error {
	v := validator.New()

	v.Check(r.ID, "id", validator.Required(), validator.GID(coredata.FindingEntityType))
	v.Check(r.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(r.Source, "source", validator.SafeText(ContentMaxLength))
	v.Check(r.RootCause, "root_cause", validator.SafeText(ContentMaxLength))
	v.Check(r.CorrectiveAction, "corrective_action", validator.SafeText(ContentMaxLength))
	v.Check(r.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))
	v.Check(r.Status, "status", validator.OneOfSlice(coredata.FindingStatuses()))
	v.Check(r.Priority, "priority", validator.OneOfSlice(coredata.FindingPriorities()))
	v.Check(r.RiskID, "risk_id", validator.GID(coredata.RiskEntityType))
	v.Check(r.EffectivenessCheck, "effectiveness_check", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s FindingService) Get(
	ctx context.Context,
	findingID gid.GID,
) (*coredata.Finding, error) {
	finding := &coredata.Finding{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return finding.LoadByID(ctx, conn, s.svc.scope, findingID)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot get finding: %w", err)
	}

	return finding, nil
}

func (s *FindingService) Create(
	ctx context.Context,
	req *CreateFindingRequest,
) (*coredata.Finding, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()

	finding := &coredata.Finding{
		ID:                 gid.New(s.svc.scope.GetTenantID(), coredata.FindingEntityType),
		OrganizationID:     req.OrganizationID,
		Kind:               req.Kind,
		Description:        req.Description,
		Source:             req.Source,
		IdentifiedOn:       req.IdentifiedOn,
		RootCause:          req.RootCause,
		CorrectiveAction:   req.CorrectiveAction,
		OwnerID:            req.OwnerID,
		DueDate:            req.DueDate,
		Status:             coredata.FindingStatusOpen,
		Priority:           coredata.FindingPriorityMedium,
		RiskID:             req.RiskID,
		EffectivenessCheck: req.EffectivenessCheck,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if req.Status != nil {
		finding.Status = *req.Status
	}

	if req.Priority != nil {
		finding.Priority = *req.Priority
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if req.OwnerID != nil {
				owner := &coredata.MembershipProfile{}
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}
			}

			if err := finding.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert finding: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return finding, nil
}

func (s *FindingService) Update(
	ctx context.Context,
	req *UpdateFindingRequest,
) (*coredata.Finding, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	finding := &coredata.Finding{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := finding.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load finding: %w", err)
			}

			if req.Description != nil {
				finding.Description = *req.Description
			}
			if req.Source != nil {
				finding.Source = *req.Source
			}
			if req.IdentifiedOn != nil {
				finding.IdentifiedOn = *req.IdentifiedOn
			}
			if req.RootCause != nil {
				finding.RootCause = *req.RootCause
			}
			if req.CorrectiveAction != nil {
				finding.CorrectiveAction = *req.CorrectiveAction
			}
			if req.OwnerID != nil {
				owner := &coredata.MembershipProfile{}
				if err := owner.LoadByID(ctx, conn, s.svc.scope, *req.OwnerID); err != nil {
					return fmt.Errorf("cannot load owner profile: %w", err)
				}
				finding.OwnerID = req.OwnerID
			}
			if req.DueDate != nil {
				finding.DueDate = *req.DueDate
			}
			if req.Status != nil {
				finding.Status = *req.Status
			}
			if req.Priority != nil {
				finding.Priority = *req.Priority
			}
			if req.RiskID != nil {
				finding.RiskID = *req.RiskID
			}
			if req.EffectivenessCheck != nil {
				finding.EffectivenessCheck = *req.EffectivenessCheck
			}

			if finding.Status == coredata.FindingStatusRiskAccepted && finding.RiskID == nil {
				return fmt.Errorf("cannot update finding: risk_id is required when status is RISK_ACCEPTED")
			}

			finding.UpdatedAt = time.Now()

			if err := finding.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update finding: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return finding, nil
}

func (s FindingService) Delete(
	ctx context.Context,
	findingID gid.GID,
) error {
	finding := coredata.Finding{ID: findingID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := finding.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete finding: %w", err)
			}
			return nil
		},
	)
}

func (s FindingService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.FindingOrderField],
	filter *coredata.FindingFilter,
) (*page.Page[*coredata.Finding, coredata.FindingOrderField], error) {
	var findings coredata.Findings

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := findings.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load findings: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(findings, cursor), nil
}

func (s FindingService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.FindingFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			findings := coredata.Findings{}
			count, err = findings.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count findings: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s FindingService) CreateAuditMapping(
	ctx context.Context,
	findingID gid.GID,
	auditID gid.GID,
	referenceID string,
) (*coredata.Finding, *coredata.Audit, error) {
	finding := &coredata.Finding{}
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := finding.LoadByID(ctx, conn, s.svc.scope, findingID); err != nil {
				return fmt.Errorf("cannot load finding: %w", err)
			}

			if err := audit.LoadByID(ctx, conn, s.svc.scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if finding.OrganizationID != audit.OrganizationID {
				return fmt.Errorf("cannot create finding audit mapping: finding and audit belong to different organizations")
			}

			findingAudit := &coredata.FindingAudit{
				FindingID:      findingID,
				AuditID:        auditID,
				ReferenceID:    referenceID,
				OrganizationID: finding.OrganizationID,
				CreatedAt:      time.Now(),
			}

			if err := findingAudit.Upsert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot create finding audit mapping: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return finding, audit, nil
}

func (s FindingService) DeleteAuditMapping(
	ctx context.Context,
	findingID gid.GID,
	auditID gid.GID,
) (*coredata.Finding, *coredata.Audit, error) {
	finding := &coredata.Finding{}
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := finding.LoadByID(ctx, conn, s.svc.scope, findingID); err != nil {
				return fmt.Errorf("cannot load finding: %w", err)
			}

			if err := audit.LoadByID(ctx, conn, s.svc.scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			findingAudit := &coredata.FindingAudit{}
			if err := findingAudit.Delete(ctx, conn, s.svc.scope, finding.ID, audit.ID); err != nil {
				return fmt.Errorf("cannot delete finding audit mapping: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete finding audit mapping: %w", err)
	}

	return finding, audit, nil
}

func (s FindingService) ListForAuditID(
	ctx context.Context,
	auditID gid.GID,
	cursor *page.Cursor[coredata.FindingOrderField],
	filter *coredata.FindingFilter,
) (*page.Page[*coredata.Finding, coredata.FindingOrderField], error) {
	var findings coredata.Findings
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := audit.LoadByID(ctx, conn, s.svc.scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}
			if err := findings.LoadByAuditID(ctx, conn, s.svc.scope, auditID, cursor, filter); err != nil {
				return fmt.Errorf("cannot load findings: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage([]*coredata.Finding(findings), cursor), nil
}

func (s FindingService) CountForAuditID(
	ctx context.Context,
	auditID gid.GID,
	filter *coredata.FindingFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			findings := coredata.Findings{}
			count, err = findings.CountByAuditID(ctx, conn, s.svc.scope, auditID, filter)
			if err != nil {
				return fmt.Errorf("cannot count findings: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}
