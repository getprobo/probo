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

type DataProtectionImpactAssessmentService struct {
	svc               *TenantService
	html2pdfConverter *html2pdf.Converter
}

type (
	CreateDataProtectionImpactAssessmentRequest struct {
		ProcessingActivityID        gid.GID
		Description                 *string
		NecessityAndProportionality *string
		PotentialRisk               *string
		Mitigations                 *string
		ResidualRisk                *coredata.DataProtectionImpactAssessmentResidualRisk
	}

	UpdateDataProtectionImpactAssessmentRequest struct {
		ID                          gid.GID
		Description                 **string
		NecessityAndProportionality **string
		PotentialRisk               **string
		Mitigations                 **string
		ResidualRisk                *coredata.DataProtectionImpactAssessmentResidualRisk
	}
)

func (req *CreateDataProtectionImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ProcessingActivityID, "processing_activity_id", validator.Required(), validator.GID(coredata.ProcessingActivityEntityType))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.NecessityAndProportionality, "necessity_and_proportionality", validator.SafeText(ContentMaxLength))
	v.Check(req.PotentialRisk, "potential_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.Mitigations, "mitigations", validator.SafeText(ContentMaxLength))
	v.Check(req.ResidualRisk, "residual_risk", validator.OneOfSlice(coredata.DataProtectionImpactAssessmentResidualRisks()))

	return v.Error()
}

func (req *UpdateDataProtectionImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.DataProtectionImpactAssessmentEntityType))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.NecessityAndProportionality, "necessity_and_proportionality", validator.SafeText(ContentMaxLength))
	v.Check(req.PotentialRisk, "potential_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.Mitigations, "mitigations", validator.SafeText(ContentMaxLength))
	v.Check(req.ResidualRisk, "residual_risk", validator.OneOfSlice(coredata.DataProtectionImpactAssessmentResidualRisks()))

	return v.Error()
}

