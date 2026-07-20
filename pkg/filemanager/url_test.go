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

package filemanager_test

import (
	"context"
	"io"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
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

	svc := filemanager.NewService(nil, base, nil, log.NewLogger(log.WithOutput(io.Discard)))
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

	svc := filemanager.NewService(nil, base, nil, log.NewLogger(log.WithOutput(io.Discard)))
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
	svc := filemanager.NewService(nil, nil, s3Client, log.NewLogger(log.WithOutput(io.Discard)))
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
