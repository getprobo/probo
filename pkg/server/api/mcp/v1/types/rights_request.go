// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewRightsRequest(rr *coredata.RightsRequest) *RightsRequest {
	rightsRequest := &RightsRequest{
		ID:             rr.ID,
		OrganizationID: rr.OrganizationID,
		RequestType:    rr.RequestType,
		RequestState:   rr.RequestState,
		Contact:        rr.Contact,
		Details:        rr.Details,
		Deadline:       rr.Deadline,
		ActionTaken:    rr.ActionTaken,
		CreatedAt:      rr.CreatedAt,
		UpdatedAt:      rr.UpdatedAt,
	}

	if rr.DataSubject != nil {
		rightsRequest.DataSubject = *rr.DataSubject
	}

	return rightsRequest
}

func NewListRightsRequestsOutput(rightsRequestPage *page.Page[*coredata.RightsRequest, coredata.RightsRequestOrderField]) ListRightsRequestsOutput {
	rightsRequests := make([]*RightsRequest, 0, len(rightsRequestPage.Data))
	for _, v := range rightsRequestPage.Data {
		rightsRequests = append(rightsRequests, NewRightsRequest(v))
	}

	var nextCursor *page.CursorKey
	if len(rightsRequestPage.Data) > 0 {
		cursorKey := rightsRequestPage.Data[len(rightsRequestPage.Data)-1].CursorKey(rightsRequestPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRightsRequestsOutput{
		NextCursor:     nextCursor,
		RightsRequests: rightsRequests,
	}
}
