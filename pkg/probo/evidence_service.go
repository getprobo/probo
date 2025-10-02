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
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/filevalidation"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
)

type (
	EvidenceService struct {
		svc           *TenantService
		fileValidator *filevalidation.FileValidator
	}

	File struct {
		Content     io.Reader
		Filename    string
		Size        int64
		ContentType string
	}

	RequestEvidenceRequest struct {
		MeasureID   *gid.GID
		TaskID      *gid.GID
		Type        coredata.EvidenceType
		Name        string
		Description string
	}

	FulfilledEvidenceRequest struct {
		EvidenceID gid.GID
		File       io.Reader
		URL        *string
		Filename   *string
	}

	UploadTaskEvidenceRequest struct {
		TaskID gid.GID
		File   File
	}

	UploadMeasureEvidenceRequest struct {
		MeasureID gid.GID
		File      File
	}
)

func (s EvidenceService) Get(
	ctx context.Context,
	evidenceID gid.GID,
) (*coredata.Evidence, error) {
	evidence := &coredata.Evidence{}

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return evidence.LoadByID(ctx, conn, s.svc.scope, evidenceID)
		},
	)

	if err != nil {
		return nil, err
	}

	return evidence, nil
}

func (s EvidenceService) Request(
	ctx context.Context,
	req RequestEvidenceRequest,
) (*coredata.Evidence, error) {
	evidenceID := gid.New(s.svc.scope.GetTenantID(), coredata.EvidenceEntityType)
	now := time.Now()

	evidence := &coredata.Evidence{
		ID:          evidenceID,
		State:       coredata.EvidenceStateRequested,
		Type:        req.Type,
		Filename:    req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			task := &coredata.Task{}
			if req.TaskID != nil {
				if err := task.LoadByID(ctx, conn, s.svc.scope, *req.TaskID); err != nil {
					return fmt.Errorf("cannot load task: %w", err)
				}

				if task.MeasureID == nil {
					return fmt.Errorf("task %q has no measure", req.TaskID)
				}

				evidence.TaskID = req.TaskID
				evidence.MeasureID = *task.MeasureID
			} else if req.MeasureID != nil {
				evidence.MeasureID = *req.MeasureID
			} else {
				return fmt.Errorf("measure id or task id is required")
			}

			return evidence.Insert(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot insert evidence: %w", err)
	}

	return evidence, nil
}

func (s EvidenceService) Fulfill(
	ctx context.Context,
	req FulfilledEvidenceRequest,
) (*coredata.Evidence, error) {
	evidence := &coredata.Evidence{}

	err := s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := evidence.LoadByID(ctx, conn, s.svc.scope, req.EvidenceID); err != nil {
				return fmt.Errorf("cannot load evidence: %w", err)
			}

			measure := &coredata.Measure{}
			if err := measure.LoadByID(ctx, conn, s.svc.scope, evidence.MeasureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}

			evidence.State = coredata.EvidenceStateFulfilled

			if req.File != nil {
				evidence.Type = coredata.EvidenceTypeFile

				var fileSize int64
				var fileContent io.ReadSeeker

				if seeker, ok := req.File.(io.Seeker); ok {
					size, err := seeker.Seek(0, io.SeekEnd)
					if err != nil {
						return fmt.Errorf("cannot determine file size: %w", err)
					}

					_, err = seeker.Seek(0, io.SeekStart)
					if err != nil {
						return fmt.Errorf("cannot reset file position: %w", err)
					}
					fileSize = size
					fileContent = req.File.(io.ReadSeeker)
				} else {
					buf, err := io.ReadAll(req.File)
					if err != nil {
						return fmt.Errorf("cannot read file: %w", err)
					}
					fileSize = int64(len(buf))
					fileContent = bytes.NewReader(buf)
				}

				contentType := "application/octet-stream"
				if req.Filename != nil {
					evidence.Filename = *req.Filename
					if detectedType := mime.TypeByExtension(filepath.Ext(*req.Filename)); detectedType != "" {
						contentType = detectedType
					}
				}

				if err := s.fileValidator.Validate(evidence.Filename, contentType, fileSize); err != nil {
					return err
				}

				objectKey, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("cannot generate object key: %w", err)
				}

				_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
					Bucket:      aws.String(s.svc.bucket),
					Key:         aws.String(objectKey.String()),
					Body:        fileContent,
					ContentType: aws.String(contentType),
					Metadata: map[string]string{
						"type":            "evidence",
						"evidence-id":     req.EvidenceID.String(),
						"organization-id": measure.OrganizationID.String(),
					},
				})
				if err != nil {
					return fmt.Errorf("cannot upload file to S3: %w", err)
				}

				headOutput, err := s.svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
					Bucket: aws.String(s.svc.bucket),
					Key:    aws.String(objectKey.String()),
				})
				if err != nil {
					return fmt.Errorf("cannot get object metadata: %w", err)
				}

				evidence.ObjectKey = objectKey.String()
				evidence.MimeType = contentType
				evidence.Size = uint64(*headOutput.ContentLength)
			} else if req.URL != nil {
				evidence.Type = coredata.EvidenceTypeLink
				evidence.URL = *req.URL
			}

			evidence.UpdatedAt = time.Now()

			return evidence.Update(ctx, conn, s.svc.scope)
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot update evidence: %w", err)
	}

	return evidence, nil
}