func (s DataProtectionImpactAssessmentService) Get(
	ctx context.Context,
	dpiaID gid.GID,
) (*coredata.DataProtectionImpactAssessment, error) {
	dpia := &coredata.DataProtectionImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := dpia.LoadByID(ctx, conn, s.svc.scope, dpiaID); err != nil {
				return fmt.Errorf("cannot load data protection impact assessment: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s DataProtectionImpactAssessmentService) GetByProcessingActivityID(
	ctx context.Context,
	processingActivityID gid.GID,
) (*coredata.DataProtectionImpactAssessment, error) {
	dpia := &coredata.DataProtectionImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := dpia.LoadByProcessingActivityID(ctx, conn, s.svc.scope, processingActivityID); err != nil {
				return fmt.Errorf("cannot load data protection impact assessment: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s DataProtectionImpactAssessmentService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DataProtectionImpactAssessmentOrderField],
	filter *coredata.DataProtectionImpactAssessmentFilter,
) (*page.Page[*coredata.DataProtectionImpactAssessment, coredata.DataProtectionImpactAssessmentOrderField], error) {
	var dpias coredata.DataProtectionImpactAssessments

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := dpias.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load data protection impact assessments: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(dpias, cursor), nil
}

func (s DataProtectionImpactAssessmentService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.DataProtectionImpactAssessmentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			dpias := coredata.DataProtectionImpactAssessments{}
			count, err = dpias.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *DataProtectionImpactAssessmentService) Create(
	ctx context.Context,
	req *CreateDataProtectionImpactAssessmentRequest,
) (*coredata.DataProtectionImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	dpia := &coredata.DataProtectionImpactAssessment{
		ID:                          gid.New(s.svc.scope.GetTenantID(), coredata.DataProtectionImpactAssessmentEntityType),
		ProcessingActivityID:        req.ProcessingActivityID,
		Description:                 req.Description,
		NecessityAndProportionality: req.NecessityAndProportionality,
		PotentialRisk:               req.PotentialRisk,
		Mitigations:                 req.Mitigations,
		ResidualRisk:                req.ResidualRisk,
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			processingActivity := &coredata.ProcessingActivity{}
			if err := processingActivity.LoadByID(ctx, conn, s.svc.scope, req.ProcessingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			dpia.OrganizationID = processingActivity.OrganizationID

			if err := dpia.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert data protection impact assessment: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s *DataProtectionImpactAssessmentService) Update(
	ctx context.Context,
	req *UpdateDataProtectionImpactAssessmentRequest,
) (*coredata.DataProtectionImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	dpia := &coredata.DataProtectionImpactAssessment{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := dpia.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load data protection impact assessment: %w", err)
			}

			if req.Description != nil {
				dpia.Description = *req.Description
			}

			if req.NecessityAndProportionality != nil {
				dpia.NecessityAndProportionality = *req.NecessityAndProportionality
			}

			if req.PotentialRisk != nil {
				dpia.PotentialRisk = *req.PotentialRisk
			}

			if req.Mitigations != nil {
				dpia.Mitigations = *req.Mitigations
			}

			if req.ResidualRisk != nil {
				dpia.ResidualRisk = req.ResidualRisk
			}

			dpia.UpdatedAt = time.Now()

			if err := dpia.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update data protection impact assessment: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return dpia, nil
}

func (s *DataProtectionImpactAssessmentService) Delete(
	ctx context.Context,
	dpiaID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			dpia := &coredata.DataProtectionImpactAssessment{}
			if err := dpia.LoadByID(ctx, conn, s.svc.scope, dpiaID); err != nil {
				return fmt.Errorf("cannot load data protection impact assessment: %w", err)
			}

			if err := dpia.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete data protection impact assessment: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s *DataProtectionImpactAssessmentService) ExportPDF(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.DataProtectionImpactAssessmentFilter,
) ([]byte, error) {
	var tableData docgen.DataProtectionImpactAssessmentTableData

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			var assessments coredata.DataProtectionImpactAssessments
			if err := assessments.LoadAllByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter); err != nil {
				return fmt.Errorf("cannot load data protection impact assessments: %w", err)
			}

			if len(assessments) == 0 {
				return fmt.Errorf("no data protection impact assessments found: %w", coredata.ErrResourceNotFound)
			}

			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
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

			var snapshots coredata.Snapshots
			snapshotType := coredata.SnapshotsTypeProcessingActivities

			var version int
			var publishedAt time.Time

			if snapshotID := filter.SnapshotID(); snapshotID != nil {
				snapshot := &coredata.Snapshot{}
				if err := snapshot.LoadByID(ctx, conn, s.svc.scope, *snapshotID); err != nil {
					return fmt.Errorf("cannot load snapshot: %w", err)
				}
				publishedAt = snapshot.CreatedAt
				snapshotFilter := coredata.NewSnapshotFilter(&snapshotType).WithBeforeDate(&snapshot.CreatedAt)
				snapshotCount, err := snapshots.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, snapshotFilter)
				if err != nil {
					return fmt.Errorf("cannot count processing activities snapshots: %w", err)
				}
				version = snapshotCount
			} else {
				publishedAt = time.Now()
				snapshotFilter := coredata.NewSnapshotFilter(&snapshotType)
				snapshotCount, err := snapshots.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, snapshotFilter)
				if err != nil {
					return fmt.Errorf("cannot count processing activities snapshots: %w", err)
				}
				version = snapshotCount + 1
			}

			assessmentRows := make([]docgen.DataProtectionImpactAssessmentRowData, len(assessments))
			for i, assessment := range assessments {
				processingActivity := &coredata.ProcessingActivity{}
				if err := processingActivity.LoadByID(ctx, conn, s.svc.scope, assessment.ProcessingActivityID); err != nil {
					return fmt.Errorf("cannot load processing activity: %w", err)
				}

				assessmentRows[i] = docgen.DataProtectionImpactAssessmentRowData{
					ProcessingActivityName:      processingActivity.Name,
					Description:                 assessment.Description,
					NecessityAndProportionality: assessment.NecessityAndProportionality,
					PotentialRisk:               assessment.PotentialRisk,
					Mitigations:                 assessment.Mitigations,
					ResidualRisk:                assessment.ResidualRisk,
				}
			}

			tableData = docgen.DataProtectionImpactAssessmentTableData{
				CompanyName:                 organization.Name,
				CompanyHorizontalLogoBase64: horizontalLogoBase64,
				Version:                     version,
				PublishedAt:                 publishedAt,
				Assessments:                 assessmentRows,
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	htmlContent, err := docgen.RenderDataProtectionImpactAssessmentsTableHTML(tableData)
	if err != nil {
		return nil, fmt.Errorf("cannot render HTML: %w", err)
	}

	cfg := html2pdf.RenderConfig{
		PageFormat:      html2pdf.PageFormatA4,
		Orientation:     html2pdf.OrientationPortrait,
		MarginTop:       html2pdf.NewMarginInches(0.98),
		MarginBottom:    html2pdf.NewMarginInches(0.98),
		MarginLeft:      html2pdf.NewMarginInches(0.98),
		MarginRight:     html2pdf.NewMarginInches(0.98),
		PrintBackground: true,
		Scale:           1.0,
	}

	pdfReader, err := s.html2pdfConverter.GeneratePDF(ctx, htmlContent, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot generate PDF: %w", err)
	}

	pdfData, err := io.ReadAll(pdfReader)
	if err != nil {
		return nil, fmt.Errorf("cannot read PDF data: %w", err)
	}

	return pdfData, nil
}
