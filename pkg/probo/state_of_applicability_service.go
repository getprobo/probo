// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

type StateOfApplicabilityService struct {
	svc               *TenantService
	html2pdfConverter *html2pdf.Converter
}

type (
	CreateStateOfApplicabilityRequest struct {
		OrganizationID gid.GID
		Name           string
		OwnerID        gid.GID
	}

	UpdateStateOfApplicabilityRequest struct {
		StateOfApplicabilityID gid.GID
		Name                   *string
		OwnerID                *gid.GID
	}
)

func (csr *CreateStateOfApplicabilityRequest) Validate() error {
	v := validator.New()

	v.Check(csr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(csr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(csr.OwnerID, "owner_id", validator.Required(), validator.GID(coredata.PeopleEntityType))

	return v.Error()
}

func (usr *UpdateStateOfApplicabilityRequest) Validate() error {
	v := validator.New()

	v.Check(usr.StateOfApplicabilityID, "state_of_applicability_id", validator.Required(), validator.GID(coredata.StateOfApplicabilityEntityType))
	v.Check(usr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(usr.OwnerID, "owner_id", validator.GID(coredata.PeopleEntityType))

	return v.Error()
}

func (s StateOfApplicabilityService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.StateOfApplicabilityOrderField],
	filter *coredata.StateOfApplicabilityFilter,
) (*page.Page[*coredata.StateOfApplicability, coredata.StateOfApplicabilityOrderField], error) {
	var statesOfApplicability coredata.StatesOfApplicability
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := statesOfApplicability.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organization.ID,
				cursor,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot load states_of_applicability: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(statesOfApplicability, cursor), nil
}

func (s StateOfApplicabilityService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.StateOfApplicabilityFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			statesOfApplicability := &coredata.StatesOfApplicability{}
			count, err = statesOfApplicability.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count states_of_applicability: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s StateOfApplicabilityService) Get(
	ctx context.Context,
	stateOfApplicabilityID gid.GID,
) (*coredata.StateOfApplicability, error) {
	stateOfApplicability := &coredata.StateOfApplicability{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return stateOfApplicability.LoadByID(ctx, conn, s.svc.scope, stateOfApplicabilityID)
		},
	)

	if err != nil {
		return nil, err
	}

	return stateOfApplicability, nil
}

func (s StateOfApplicabilityService) Create(
	ctx context.Context,
	req CreateStateOfApplicabilityRequest,
) (*coredata.StateOfApplicability, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	now := time.Now()
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load organization: %w", err)
	}

	stateOfApplicabilityID := gid.New(organization.ID.TenantID(), coredata.StateOfApplicabilityEntityType)
	stateOfApplicability := &coredata.StateOfApplicability{
		ID:             stateOfApplicabilityID,
		OrganizationID: organization.ID,
		Name:           req.Name,
		OwnerID:        req.OwnerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := stateOfApplicability.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert state_of_applicability: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return stateOfApplicability, nil
}

func (s StateOfApplicabilityService) Update(
	ctx context.Context,
	req UpdateStateOfApplicabilityRequest,
) (*coredata.StateOfApplicability, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	stateOfApplicability := &coredata.StateOfApplicability{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := stateOfApplicability.LoadByID(ctx, conn, s.svc.scope, req.StateOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load state_of_applicability: %w", err)
			}

			if req.Name != nil {
				stateOfApplicability.Name = *req.Name
			}
			if req.OwnerID != nil {
				stateOfApplicability.OwnerID = *req.OwnerID
			}

			stateOfApplicability.UpdatedAt = time.Now()

			if err := stateOfApplicability.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update state_of_applicability: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return stateOfApplicability, nil
}

func (s StateOfApplicabilityService) Delete(
	ctx context.Context,
	stateOfApplicabilityID gid.GID,
) error {
	stateOfApplicability := &coredata.StateOfApplicability{ID: stateOfApplicabilityID}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := stateOfApplicability.LoadByID(ctx, conn, s.svc.scope, stateOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load state_of_applicability: %w", err)
			}

			if err := stateOfApplicability.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete state_of_applicability: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (s StateOfApplicabilityService) ListAvailableControls(
	ctx context.Context,
	stateOfApplicabilityID gid.GID,
) ([]*coredata.AvailableStateOfApplicabilityControl, error) {
	var availableControls coredata.AvailableStateOfApplicabilityControls

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := availableControls.LoadAvailableByStateOfApplicabilityID(ctx, conn, s.svc.scope, stateOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load available controls: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return availableControls, nil
}

func (s StateOfApplicabilityService) LinkControl(
	ctx context.Context,
	stateOfApplicabilityID gid.GID,
	controlID gid.GID,
	applicability bool,
	justification *string,
) (*coredata.StateOfApplicabilityControl, error) {
	stateOfApplicability := &coredata.StateOfApplicability{}
	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return stateOfApplicability.LoadByID(ctx, conn, s.svc.scope, stateOfApplicabilityID)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot load state of applicability: %w", err)
	}

	now := time.Now()
	control := &coredata.StateOfApplicabilityControl{
		ID:                     gid.New(s.svc.scope.GetTenantID(), coredata.StateOfApplicabilityControlEntityType),
		StateOfApplicabilityID: stateOfApplicabilityID,
		ControlID:              controlID,
		OrganizationID:         stateOfApplicability.OrganizationID,
		Applicability:          applicability,
		Justification:          justification,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err = s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		return control.Upsert(ctx, conn, s.svc.scope)
	})
	if err != nil {
		return nil, err
	}

	return control, nil
}

func (s StateOfApplicabilityService) DeleteControlLink(
	ctx context.Context,
	stateOfApplicabilityID gid.GID,
	controlID gid.GID,
) (gid.GID, error) {
	control := &coredata.StateOfApplicabilityControl{}

	err := s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		if err := control.LoadByStateOfApplicabilityIDAndControlID(ctx, conn, s.svc.scope, stateOfApplicabilityID, controlID); err != nil {
			return err
		}
		return control.Delete(ctx, conn, s.svc.scope)
	})
	if err != nil {
		return gid.GID{}, err
	}

	return control.ID, nil
}

