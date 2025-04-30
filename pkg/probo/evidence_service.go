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
	"mime"
	"net/url"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
)

type (
	EvidenceService struct {
		svc *TenantService
	}

	RequestEvidenceRequest struct {
		TaskID      gid.GID
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

	CreateEvidenceRequest struct {
		TaskID      gid.GID
		Name        string
		Type        coredata.EvidenceType
		File        io.Reader
		URL         string
		Description string
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
	evidenceID, err := gid.NewGID(s.svc.scope.GetTenantID(), coredata.EvidenceEntityType)
	if err != nil {
		return nil, fmt.Errorf("cannot create evidence global id: %w", err)
	}

	now := time.Now()

	evidence := &coredata.Evidence{
		ID:          evidenceID,
		TaskID:      req.TaskID,
		State:       coredata.EvidenceStateRequested,
		Type:        req.Type,
		Filename:    req.Name,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = s.svc.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
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

			evidence.State = coredata.EvidenceStateFulfilled

			if req.File != nil {
				evidence.Type = coredata.EvidenceTypeFile

				contentType := "application/octet-stream"
				if req.Filename != nil {
					evidence.Filename = *req.Filename
					if detectedType := mime.TypeByExtension(filepath.Ext(*req.Filename)); detectedType != "" {
						contentType = detectedType
					}
				}

				objectKey, err := uuid.NewV7()
				if err != nil {
					return fmt.Errorf("cannot generate object key: %w", err)
				}

				_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
					Bucket:      aws.String(s.svc.bucket),
					Key:         aws.String(objectKey.String()),
					Body:        req.File,
					ContentType: aws.String(contentType),
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

func (s EvidenceService) Create(
	ctx context.Context,
	req CreateEvidenceRequest,
) (*coredata.Evidence, error) {
	now := time.Now()
	evidenceID, err := gid.NewGID(s.svc.scope.GetTenantID(), coredata.EvidenceEntityType)
	if err != nil {
		return nil, fmt.Errorf("cannot create evidence global id: %w", err)
	}

	referenceID, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("cannot generate reference id: %w", err)
	}

	evidence := &coredata.Evidence{
		ID:          evidenceID,
		TaskID:      req.TaskID,
		State:       coredata.EvidenceStateFulfilled,
		ReferenceID: "custom-evidence-" + referenceID.String(),
		Type:        req.Type,
		Filename:    req.Name,
		URL:         req.URL,
		Description: req.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if req.Type == coredata.EvidenceTypeFile {
		contentType := "application/octet-stream"
		if req.Name != "" {
			if detectedType := mime.TypeByExtension(filepath.Ext(req.Name)); detectedType != "" {
				contentType = detectedType
			}
		}

		objectKey, err := uuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("cannot generate object key: %w", err)
		}

		_, err = s.svc.s3.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      aws.String(s.svc.bucket),
			Key:         aws.String(objectKey.String()),
			Body:        req.File,
			ContentType: aws.String(contentType),
		})
		if err != nil {
			return nil, fmt.Errorf("cannot upload file to S3: %w", err)
		}

		headOutput, err := s.svc.s3.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(s.svc.bucket),
			Key:    aws.String(objectKey.String()),
		})
		if err != nil {
			return nil, fmt.Errorf("cannot get object metadata: %w", err)
		}

		evidence.ObjectKey = objectKey.String()
		evidence.MimeType = contentType
		evidence.Size = uint64(*headOutput.ContentLength)
	} else if req.Type == coredata.EvidenceTypeLink {
		evidence.MimeType = "text/uri-list"
		evidence.Size = uint64(len(req.URL))
		evidence.ObjectKey = ""
	}

	task := &coredata.Task{}

	err = s.svc.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := task.LoadByID(ctx, conn, s.svc.scope, req.TaskID); err != nil {
				return fmt.Errorf("cannot load task %q: %w", req.TaskID, err)
			}

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
