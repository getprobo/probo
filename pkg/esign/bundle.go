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
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const signedDocumentFallback = "signed-document.pdf"

// zipEpoch is the earliest timestamp the ZIP format can store.
var zipEpoch = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type bundleFile struct {
	Name string
	Body []byte
}

// WriteBundle writes a zip of the signed file and its certificate of
// completion. The returned name is the download filename. ErrBundleNotReady
// means the certificate has not been generated yet.
func (s *Service) WriteBundle(
	ctx context.Context,
	signatureID gid.GID,
	w io.Writer,
) (string, error) {
	var (
		scope           = coredata.NewScopeFromObjectID(signatureID)
		signature       coredata.ElectronicSignature
		signedFile      coredata.File
		certificateFile coredata.File
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := signature.LoadByID(ctx, conn, scope, signatureID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrElectronicSignatureNotFound
				}

				return fmt.Errorf("cannot load electronic signature: %w", err)
			}

			if signature.CertificateFileID == nil {
				return ErrBundleNotReady
			}

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
		return "", err
	}

	signedBody, err := s.fileManager.GetFileBytes(ctx, &signedFile)
	if err != nil {
		return "", fmt.Errorf("cannot read signed file: %w", err)
	}

	certificateBody, err := s.fileManager.GetFileBytes(ctx, &certificateFile)
	if err != nil {
		return "", fmt.Errorf("cannot read certificate file: %w", err)
	}

	if err := writeSignatureBundle(
		w,
		bundleFile{Name: signedFile.FileName, Body: signedBody},
		bundleFile{Name: certificateFile.FileName, Body: certificateBody},
	); err != nil {
		return "", fmt.Errorf("cannot write signature bundle: %w", err)
	}

	return bundleArchiveName(signature.DocumentName, signature.DocumentType), nil
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
