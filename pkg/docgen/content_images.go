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

package docgen

import (
	"context"
	"encoding/json"
	"strings"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

type (
	// ImageContentReader loads an image file as base64 for PDF embedding.
	ImageContentReader interface {
		GetFileBase64(ctx context.Context, file *coredata.File) (string, string, error)
	}
)

// ContentImageFiles loads the active image files referenced by document content.
// A file that cannot be parsed, loaded, or is not an image is skipped.
func ContentImageFiles(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	content string,
	logger *log.Logger,
) []*coredata.File {
	if strings.TrimSpace(content) == "" {
		return nil
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		logger.WarnCtx(ctx, "cannot parse document content for PDF images", log.Error(err))

		return nil
	}

	var files []*coredata.File

	seen := map[string]struct{}{}

	for _, ref := range prosemirror.ReferencedFiles(node) {
		if !ref.Image {
			continue
		}

		if _, ok := seen[ref.ID]; ok {
			continue
		}

		seen[ref.ID] = struct{}{}

		fileID, err := gid.ParseGID(ref.ID)
		if err != nil {
			logger.WarnCtx(ctx, "cannot parse image file id", log.Error(err))

			continue
		}

		file := &coredata.File{}
		if err := file.LoadActiveByID(ctx, conn, scope, fileID); err != nil {
			logger.WarnCtx(
				ctx,
				"cannot load image file for PDF",
				log.Error(err),
				log.String("file_id", fileID.String()),
			)

			continue
		}

		if !prosemirror.IsImageMime(file.MimeType) {
			logger.WarnCtx(
				ctx,
				"skipping non-image file in PDF",
				log.String("file_id", file.ID.String()),
				log.String("mime_type", file.MimeType),
			)

			continue
		}

		files = append(files, file)
	}

	return files
}

// EmbedContentImages replaces referenced image files with data URLs and drops
// remote image sources. Malformed content is returned unchanged.
func EmbedContentImages(
	ctx context.Context,
	logger *log.Logger,
	content string,
	files []*coredata.File,
	reader ImageContentReader,
) string {
	if strings.TrimSpace(content) == "" {
		return content
	}

	node, err := prosemirror.Parse(content)
	if err != nil {
		logger.WarnCtx(ctx, "cannot parse document content for PDF images", log.Error(err))

		return content
	}

	dataURLs := map[string]string{}

	for _, file := range files {
		encoded, mimeType, err := reader.GetFileBase64(ctx, file)
		if err != nil {
			logger.WarnCtx(
				ctx,
				"cannot read image file for PDF",
				log.Error(err),
				log.String("file_id", file.ID.String()),
			)

			continue
		}

		dataURLs[file.ID.String()] = "data:" + mimeType + ";base64," + encoded
	}

	prosemirror.EmbedImageDataURLs(&node, dataURLs)
	prosemirror.DropRemoteImageSrcs(&node)

	out, err := json.Marshal(node)
	if err != nil {
		logger.WarnCtx(ctx, "cannot marshal document content with embedded images", log.Error(err))

		return content
	}

	return string(out)
}
