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

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	ApplyCollaborationEditsRequest struct {
		DocumentID gid.GID
		BaseHeads  []automerge.Hash
		Edits      []DocumentCollaborationEdit
		IdentityID gid.GID
		ProfileID  *gid.GID
	}

	ApplyCollaborationContentRequest struct {
		DocumentID gid.GID
		Content    string
		IdentityID gid.GID
		ProfileID  *gid.GID
	}
)

func (s *DocumentService) ApplyCollaborationContent(
	ctx context.Context,
	scope coredata.Scoper,
	req ApplyCollaborationContentRequest,
) (*CollaborationWriteResult, error) {
	return s.collaboration.ApplyContent(
		ctx,
		scope,
		CollaborationApplyContentRequest{
			DocumentID: req.DocumentID,
			Content:    req.Content,
			IdentityID: req.IdentityID,
			ProfileID:  req.ProfileID,
		},
	)
}

func (s *DocumentService) ApplyCollaborationEdits(
	ctx context.Context,
	scope coredata.Scoper,
	req ApplyCollaborationEditsRequest,
) (*CollaborationWriteResult, error) {
	return s.collaboration.ApplyEdits(
		ctx,
		scope,
		CollaborationApplyEditsRequest{
			DocumentID: req.DocumentID,
			BaseHeads:  req.BaseHeads,
			Edits:      req.Edits,
			IdentityID: req.IdentityID,
			ProfileID:  req.ProfileID,
		},
	)
}
