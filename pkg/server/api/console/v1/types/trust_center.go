// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	ID                   gid.GID                          `json:"id"`
	Active               bool                             `json:"active"`
	SearchEngineIndexing coredata.SearchEngineIndexing    `json:"searchEngineIndexing"`
	Description          *string                          `json:"description,omitempty"`
	WebsiteURL           *string                          `json:"websiteUrl,omitempty"`
	Email                *string                          `json:"email,omitempty"`
	HeadquarterAddress   *string                          `json:"headquarterAddress,omitempty"`
	CustomDomain         *CustomDomain                    `json:"customDomain,omitempty"`
	Logo                 *File                            `json:"logo,omitempty"`
	DarkLogo             *File                            `json:"darkLogo,omitempty"`
	Nda                  *File                            `json:"nda,omitempty"`
	CreatedAt            time.Time                        `json:"createdAt"`
	UpdatedAt            time.Time                        `json:"updatedAt"`
	Organization         *Organization                    `json:"organization"`
	Accesses             *TrustCenterAccessConnection     `json:"accesses"`
	References           *TrustCenterReferenceConnection  `json:"references"`
	ComplianceFrameworks *ComplianceFrameworkConnection   `json:"complianceFrameworks"`
	ExternalUrls         *ComplianceExternalURLConnection `json:"externalUrls"`
	MailingList          *MailingList                     `json:"mailingList,omitempty"`
	Permission           bool                             `json:"permission"`
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

	if tc.CustomDomainID != nil {
		trustCenter.CustomDomain = &CustomDomain{ID: *tc.CustomDomainID}
	}

	return trustCenter
}
