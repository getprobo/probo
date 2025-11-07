// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type TrustCenter struct {
	ID             gid.GID                         `json:"id"`
	OrganizationID gid.GID                         `json:"-"`
	Active         bool                            `json:"active"`
	NdaFileName    *string                         `json:"ndaFileName,omitempty"`
	NdaFileURL     *string                         `json:"ndaFileUrl,omitempty"`
	CreatedAt      time.Time                       `json:"createdAt"`
	UpdatedAt      time.Time                       `json:"updatedAt"`
	Organization   *Organization                   `json:"organization"`
	Accesses       *TrustCenterAccessConnection    `json:"accesses"`
	References     *TrustCenterReferenceConnection `json:"references"`
}

func (TrustCenter) IsNode()             {}
func (this TrustCenter) GetID() gid.GID { return this.ID }

func NewTrustCenter(tc *coredata.TrustCenter, file *coredata.File) *TrustCenter {
	var ndaFileName *string
	if file != nil {
		ndaFileName = &file.FileName
	}

	return &TrustCenter{
		ID:             tc.ID,
		OrganizationID: tc.OrganizationID,
		Active:         tc.Active,
		NdaFileName:    ndaFileName,
		CreatedAt:      tc.CreatedAt,
		UpdatedAt:      tc.UpdatedAt,
	}
}
