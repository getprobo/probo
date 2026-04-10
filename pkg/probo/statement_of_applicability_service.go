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

type StatementOfApplicabilityService struct {
	svc *TenantService
}

type (
	CreateStatementOfApplicabilityRequest struct {
		OrganizationID     gid.GID
		Name               string
		DefaultApproverIDs []gid.GID
	}

	UpdateStatementOfApplicabilityRequest struct {
		StatementOfApplicabilityID gid.GID
		Name                       *string
		DefaultApproverIDs         *[]gid.GID
	}
)

func (csr *CreateStatementOfApplicabilityRequest) Validate() error {
	v := validator.New()

	v.Check(csr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(csr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(len(csr.DefaultApproverIDs), "default_approver_ids", validator.Max(100))
	v.Check(csr.DefaultApproverIDs, "default_approver_ids", validator.NoDuplicates())
	v.CheckEach(csr.DefaultApproverIDs, "default_approver_ids", func(_ int, item any) {
		v.Check(item, "default_approver_ids", validator.GID(coredata.MembershipProfileEntityType))
	})

	return v.Error()
}

func (usr *UpdateStatementOfApplicabilityRequest) Validate() error {
	v := validator.New()

	v.Check(usr.StatementOfApplicabilityID, "statement_of_applicability_id", validator.Required(), validator.GID(coredata.StatementOfApplicabilityEntityType))
	v.Check(usr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	if usr.DefaultApproverIDs != nil {
		v.Check(len(*usr.DefaultApproverIDs), "default_approver_ids", validator.Max(100))
		v.Check(*usr.DefaultApproverIDs, "default_approver_ids", validator.NoDuplicates())
		v.CheckEach(*usr.DefaultApproverIDs, "default_approver_ids", func(_ int, item any) {
			v.Check(item, "default_approver_ids", validator.GID(coredata.MembershipProfileEntityType))
		})
	}

	return v.Error()
}

func (s StatementOfApplicabilityService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.StatementOfApplicabilityOrderField],
) (*page.Page[*coredata.StatementOfApplicability, coredata.StatementOfApplicabilityOrderField], error) {
	var statementsOfApplicability coredata.StatementsOfApplicability
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := statementsOfApplicability.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organization.ID,
				cursor,
			)
			if err != nil {
				return fmt.Errorf("cannot load statements_of_applicability: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(statementsOfApplicability, cursor), nil
}

func (s StatementOfApplicabilityService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			statementsOfApplicability := &coredata.StatementsOfApplicability{}
			count, err = statementsOfApplicability.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count statements_of_applicability: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s StatementOfApplicabilityService) Get(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
) (*coredata.StatementOfApplicability, error) {
	statementOfApplicability := &coredata.StatementOfApplicability{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return statementOfApplicability.LoadByID(ctx, conn, s.svc.scope, statementOfApplicabilityID)
		},
	)

	if err != nil {
		return nil, err
	}

	return statementOfApplicability, nil
}

func (s StatementOfApplicabilityService) GetDefaultApprovers(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
) (coredata.MembershipProfiles, error) {
	var approvers coredata.StatementOfApplicabilityDefaultApprovers

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return approvers.LoadByStatementOfApplicabilityID(ctx, conn, s.svc.scope, statementOfApplicabilityID)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot load default approvers: %w", err)
	}

	if len(approvers) == 0 {
		return nil, nil
	}

	profileIDs := make([]gid.GID, len(approvers))
	for i, a := range approvers {
		profileIDs[i] = a.ApproverProfileID
	}

	var profiles coredata.MembershipProfiles

	err = s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return profiles.LoadByIDs(ctx, conn, s.svc.scope, profileIDs)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot load approver profiles: %w", err)
	}

	return profiles, nil
}

