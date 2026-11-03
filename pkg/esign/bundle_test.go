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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestWriteSignatureBundle_IncludesBothFiles(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	err := writeSignatureBundle(
		&buf,
		bundleFile{Name: "Mutual NDA.pdf", Body: []byte("signed-bytes")},
		bundleFile{Name: certificateFilename, Body: []byte("certificate-bytes")},
	)
	require.NoError(t, err)

	entries := readZipEntries(t, &buf)
	assert.Equal(t, "signed-bytes", entries["Mutual NDA.pdf"])
	assert.Equal(t, "certificate-bytes", entries[certificateFilename])
	assert.Len(t, entries, 2)
}

func TestWriteSignatureBundle_DisambiguatesCollidingNames(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	err := writeSignatureBundle(
		&buf,
		bundleFile{Name: certificateFilename, Body: []byte("signed")},
		bundleFile{Name: certificateFilename, Body: []byte("certificate")},
	)
	require.NoError(t, err)

	entries := readZipEntries(t, &buf)
	assert.Equal(t, "signed", entries[certificateFilename])
	assert.Equal(t, "certificate", entries["certificate-of-completion-2.pdf"])
}

func TestWriteSignatureBundle_FlattensPathNames(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	err := writeSignatureBundle(
		&buf,
		bundleFile{Name: "../secrets/nda.pdf", Body: []byte("signed")},
		bundleFile{Name: "certs/certificate-of-completion.pdf", Body: []byte("certificate")},
	)
	require.NoError(t, err)

	entries := readZipEntries(t, &buf)
	assert.Equal(t, "signed", entries["_secrets_nda.pdf"])
	assert.Equal(t, "certificate", entries["certs_certificate-of-completion.pdf"])
}

func TestBundleArchiveName(t *testing.T) {
	t.Parallel()

	t.Run("uses the document name without its extension", func(t *testing.T) {
		t.Parallel()

		name := "Mutual NDA.pdf"
		assert.Equal(
			t,
			"Mutual NDA.zip",
			bundleArchiveName(&name, coredata.ElectronicSignatureDocumentTypeNDA),
		)
	})

	t.Run("falls back to the document type", func(t *testing.T) {
		t.Parallel()

		assert.Equal(
			t,
			"Non-Disclosure Agreement.zip",
			bundleArchiveName(nil, coredata.ElectronicSignatureDocumentTypeNDA),
		)
	})
}

func TestZipEntryName_UsesFallbackWhenBlank(t *testing.T) {
	t.Parallel()

	assert.Equal(t, signedDocumentFallback, zipEntryName("   ", signedDocumentFallback))
	assert.Equal(t, signedDocumentFallback, zipEntryName("..", signedDocumentFallback))
}

func readZipEntries(t *testing.T, buf *bytes.Buffer) map[string]string {
	t.Helper()

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	entries := make(map[string]string, len(reader.File))
	for _, file := range reader.File {
		rc, err := file.Open()
		require.NoError(t, err)

		body, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())

		entries[file.Name] = string(body)
	}

	return entries
}
