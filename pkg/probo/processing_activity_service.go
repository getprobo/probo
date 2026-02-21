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
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/docgen"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/html2pdf"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type ProcessingActivityService struct {
	svc               *TenantService
	html2pdfConverter *html2pdf.Converter
}

type (
	CreateProcessingActivityRequest struct {
		OrganizationID                       gid.GID
		Name                                 string
		Purpose                              *string
		DataSubjectCategory                  *string
		PersonalDataCategory                 *string
		SpecialOrCriminalData                coredata.ProcessingActivitySpecialOrCriminalDatum
		ConsentEvidenceLink                  *string
		LawfulBasis                          coredata.ProcessingActivityLawfulBasis
		Recipients                           *string
		Location                             *string
		InternationalTransfers               bool
		TransferSafeguard                    *coredata.ProcessingActivityTransferSafeguard
		RetentionPeriod                      *string
		SecurityMeasures                     *string
		DataProtectionImpactAssessmentNeeded coredata.ProcessingActivityDataProtectionImpactAssessment
		TransferImpactAssessmentNeeded       coredata.ProcessingActivityTransferImpactAssessment
		LastReviewDate                       *time.Time
		NextReviewDate                       *time.Time
		Role                                 coredata.ProcessingActivityRole
		DataProtectionOfficerID              *gid.GID
		VendorIDs                            []gid.GID
	}

	UpdateProcessingActivityRequest struct {
		ID                                   gid.GID
		Name                                 *string
		Purpose                              **string
		DataSubjectCategory                  **string
		PersonalDataCategory                 **string
		SpecialOrCriminalData                *coredata.ProcessingActivitySpecialOrCriminalDatum
		ConsentEvidenceLink                  **string
		LawfulBasis                          *coredata.ProcessingActivityLawfulBasis
		Recipients                           **string
		Location                             **string
		InternationalTransfers               *bool
		TransferSafeguard                    **coredata.ProcessingActivityTransferSafeguard
		RetentionPeriod                      **string
		SecurityMeasures                     **string
		DataProtectionImpactAssessmentNeeded *coredata.ProcessingActivityDataProtectionImpactAssessment
		TransferImpactAssessmentNeeded       *coredata.ProcessingActivityTransferImpactAssessment
		LastReviewDate                       **time.Time
		NextReviewDate                       **time.Time
		Role                                 *coredata.ProcessingActivityRole
		DataProtectionOfficerID              **gid.GID
		VendorIDs                            *[]gid.GID
	}
)

