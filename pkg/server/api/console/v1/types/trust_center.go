// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

type TrustCenter struct {
	ID                   gid.GID                         `json:"id"`
	Active               bool                            `json:"active"`
	SearchEngineIndexing coredata.SearchEngineIndexing   `json:"searchEngineIndexing"`
	Logo                 *File                           `json:"logo,omitempty"`
	DarkLogo             *File                           `json:"darkLogo,omitempty"`
	Nda                  *File                           `json:"nda,omitempty"`
	Description          *string                         `json:"description,omitempty"`
	WebsiteURL           *string                         `json:"websiteUrl,omitempty"`
	Email                *string                         `json:"email,omitempty"`
	HeadquarterAddress   *string                         `json:"headquarterAddress,omitempty"`
	CreatedAt            time.Time                       `json:"createdAt"`
	UpdatedAt            time.Time                       `json:"updatedAt"`
	Organization         *Organization                   `json:"organization"`
	Accesses             *TrustCenterAccessConnection    `json:"accesses"`
	References           *TrustCenterReferenceConnection `json:"references"`
	ComplianceFrameworks *ComplianceFrameworkConnection  `json:"complianceFrameworks"`
	CustomLinks          *ComplianceCustomLinkConnection `json:"customLinks"`
	MailingList          *MailingList                    `json:"mailingList,omitempty"`
	Permission           bool                            `json:"permission"`
}

func (TrustCenter) IsNode()          {}
func (t TrustCenter) GetID() gid.GID { return t.ID }

func NewTrustCenter(tc *coredata.TrustCenter) *TrustCenter {
	trustCenter := &TrustCenter{
		ID: tc.ID,
		Organization: &Organization{
			ID: tc.OrganizationID,
		},
		Active:               tc.Active,
		SearchEngineIndexing: tc.SearchEngineIndexing,
		Description:          tc.Description,
		WebsiteURL:           tc.WebsiteURL,
		Email:                tc.Email,
		HeadquarterAddress:   tc.HeadquarterAddress,
		CreatedAt:            tc.CreatedAt,
		UpdatedAt:            tc.UpdatedAt,
	}

	if tc.LogoFileID != nil {
		trustCenter.Logo = &File{ID: *tc.LogoFileID}
	}

	if tc.DarkLogoFileID != nil {
		trustCenter.DarkLogo = &File{ID: *tc.DarkLogoFileID}
	}

	if tc.NonDisclosureAgreementFileID != nil {
		trustCenter.Nda = &File{ID: *tc.NonDisclosureAgreementFileID}
	}

	return trustCenter
}