func (s StateOfApplicabilityService) ListControlLinks(
	ctx context.Context,
	controlID gid.GID,
	cursor *page.Cursor[coredata.StateOfApplicabilityOrderField],
) (*page.Page[*coredata.StateOfApplicabilityControl, coredata.StateOfApplicabilityOrderField], error) {
	var controls coredata.StateOfApplicabilityControls

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return controls.LoadByControlID(ctx, conn, s.svc.scope, controlID, cursor)
	})
	if err != nil {
		return nil, err
	}

	return page.NewPage(controls, cursor), nil
}

func (s StateOfApplicabilityService) ExportPDF(
	ctx context.Context,
	stateOfApplicabilityID gid.GID,
) ([]byte, error) {
	var documentData docgen.StateOfApplicabilityData

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			stateOfApplicability := &coredata.StateOfApplicability{}
			if err := stateOfApplicability.LoadByID(ctx, conn, s.svc.scope, stateOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load state of applicability: %w", err)
			}

			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, stateOfApplicability.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			owner := &coredata.People{}
			if err := owner.LoadByID(ctx, conn, s.svc.scope, stateOfApplicability.OwnerID); err != nil {
				return fmt.Errorf("cannot load owner: %w", err)
			}

			var availableControls coredata.AvailableStateOfApplicabilityControls
			if err := availableControls.LoadAvailableByStateOfApplicabilityID(ctx, conn, s.svc.scope, stateOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load available controls: %w", err)
			}

			linkedControls := make([]*coredata.AvailableStateOfApplicabilityControl, 0)
			for _, ctrl := range availableControls {
				if ctrl.StateOfApplicabilityID != nil {
					linkedControls = append(linkedControls, ctrl)
				}
			}

			obligationsByControl := make(map[gid.GID][]coredata.ObligationType)
			if len(linkedControls) > 0 {
				controlIDs := make([]gid.GID, len(linkedControls))
				for i, ctrl := range linkedControls {
					controlIDs[i] = ctrl.ControlID
				}

				var controlObligationTypes coredata.ControlObligationTypes
				if err := controlObligationTypes.LoadTypesByControlIDs(ctx, conn, s.svc.scope, controlIDs); err != nil {
					return fmt.Errorf("cannot load control obligations: %w", err)
				}

				for _, cot := range controlObligationTypes {
					obligationsByControl[cot.ControlID] = append(obligationsByControl[cot.ControlID], cot.Type)
				}
			}

			controlsWithRisks := make(map[gid.GID]bool)
			if len(linkedControls) > 0 {
				controlIDs := make([]gid.GID, len(linkedControls))
				for i, ctrl := range linkedControls {
					controlIDs[i] = ctrl.ControlID
				}

				var controlsWithRisk coredata.ControlsWithRisk
				if err := controlsWithRisk.LoadByControlIDs(ctx, conn, s.svc.scope, controlIDs); err != nil {
					return fmt.Errorf("cannot load controls with risks: %w", err)
				}

				for _, cwr := range controlsWithRisk {
					controlsWithRisks[cwr.ControlID] = true
				}
			}

			frameworkControlsMap := make(map[string][]docgen.ControlData)
			frameworkOrder := []string{}

			for _, ctrl := range linkedControls {
				if _, exists := frameworkControlsMap[ctrl.FrameworkName]; !exists {
					frameworkOrder = append(frameworkOrder, ctrl.FrameworkName)
					frameworkControlsMap[ctrl.FrameworkName] = []docgen.ControlData{}
				}

				var regulatory *bool
				var contractual *bool
				var riskAssessment *bool

				if ctrl.Applicability != nil && *ctrl.Applicability {
					falseVal := false
					trueVal := true

					regulatory = &falseVal
					contractual = &falseVal
					riskAssessment = &falseVal

					obligations := obligationsByControl[ctrl.ControlID]
					for _, obligationType := range obligations {
						if obligationType == coredata.ObligationTypeLegal {
							regulatory = &trueVal
						}
						if obligationType == coredata.ObligationTypeContractual {
							contractual = &trueVal
						}
					}

					if controlsWithRisks[ctrl.ControlID] {
						riskAssessment = &trueVal
					}
				}

				frameworkControlsMap[ctrl.FrameworkName] = append(
					frameworkControlsMap[ctrl.FrameworkName],
					docgen.ControlData{
						FrameworkName:  ctrl.FrameworkName,
						SectionTitle:   ctrl.SectionTitle,
						Name:           ctrl.Name,
						Applicability:  ctrl.Applicability,
						Justification:  ctrl.Justification,
						BestPractice:   ctrl.BestPractice,
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
			snapshotType := coredata.SnapshotsTypeStatesOfApplicability

			var version int
			var publishedAt time.Time

			if stateOfApplicability.SnapshotID != nil {
				snapshot := &coredata.Snapshot{}
				if err := snapshot.LoadByID(ctx, conn, s.svc.scope, *stateOfApplicability.SnapshotID); err != nil {
					return fmt.Errorf("cannot load snapshot: %w", err)
				}
				publishedAt = snapshot.CreatedAt
				snapshotFilter := coredata.NewSnapshotFilter(&snapshotType).WithBeforeDate(&snapshot.CreatedAt)
				snapshotCount, err := snapshots.CountByOrganizationID(ctx, conn, s.svc.scope, stateOfApplicability.OrganizationID, snapshotFilter)
				if err != nil {
					return fmt.Errorf("cannot count states of applicability snapshots: %w", err)
				}
				version = snapshotCount
			} else {
				publishedAt = time.Now()
				snapshotFilter := coredata.NewSnapshotFilter(&snapshotType)
				snapshotCount, err := snapshots.CountByOrganizationID(ctx, conn, s.svc.scope, stateOfApplicability.OrganizationID, snapshotFilter)
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

			documentData = docgen.StateOfApplicabilityData{
				Title:                       stateOfApplicability.Name,
				OrganizationName:            organization.Name,
				CreatedAt:                   stateOfApplicability.CreatedAt,
				TotalControls:               len(linkedControls),
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

	htmlData, err := docgen.RenderStateOfApplicabilityHTML(documentData)
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