func (s StatementOfApplicabilityService) Create(
	ctx context.Context,
	req CreateStatementOfApplicabilityRequest,
) (*coredata.StatementOfApplicability, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load organization: %w", err)
	}

	statementOfApplicabilityID := gid.New(organization.ID.TenantID(), coredata.StatementOfApplicabilityEntityType)
	statementOfApplicability := &coredata.StatementOfApplicability{
		ID:             statementOfApplicabilityID,
		OrganizationID: organization.ID,
		Name:           req.Name,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := statementOfApplicability.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert statement_of_applicability: %w", err)
			}

			if len(req.DefaultApproverIDs) > 0 {
				approvers := &coredata.StatementOfApplicabilityDefaultApprovers{}
				if err := approvers.MergeByStatementOfApplicabilityID(
					ctx,
					conn,
					s.svc.scope,
					statementOfApplicabilityID,
					organization.ID,
					req.DefaultApproverIDs,
				); err != nil {
					return fmt.Errorf("cannot set default approvers: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return statementOfApplicability, nil
}

func (s StatementOfApplicabilityService) Update(
	ctx context.Context,
	req UpdateStatementOfApplicabilityRequest,
) (*coredata.StatementOfApplicability, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	statementOfApplicability := &coredata.StatementOfApplicability{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := statementOfApplicability.LoadByID(ctx, conn, s.svc.scope, req.StatementOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load statement_of_applicability: %w", err)
			}

			if req.Name != nil {
				statementOfApplicability.Name = *req.Name
			}

			statementOfApplicability.UpdatedAt = time.Now()

			if err := statementOfApplicability.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update statement_of_applicability: %w", err)
			}

			if req.DefaultApproverIDs != nil {
				approvers := &coredata.StatementOfApplicabilityDefaultApprovers{}
				if err := approvers.MergeByStatementOfApplicabilityID(
					ctx,
					conn,
					s.svc.scope,
					req.StatementOfApplicabilityID,
					statementOfApplicability.OrganizationID,
					*req.DefaultApproverIDs,
				); err != nil {
					return fmt.Errorf("cannot update default approvers: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return statementOfApplicability, nil
}

func (s StatementOfApplicabilityService) Delete(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
) error {
	statementOfApplicability := &coredata.StatementOfApplicability{ID: statementOfApplicabilityID}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := statementOfApplicability.LoadByID(ctx, conn, s.svc.scope, statementOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load statement_of_applicability: %w", err)
			}

			if err := statementOfApplicability.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete statement_of_applicability: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (s StatementOfApplicabilityService) GetApplicabilityStatement(
	ctx context.Context,
	applicabilityStatementID gid.GID,
) (*coredata.ApplicabilityStatement, error) {
	applicabilityStatement := &coredata.ApplicabilityStatement{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return applicabilityStatement.LoadByID(ctx, conn, s.svc.scope, applicabilityStatementID)
		},
	)
	if err != nil {
		return nil, err
	}

	return applicabilityStatement, nil
}

func (s StatementOfApplicabilityService) ListApplicabilityStatements(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
	cursor *page.Cursor[coredata.ApplicabilityStatementOrderField],
) (*page.Page[*coredata.ApplicabilityStatement, coredata.ApplicabilityStatementOrderField], error) {
	var statements coredata.ApplicabilityStatements

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := statements.LoadByStatementOfApplicabilityID(ctx, conn, s.svc.scope, statementOfApplicabilityID, cursor); err != nil {
				return fmt.Errorf("cannot load applicability statements: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(statements, cursor), nil
}

func (s StatementOfApplicabilityService) CountApplicabilityStatements(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			statements := &coredata.ApplicabilityStatements{}
			count, err = statements.CountByStatementOfApplicabilityID(ctx, conn, s.svc.scope, statementOfApplicabilityID)
			if err != nil {
				return fmt.Errorf("cannot count applicability statements: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s StatementOfApplicabilityService) CreateApplicabilityStatement(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
	controlID gid.GID,
	applicability bool,
	justification *string,
) (*coredata.ApplicabilityStatement, error) {
	var (
		statementOfApplicability = &coredata.StatementOfApplicability{}
		applicabilityStatement   = &coredata.ApplicabilityStatement{}
		now                      = time.Now()
	)

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := statementOfApplicability.LoadByID(ctx, conn, s.svc.scope, statementOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load statement of applicability: %w", err)
			}

			applicabilityStatement = &coredata.ApplicabilityStatement{
				ID:                         gid.New(s.svc.scope.GetTenantID(), coredata.ApplicabilityStatementEntityType),
				StatementOfApplicabilityID: statementOfApplicabilityID,
				ControlID:                  controlID,
				OrganizationID:             statementOfApplicability.OrganizationID,
				Applicability:              applicability,
				Justification:              justification,
				CreatedAt:                  now,
				UpdatedAt:                  now,
			}

			if err := applicabilityStatement.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert applicability statement: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return applicabilityStatement, nil
}

func (s StatementOfApplicabilityService) UpdateApplicabilityStatement(
	ctx context.Context,
	applicabilityStatementID gid.GID,
	applicability bool,
	justification *string,
) (*coredata.ApplicabilityStatement, error) {
	applicabilityStatement := &coredata.ApplicabilityStatement{}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := applicabilityStatement.LoadByID(ctx, conn, s.svc.scope, applicabilityStatementID); err != nil {
				return err
			}

			applicabilityStatement.Applicability = applicability
			applicabilityStatement.Justification = justification
			applicabilityStatement.UpdatedAt = time.Now()

			return applicabilityStatement.UpdateByID(ctx, conn, s.svc.scope)
		},
	)
	if err != nil {
		return nil, err
	}

	return applicabilityStatement, nil
}

func (s StatementOfApplicabilityService) DeleteApplicabilityStatement(
	ctx context.Context,
	applicabilityStatementID gid.GID,
) error {
	applicabilityStatement := &coredata.ApplicabilityStatement{}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			return applicabilityStatement.DeleteByID(ctx, conn, s.svc.scope, applicabilityStatementID)
		},
	)
}

func (s StatementOfApplicabilityService) ListControlLinks(
	ctx context.Context,
	controlID gid.GID,
	cursor *page.Cursor[coredata.ApplicabilityStatementOrderField],
) (*page.Page[*coredata.ApplicabilityStatement, coredata.ApplicabilityStatementOrderField], error) {
	var controls coredata.ApplicabilityStatements

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return controls.LoadByControlID(ctx, conn, s.svc.scope, controlID, cursor)
	})
	if err != nil {
		return nil, err
	}

	return page.NewPage(controls, cursor), nil
}
