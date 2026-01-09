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
		Description    *string
		SourceID       *gid.GID
		SnapshotID     *gid.GID
	}

	UpdateStateOfApplicabilityRequest struct {
		StateOfApplicabilityID gid.GID
		Name                   *string
		Description            *string
	}
)

func (csr *CreateStateOfApplicabilityRequest) Validate() error {
	v := validator.New()

	v.Check(csr.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(csr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(csr.SourceID, "source_id", validator.GID(coredata.FrameworkEntityType))
	v.Check(csr.SnapshotID, "snapshot_id", validator.GID(coredata.SnapshotEntityType))

	return v.Error()
}

func (usr *UpdateStateOfApplicabilityRequest) Validate() error {
	v := validator.New()

	v.Check(usr.StateOfApplicabilityID, "state_of_applicability_id", validator.Required(), validator.GID(coredata.StateOfApplicabilityEntityType))
	v.Check(usr.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))

	return v.Error()
}

func (s StateOfApplicabilityService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.StateOfApplicabilityOrderField],
	filter *coredata.StateOfApplicabilityFilter,
) (*page.Page[*coredata.StateOfApplicability, coredata.StateOfApplicabilityOrderField], error) {
	var stateOfApplicabilities coredata.StateOfApplicabilities
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			err := stateOfApplicabilities.LoadByOrganizationID(
				ctx,
				conn,
				s.svc.scope,
				organization.ID,
				cursor,
				filter,
			)
			if err != nil {
				return fmt.Errorf("cannot load state_of_applicabilities: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(stateOfApplicabilities, cursor), nil
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
			stateOfApplicabilities := &coredata.StateOfApplicabilities{}
			count, err = stateOfApplicabilities.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count state_of_applicabilities: %w", err)
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
		Description:    req.Description,
		SourceID:       req.SourceID,
		SnapshotID:     req.SnapshotID,
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
			if req.Description != nil {
				stateOfApplicability.Description = req.Description
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
) ([]*coredata.AvailableControlForStateOfApplicability, error) {
	var availableControls coredata.AvailableControlsForStateOfApplicability

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
	state coredata.StateOfApplicabilityControlState,
	exclusionJustification *string,
) error {
	stateOfApplicability := &coredata.StateOfApplicability{}
	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return stateOfApplicability.LoadByID(ctx, conn, s.svc.scope, stateOfApplicabilityID)
	})
	if err != nil {
		return fmt.Errorf("cannot load state of applicability: %w", err)
	}

	control := &coredata.StateOfApplicabilityControl{
		StateOfApplicabilityID: stateOfApplicabilityID,
		ControlID:              controlID,
		State:                  state,
		ExclusionJustification: exclusionJustification,
		CreatedAt:              time.Now(),
	}

	err = s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		return control.Upsert(ctx, conn, s.svc.scope)
	})

	return err
}

func (s StateOfApplicabilityService) UnlinkControl(
	ctx context.Context,
	stateOfApplicabilityID gid.GID,
	controlID gid.GID,
) error {
	control := &coredata.StateOfApplicabilityControl{
		StateOfApplicabilityID: stateOfApplicabilityID,
		ControlID:              controlID,
	}

	return s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		return control.Delete(ctx, conn, s.svc.scope)
	})
}

func (s StateOfApplicabilityService) ExportPDF(
	ctx context.Context,
	stateOfApplicabilityID gid.GID,
) ([]byte, error) {
	// Fetch all required data within a transaction
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

			var availableControls coredata.AvailableControlsForStateOfApplicability
			if err := availableControls.LoadAvailableByStateOfApplicabilityID(ctx, conn, s.svc.scope, stateOfApplicabilityID); err != nil {
				return fmt.Errorf("cannot load available controls: %w", err)
			}

			linkedControls := make([]*coredata.AvailableControlForStateOfApplicability, 0)
			for _, ctrl := range availableControls {
				if ctrl.StateOfApplicabilityID != nil {
					linkedControls = append(linkedControls, ctrl)
				}
			}

			frameworkControlsMap := make(map[string][]docgen.ControlData)
			frameworkOrder := []string{}

			for _, ctrl := range linkedControls {
				if _, exists := frameworkControlsMap[ctrl.FrameworkName]; !exists {
					frameworkOrder = append(frameworkOrder, ctrl.FrameworkName)
					frameworkControlsMap[ctrl.FrameworkName] = []docgen.ControlData{}
				}

				stateStr := ""
				if ctrl.State != nil {
					stateStr = string(*ctrl.State)
				}
				exclusionJustification := ""
				if ctrl.ExclusionJustification != nil {
					exclusionJustification = *ctrl.ExclusionJustification
				}

				frameworkControlsMap[ctrl.FrameworkName] = append(
					frameworkControlsMap[ctrl.FrameworkName],
					docgen.ControlData{
						FrameworkName:          ctrl.FrameworkName,
						SectionTitle:           ctrl.SectionTitle,
						Name:                   ctrl.Name,
						State:                  stateStr,
						ExclusionJustification: exclusionJustification,
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

			// Calculate version and published date
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

			// Get company logo
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

			description := ""
			if stateOfApplicability.Description != nil {
				description = *stateOfApplicability.Description
			}

			// Build document data (all data fetching done, transaction will close after this)
			documentData = docgen.StateOfApplicabilityData{
				Title:                       stateOfApplicability.Name,
				OrganizationName:            organization.Name,
				Description:                 description,
				CreatedAt:                   stateOfApplicability.CreatedAt,
				TotalControls:               len(linkedControls),
				FrameworkGroups:             frameworkGroups,
				CompanyHorizontalLogoBase64: horizontalLogoBase64,
				Version:                     version,
				PublishedAt:                 publishedAt,
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Transaction closed - now perform slow I/O operations outside transaction
	// Render HTML
	htmlData, err := docgen.RenderStateOfApplicabilityHTML(documentData)
	if err != nil {
		return nil, fmt.Errorf("cannot render HTML: %w", err)
	}

	// Convert to PDF
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