func (cpar *CreateProcessingActivityRequest) Validate() error {
	v := validator.New()

	v.Check(cpar.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(cpar.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(cpar.Purpose, "purpose", validator.SafeText(TitleMaxLength))
	v.Check(cpar.DataSubjectCategory, "data_subject_category", validator.SafeText(TitleMaxLength))
	v.Check(cpar.PersonalDataCategory, "personal_data_category", validator.SafeText(TitleMaxLength))
	v.Check(cpar.SpecialOrCriminalData, "special_or_criminal_data", validator.Required(), validator.OneOfSlice(coredata.ProcessingActivitySpecialOrCriminalData()))
	v.Check(cpar.ConsentEvidenceLink, "consent_evidence_link", validator.SafeText(2048))
	v.Check(cpar.LawfulBasis, "lawful_basis", validator.Required(), validator.OneOfSlice(coredata.ProcessingActivityLawfulBases()))
	v.Check(cpar.Recipients, "recipients", validator.SafeText(TitleMaxLength))
	v.Check(cpar.Location, "location", validator.SafeText(TitleMaxLength))
	v.Check(cpar.InternationalTransfers, "international_transfers", validator.Required())
	v.Check(cpar.TransferSafeguard, "transfer_safeguard", validator.OneOfSlice(coredata.ProcessingActivityTransferSafeguards()))
	v.Check(cpar.RetentionPeriod, "retention_period", validator.SafeText(TitleMaxLength))
	v.Check(cpar.SecurityMeasures, "security_measures", validator.SafeText(TitleMaxLength))
	v.Check(cpar.DataProtectionImpactAssessmentNeeded, "data_protection_impact_assessment_needed", validator.Required(), validator.OneOfSlice(coredata.ProcessingActivityDataProtectionImpactAssessments()))
	v.Check(cpar.TransferImpactAssessmentNeeded, "transfer_impact_assessment_needed", validator.Required(), validator.OneOfSlice(coredata.ProcessingActivityTransferImpactAssessments()))
	v.Check(cpar.Role, "role", validator.Required(), validator.OneOfSlice(coredata.ProcessingActivityRoles()))
	v.Check(cpar.DataProtectionOfficerID, "data_protection_officer_id", validator.GID(coredata.MembershipProfileEntityType))
	v.CheckEach(cpar.VendorIDs, "vendor_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("vendor_ids[%d]", index), validator.Required(), validator.GID(coredata.VendorEntityType))
	})

	return v.Error()
}

func (upar *UpdateProcessingActivityRequest) Validate() error {
	v := validator.New()

	v.Check(upar.ID, "id", validator.Required(), validator.GID(coredata.ProcessingActivityEntityType))
	v.Check(upar.Name, "name", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(upar.Purpose, "purpose", validator.SafeText(TitleMaxLength))
	v.Check(upar.DataSubjectCategory, "data_subject_category", validator.SafeText(TitleMaxLength))
	v.Check(upar.PersonalDataCategory, "personal_data_category", validator.SafeText(TitleMaxLength))
	v.Check(upar.SpecialOrCriminalData, "special_or_criminal_data", validator.OneOfSlice(coredata.ProcessingActivitySpecialOrCriminalData()))
	v.Check(upar.ConsentEvidenceLink, "consent_evidence_link", validator.SafeText(2048))
	v.Check(upar.LawfulBasis, "lawful_basis", validator.OneOfSlice(coredata.ProcessingActivityLawfulBases()))
	v.Check(upar.Recipients, "recipients", validator.SafeText(TitleMaxLength))
	v.Check(upar.Location, "location", validator.SafeText(TitleMaxLength))
	v.Check(upar.TransferSafeguard, "transfer_safeguards", validator.OneOfSlice(coredata.ProcessingActivityTransferSafeguards()))
	v.Check(upar.RetentionPeriod, "retention_period", validator.SafeText(TitleMaxLength))
	v.Check(upar.SecurityMeasures, "security_measures", validator.SafeText(TitleMaxLength))
	v.Check(upar.DataProtectionImpactAssessmentNeeded, "data_protection_impact_assessment_needed", validator.OneOfSlice(coredata.ProcessingActivityDataProtectionImpactAssessments()))
	v.Check(upar.TransferImpactAssessmentNeeded, "transfer_impact_assessment_needed", validator.OneOfSlice(coredata.ProcessingActivityTransferImpactAssessments()))
	v.Check(upar.Role, "role", validator.OneOfSlice(coredata.ProcessingActivityRoles()))
	v.Check(upar.DataProtectionOfficerID, "data_protection_officer_id", validator.GID(coredata.MembershipProfileEntityType))
	v.CheckEach(upar.VendorIDs, "vendor_ids", func(index int, item any) {
		v.Check(item, fmt.Sprintf("vendor_ids[%d]", index), validator.GID(coredata.VendorEntityType))
	})

	return v.Error()
}

func (s ProcessingActivityService) Get(
	ctx context.Context,
	processingActivityID gid.GID,
) (*coredata.ProcessingActivity, error) {
	processingActivity := &coredata.ProcessingActivity{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return processingActivity.LoadByID(ctx, conn, s.svc.scope, processingActivityID)
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivity, nil
}

func (s *ProcessingActivityService) Create(
	ctx context.Context,
	req *CreateProcessingActivityRequest,
) (*coredata.ProcessingActivity, error) {
	now := time.Now()
	processingActivityVendors := &coredata.ProcessingActivityVendors{}

	processingActivity := &coredata.ProcessingActivity{
		ID:                                   gid.New(s.svc.scope.GetTenantID(), coredata.ProcessingActivityEntityType),
		OrganizationID:                       req.OrganizationID,
		Name:                                 req.Name,
		Purpose:                              req.Purpose,
		DataSubjectCategory:                  req.DataSubjectCategory,
		PersonalDataCategory:                 req.PersonalDataCategory,
		SpecialOrCriminalData:                req.SpecialOrCriminalData,
		ConsentEvidenceLink:                  req.ConsentEvidenceLink,
		LawfulBasis:                          req.LawfulBasis,
		Recipients:                           req.Recipients,
		Location:                             req.Location,
		InternationalTransfers:               req.InternationalTransfers,
		TransferSafeguard:                    req.TransferSafeguard,
		RetentionPeriod:                      req.RetentionPeriod,
		SecurityMeasures:                     req.SecurityMeasures,
		DataProtectionImpactAssessmentNeeded: req.DataProtectionImpactAssessmentNeeded,
		TransferImpactAssessmentNeeded:       req.TransferImpactAssessmentNeeded,
		LastReviewDate:                       req.LastReviewDate,
		NextReviewDate:                       req.NextReviewDate,
		Role:                                 req.Role,
		DataProtectionOfficerID:              req.DataProtectionOfficerID,
		CreatedAt:                            now,
		UpdatedAt:                            now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := processingActivity.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert processing activity: %w", err)
			}

			if len(req.VendorIDs) > 0 {
				if err := processingActivityVendors.Insert(ctx, conn, s.svc.scope, processingActivity.ID, req.OrganizationID, req.VendorIDs); err != nil {
					return fmt.Errorf("cannot create processing activity vendors: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivity, nil
}

func (s *ProcessingActivityService) Update(
	ctx context.Context,
	req *UpdateProcessingActivityRequest,
) (*coredata.ProcessingActivity, error) {
	processingActivity := &coredata.ProcessingActivity{}
	processingActivityVendors := &coredata.ProcessingActivityVendors{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := processingActivity.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
				return fmt.Errorf("cannot load processing activity: %w", err)
			}

			if req.Name != nil {
				processingActivity.Name = *req.Name
			}
			if req.Purpose != nil {
				processingActivity.Purpose = *req.Purpose
			}
			if req.DataSubjectCategory != nil {
				processingActivity.DataSubjectCategory = *req.DataSubjectCategory
			}
			if req.PersonalDataCategory != nil {
				processingActivity.PersonalDataCategory = *req.PersonalDataCategory
			}
			if req.SpecialOrCriminalData != nil {
				processingActivity.SpecialOrCriminalData = *req.SpecialOrCriminalData
			}
			if req.ConsentEvidenceLink != nil {
				processingActivity.ConsentEvidenceLink = *req.ConsentEvidenceLink
			}
			if req.LawfulBasis != nil {
				processingActivity.LawfulBasis = *req.LawfulBasis
			}
			if req.Recipients != nil {
				processingActivity.Recipients = *req.Recipients
			}
			if req.Location != nil {
				processingActivity.Location = *req.Location
			}
			if req.InternationalTransfers != nil {
				processingActivity.InternationalTransfers = *req.InternationalTransfers
			}
			if req.TransferSafeguard != nil {
				processingActivity.TransferSafeguard = *req.TransferSafeguard
			}
			if req.RetentionPeriod != nil {
				processingActivity.RetentionPeriod = *req.RetentionPeriod
			}
			if req.SecurityMeasures != nil {
				processingActivity.SecurityMeasures = *req.SecurityMeasures
			}
			if req.DataProtectionImpactAssessmentNeeded != nil {
				processingActivity.DataProtectionImpactAssessmentNeeded = *req.DataProtectionImpactAssessmentNeeded
			}
			if req.TransferImpactAssessmentNeeded != nil {
				processingActivity.TransferImpactAssessmentNeeded = *req.TransferImpactAssessmentNeeded
			}
			if req.LastReviewDate != nil {
				processingActivity.LastReviewDate = *req.LastReviewDate
			}
			if req.NextReviewDate != nil {
				processingActivity.NextReviewDate = *req.NextReviewDate
			}
			if req.Role != nil {
				processingActivity.Role = *req.Role
			}
			if req.DataProtectionOfficerID != nil {
				processingActivity.DataProtectionOfficerID = *req.DataProtectionOfficerID
			}

			processingActivity.UpdatedAt = time.Now()

			if err := processingActivity.Update(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update processing activity: %w", err)
			}

			if req.VendorIDs != nil {
				if err := processingActivityVendors.Merge(ctx, conn, s.svc.scope, processingActivity.ID, processingActivity.OrganizationID, *req.VendorIDs); err != nil {
					return fmt.Errorf("cannot update processing activity vendors: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return processingActivity, nil
}

func (s ProcessingActivityService) Delete(
	ctx context.Context,
	processingActivityID gid.GID,
) error {
	processingActivity := coredata.ProcessingActivity{ID: processingActivityID}
	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := processingActivity.Delete(ctx, conn, s.svc.scope)
			if err != nil {
				return fmt.Errorf("cannot delete processing activity: %w", err)
			}
			return nil
		},
	)
}

func (s ProcessingActivityService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.ProcessingActivityOrderField],
	filter *coredata.ProcessingActivityFilter,
) (*page.Page[*coredata.ProcessingActivity, coredata.ProcessingActivityOrderField], error) {
	var processingActivities coredata.ProcessingActivities

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := processingActivities.LoadByOrganizationID(ctx, conn, s.svc.scope, organizationID, cursor, filter)
			if err != nil {
				return fmt.Errorf("cannot load processing activities: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(processingActivities, cursor), nil
}

func (s ProcessingActivityService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.ProcessingActivityFilter,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			processingActivities := coredata.ProcessingActivities{}
			count, err = processingActivities.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot count processing activities: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *ProcessingActivityService) ExportPDF(
	ctx context.Context,
	organizationID gid.GID,
	filter *coredata.ProcessingActivityFilter,
) ([]byte, error) {
	var tableData docgen.ProcessingActivityTableData

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			var processingActivities coredata.ProcessingActivities
			if err := processingActivities.LoadAllByOrganizationID(ctx, conn, s.svc.scope, organizationID, filter); err != nil {
				return fmt.Errorf("cannot load processing activities: %w", err)
			}

			if len(processingActivities) == 0 {
				return fmt.Errorf("no processing activities found: %w", coredata.ErrResourceNotFound)
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

			var vendors coredata.Vendors
			vendorMap, err := vendors.LoadAllByProcessingActivities(ctx, conn, s.svc.scope, organizationID, filter)
			if err != nil {
				return fmt.Errorf("cannot load vendors: %w", err)
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

			activities := make([]docgen.ProcessingActivityRowData, len(processingActivities))
			for i, pa := range processingActivities {
				dpoFullName := (*string)(nil)
				if pa.DataProtectionOfficerID != nil {
					dpo := &coredata.MembershipProfile{}
					if err := dpo.LoadByID(ctx, conn, s.svc.scope, *pa.DataProtectionOfficerID); err == nil {
						dpoFullName = &dpo.FullName
					}
				}

				vendorsList := ""
				if vendorNames, ok := vendorMap[pa.ID]; ok && len(vendorNames) > 0 {
					vendorsList = strings.Join(vendorNames, ", ")
				}

				activities[i] = docgen.ProcessingActivityRowData{
					Name:                                 pa.Name,
					Purpose:                              pa.Purpose,
					DataSubjectCategory:                  pa.DataSubjectCategory,
					PersonalDataCategory:                 pa.PersonalDataCategory,
					SpecialOrCriminalData:                pa.SpecialOrCriminalData,
					ConsentEvidenceLink:                  pa.ConsentEvidenceLink,
					LawfulBasis:                          pa.LawfulBasis,
					Recipients:                           pa.Recipients,
					Location:                             pa.Location,
					InternationalTransfers:               pa.InternationalTransfers,
					TransferSafeguards:                   pa.TransferSafeguard,
					RetentionPeriod:                      pa.RetentionPeriod,
					SecurityMeasures:                     pa.SecurityMeasures,
					DataProtectionImpactAssessmentNeeded: pa.DataProtectionImpactAssessmentNeeded,
					TransferImpactAssessmentNeeded:       pa.TransferImpactAssessmentNeeded,
					LastReviewDate:                       pa.LastReviewDate,
					NextReviewDate:                       pa.NextReviewDate,
					Role:                                 pa.Role,
					DataProtectionOfficerFullName:        dpoFullName,
					Vendors:                              vendorsList,
				}
			}

			tableData = docgen.ProcessingActivityTableData{
				CompanyName:                 organization.Name,
				CompanyHorizontalLogoBase64: horizontalLogoBase64,
				Version:                     version,
				PublishedAt:                 publishedAt,
				Activities:                  activities,
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	htmlContent, err := docgen.RenderProcessingActivitiesTableHTML(tableData)
	if err != nil {
		return nil, fmt.Errorf("cannot generate HTML: %w", err)
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