func (s EvidenceService) UploadTaskEvidence(
	ctx context.Context,
	req UploadTaskEvidenceRequest,
) (*coredata.Evidence, error) {
	now := time.Now()
	evidenceID := gid.New(s.svc.scope.GetTenantID(), coredata.EvidenceEntityType)

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	evidence := &coredata.Evidence{
		ID:          evidenceID,
		TaskID:      &req.TaskID,
		State:       coredata.EvidenceStateFulfilled,
		ReferenceID: "custom-evidence-" + referenceID.String(),
		Type:        coredata.EvidenceTypeFile,
		Filename:    req.File.Filename,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if req.File.ContentType == "" {
		req.File.ContentType = "application/octet-stream"
	}

	if err := s.fileValidator.Validate(req.File.Filename, req.File.ContentType, req.File.Size); err != nil {
		return nil, err
	}

	var measureOrganizationID gid.GID

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			task := &coredata.Task{}

			if err := task.LoadByID(ctx, conn, s.svc.scope, req.TaskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", req.TaskID, err)
			}

			if task.MeasureID == nil {
				return fmt.Errorf("task %q has no measure", req.TaskID)
			}

			measure := &coredata.Measure{}
			if err := measure.LoadByID(ctx, conn, s.svc.scope, *task.MeasureID); err != nil {
				return fmt.Errorf("cannot load measure: %w", err)
			}
			measureOrganizationID = measure.OrganizationID

			objectKey, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("cannot generate object key: %w", err)
			}

			_, err = s.svc.s3.PutObject(
				ctx,
				&s3.PutObjectInput{
					Bucket:      aws.String(s.svc.bucket),
					Key:         aws.String(objectKey.String()),
					Body:        req.File.Content,
					ContentType: aws.String(req.File.ContentType),
					Metadata: map[string]string{
						"type":            "evidence",
						"evidence-id":     evidenceID.String(),
						"organization-id": measureOrganizationID.String(),
					},
				},
			)
			if err != nil {
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			evidence.ObjectKey = objectKey.String()
			evidence.MimeType = req.File.ContentType
			evidence.Size = uint64(req.File.Size)
			evidence.MeasureID = *task.MeasureID

			if err := evidence.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert evidence: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		// TODO try do delete file from s3 if it's a file type
		return nil, err
	}

	return evidence, nil
}

