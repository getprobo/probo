// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package filemanager_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
)

func TestGenerateFileURL_PublicFile(t *testing.T) {
	t.Parallel()

	base, err := baseurl.Parse("https://app.example.com")
	if err != nil {
		t.Fatalf("cannot parse base URL: %v", err)
	}

	svc := filemanager.NewService(nil, base, nil)
	file := &coredata.File{
		ID:         gid.New(gid.NilTenant, coredata.FileEntityType),
		Visibility: coredata.FileVisibilityPublic,
	}

	assert.Equal(
		t,
		"https://app.example.com/api/files/v1/public/"+file.ID.String(),
		svc.GenerateFileURL(file),
	)
}

func TestGenerateFileURL_PrivateFile(t *testing.T) {
	t.Parallel()

	base, err := baseurl.Parse("https://app.example.com")
	if err != nil {
		t.Fatalf("cannot parse base URL: %v", err)
	}

	svc := filemanager.NewService(nil, base, nil)
	file := &coredata.File{
		ID:         gid.New(gid.NilTenant, coredata.FileEntityType),
		Visibility: coredata.FileVisibilityPrivate,
	}

	assert.Equal(
		t,
		"https://app.example.com/api/files/v1/"+file.ID.String(),
		svc.GenerateFileURL(file),
	)
}

func TestGeneratePresignedURL_EscapesContentDispositionFilename(t *testing.T) {
	t.Parallel()

	s3Client := awss3.NewFromConfig(
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("access-key", "secret-key", ""),
		},
	)
	svc := filemanager.NewService(nil, nil, s3Client)
	file := &coredata.File{
		BucketName: "uploads",
		FileKey:    "tenant/file",
		FileName:   `report "Q2"/résumé 100%.pdf`,
		MimeType:   "application/pdf",
	}

	rawURL, err := svc.GeneratePresignedURL(context.Background(), file, time.Hour)
	require.NoError(t, err)

	parsedURL, err := url.Parse(rawURL)
	require.NoError(t, err)

	assert.Equal(
		t,
		`attachment; filename="report \"Q2\"/r_sum_ 100%.pdf"; filename*=UTF-8''report%20%22Q2%22%2Fr%C3%A9sum%C3%A9%20100%25.pdf`,
		parsedURL.Query().Get("response-content-disposition"),
	)
}
