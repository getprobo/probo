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
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
	"go.probo.inc/probo/pkg/validator"
)

func (r *Resolver) uploadFile(
	ctx context.Context,
	input *types.UploadFileInput,
) (*mcp.CallToolResult, types.UploadFileOutput, error) {
	identity := authn.IdentityFromContext(ctx)
	if identity == nil {
		return nil, types.UploadFileOutput{}, fmt.Errorf("missing identity")
	}

	scope, err := r.Authorize(ctx, input.OrganizationID, probo.ActionFileUpload)
	if err != nil {
		return nil, types.UploadFileOutput{}, err
	}

	payload, err := decodeFileContent(input.ContentBase64, probo.AttachmentMaxFileSize())
	if err != nil {
		return nil, types.UploadFileOutput{}, err
	}

	mimeType := ""
	if input.MimeType != nil {
		mimeType = strings.TrimSpace(*input.MimeType)
	}

	if mimeType == "" {
		inferred, ok := probo.ContentTypeForAttachment(input.FileName)
		if !ok {
			return nil, types.UploadFileOutput{}, fmt.Errorf("cannot infer mime_type for %q: file extension is missing or not supported", input.FileName)
		}

		mimeType = inferred
	}

	file, err := r.proboSvc.Files.UploadFile(
		ctx,
		scope,
		probo.AttachmentUpload{
			OrganizationID: input.OrganizationID,
			File: &probo.FileUpload{
				Content:     bytes.NewReader(payload),
				Filename:    input.FileName,
				Size:        int64(len(payload)),
				ContentType: mimeType,
			},
		},
	)
	if err != nil {
		if validationErrors, ok := errors.AsType[validator.ValidationErrors](err); ok {
			return nil, types.UploadFileOutput{}, fmt.Errorf("%s", validationErrors.Error())
		}

		r.logger.ErrorCtx(ctx, "cannot upload file", log.Error(err))

		return nil, types.UploadFileOutput{}, fmt.Errorf("internal error")
	}

	return nil, types.UploadFileOutput{
		ID:       file.ID,
		FileName: file.FileName,
		MimeType: file.MimeType,
		Size:     int(file.FileSize),
	}, nil
}

// decodeFileContent reads at most maxSize+1 decoded bytes. UploadFile
// checks the size only after the caller already holds the bytes, so a full
// DecodeString would allocate the payload before that check can reject it.
func decodeFileContent(encoded string, maxSize int64) ([]byte, error) {
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encoded))

	payload, err := io.ReadAll(io.LimitReader(decoder, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("content_base64 is not valid base64")
	}

	if int64(len(payload)) > maxSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", maxSize)
	}

	return payload, nil
}
