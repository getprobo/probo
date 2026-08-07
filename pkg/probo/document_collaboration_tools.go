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

package probo

import (
	"context"
	"encoding/base64"
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	readLiveDocumentParams struct{}

	readLiveDocumentResult struct {
		Text     string `json:"text"`
		Revision int64  `json:"revision"`
	}

	editLiveDocumentParams struct {
		ExpectedRevision int64  `json:"expected_revision" jsonschema:"Revision returned by read_live_document. The edit is rejected if the document changed."`
		Index            uint32 `json:"index" jsonschema:"UTF-16 sequence position at which to apply the edit."`
		Cursor           string `json:"cursor,omitempty" jsonschema:"Optional base64 stable Automerge cursor. When provided it takes precedence over index."`
		DeleteCount      int32  `json:"delete_count" jsonschema:"Number of UTF-16 sequence units to delete."`
		Text             string `json:"text" jsonschema:"Text to insert at the edit position."`
	}

	editLiveDocumentResult struct {
		Revision int64 `json:"revision"`
	}
)

func (s *DocumentService) CollaborationTools(
	scope coredata.Scoper,
	documentVersionID gid.GID,
) []agent.Tool {
	return []agent.Tool{
		agent.FunctionTool(
			"read_live_document",
			"Read the current collaborative document text and revision before proposing an edit.",
			func(ctx context.Context, _ readLiveDocumentParams) (agent.ToolResult, error) {
				text, revision, err := s.ReadCollaborationText(
					ctx,
					scope,
					documentVersionID,
				)
				if err != nil {
					return agent.ToolResult{}, fmt.Errorf("cannot read live document: %w", err)
				}

				return agent.ResultJSON(
					readLiveDocumentResult{
						Text:     text,
						Revision: revision,
					},
				), nil
			},
		),
		agent.FunctionTool(
			"edit_live_document_text",
			"Edit the collaborative document using the revision returned by read_live_document. Re-read and retry if the document changed concurrently.",
			func(ctx context.Context, params editLiveDocumentParams) (agent.ToolResult, error) {
				var cursor automerge.Cursor
				if params.Cursor != "" {
					decoded, err := base64.StdEncoding.DecodeString(params.Cursor)
					if err != nil {
						return agent.ToolResult{}, fmt.Errorf("cannot decode live document cursor: %w", err)
					}
					cursor = automerge.Cursor(decoded)
				}

				revision, err := s.ApplyCollaborationTextEdit(
					ctx,
					scope,
					documentVersionID,
					DocumentCollaborationTextEdit{
						ExpectedRevision: params.ExpectedRevision,
						Index:            params.Index,
						Cursor:           cursor,
						DeleteCount:      params.DeleteCount,
						Text:             params.Text,
					},
				)
				if err != nil {
					return agent.ToolResult{}, fmt.Errorf("cannot edit live document: %w", err)
				}

				return agent.ResultJSON(editLiveDocumentResult{Revision: revision}), nil
			},
		),
	}
}
