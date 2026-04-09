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
	"io"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type StatementOfApplicabilityService struct {
	svc               *TenantService
	html2pdfConverter *html2pdf.Converter
}

type (
	CreateStatementOfApplicabilityRequest struct {
		OrganizationID gid.GID
		Name           string
		OwnerID        gid.GID
	}

	UpdateStatementOfApplicabilityRequest struct {
		StatementOfApplicabilityID gid.GID
		Name                       *string
		OwnerID                    *gid.GID
	}
)

func (csr *CreateStatementOfApplicabilityRequest) Validate() error {
	v := validator.New()

	v.Check(csr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(csr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(csr.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (usr *UpdateStatementOfApplicabilityRequest) Validate() error {
	v := validator.New()

	v.Check(usr.StatementOfApplicabilityID, "statement_of_applicability_id", validator.Required(), validator.GID(coredata.StatementOfApplicabilityEntityType))
	v.Check(usr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(usr.OwnerID, "owner_id", validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (s StatementOfApplicabilityService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.StatementOfApplicabilityOrderField],
	filter *coredata.StatementOfApplicabilityFilter,
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
				filter,
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
	filter *coredata.StatementOfApplicabilityFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
			statementsOfApplicability := &coredata.StatementsOfApplicability{}
			count, err = statementsOfApplicability.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
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
		OwnerID:        req.OwnerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := statementOfApplicability.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert statement_of_applicability: %w", err)
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
			if req.OwnerID != nil {
				statementOfApplicability.OwnerID = *req.OwnerID
			}

			statementOfApplicability.UpdatedAt = time.Now()

			if err := statementOfApplicability.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update statement_of_applicability: %w", err)
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

func (s StatementOfApplicabilityService) ExportPDF(
	ctx context.Context,
	statementOfApplicabilityID gid.GID,
) ([]byte, error) {
	var documentData docgen.StatementOfApplicabilityData

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			statementOfApplicability := &coredata.StatementOfApplicability{}
			if err := statementOfApplicability.LoadByID(ctx, conn, s.svc.scope, statementOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load statement of applicability: %w", err)
			}

			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, statementOfApplicability.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			owner := &coredata.MembershipProfile{}
			if err := owner.LoadByID(ctx, conn, s.svc.scope, statementOfApplicability.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner profile: %w", err)
			}

			// Load applicability statements
			var applicabilityStatements coredata.ApplicabilityStatements
			cursor := page.NewCursor(
				10000,
				nil,
				page.Head,
				page.OrderBy[coredata.ApplicabilityStatementOrderField]{
					Field:     coredata.ApplicabilityStatementOrderFieldControlSectionTitle,
					Direction: page.OrderDirectionAsc,
				},
			)
			if err := applicabilityStatements.LoadByStatementOfApplicabilityID(ctx, conn, s.svc.scope, statementOfApplicabilityID, cursor); err != nil {
				return fmt.Errorf("cannot load applicability statements: %w", err)
			}

			if len(applicabilityStatements) == 0 {
				// No linked controls, skip loading additional data
				documentData = docgen.StatementOfApplicabilityData{
					Title:            statementOfApplicability.Name,
					OrganizationName: organization.Name,
					CreatedAt:        statementOfApplicability.CreatedAt,
					TotalControls:    0,
					FrameworkGroups:  []docgen.FrameworkControlGroup{},
				}
				return nil
			}

			frameworkControlsMap := make(map[string][]docgen.ControlData)
			frameworkOrder := []string{}

			for _, stmt := range applicabilityStatements {
				// Load control
				control := &coredata.Control{}
				if err := control.LoadByID(ctx, conn, s.svc.scope, stmt.ControlID); err != nil {
					return fmt.Errorf("cannot load control: %w", err)
				}

				// Load framework
				framework := &coredata.Framework{}
				if err := framework.LoadByID(ctx, conn, s.svc.scope, control.FrameworkID); err != nil {
					return fmt.Errorf("cannot load framework: %w", err)
				}

				// Count legal obligations
				var controlObligations coredata.ControlObligations
				legalType := coredata.ObligationTypeLegal
				legalFilter := coredata.NewControlObligationFilter(&legalType)
				legalCount, err := controlObligations.CountByControlID(ctx, conn, s.svc.scope, stmt.ControlID, legalFilter)
				if err != nil {
					return fmt.Errorf("cannot count legal obligations: %w", err)
				}

				// Count contractual obligations
				contractualType := coredata.ObligationTypeContractual
				contractualFilter := coredata.NewControlObligationFilter(&contractualType)
				contractualCount, err := controlObligations.CountByControlID(ctx, conn, s.svc.scope, stmt.ControlID, contractualFilter)
				if err != nil {
					return fmt.Errorf("cannot count contractual obligations: %w", err)
				}

				// Check if control has risk
				var controlsWithRisk coredata.ControlsWithRisk
				if err := controlsWithRisk.LoadByControlIDs(ctx, conn, s.svc.scope, []gid.GID{stmt.ControlID}); err != nil {
					return fmt.Errorf("cannot load controls with risks: %w", err)
				}
				hasRisk := len(controlsWithRisk) > 0

				if _, exists := frameworkControlsMap[framework.Name]; !exists {
					frameworkOrder = append(frameworkOrder, framework.Name)
					frameworkControlsMap[framework.Name] = []docgen.ControlData{}
				}

				var regulatory *bool
				var contractual *bool
				var bestPractice *bool
				var riskAssessment *bool

				if stmt.Applicability {
					falseVal := false
					trueVal := true

					regulatory = &falseVal
					contractual = &falseVal
					riskAssessment = &falseVal

					if legalCount > 0 {
						regulatory = &trueVal
					}
					if contractualCount > 0 {
						contractual = &trueVal
					}
					if hasRisk {
						riskAssessment = &trueVal
					}

					bestPractice = &control.BestPractice
				}

				applicability := stmt.Applicability

				implemented := control.Implemented.String()
				frameworkControlsMap[framework.Name] = append(
					frameworkControlsMap[framework.Name],
					docgen.ControlData{
						FrameworkName: framework.Name,
						SectionTitle:  control.SectionTitle,
						Name:          control.Name,
						Applicability: &applicability,
						Justification: stmt.Justification,
						BestPractice:  bestPractice,
						Implemented:   &implemented,
						NotImplementedJustification: func() *string {
							if control.Implemented == coredata.ControlImplementationStateImplemented {
								return nil
							}
							return control.NotImplementedJustification
						}(),
						Regulatory:     regulatory,
						Contractual:    contractual,
						RiskAssessment: riskAssessment,
					},
				)
			}

			frameworkGroups := make([]docgen.FrameworkControlGroup, len(frameworkOrder))
			for i, frameworkName := range frameworkOrder {
				frameworkGroups[i] = docgen.FrameworkControlGroup{
					FrameworkName: frameworkName,
					Controls:      frameworkControlsMap[frameworkName],
				}
			}

			var snapshots coredata.Snapshots
			snapshotType := coredata.SnapshotsTypeStatementsOfApplicability

			var version int
			var publishedAt time.Time

			if statementOfApplicability.SnapshotID != nil {
				snapshot := &coredata.Snapshot{}
				if err := snapshot.LoadByID(ctx, conn, s.svc.scope, *statementOfApplicability.SnapshotID); err != nil {
					return fmt.Errorf("cannot load snapshot: %w", err)
				}
				publishedAt = snapshot.CreatedAt
				snapshotFilter := coredata.NewSnapshotFilter(&snapshotType).WithBeforeDate(&snapshot.CreatedAt)
				snapshotCount, err := snapshots.CountByOrganizationID(ctx, conn, s.svc.scope, statementOfApplicability.OrganizationID, snapshotFilter)
				if err != nil {
					return fmt.Errorf("cannot count states of applicability snapshots: %w", err)
				}
				version = snapshotCount
			} else {
				publishedAt = time.Now()
				snapshotFilter := coredata.NewSnapshotFilter(&snapshotType)
				snapshotCount, err := snapshots.CountByOrganizationID(ctx, conn, s.svc.scope, statementOfApplicability.OrganizationID, snapshotFilter)
				if err != nil {
					return fmt.Errorf("cannot count states of applicability snapshots: %w", err)
				}
				version = snapshotCount + 1
			}

			horizontalLogoBase64 := ""
			if organization.HorizontalLogoFileID != nil {
				fileRecord := &coredata.File{}
				fileErr := fileRecord.LoadByID(ctx, conn, s.svc.scope, *organization.HorizontalLogoFileID)
				if fileErr == nil {
					base64Data, mimeType, logoErr := s.svc.fileManager.GetFileBase64(ctx, fileRecord)
					if logoErr == nil {
						horizontalLogoBase64 = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
					}
				}
			}

			documentData = docgen.StatementOfApplicabilityData{
				Title:                       statementOfApplicability.Name,
				OrganizationName:            organization.Name,
				CreatedAt:                   statementOfApplicability.CreatedAt,
				TotalControls:               len(applicabilityStatements),
				FrameworkGroups:             frameworkGroups,
				CompanyHorizontalLogoBase64: horizontalLogoBase64,
				Version:                     version,
				PublishedAt:                 publishedAt,
				Approver:                    owner.FullName,
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	htmlData, err := docgen.RenderStatementOfApplicabilityHTML(documentData)
	if err != nil {
		return nil, fmt.Errorf("cannot render HTML: %w", err)
	}

	cfg := html2pdf.RenderConfig{
		PageFormat:      html2pdf.PageFormatA4,
		Orientation:     html2pdf.OrientationPortrait,
		MarginTop:       html2pdf.NewMarginInches(1.0),
		MarginBottom:    html2pdf.NewMarginInches(1.0),
		MarginLeft:      html2pdf.NewMarginInches(1.0),
		MarginRight:     html2pdf.NewMarginInches(1.0),
		PrintBackground: true,
		Scale:           1.0,
	}

	pdfReader, err := s.html2pdfConverter.GeneratePDF(ctx, htmlData, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot generate PDF: %w", err)
	}

	pdfData, err := io.ReadAll(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("cannot read PDF data: %w", err)
	}

	return pdfData, nil
}
