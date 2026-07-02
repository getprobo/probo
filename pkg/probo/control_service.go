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

type (
	ControlService struct {
		svc *Service
	}

	CreateControlRequest struct {
		ID                          gid.GID
		FrameworkID                 gid.GID
		Name                        string
		Description                 *string
		SectionTitle                string
		BestPractice                bool
		MaturityLevel               coredata.ControlMaturityLevel
		NotImplementedJustification *string
	}

	UpdateControlRequest struct {
		ID                          gid.GID
		Name                        *string
		Description                 **string
		SectionTitle                *string
		BestPractice                *bool
		MaturityLevel               *coredata.ControlMaturityLevel
		NotImplementedJustification **string
	}
)

func (ccr *CreateControlRequest) Validate() error {
	v := validator.New()

	v.Check(ccr.FrameworkID, "framework_id", validator.Required(), validator.GID(coredata.FrameworkEntityType))
	v.Check(ccr.Name, "name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ccr.Description, "description", validator.Required(), validator.SafeText(ContentMaxLength))
	v.Check(ccr.SectionTitle, "section_title", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ccr.NotImplementedJustification, "not_implemented_justification", validator.SafeText(ContentMaxLength))

	v.Check(
		ccr.MaturityLevel,
		"maturity_level",
		validator.Required(),
		validator.OneOfSlice(coredata.ControlMaturityLevels()),
	)

	return v.Error()
}

func (ucr *UpdateControlRequest) Validate() error {
	v := validator.New()

	v.Check(ucr.ID, "id", validator.Required(), validator.GID(coredata.ControlEntityType))
	v.Check(ucr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ucr.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(ucr.SectionTitle, "section_title", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(ucr.NotImplementedJustification, "not_implemented_justification", validator.SafeText(ContentMaxLength))

	if ucr.MaturityLevel != nil {
		v.Check(
			*ucr.MaturityLevel,
			"maturity_level",
			validator.OneOfSlice(coredata.ControlMaturityLevels()),
		)
	}

	return v.Error()
}

func (s ControlService) CountForDocumentID(
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
	filter *coredata.ControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			controls := &coredata.Controls{}

			count, err = controls.CountByDocumentID(ctx, conn, scope, documentID, filter)
			if err != nil {
				return fmt.Errorf("cannot count controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count controls: %w", err)
	}

	return count, nil
}

func (s ControlService) ListForDocumentID(
	ctx context.Context, scope coredata.Scoper,
	documentID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
	filter *coredata.ControlFilter,
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	document := &coredata.Document{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := document.LoadByID(ctx, conn, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			return controls.LoadByDocumentID(ctx, conn, scope, documentID, cursor, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}

func (s ControlService) CountForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
	filter *coredata.ControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			controls := &coredata.Controls{}

			count, err = controls.CountByMeasureID(ctx, conn, scope, measureID, filter)
			if err != nil {
				return fmt.Errorf("cannot count controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count controls: %w", err)
	}

	return count, nil
}

func (s ControlService) ListForMeasureID(
	ctx context.Context, scope coredata.Scoper,
	measureID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
	filter *coredata.ControlFilter,
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	measure := &coredata.Measure{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := measure.LoadByID(ctx, conn, scope, measureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			return controls.LoadByMeasureID(ctx, conn, scope, measureID, cursor, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}

func (s ControlService) CountForFrameworkID(
	ctx context.Context, scope coredata.Scoper,
	frameworkID gid.GID,
	filter *coredata.ControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			controls := &coredata.Controls{}

			count, err = controls.CountByFrameworkID(ctx, conn, scope, frameworkID, filter)
			if err != nil {
				return fmt.Errorf("cannot count controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count controls: %w", err)
	}

	return count, nil
}

func (s ControlService) ListForFrameworkID(
	ctx context.Context, scope coredata.Scoper,
	frameworkID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
	filter *coredata.ControlFilter,
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	framework := &coredata.Framework{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := framework.LoadByID(ctx, conn, scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			return controls.LoadByFrameworkID(
				ctx,
				conn,
				scope,
				framework.ID,
				cursor,
				filter,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}

func (s ControlService) CountForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	filter *coredata.ControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			controls := &coredata.Controls{}

			count, err = controls.CountByOrganizationID(ctx, conn, scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count controls: %w", err)
	}

	return count, nil
}

func (s ControlService) ListForOrganizationID(
	ctx context.Context, scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
	filter *coredata.ControlFilter,
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			return controls.LoadByOrganizationID(
				ctx,
				conn,
				scope,
				organization.ID,
				cursor,
				filter,
			)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}

func (s ControlService) CountForRiskID(
	ctx context.Context, scope coredata.Scoper,
	riskID gid.GID,
	filter *coredata.ControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			controls := &coredata.Controls{}

			count, err = controls.CountByRiskID(ctx, conn, scope, riskID, filter)
			if err != nil {
				return fmt.Errorf("cannot count controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count controls: %w", err)
	}

	return count, nil
}

func (s ControlService) ListForRiskID(
	ctx context.Context, scope coredata.Scoper,
	riskID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
	filter *coredata.ControlFilter,
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	risk := &coredata.Risk{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := risk.LoadByID(ctx, conn, scope, riskID); err != nil {
				return fmt.Errorf("cannot load risk: %w", err)
			}

			return controls.LoadByRiskID(ctx, conn, scope, risk.ID, cursor, filter)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list controls: %w", err)
	}

	return page.NewPage(controls, cursor), nil
}

func (s ControlService) CreateMeasureMapping(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	measureID gid.GID,
) (*coredata.Control, *coredata.Measure, error) {
	control := &coredata.Control{}
	measure := &coredata.Measure{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := control.LoadByID(ctx, conn, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if err := measure.LoadByID(ctx, conn, scope, measureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			controlMeasure := &coredata.ControlMeasure{
				ControlID:      controlID,
				MeasureID:      measureID,
				OrganizationID: control.OrganizationID,
				TenantID:       scope.GetTenantID(),
				CreatedAt:      time.Now(),
			}

			return controlMeasure.Upsert(ctx, conn, scope)
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create control measure mapping: %w", err)
	}

	return control, measure, nil
}

func (s ControlService) DeleteMeasureMapping(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	measureID gid.GID,
) (*coredata.Control, *coredata.Measure, error) {
	control := &coredata.Control{}
	measure := &coredata.Measure{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := control.LoadByID(ctx, tx, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if err := measure.LoadByID(ctx, tx, scope, measureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			controlMeasure := &coredata.ControlMeasure{}
			if err := controlMeasure.Delete(ctx, tx, scope, control.ID, measure.ID); err != nil {
				return fmt.Errorf("cannot delete control measure mapping: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete control measure mapping: %w", err)
	}

	return control, measure, nil
}

func (s ControlService) CreateDocumentMapping(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	documentID gid.GID,
) (*coredata.Control, *coredata.Document, error) {
	control := &coredata.Control{}
	document := &coredata.Document{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := control.LoadByID(ctx, tx, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			controlDocument := &coredata.ControlDocument{
				ControlID:      control.ID,
				DocumentID:     document.ID,
				OrganizationID: control.OrganizationID,
				TenantID:       scope.GetTenantID(),
				CreatedAt:      time.Now(),
			}

			if err := controlDocument.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert control document: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create control document mapping: %w", err)
	}

	return control, document, nil
}

func (s ControlService) DeleteDocumentMapping(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	documentID gid.GID,
) (*coredata.Control, *coredata.Document, error) {
	control := &coredata.Control{}
	document := &coredata.Document{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := control.LoadByID(ctx, tx, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if err := document.LoadByID(ctx, tx, scope, documentID); err != nil {
				return fmt.Errorf("cannot load document: %w", err)
			}

			controlDocument := &coredata.ControlDocument{}
			if err := controlDocument.Delete(ctx, tx, scope, control.ID, document.ID); err != nil {
				return fmt.Errorf("cannot delete control document mapping: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete control document mapping: %w", err)
	}

	return control, document, nil
}

func (s ControlService) CreateAuditMapping(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	auditID gid.GID,
) (*coredata.Control, *coredata.Audit, error) {
	control := &coredata.Control{}
	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := control.LoadByID(ctx, conn, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if err := audit.LoadByID(ctx, conn, scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			controlAudit := &coredata.ControlAudit{
				ControlID:      controlID,
				AuditID:        auditID,
				OrganizationID: control.OrganizationID,
				CreatedAt:      time.Now(),
			}

			if err := controlAudit.Upsert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot create control audit mapping: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return control, audit, nil
}

func (s ControlService) DeleteAuditMapping(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	auditID gid.GID,
) (*coredata.Control, *coredata.Audit, error) {
	control := &coredata.Control{}
	audit := &coredata.Audit{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := control.LoadByID(ctx, tx, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if err := audit.LoadByID(ctx, tx, scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			controlAudit := &coredata.ControlAudit{}
			if err := controlAudit.Delete(ctx, tx, scope, control.ID, audit.ID); err != nil {
				return fmt.Errorf("cannot delete control audit mapping: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete control audit mapping: %w", err)
	}

	return control, audit, nil
}

func (s ControlService) CreateObligationMapping(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	obligationID gid.GID,
) (*coredata.Control, *coredata.Obligation, error) {
	control := &coredata.Control{}
	obligation := &coredata.Obligation{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := control.LoadByID(ctx, conn, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if err := obligation.LoadByID(ctx, conn, scope, obligationID); err != nil {
				return fmt.Errorf("cannot load obligation: %w", err)
			}

			controlObligation := &coredata.ControlObligation{
				ControlID:    controlID,
				ObligationID: obligationID,
				CreatedAt:    time.Now(),
			}

			if err := controlObligation.Upsert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot create control obligation mapping: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return control, obligation, nil
}

func (s ControlService) DeleteObligationMapping(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
	obligationID gid.GID,
) (*coredata.Control, *coredata.Obligation, error) {
	control := &coredata.Control{}
	obligation := &coredata.Obligation{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := control.LoadByID(ctx, tx, scope, controlID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if err := obligation.LoadByID(ctx, tx, scope, obligationID); err != nil {
				return fmt.Errorf("cannot load obligation: %w", err)
			}

			controlObligation := &coredata.ControlObligation{}
			if err := controlObligation.Delete(ctx, tx, scope, control.ID, obligation.ID); err != nil {
				return fmt.Errorf("cannot delete control obligation mapping: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot delete control obligation mapping: %w", err)
	}

	return control, obligation, nil
}

func (s ControlService) ListForAuditID(
	ctx context.Context, scope coredata.Scoper,
	auditID gid.GID,
	cursor *page.Cursor[coredata.ControlOrderField],
	filter *coredata.ControlFilter,
) (*page.Page[*coredata.Control, coredata.ControlOrderField], error) {
	var controls coredata.Controls

	audit := &coredata.Audit{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := audit.LoadByID(ctx, conn, scope, auditID); err != nil {
				return fmt.Errorf("cannot load audit: %w", err)
			}

			if err := controls.LoadByAuditID(ctx, conn, scope, auditID, cursor, filter); err != nil {
				return fmt.Errorf("cannot load controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage([]*coredata.Control(controls), cursor), nil
}

func (s ControlService) CountForStatementOfApplicabilityID(
	ctx context.Context, scope coredata.Scoper,
	statementOfApplicabilityID gid.GID,
	filter *coredata.ControlFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			controls := &coredata.Controls{}

			count, err = controls.CountByStatementOfApplicabilityID(ctx, conn, scope, statementOfApplicabilityID, filter)
			if err != nil {
				return fmt.Errorf("cannot count controls: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count controls: %w", err)
	}

	return count, nil
}

func (s ControlService) Create(
	ctx context.Context, scope coredata.Scoper,
	req CreateControlRequest,
) (*coredata.Control, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()
	framework := &coredata.Framework{}

	notImplementedJustification := req.NotImplementedJustification
	if req.MaturityLevel != coredata.ControlMaturityLevelNone {
		notImplementedJustification = nil
	}

	control := &coredata.Control{
		ID:                          gid.New(scope.GetTenantID(), coredata.ControlEntityType),
		FrameworkID:                 req.FrameworkID,
		Name:                        req.Name,
		Description:                 req.Description,
		SectionTitle:                req.SectionTitle,
		BestPractice:                req.BestPractice,
		MaturityLevel:               req.MaturityLevel,
		NotImplementedJustification: notImplementedJustification,
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := framework.LoadByID(ctx, conn, scope, req.FrameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			control.FrameworkID = framework.ID
			control.OrganizationID = framework.OrganizationID

			return control.Insert(ctx, conn, scope)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create control: %w", err)
	}

	return control, nil
}

func (s ControlService) Get(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
) (*coredata.Control, error) {
	control := &coredata.Control{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return control.LoadByID(ctx, conn, scope, controlID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get control: %w", err)
	}

	return control, nil
}

func (s ControlService) GetByIDs(
	ctx context.Context, scope coredata.Scoper,
	controlIDs ...gid.GID,
) (coredata.Controls, error) {
	var controls coredata.Controls

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := controls.LoadByIDs(
				ctx,
				conn,
				scope,
				controlIDs,
			); err != nil {
				return fmt.Errorf("cannot load controls by ids: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return controls, nil
}

func (s ControlService) Update(
	ctx context.Context, scope coredata.Scoper,
	req UpdateControlRequest,
) (*coredata.Control, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	control := &coredata.Control{ID: req.ID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := control.LoadByID(ctx, conn, scope, req.ID); err != nil {
				return fmt.Errorf("cannot load control: %w", err)
			}

			if req.Name != nil {
				control.Name = *req.Name
			}

			if req.Description != nil {
				control.Description = *req.Description
			}

			if req.SectionTitle != nil {
				control.SectionTitle = *req.SectionTitle
			}

			if req.BestPractice != nil {
				control.BestPractice = *req.BestPractice
			}

			if req.MaturityLevel != nil {
				control.MaturityLevel = *req.MaturityLevel
				if *req.MaturityLevel != coredata.ControlMaturityLevelNone {
					control.NotImplementedJustification = nil
				}
			}

			if req.NotImplementedJustification != nil && control.MaturityLevel == coredata.ControlMaturityLevelNone {
				control.NotImplementedJustification = *req.NotImplementedJustification
			}

			control.UpdatedAt = time.Now()

			return control.Update(ctx, conn, scope)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot update control: %w", err)
	}

	return control, nil
}

func (s ControlService) Delete(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
) error {
	control := &coredata.Control{ID: controlID}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return control.Delete(ctx, tx, scope)
		},
	)
}

func (s ControlService) HasRegulatoryObligation(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
) (bool, error) {
	var hasRegulatory bool

	obligationType := coredata.ObligationTypeLegal
	filter := coredata.NewControlObligationFilter(&obligationType)

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var controlObligations coredata.ControlObligations

			count, err := controlObligations.CountByControlID(ctx, conn, scope, controlID, filter)
			if err != nil {
				return fmt.Errorf("cannot count regulatory obligations: %w", err)
			}

			hasRegulatory = count > 0

			return nil
		},
	)

	return hasRegulatory, err
}

func (s ControlService) HasContractualObligation(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
) (bool, error) {
	var hasContractual bool

	obligationType := coredata.ObligationTypeContractual
	filter := coredata.NewControlObligationFilter(&obligationType)

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var controlObligations coredata.ControlObligations

			count, err := controlObligations.CountByControlID(ctx, conn, scope, controlID, filter)
			if err != nil {
				return fmt.Errorf("cannot count contractual obligations: %w", err)
			}

			hasContractual = count > 0

			return nil
		},
	)

	return hasContractual, err
}

func (s ControlService) HasRiskAssessment(
	ctx context.Context, scope coredata.Scoper,
	controlID gid.GID,
) (bool, error) {
	var hasRisk bool

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var controlsWithRisk coredata.ControlsWithRisk
			if err := controlsWithRisk.LoadByControlIDs(ctx, conn, scope, []gid.GID{controlID}); err != nil {
				return fmt.Errorf("cannot load controls with risk: %w", err)
			}

			hasRisk = len(controlsWithRisk) > 0

			return nil
		},
	)

	return hasRisk, err
}