func (s EvidenceService) UploadMeasureEvidence(
	ctx context.Context,
	req UploadMeasureEvidenceRequest,
) (*coredata.Evidence, error) {
	now := time.Now()
	evidenceID := gid.New(s.svc.scope.GetTenantID(), coredata.EvidenceEntityType)

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	evidence := &coredata.Evidence{
		ID:          evidenceID,
		MeasureID:   req.MeasureID,
		State:       coredata.EvidenceStateFulfilled,
		ReferenceID: "custom-evidence-" + referenceID.String(),
		Type:        coredata.EvidenceTypeFile,
		Filename:    req.File.Filename,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if req.File.ContentType == "" {
		req.File.ContentType = "application/octet-stream"
	}

	if err := s.fileValidator.Validate(req.File.Filename, req.File.ContentType, req.File.Size); err != nil {
		return nil, err
	}

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			measure := &coredata.Measure{}
			if err := measure.LoadByID(ctx, conn, s.svc.scope, req.MeasureID); err != nil {
				return fmt.Errorf("cannot load measure %q: %w", req.MeasureID, err)
			}

			objectKey, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("cannot generate object key: %w", err)
			}

			_, err = s.svc.s3.PutObject(
				ctx,
				&s3.PutObjectInput{
					Bucket:      aws.String(s.svc.bucket),
					Key:         aws.String(objectKey.String()),
					Body:        req.File.Content,
					ContentType: aws.String(req.File.ContentType),
					Metadata: map[string]string{
						"type":            "evidence",
						"evidence-id":     evidenceID.String(),
						"organization-id": measure.OrganizationID.String(),
					},
				},
			)
			if err != nil {
				return fmt.Errorf("cannot upload file to S3: %w", err)
			}

			evidence.ObjectKey = objectKey.String()
			evidence.MimeType = req.File.ContentType
			evidence.Size = uint64(req.File.Size)
			evidence.MeasureID = req.MeasureID

			if err := evidence.Insert(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot insert evidence: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		// TODO try do delete file from s3 if it's a file type
		return nil, err
	}

	return evidence, nil
}

func (s EvidenceService) GenerateFileURL(
	ctx context.Context,
	evidenceID gid.GID,
	expiresIn time.Duration,
) (*string, error) {
	evidence, err := s.Get(ctx, evidenceID)
	if err != nil {
		return nil, fmt.Errorf("cannot get evidence: %w", err)
	}

	if evidence.Type == coredata.EvidenceTypeLink {
		return nil, fmt.Errorf("cannot generate file URL for link type evidence")
	}

	presignClient := s3.NewPresignClient(s.svc.s3)

	// Use RFC 6266/5987 encoding for filename with UTF-8 support
	encodedFilename := url.QueryEscape(evidence.Filename)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
		encodedFilename, encodedFilename)

	presignedReq, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     aws.String(s.svc.bucket),
		Key:                        aws.String(evidence.ObjectKey),
		ResponseCacheControl:       aws.String("max-age=3600, public"),
		ResponseContentType:        aws.String(evidence.MimeType),
		ResponseContentDisposition: aws.String(contentDisposition),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return nil, fmt.Errorf("cannot presign GetObject request: %w", err)
	}

	return &presignedReq.URL, nil
}

func (s EvidenceService) CountForMeasureID(
	ctx context.Context,
	measureID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			evidences := coredata.Evidences{}
			count, err = evidences.CountByMeasureID(ctx, conn, s.svc.scope, measureID)
			if err != nil {
				return fmt.Errorf("cannot count evidences: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s EvidenceService) ListForMeasureID(
	ctx context.Context,
	measureID gid.GID,
	cursor *page.Cursor[coredata.EvidenceOrderField],
) (*page.Page[*coredata.Evidence, coredata.EvidenceOrderField], error) {
	var evidences coredata.Evidences

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return evidences.LoadByMeasureID(
				ctx,
				conn,
				s.svc.scope,
				measureID,
				cursor,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(evidences, cursor), nil
}

func (s EvidenceService) CountForTaskID(
	ctx context.Context,
	taskID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			evidences := coredata.Evidences{}
			count, err = evidences.CountByTaskID(ctx, conn, s.svc.scope, taskID)
			if err != nil {
				return fmt.Errorf("cannot count evidences: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s EvidenceService) ListForTaskID(
	ctx context.Context,
	taskID gid.GID,
	cursor *page.Cursor[coredata.EvidenceOrderField],
) (*page.Page[*coredata.Evidence, coredata.EvidenceOrderField], error) {
	var evidences coredata.Evidences

	err := s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return evidences.LoadByTaskID(
				ctx,
				conn,
				s.svc.scope,
				taskID,
				cursor,
			)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(evidences, cursor), nil
}

func (s *EvidenceService) Delete(
	ctx context.Context,
	evidenceID gid.GID,
) error {
	evidence := &coredata.Evidence{ID: evidenceID}

	return s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := evidence.Delete(ctx, conn, s.svc.scope); err != nil {
				return fmt.Errorf("cannot delete evidence: %w", err)
			}

			return nil
		},
	)
}
