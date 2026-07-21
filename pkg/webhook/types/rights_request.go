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

package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type RightsRequest struct {
	ID             gid.GID                     `json:"id"`
	OrganizationID gid.GID                     `json:"organizationId"`
	RequestType    coredata.RightsRequestType  `json:"requestType"`
	RequestState   coredata.RightsRequestState `json:"requestState"`
	DataSubject    *string                     `json:"dataSubject"`
	Contact        *string                     `json:"contact"`
	Details        *string                     `json:"details"`
	Deadline       *time.Time                  `json:"deadline"`
	ActionTaken    *string                     `json:"actionTaken"`
	CreatedAt      time.Time                   `json:"createdAt"`
	UpdatedAt      time.Time                   `json:"updatedAt"`
}

func NewRightsRequest(r *coredata.RightsRequest) *RightsRequest {
	return &RightsRequest{
		ID:             r.ID,
		OrganizationID: r.OrganizationID,
		RequestType:    r.RequestType,
		RequestState:   r.RequestState,
		DataSubject:    r.DataSubject,
		Contact:        r.Contact,
		Details:        r.Details,
		Deadline:       r.Deadline,
		ActionTaken:    r.ActionTaken,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}
