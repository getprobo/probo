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

type TransferImpactAssessmentService struct {
	svc               *TenantService
	html2pdfConverter *html2pdf.Converter
}

type (
	CreateTransferImpactAssessmentRequest struct {
		ProcessingActivityID  gid.GID
		DataSubjects          *string
		LegalMechanism        *string
		Transfer              *string
		LocalLawRisk          *string
		SupplementaryMeasures *string
	}

	UpdateTransferImpactAssessmentRequest struct {
		ID                    gid.GID
		DataSubjects          **string
		LegalMechanism        **string
		Transfer              **string
		LocalLawRisk          **string
		SupplementaryMeasures **string
	}
)

func (req *CreateTransferImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ProcessingActivityID, "processing_activity_id", validator.Required(), validator.GID(coredata.ProcessingActivityEntityType))
	v.Check(req.DataSubjects, "data_subjects", validator.SafeText(ContentMaxLength))
	v.Check(req.LegalMechanism, "legal_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(req.Transfer, "transfer", validator.SafeText(ContentMaxLength))
	v.Check(req.LocalLawRisk, "local_law_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.SupplementaryMeasures, "supplementary_measures", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (req *UpdateTransferImpactAssessmentRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.TransferImpactAssessmentEntityType))
	v.Check(req.DataSubjects, "data_subjects", validator.SafeText(ContentMaxLength))
	v.Check(req.LegalMechanism, "legal_mechanism", validator.SafeText(ContentMaxLength))
	v.Check(req.Transfer, "transfer", validator.SafeText(ContentMaxLength))
	v.Check(req.LocalLawRisk, "local_law_risk", validator.SafeText(ContentMaxLength))
	v.Check(req.SupplementaryMeasures, "supplementary_measures", validator.SafeText(ContentMaxLength))

	return v.Error()
}

func (s TransferImpactAssessmentService) Get(
	ctx context.Context,
	tiaID gid.GID,
) (*coredata.TransferImpactAssessment, error) {
	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := tia.LoadByID(ctx, conn, s.svc.scope, tiaID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s TransferImpactAssessmentService) GetByProcessingActivityID(
	ctx context.Context,
	processingActivityID gid.GID,
) (*coredata.TransferImpactAssessment, error) {
	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := tia.LoadByProcessingActivityID(ctx, conn, s.svc.scope, processingActivityID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s TransferImpactAssessmentService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.TransferImpactAssessmentOrderField],
	filter *coredata.TransferImpactAssessmentFilter,
) (*page.Page[*coredata.TransferImpactAssessment, coredata.TransferImpactAssessmentOrderField], error) {
	var tias coredata.TransferImpactAssessments

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := tias.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load transfer impact assessments: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(tias, cursor), nil
}

func (s TransferImpactAssessmentService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.TransferImpactAssessmentFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			tias := coredata.TransferImpactAssessments{}
			count, err = tias.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			return err
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *TransferImpactAssessmentService) Create(
	ctx context.Context,
	req *CreateTransferImpactAssessmentRequest,
) (*coredata.TransferImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	now := time.Now()

	tia := &coredata.TransferImpactAssessment{
		ID:                    gid.New(s.svc.scope.GetTenantID(), coredata.TransferImpactAssessmentEntityType),
		ProcessingActivityID:  req.ProcessingActivityID,
		DataSubjects:          req.DataSubjects,
		LegalMechanism:        req.LegalMechanism,
		Transfer:              req.Transfer,
		LocalLawRisk:          req.LocalLawRisk,
		SupplementaryMeasures: req.SupplementaryMeasures,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			processingActivity := &coredata.ProcessingActivity{}
			if err := processingActivity.LoadByID(ctx, conn, s.svc.scope, req.ProcessingActivityID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			tia.OrganizationID = processingActivity.OrganizationID

			if err := tia.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert transfer impact assessment: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s *TransferImpactAssessmentService) Update(
	ctx context.Context,
	req *UpdateTransferImpactAssessmentRequest,
) (*coredata.TransferImpactAssessment, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	tia := &coredata.TransferImpactAssessment{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := tia.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			if req.DataSubjects != nil {
				tia.DataSubjects = *req.DataSubjects
			}

			if req.LegalMechanism != nil {
				tia.LegalMechanism = *req.LegalMechanism
			}

			if req.Transfer != nil {
				tia.Transfer = *req.Transfer
			}

			if req.LocalLawRisk != nil {
				tia.LocalLawRisk = *req.LocalLawRisk
			}

			if req.SupplementaryMeasures != nil {
				tia.SupplementaryMeasures = *req.SupplementaryMeasures
			}

			tia.UpdatedAt = time.Now()

			if err := tia.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update transfer impact assessment: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tia, nil
}

func (s *TransferImpactAssessmentService) Delete(
	ctx context.Context,
	tiaID gid.GID,
) error {
	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			tia := &coredata.TransferImpactAssessment{}
			if err := tia.LoadByID(ctx, conn, s.svc.scope, tiaID); err != nil {
				return fmt.Errorf("cannot load transfer impact assessment: %w", err)
			}

			if err := tia.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete transfer impact assessment: %w", err)
			}

			return nil
		},
	)

	return err
}

func (s *TransferImpactAssessmentService) ExportPDF(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.TransferImpactAssessmentFilter,
) ([]byte, error) {
	var tableData docgen.TransferImpactAssessmentTableData

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			var assessments coredata.TransferImpactAssessments
			if err := assessments.LoadAllByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter); err != nil {
				return fmt.Errorf("cannot load transfer impact assessments: %w", err)
			}

			if len(assessments) == 0 {
				return fmt.Errorf("no transfer impact assessments found: %w", coredata.ErrResourceNotFound)
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

			assessmentRows := make([]docgen.TransferImpactAssessmentRowData, len(assessments))
			for i, assessment := range assessments {
				processingActivity := &coredata.ProcessingActivity{}
				if err := processingActivity.LoadByID(ctx, conn, s.svc.scope, assessment.ProcessingActivityID); err != nil {
					return fmt.Errorf("cannot load processing activity: %w", err)
				}

				assessmentRows[i] = docgen.TransferImpactAssessmentRowData{
					ProcessingActivityName: processingActivity.Name,
					DataSubjects:           assessment.DataSubjects,
					LegalMechanism:         assessment.LegalMechanism,
					Transfer:               assessment.Transfer,
					LocalLawRisk:           assessment.LocalLawRisk,
					SupplementaryMeasures:  assessment.SupplementaryMeasures,
				}
			}

			tableData = docgen.TransferImpactAssessmentTableData{
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

	htmlContent, err := docgen.RenderTransferImpactAssessmentsTableHTML(tableData)
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
