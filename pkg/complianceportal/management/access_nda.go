// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package management

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

// RecordedNDAMaxBytes is the upload cap for an externally signed portal NDA PDF.
const RecordedNDAMaxBytes int64 = 10 * 1024 * 1024

var recordedNDAPDFMagic = []byte("%PDF-")

type UploadRecordedNDARequest struct {
	AccessID    gid.GID
	File        io.Reader
	FileName    string
	ContentType string
	Size        int64
	SignedAt    time.Time
}

func (req *UploadRecordedNDARequest) Validate() error {
	v := validator.New()

	v.Check(
		req.AccessID,
		"compliance_portal_access_id",
		validator.Required(),
		validator.GID(coredata.CompliancePortalAccessEntityType),
	)
	v.Check(req.FileName, "file_name", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))

	var signedAt *time.Time
	if !req.SignedAt.IsZero() {
		signedAt = &req.SignedAt
	}

	// Noon UTC of the chosen calendar day can land a few hours ahead of now.
	v.Check(signedAt, "signed_at", validator.Required(), validator.Before(time.Now().Add(48*time.Hour)))

	if err := v.Error(); err != nil {
		return err
	}

	contentType := filevalidation.ResolveContentType(req.ContentType, req.FileName)
	req.ContentType = contentType

	fv := filevalidation.NewValidator(
		filevalidation.WithMimeTypes("application/pdf"),
		filevalidation.WithMaxFileSize(RecordedNDAMaxBytes),
	)

	if err := fv.Validate(req.FileName, contentType, req.Size); err != nil {
		return fmt.Errorf("invalid recorded NDA file: %w", err)
	}

	return nil
}

func (s *Service) SetAccessNDABypass(
	ctx context.Context,
	scope coredata.Scoper,
	accessID gid.GID,
	bypassed bool,
) (*coredata.CompliancePortalAccess, error) {
	var access coredata.CompliancePortalAccess

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := access.LoadByIDForUpdate(ctx, tx, scope, accessID); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			access.NDABypassed = bypassed
			access.UpdatedAt = time.Now()

			if err := access.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal access: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &access, nil
}

func (s *Service) UploadRecordedNDA(
	ctx context.Context,
	scope coredata.Scoper,
	req *UploadRecordedNDARequest,
) (*coredata.CompliancePortalAccess, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	fileReader, err := ensureRecordedNDAPDF(req.File)
	if err != nil {
		return nil, err
	}

	var (
		access coredata.CompliancePortalAccess
		s3Key  string
	)

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := access.LoadByIDForUpdate(ctx, tx, scope, req.AccessID); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			objectKey, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("cannot generate object key: %w", err)
			}

			now := time.Now()
			fileID := gid.New(scope.GetTenantID(), coredata.FileEntityType)

			file := &coredata.File{
				ID:             fileID,
				OrganizationID: access.OrganizationID,
				BucketName:     s.bucket,
				MimeType:       req.ContentType,
				FileName:       req.FileName,
				FileKey:        objectKey.String(),
				Visibility:     coredata.FileVisibilityPrivate,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			fileSize, err := s.fileManager.PutFile(
				ctx,
				file,
				fileReader,
				map[string]string{
					"type":                        "compliance-portal-recorded-nda",
					"compliance-portal-access-id": access.ID.String(),
					"organization-id":             access.OrganizationID.String(),
				},
			)
			if err != nil {
				return fmt.Errorf("cannot upload recorded NDA: %w", err)
			}

			s3Key = objectKey.String()
			file.FileSize = fileSize

			if err := file.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert recorded NDA file: %w", err)
			}

			signedAt := req.SignedAt
			access.RecordedNDAFileID = &fileID
			access.RecordedNDASignedAt = &signedAt
			access.UpdatedAt = now

			if err := access.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal access: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		s.cleanupFileS3Object(ctx, scope, s3Key)
		return nil, err
	}

	return &access, nil
}

func (s *Service) DeleteRecordedNDA(
	ctx context.Context,
	scope coredata.Scoper,
	accessID gid.GID,
) (*coredata.CompliancePortalAccess, error) {
	var access coredata.CompliancePortalAccess

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := access.LoadByIDForUpdate(ctx, tx, scope, accessID); err != nil {
				return fmt.Errorf("cannot load compliance portal access: %w", err)
			}

			access.RecordedNDAFileID = nil
			access.RecordedNDASignedAt = nil
			access.UpdatedAt = time.Now()

			if err := access.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update compliance portal access: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &access, nil
}

func ensureRecordedNDAPDF(file io.Reader) (io.Reader, error) {
	header := make([]byte, len(recordedNDAPDFMagic))
	n, err := io.ReadFull(file, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("cannot read recorded NDA: %w", err)
	}

	if n < len(recordedNDAPDFMagic) || !bytes.Equal(header[:n], recordedNDAPDFMagic) {
		return nil, validator.ValidationErrors{
			{
				Field:   "file",
				Code:    validator.ErrorCodeInvalidFormat,
				Message: "must be a PDF",
			},
		}
	}

	return io.MultiReader(bytes.NewReader(header[:n]), file), nil
}
