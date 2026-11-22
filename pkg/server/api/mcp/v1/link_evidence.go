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
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
	"go.probo.inc/probo/pkg/validator"
)

// internalControlLinkEvidenceUpload builds the same text/uri-list file the
// console link form uploads. The console GraphQL API stays a file upload.
func internalControlLinkEvidenceUpload(input *types.AddInternalControlLinkEvidenceInput) (probo.UploadInternalControlEvidenceRequest, error) {
	name := strings.TrimSpace(input.Name)
	rawURL := strings.TrimSpace(input.URL)

	v := validator.New()
	v.Check(name, "name", validator.Required(), validator.SafeTextNoNewLine(probo.TitleMaxLength))
	v.Check(rawURL, "url", validator.Required(), validator.URL(), validator.MaxLen(probo.ContentMaxLength))

	if err := v.Error(); err != nil {
		return probo.UploadInternalControlEvidenceRequest{}, fmt.Errorf("%s", err)
	}

	content := []byte(rawURL)

	return probo.UploadInternalControlEvidenceRequest{
		InternalControlID: input.InternalControlID,
		File: probo.FileUpload{
			Content:     bytes.NewReader(content),
			Filename:    name + ".uri",
			Size:        int64(len(content)),
			ContentType: "text/uri-list",
		},
	}, nil
}
