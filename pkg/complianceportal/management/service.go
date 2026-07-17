// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

// Package management holds the scoped, admin-facing compliance portal services
// (compliance page CRUD, domains, frameworks, external URLs, references, files and
// accesses). It is the write side of the compliance portal feature.
package management

import (
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/certmanager"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/slack"
)

const (
	NameMaxLength    = 100
	TitleMaxLength   = 1000
	ContentMaxLength = 5000
)

type (
	// Service is the admin-facing compliance portal service. It exposes the
	// scoped CRUD operations for the compliance page and its related resources as
	// methods on a single type.
	Service struct {
		pg            *pg.Client
		s3            *s3.Client
		bucket        string
		baseURL       string
		baseDomain    string
		fileManager   *filemanager.Service
		certManager   *certmanager.Service
		logger        *log.Logger
		SlackMessages *slack.Service
		fileValidator *filevalidation.FileValidator
	}

	// FileUpload is an in-memory file supplied by a caller for upload.
	FileUpload struct {
		Content     io.Reader
		Filename    string
		Size        int64
		ContentType string
	}

	// File is an in-memory file supplied by a caller for upload.
	File struct {
		Content     io.Reader
		Filename    string
		Size        int64
		ContentType string
	}
)

func NewService(
	pgClient *pg.Client,
	s3Client *s3.Client,
	bucket string,
	baseURL string,
	baseDomain string,
	fileManagerService *filemanager.Service,
	certManagerService *certmanager.Service,
	slackService *slack.Service,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:            pgClient,
		s3:            s3Client,
		bucket:        bucket,
		baseURL:       baseURL,
		baseDomain:    baseDomain,
		fileManager:   fileManagerService,
		certManager:   certManagerService,
		logger:        logger,
		SlackMessages: slackService,
		fileValidator: filevalidation.NewValidator(
			filevalidation.WithCategories(
				filevalidation.CategoryData,
				filevalidation.CategoryDocument,
				filevalidation.CategoryImage,
				filevalidation.CategoryPresentation,
				filevalidation.CategorySpreadsheet,
				filevalidation.CategoryText,
			),
			filevalidation.WithMaxFileSize(10*1024*1024), // 10MB
		),
	}
}
