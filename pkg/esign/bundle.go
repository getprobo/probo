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
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"go.gearno.de/crypto/uuid"
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

// UploadSignatureBundle writes a private zip of the signed file and its
// certificate. The caller inserts the returned file row.
func UploadSignatureBundle(
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

	reader, writer := io.Pipe()

	go func() {
		err := writeSignatureBundle(
			writer,
			bundleFile{Name: signedFile.FileName, Body: signedBody},
			bundleFile{Name: certificateFile.FileName, Body: certificateBody},
		)
		_ = writer.CloseWithError(err)
	}()

	size, err := files.PutFile(
		ctx,
		&bundle,
		reader,
		map[string]string{
			"type":         "electronic-signature-bundle",
			"signature-id": signature.ID.String(),
		},
		filemanager.WithAttachmentContentDisposition(),
	)
	if err != nil {
		_ = reader.Close()

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
