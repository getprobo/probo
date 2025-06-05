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
	"os"
	"path/filepath"
	"time"

	"archive/tar"
	"compress/gzip"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/slug"
	"go.gearno.de/kit/pg"
)

type (
	FrameworkService struct {
		svc *TenantService
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

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := organization.LoadByID(ctx, conn, s.svc.scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			framework.OrganizationID = organization.ID

			if err := framework.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert framework: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return framework, nil
}

func (s FrameworkService) ListForOrganizationID(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.FrameworkOrderField],
) (*page.Page[*coredata.Framework, coredata.FrameworkOrderField], error) {
	var frameworks coredata.Frameworks
	organization := &coredata.Organization{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
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
		},
	)

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

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return framework.LoadByID(ctx, conn, s.svc.scope, frameworkID)
		},
	)

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

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
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
		},
	)
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

	return s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return framework.Delete(ctx, conn, s.svc.scope, frameworkID)
		},
	)
}

func (s FrameworkService) Import(
	ctx context.Context,
	organizationID gid.GID,
	req ImportFrameworkRequest,
) (*coredata.Framework, error) {
	var framework *coredata.Framework
	frameworkID := gid.New(organizationID.TenantID(), coredata.FrameworkEntityType)
	now := time.Now()

	err := s.svc.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
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
					TenantID:     organizationID.TenantID(),
					FrameworkID:  frameworkID,
					SectionTitle: control.ID,
					Name:         control.Name,
					Description:  control.Description,
					CreatedAt:    now,
					UpdatedAt:    now,
				}

				if err := control.Insert(ctx, tx, s.svc.scope); err != nil {
					return fmt.Errorf("cannot insert control: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return framework, nil
}

func (s FrameworkService) ExportAudit(
	ctx context.Context,
	frameworkID gid.GID,
) (string, error) {
	var archivePath string
	var objectKey string

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			framework := &coredata.Framework{}
			if err := framework.LoadByID(ctx, conn, s.svc.scope, frameworkID); err != nil {
				return fmt.Errorf("cannot load framework: %w", err)
			}

			now := time.Now()
			exportDir := filepath.Join(os.TempDir(), "probo-export", framework.Name, now.Format("2006-01-02-15-04-05"))

			if err := os.MkdirAll(exportDir, 0755); err != nil {
				return fmt.Errorf("cannot create export directory: %w", err)
			}

			fmt.Println("Exporting framework", framework.Name, "to", exportDir)

			cursor := page.NewCursor(
				0,
				nil,
				page.Head,
				page.OrderBy[coredata.ControlOrderField]{
					Field:     coredata.ControlOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
			)

			controls := coredata.Controls{}
			if err := controls.LoadByFrameworkID(ctx, conn, s.svc.scope, frameworkID, cursor); err != nil {
				return fmt.Errorf("cannot load controls: %w", err)
			}

			for _, control := range controls {
				controlDir := filepath.Join(exportDir, control.SectionTitle)
				if err := os.MkdirAll(controlDir, 0755); err != nil {
					return fmt.Errorf("cannot create control directory: %w", err)
				}

				measures := coredata.Measures{}
				cursor := page.NewCursor(
					0,
					nil,
					page.Head,
					page.OrderBy[coredata.MeasureOrderField]{
						Field:     coredata.MeasureOrderFieldCreatedAt,
						Direction: page.OrderDirectionAsc,
					},
				)

				if err := measures.LoadByControlID(ctx, conn, s.svc.scope, control.ID, cursor); err != nil {
					return fmt.Errorf("cannot load measures: %w", err)
				}

				documents := coredata.Documents{}

				cursor2 := page.NewCursor(
					0,
					nil,
					page.Head,
					page.OrderBy[coredata.DocumentOrderField]{
						Field:     coredata.DocumentOrderFieldCreatedAt,
						Direction: page.OrderDirectionAsc,
					},
				)

				if err := documents.LoadByControlID(ctx, conn, s.svc.scope, control.ID, cursor2); err != nil {
					return fmt.Errorf("cannot load documents: %w", err)
				}

				for _, document := range documents {
					documentDir := filepath.Join(controlDir, document.Title)
					if err := os.MkdirAll(documentDir, 0755); err != nil {
						return fmt.Errorf("cannot create document directory: %w", err)
					}

					version := coredata.DocumentVersion{}
					if err := version.LoadLatestVersion(ctx, conn, s.svc.scope, document.ID); err != nil {
						return fmt.Errorf("cannot load document version: %w", err)
					}

					documentFile := filepath.Join(documentDir, "document.md")
					if err := os.WriteFile(documentFile, []byte(version.Content), 0644); err != nil {
						return fmt.Errorf("cannot write document file: %w", err)
					}
				}

				for _, measure := range measures {
					measureDir := filepath.Join(controlDir, measure.Name)
					if err := os.MkdirAll(measureDir, 0755); err != nil {
						return fmt.Errorf("cannot create measure directory: %w", err)
					}

					evidences := coredata.Evidences{}
					cursor := page.NewCursor(
						0,
						nil,
						page.Head,
						page.OrderBy[coredata.EvidenceOrderField]{
							Field:     coredata.EvidenceOrderFieldCreatedAt,
							Direction: page.OrderDirectionAsc,
						},
					)

					if err := evidences.LoadByMeasureID(ctx, conn, s.svc.scope, measure.ID, cursor); err != nil {
						return fmt.Errorf("cannot load evidences: %w", err)
					}

					for _, evidence := range evidences {
						evidenceFile := filepath.Join(measureDir, evidence.Filename)

						if evidence.Type == coredata.EvidenceTypeFile && evidence.ObjectKey != "" {
							output, err := s.svc.s3.GetObject(
								ctx,
								&s3.GetObjectInput{
									Bucket: aws.String(s.svc.bucket),
									Key:    aws.String(evidence.ObjectKey),
								},
							)
							if err != nil {
								return fmt.Errorf("cannot download evidence file: %w", err)
							}

							defer output.Body.Close()

							file, err := os.Create(evidenceFile)
							if err != nil {
								return fmt.Errorf("cannot create evidence file: %w", err)
							}
							defer file.Close()

							_, err = io.Copy(file, output.Body)
							if err != nil {
								return fmt.Errorf("cannot write evidence file: %w", err)
							}
						}
					}
				}
			}

			archivePath = exportDir + ".tar.gz"
			if err := createTarGzArchive(exportDir, archivePath); err != nil {
				return fmt.Errorf("cannot create archive: %w", err)
			}
			defer os.Remove(archivePath)

			file, err := os.Open(archivePath)
			if err != nil {
				return fmt.Errorf("cannot open archive file: %w", err)
			}
			defer file.Close()

			objectKey = fmt.Sprintf("exports/%s/%s.tar.gz", frameworkID, now.Format("2006-01-02-15-04-05"))
			_, err = s.svc.s3.PutObject(
				ctx,
				&s3.PutObjectInput{
					Bucket: aws.String(s.svc.bucket),
					Key:    aws.String(objectKey),
					Body:   file,
				},
			)
			if err != nil {
				return fmt.Errorf("cannot upload archive to S3: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return "", err
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.svc.bucket),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		return "", fmt.Errorf("cannot generate presigned URL: %w", err)
	}

	return presignedReq.URL, nil
}

func createTarGzArchive(sourceDir, targetFile string) error {
	tarFile, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("could not create archive file: %w", err)
	}
	defer tarFile.Close()

	gzipWriter := io.Writer(tarFile)
	gzw := gzip.NewWriter(gzipWriter)
	defer gzw.Close()

	tarWriter := tar.NewWriter(gzw)
	defer tarWriter.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return fmt.Errorf("could not create tar header: %w", err)
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("could not get relative path: %w", err)
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("could not write tar header: %w", err)
		}

		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("could not open file %s: %w", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return fmt.Errorf("could not copy file content: %w", err)
			}
		}

		return nil
	})

	return err
}
