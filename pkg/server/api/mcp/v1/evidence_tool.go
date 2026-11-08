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

package mcp_v1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
	"go.probo.inc/probo/pkg/validator"
)

// uploadInternalControlEvidence attaches a base64-encoded file to an
// internal control, matching the uploadInternalControlEvidence GraphQL
// mutation. MCP carries no multipart uploads, so the file arrives inline the
// same way uploadFile receives it.
func (r *Resolver) uploadInternalControlEvidence(
	ctx context.Context,
	input *types.UploadInternalControlEvidenceInput,
) (*mcp.CallToolResult, types.UploadInternalControlEvidenceOutput, error) {
	scope, err := r.Authorize(ctx, input.InternalControlID, probo.ActionInternalControlEvidenceUpload)
	if err != nil {
		return nil, types.UploadInternalControlEvidenceOutput{}, err
	}

	payload, err := decodeFileContent(input.ContentBase64, r.proboSvc.Evidences.MaxFileSize())
	if err != nil {
		return nil, types.UploadInternalControlEvidenceOutput{}, err
	}

	mimeType := ""
	if input.MimeType != nil {
		mimeType = strings.TrimSpace(*input.MimeType)
	}

	if mimeType == "" {
		inferred, ok := r.proboSvc.Evidences.ContentTypeForFilename(input.FileName)
		if !ok {
			return nil, types.UploadInternalControlEvidenceOutput{}, fmt.Errorf("cannot infer mime_type for %q: file extension is missing or not supported", input.FileName)
		}

		mimeType = inferred
	}

	evidence, err := r.proboSvc.Evidences.UploadInternalControlEvidence(
		ctx,
		scope,
		probo.UploadInternalControlEvidenceRequest{
			InternalControlID: input.InternalControlID,
			File: probo.FileUpload{
				Content:     bytes.NewReader(payload),
				Filename:    input.FileName,
				Size:        int64(len(payload)),
				ContentType: mimeType,
			},
		},
	)
	if err != nil {
		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			return nil, types.UploadInternalControlEvidenceOutput{}, fmt.Errorf("%s", validationErrors.Error())
		}

		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, types.UploadInternalControlEvidenceOutput{}, fmt.Errorf("resource not found")
		}

		r.logger.ErrorCtx(ctx, "cannot upload internal control evidence", log.Error(err))

		return nil, types.UploadInternalControlEvidenceOutput{}, fmt.Errorf("internal error")
	}

	filesByID, err := r.evidenceFilesByID(ctx, scope, []*coredata.Evidence{evidence})
	if err != nil {
		r.logger.ErrorCtx(ctx, "cannot load uploaded evidence file", log.Error(err))

		return nil, types.UploadInternalControlEvidenceOutput{}, fmt.Errorf("internal error")
	}

	var file *types.File
	if evidence.EvidenceFileId != nil {
		file = filesByID[*evidence.EvidenceFileId]
	}

	return nil, types.UploadInternalControlEvidenceOutput{
		Evidence: types.NewEvidence(evidence, file),
	}, nil
}
