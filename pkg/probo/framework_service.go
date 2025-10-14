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
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/html2pdf"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/slug"
	"github.com/getprobo/probo/pkg/soagen"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
)

const (
	maxStateOfApplicabilityLimit  = 10_000
	frameworkExportEmailExpiresIn = 24 * time.Hour
	frameworkExportEmailSubject   = "Your framework export is ready"
	frameworkExportEmailBody      = `
Your framework export has been completed successfully.

You can download the export using the link below:
[1] %s

This link will expire in 24 hours.`
)

type (
	FrameworkService struct {
		svc               *TenantService
		html2pdfConverter *html2pdf.Converter
	}

	CreateFrameworkRequest struct {
		OrganizationID gid.GID
		Name           string
		Description    string
	}

	UpdateFrameworkRequest struct {
		ID          gid.GID
		Name        *string
		Description *string
	}

	ImportFrameworkRequest struct {
		Framework struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Controls []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"controls"`
		}
	}
)

func (s FrameworkService) RequestExport(
	ctx context.Context,
	frameworkID gid.GID,
	recipientEmail string,
	recipientName string,
) (error, *coredata.ExportJob) {
	var exportJobID gid.GID
	exportJob := &coredata.ExportJob{}

	err := s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		framework := &coredata.Framework{}
		if err := framework.LoadByID(ctx, conn, s.svc.scope, frameworkID); err != nil {
			return fmt.Errorf("cannot load framework: %w", err)
		}

		now := time.Now()
		exportJobID = gid.New(s.svc.scope.GetTenantID(), coredata.ExportJobEntityType)

		args := coredata.FrameworkExportArguments{
			FrameworkID: frameworkID,
		}
		argsJSON, err := json.Marshal(args)
		if err != nil {
			return fmt.Errorf("cannot marshal framework export arguments: %w", err)
		}

		exportJob = &coredata.ExportJob{
			ID:             exportJobID,
			Type:           coredata.ExportJobTypeFramework,
			Arguments:      argsJSON,
			Status:         coredata.ExportJobStatusPending,
			RecipientEmail: recipientEmail,
			RecipientName:  recipientName,
			CreatedAt:      now,
		}

		if err := exportJob.Insert(ctx, conn, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert export job: %w", err)
		}

		return nil
	})

	if err != nil {
		return err, nil
	}

	return nil, exportJob
}

func (s FrameworkService) Export(
	ctx context.Context,
	frameworkID gid.GID,
	file io.Writer,
) error {
	archive := zip.NewWriter(file)
	defer archive.Close()

	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, s.svc.scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			controls := coredata.Controls{}
			err := controls.LoadByFrameworkID(
				ctx,
				conn,
				s.svc.scope,
				frameworkID,
				page.NewCursor(
					10_000,
					nil,
					page.Head,
					page.OrderBy[coredata.ControlOrderField]{
						Field:     coredata.ControlOrderFieldSectionTitle,
						Direction: page.OrderDirectionAsc,
					},
				),
				coredata.NewControlFilter(nil),
			)
			if err != nil {
				return fmt.Errorf("cannot load controls: %w", err)
			}

			for _, control := range controls {
				_, err := archive.Create(fmt.Sprintf("%s/%s/", framework.Name, control.SectionTitle))
				if err != nil {
					return fmt.Errorf("cannot create control directory in archive: %w", err)
				}

				measures := coredata.Measures{}
				err = measures.LoadByControlID(
					ctx,
					conn,
					s.svc.scope,
					control.ID,
					page.NewCursor(
						10_000,
						nil,
						page.Head,
						page.OrderBy[coredata.MeasureOrderField]{
							Field:     coredata.MeasureOrderFieldCreatedAt,
							Direction: page.OrderDirectionAsc,
						},
					),
					coredata.NewMeasureFilter(nil, nil),
				)
				if err != nil {
					return fmt.Errorf("cannot load measures: %w", err)
				}

				for _, measure := range measures {
					_, err := archive.Create(fmt.Sprintf("%s/%s/%s/", framework.Name, control.SectionTitle, measure.Name))
					if err != nil {
						return fmt.Errorf("cannot create measure directory in archive: %w", err)
					}

					evidences := coredata.Evidences{}
					err = evidences.LoadByMeasureID(
						ctx,
						conn,
						s.svc.scope,
						measure.ID,
						page.NewCursor(
							10_000,
							nil,
							page.Head,
							page.OrderBy[coredata.EvidenceOrderField]{
								Field:     coredata.EvidenceOrderFieldCreatedAt,
								Direction: page.OrderDirectionAsc,
							},
						),
					)
					if err != nil {
						return fmt.Errorf("cannot load evidences: %w", err)
					}

					for _, evidence := range evidences {
						if evidence.Type != coredata.EvidenceTypeFile ||
							evidence.State != coredata.EvidenceStateFulfilled ||
							evidence.EvidenceFileId == nil {
							continue
						}

						evidence_file := &coredata.File{}
						if err := evidence_file.LoadByID(ctx, conn, s.svc.scope, *evidence.EvidenceFileId); err != nil {
							return fmt.Errorf("cannot load evidence file: %w", err)
						}

						object, err := s.svc.s3.GetObject(
							ctx,
							&s3.GetObjectInput{
								Bucket: aws.String(s.svc.bucket),
								Key:    aws.String(evidence_file.FileKey),
							},
						)
						if err != nil {
							return fmt.Errorf("cannot download evidence: %w", err)
						}
						defer object.Body.Close()

						w, err := archive.Create(fmt.Sprintf("%s/%s/%s/%s", framework.Name, control.SectionTitle, measure.Name, evidence_file.FileName))
						if err != nil {
							return fmt.Errorf("cannot create evidence in archive: %w", err)
						}

						_, err = io.Copy(w, object.Body)
						if err != nil {
							return fmt.Errorf("cannot write evidence to archive: %w", err)
						}
					}
				}

				documents := coredata.Documents{}
				err = documents.LoadByControlID(
					ctx,
					conn,
					s.svc.scope,
					control.ID,
					page.NewCursor(
						10_000,
						nil,
						page.Head,
						page.OrderBy[coredata.DocumentOrderField]{
							Field:     coredata.DocumentOrderFieldCreatedAt,
							Direction: page.OrderDirectionAsc,
						},
					),
					coredata.NewDocumentFilter(nil),
				)
				if err != nil {
					return fmt.Errorf("cannot load documents: %w", err)
				}

				for _, document := range documents {
					documentVersion := &coredata.DocumentVersion{}
					if err := documentVersion.LoadLatestPublishedVersion(ctx, conn, s.svc.scope, document.ID); err != nil {
						return fmt.Errorf("cannot load document version: %w", err)
					}

					exportedPDF, err := exportDocumentPDF(
						ctx,
						s.svc,
						s.html2pdfConverter,
						conn,
						s.svc.scope,
						documentVersion.ID,
						ExportPDFOptions{WithSignatures: true},
					)
					if err != nil {
						return fmt.Errorf("cannot export document PDF: %w", err)
					}

					w, err := archive.Create(fmt.Sprintf("%s/%s/%s.pdf", framework.Name, control.SectionTitle, document.Title))
					if err != nil {
						return fmt.Errorf("cannot create document in archive: %w", err)
					}

					_, err = w.Write(exportedPDF)
					if err != nil {
						return fmt.Errorf("cannot write document to archive: %w", err)
					}
				}
			}

			return nil
		},
	)
}

func (s FrameworkService) Create(
	ctx context.Context,
	req CreateFrameworkRequest,
) (*coredata.Framework, error) {
	now := time.Now()
	organization := &coredata.Organization{}

	framework := &coredata.Framework{
		ID:          gid.New(s.svc.scope.GetTenantID(), coredata.FrameworkEntityType),
		Name:        req.Name,
		Description: req.Description,
		ReferenceID: slug.Make(req.Name),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
			return fmt.Errorf("cannot load organization: %w", err)
		}

		framework.OrganizationID = organization.ID

		if err := framework.Insert(ctx, conn, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert framework: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return framework, nil
}

func (s FrameworkService) CountForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) (err error) {
		frameworks := &coredata.Frameworks{}
		count, err = frameworks.CountByOrganizationID(ctx, conn, s.svc.scope, organizationID)
		if err != nil {
			return fmt.Errorf("cannot count frameworks: %w", err)
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("cannot count frameworks: %w", err)
	}

	return count, nil
}

func (s FrameworkService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.FrameworkOrderField],
) (*page.Page[*coredata.Framework, coredata.FrameworkOrderField], error) {
	var frameworks coredata.Frameworks
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		if err := organization.LoadByID(ctx, conn, s.svc.scope, organizationID); err != nil {
			return fmt.Errorf("cannot load organization: %w", err)
		}

		err := frameworks.LoadByOrganizationID(
			ctx,
			conn,
			s.svc.scope,
			organization.ID,
			cursor,
		)
		if err != nil {
			return fmt.Errorf("cannot load frameworks: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return page.NewPage(frameworks, cursor), nil
}

func (s FrameworkService) Get(
	ctx context.Context,
	frameworkID gid.GID,
) (*coredata.Framework, error) {
	framework := &coredata.Framework{}

	err := s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return framework.LoadByID(ctx, conn, s.svc.scope, frameworkID)
	})

	if err != nil {
		return nil, err
	}

	return framework, nil
}

func (s FrameworkService) Update(
	ctx context.Context,
	req UpdateFrameworkRequest,
) (*coredata.Framework, error) {
	framework := &coredata.Framework{ID: req.ID}

	err := s.svc.pg.WithTx(ctx, func(conn pg.Conn) error {
		if err := framework.LoadByID(ctx, conn, s.svc.scope, req.ID); err != nil {
			return fmt.Errorf("cannot load framework: %w", err)
		}

		if req.Name != nil {
			framework.Name = *req.Name
		}

		if req.Description != nil {
			framework.Description = *req.Description
		}

		return framework.Update(ctx, conn, s.svc.scope)
	})
	if err != nil {
		return nil, err
	}

	return framework, nil
}

func (s FrameworkService) Delete(
	ctx context.Context,
	frameworkID gid.GID,
) error {
	framework := &coredata.Framework{}

	return s.svc.pg.WithConn(ctx, func(conn pg.Conn) error {
		return framework.Delete(ctx, conn, s.svc.scope, frameworkID)
	})
}

func (s FrameworkService) Import(
	ctx context.Context,
	organizationID gid.GID,
	req ImportFrameworkRequest,
) (*coredata.Framework, error) {
	var framework *coredata.Framework
	frameworkID := gid.New(organizationID.TenantID(), coredata.FrameworkEntityType)
	now := time.Now()

	err := s.svc.pg.WithTx(ctx, func(tx pg.Conn) error {
		organization := &coredata.Organization{}
		if err := organization.LoadByID(ctx, tx, s.svc.scope, organizationID); err != nil {
			return fmt.Errorf("cannot load organization: %w", err)
		}

		framework = &coredata.Framework{
			ID:             frameworkID,
			OrganizationID: organization.ID,
			ReferenceID:    req.Framework.ID,
			Name:           req.Framework.Name,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := framework.Insert(ctx, tx, s.svc.scope); err != nil {
			return fmt.Errorf("cannot insert framework: %w", err)
		}

		for _, control := range req.Framework.Controls {
			controlID := gid.New(organization.ID.TenantID(), coredata.ControlEntityType)

			now := time.Now()
			control := &coredata.Control{
				ID:           controlID,
				FrameworkID:  frameworkID,
				SectionTitle: control.ID,
				Name:         control.Name,
				Description:  control.Description,
				Status:       coredata.ControlStatusIncluded,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			if err := control.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert control: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return framework, nil
}

func (s FrameworkService) StateOfApplicability(ctx context.Context, frameworkID gid.GID) ([]byte, error) {
	rows := []soagen.SOARowData{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, s.svc.scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			controls := coredata.Controls{}
			err := controls.LoadByFrameworkID(
				ctx,
				conn,
				s.svc.scope,
				frameworkID,
				page.NewCursor(
					maxStateOfApplicabilityLimit,
					nil,
					page.Head,
					page.OrderBy[coredata.ControlOrderField]{
						Field:     coredata.ControlOrderFieldSectionTitle,
						Direction: page.OrderDirectionAsc,
					},
				),
				coredata.NewControlFilter(nil),
			)

			if err != nil {
				return fmt.Errorf("cannot load controls: %w", err)
			}

			for _, control := range controls {
				exclusionJustification := ""
				if control.Status == coredata.ControlStatusExcluded {
					if control.ExclusionJustification == nil {
						return fmt.Errorf("exclusion justification is required for excluded controls")
					}
					exclusionJustification = *control.ExclusionJustification
				}

				applicability := soagen.NewApplicability("Yes", true)
				if control.Status == coredata.ControlStatusExcluded {
					applicability = soagen.NewApplicability("No", false)
				}

				bestPractice := ref.Ref(true)
				if control.Status == coredata.ControlStatusExcluded {
					bestPractice = ref.Ref(false)
				}

				row := soagen.SOARowData{
					SectionTitle:           control.SectionTitle,
					ControlName:            control.Name,
					Applicability:          applicability,
					ExclusionJustification: exclusionJustification,
					Regulatory:             ref.Ref(false),
					Contractual:            ref.Ref(false),
					BestPractice:           bestPractice,
					RiskAssessment:         ref.Ref(false),
					SecurityMeasures:       []string{},
				}

				if control.Status == coredata.ControlStatusExcluded {
					rows = append(rows, row)
					continue
				}

				measures := coredata.Measures{}
				err = measures.LoadByControlID(
					ctx,
					conn,
					s.svc.scope,
					control.ID,
					page.NewCursor(
						maxStateOfApplicabilityLimit,
						nil,
						page.Head,
						page.OrderBy[coredata.MeasureOrderField]{
							Field:     coredata.MeasureOrderFieldCreatedAt,
							Direction: page.OrderDirectionAsc,
						},
					),
					coredata.NewMeasureFilter(nil, nil),
				)
				if err != nil {
					return fmt.Errorf("cannot load measures: %w", err)
				}

				for _, measure := range measures {
					risks := coredata.Risks{}
					var nilSnapshotID *gid.GID = nil
					risksCount, err := risks.CountByMeasureID(
						ctx,
						conn,
						s.svc.scope,
						measure.ID,
						coredata.NewRiskFilter(nil, &nilSnapshotID),
					)
					if err != nil {
						return fmt.Errorf("cannot count risks: %w", err)
					}

					if risksCount > 0 {
						row.RiskAssessment = ref.Ref(true)
					}

					row.SecurityMeasures = append(row.SecurityMeasures, measure.Name)
				}

				documents := coredata.Documents{}
				err = documents.LoadByControlID(
					ctx,
					conn,
					s.svc.scope,
					control.ID,
					page.NewCursor(
						0,
						nil,
						page.Head,
						page.OrderBy[coredata.DocumentOrderField]{
							Field:     coredata.DocumentOrderFieldCreatedAt,
							Direction: page.OrderDirectionAsc,
						},
					),
					coredata.NewDocumentFilter(nil),
				)
				if err != nil {
					return fmt.Errorf("cannot load documents: %w", err)
				}

				for _, document := range documents {
					risks := coredata.Risks{}
					var nilSnapshotID *gid.GID = nil
					risksCount, err := risks.CountByDocumentID(
						ctx,
						conn,
						s.svc.scope,
						document.ID,
						coredata.NewRiskFilter(nil, &nilSnapshotID),
					)
					if err != nil {
						return fmt.Errorf("cannot count risks: %w", err)
					}

					if risksCount > 0 {
						row.RiskAssessment = ref.Ref(true)
					}

					row.SecurityMeasures = append(row.SecurityMeasures, document.Title)
				}

				rows = append(rows, row)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	output, err := soagen.GenerateExcel(
		soagen.SOAData{
			Rows: rows,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot generate Excel file: %w", err)
	}

	return output, nil
}

func (s FrameworkService) SendExportEmail(
	ctx context.Context,
	fileID gid.GID,
	recipientName string,
	recipientEmail string,
) error {
	return s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			file := &coredata.File{}
			if err := file.LoadByID(ctx, tx, s.svc.scope, fileID); err != nil {
				return fmt.Errorf("cannot load file: %w", err)
			}

			downloadURL, err := s.GenerateFrameworkExportDownloadURL(ctx, file)
			if err != nil {
				return fmt.Errorf("cannot generate download URL: %w", err)
			}

			email := coredata.NewEmail(
				recipientName,
				recipientEmail,
				frameworkExportEmailSubject,
				fmt.Sprintf(frameworkExportEmailBody, downloadURL),
			)

			if err := email.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)
}

func (s FrameworkService) GenerateFrameworkExportDownloadURL(
	ctx context.Context,
	file *coredata.File,
) (string, error) {
	presignClient := s3.NewPresignClient(s.svc.s3)

	presignedReq, err := presignClient.PresignGetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket:                     ref.Ref(s.svc.bucket),
			Key:                        ref.Ref(file.FileKey),
			ResponseCacheControl:       ref.Ref("max-age=3600, public"),
			ResponseContentType:        ref.Ref(file.MimeType),
			ResponseContentDisposition: ref.Ref(fmt.Sprintf("attachment; filename=\"%s\"", file.FileName)),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = frameworkExportEmailExpiresIn
		},
	)

	if err != nil {
		return "", fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return presignedReq.URL, nil
}

func (s *FrameworkService) BuildAndUploadExport(ctx context.Context, exportJobID gid.GID) (*coredata.ExportJob, error) {
	exportJob := &coredata.ExportJob{}
	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := exportJob.LoadByID(ctx, tx, s.svc.scope, exportJobID); err != nil {
				return fmt.Errorf("cannot load export job: %w", err)
			}

			frameworkID, err := exportJob.GetFrameworkID()
			if err != nil {
				return fmt.Errorf("cannot get framework ID: %w", err)
			}

			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, tx, s.svc.scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			tempDir := os.TempDir()
			tempFile, err := os.CreateTemp(tempDir, "probo-framework-export-*.zip")
			if err != nil {
				return fmt.Errorf("cannot create temp file: %w", err)
			}
			defer tempFile.Close()
			defer os.Remove(tempFile.Name())

			err = s.Export(ctx, frameworkID, tempFile)
			if err != nil {
				return fmt.Errorf("cannot export framework: %w", err)
			}

			uuid, err := uuid.NewV4()
			if err != nil {
				return fmt.Errorf("cannot generate uuid: %w", err)
			}

			if _, err := tempFile.Seek(0, 0); err != nil {
				return fmt.Errorf("cannot seek temp file: %w", err)
			}

			fileInfo, err := tempFile.Stat()
			if err != nil {
				return fmt.Errorf("cannot stat temp file: %w", err)
			}

			_, err = s.svc.s3.PutObject(
				ctx,
				&s3.PutObjectInput{
					Bucket:        ref.Ref(s.svc.bucket),
					Key:           ref.Ref(uuid.String()),
					Body:          tempFile,
					ContentLength: ref.Ref(fileInfo.Size()),
					ContentType:   ref.Ref("application/zip"),
					Metadata: map[string]string{
						"type":            "framework-export",
						"export-job-id":   exportJob.ID.String(),
						"organization-id": framework.OrganizationID.String(),
					},
				},
			)
			if err != nil {
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			now := time.Now()

			file := coredata.File{
				ID:         gid.New(exportJob.ID.TenantID(), coredata.FileEntityType),
				BucketName: s.svc.bucket,
				MimeType:   "application/zip",
				FileName:   fmt.Sprintf("Framework Export %s.zip", now.Format("2006-01-02")),
				FileKey:    uuid.String(),
				FileSize:   fileInfo.Size(),
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := file.Insert(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert file: %w", err)
			}

			exportJob.FileID = &file.ID
			if err := exportJob.Update(ctx, tx, s.svc.scope); err != nil {
				return fmt.Errorf("cannot update export job: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return exportJob, nil
}
