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

package esign

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
)

const signedDocumentFallback = "signed-document.pdf"

// zipEpoch is the earliest timestamp the ZIP format can store.
var zipEpoch = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type bundleFile struct {
	Name string
	Body []byte
}

// EnsureBundle returns the zip of the signed file and its certificate.
// A signature that already has a certificate and no bundle gets one stored
// on the first call.
func (s *Service) EnsureBundle(
	ctx context.Context,
	scope coredata.Scoper,
	signatureID gid.GID,
) (*coredata.File, error) {
	var signature coredata.ElectronicSignature

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := signature.LoadByID(ctx, conn, scope, signatureID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrElectronicSignatureNotFound
				}

				return fmt.Errorf("cannot load electronic signature: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if signature.BundleFileID != nil {
		return loadFile(ctx, s.pg, scope, *signature.BundleFileID)
	}

	if signature.CertificateFileID == nil {
		return nil, nil
	}

	var (
		signedFile      coredata.File
		certificateFile coredata.File
	)

	err = s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := signedFile.LoadByID(ctx, conn, scope, signature.FileID); err != nil {
				return fmt.Errorf("cannot load signed file: %w", err)
			}

			if err := certificateFile.LoadByID(ctx, conn, scope, *signature.CertificateFileID); err != nil {
				return fmt.Errorf("cannot load certificate file: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	uploaded, err := uploadSignatureBundle(
		ctx,
		s.fileManager,
		s.bucket,
		&signature,
		&signedFile,
		&certificateFile,
	)
	if err != nil {
		return nil, err
	}

	var stored coredata.File

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var locked coredata.ElectronicSignature

			if err := locked.LoadByIDForUpdate(ctx, tx, scope, signatureID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrElectronicSignatureNotFound
				}

				return fmt.Errorf("cannot lock electronic signature: %w", err)
			}

			if locked.BundleFileID != nil {
				if err := stored.LoadByID(ctx, tx, scope, *locked.BundleFileID); err != nil {
					return fmt.Errorf("cannot load signature bundle: %w", err)
				}

				return nil
			}

			if err := uploaded.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert signature bundle: %w", err)
			}

			locked.BundleFileID = &uploaded.ID
			locked.UpdatedAt = time.Now()

			if err := locked.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update signature bundle: %w", err)
			}

			stored = uploaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &stored, nil
}

func loadFile(
	ctx context.Context,
	pgClient *pg.Client,
	scope coredata.Scoper,
	fileID gid.GID,
) (*coredata.File, error) {
	var file coredata.File

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := file.LoadByID(ctx, conn, scope, fileID); err != nil {
				return fmt.Errorf("cannot load signature bundle: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &file, nil
}

func uploadSignatureBundle(
	ctx context.Context,
	files *filemanager.Service,
	bucket string,
	signature *coredata.ElectronicSignature,
	signedFile *coredata.File,
	certificateFile *coredata.File,
) (coredata.File, error) {
	signedBody, err := files.GetFileBytes(ctx, signedFile)
	if err != nil {
		return coredata.File{}, fmt.Errorf("cannot read signed file: %w", err)
	}

	certificateBody, err := files.GetFileBytes(ctx, certificateFile)
	if err != nil {
		return coredata.File{}, fmt.Errorf("cannot read certificate file: %w", err)
	}

	var body bytes.Buffer

	if err := writeSignatureBundle(
		&body,
		bundleFile{Name: signedFile.FileName, Body: signedBody},
		bundleFile{Name: certificateFile.FileName, Body: certificateBody},
	); err != nil {
		return coredata.File{}, fmt.Errorf("cannot write signature bundle: %w", err)
	}

	now := time.Now()
	bundle := coredata.File{
		ID:             gid.New(signature.ID.TenantID(), coredata.FileEntityType),
		OrganizationID: signature.OrganizationID,
		BucketName:     bucket,
		MimeType:       "application/zip",
		FileName:       bundleArchiveName(signature.DocumentName, signature.DocumentType),
		FileKey:        uuid.MustNewV4().String(),
		Visibility:     coredata.FileVisibilityPrivate,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	size, err := files.PutFile(
		ctx,
		&bundle,
		bytes.NewReader(body.Bytes()),
		map[string]string{
			"type":         "electronic-signature-bundle",
			"signature-id": signature.ID.String(),
		},
		filemanager.WithAttachmentContentDisposition(),
	)
	if err != nil {
		return coredata.File{}, fmt.Errorf("cannot upload signature bundle: %w", err)
	}

	bundle.FileSize = size

	return bundle, nil
}

func writeSignatureBundle(w io.Writer, signed, certificate bundleFile) (err error) {
	signedName, certificateName := distinctEntryNames(signed.Name, certificate.Name)

	archive := zip.NewWriter(w)

	defer func() {
		if closeErr := archive.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("cannot close archive: %w", closeErr)
		}
	}()

	if err := writeZipEntry(archive, signedName, signed.Body); err != nil {
		return fmt.Errorf("cannot write signed file to archive: %w", err)
	}

	if err := writeZipEntry(archive, certificateName, certificate.Body); err != nil {
		return fmt.Errorf("cannot write certificate to archive: %w", err)
	}

	return nil
}

func writeZipEntry(archive *zip.Writer, name string, body []byte) error {
	header := &zip.FileHeader{
		Name:     name,
		Method:   zip.Store,
		Modified: zipEpoch,
	}

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("cannot create archive entry: %w", err)
	}

	if _, err := writer.Write(body); err != nil {
		return fmt.Errorf("cannot write archive entry: %w", err)
	}

	return nil
}

func distinctEntryNames(signedName, certificateName string) (string, string) {
	signedName = zipEntryName(signedName, signedDocumentFallback)
	certificateName = zipEntryName(certificateName, certificateFilename)

	if strings.EqualFold(signedName, certificateName) {
		certificateName = disambiguateEntryName(certificateName)
	}

	return signedName, certificateName
}

func disambiguateEntryName(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	return base + "-2" + ext
}

func bundleArchiveName(
	documentName *string,
	documentType coredata.ElectronicSignatureDocumentType,
) string {
	base := zipEntryName(ref.UnrefOrZero(documentName), documentType.DisplayName())
	ext := filepath.Ext(base)

	if ext != "" {
		base = strings.Trim(strings.TrimSuffix(base, ext), " .")
	}

	if base == "" {
		base = "signed-document"
	}

	return base + ".zip"
}

func zipEntryName(name, fallback string) string {
	name = strings.Map(func(r rune) rune {
		if r < 0x20 || r == unicode.ReplacementChar || r == 0x7f {
			return '_'
		}

		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		default:
			return r
		}
	}, name)

	name = strings.Trim(name, " .")
	if name == "" || name == "_" {
		return fallback
	}

	return name
}
