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

type Vendor struct {
	ID                            gid.GID                 `json:"id"`
	Name                          string                  `json:"name"`
	Category                      coredata.VendorCategory `json:"category"`
	Description                   *string                 `json:"description"`
	StatusPageURL                 *string                 `json:"statusPageUrl"`
	TermsOfServiceURL             *string                 `json:"termsOfServiceUrl"`
	PrivacyPolicyURL              *string                 `json:"privacyPolicyUrl"`
	ServiceLevelAgreementURL      *string                 `json:"serviceLevelAgreementUrl"`
	DataProcessingAgreementURL    *string                 `json:"dataProcessingAgreementUrl"`
	BusinessAssociateAgreementURL *string                 `json:"businessAssociateAgreementUrl"`
	SubprocessorsListURL          *string                 `json:"subprocessorsListUrl"`
	Certifications                []string                `json:"certifications"`
	Countries                     []coredata.CountryCode  `json:"countries"`
	SecurityPageURL               *string                 `json:"securityPageUrl"`
	TrustPageURL                  *string                 `json:"trustPageUrl"`
	HeadquarterAddress            *string                 `json:"headquarterAddress"`
	LegalName                     *string                 `json:"legalName"`
	WebsiteURL                    *string                 `json:"websiteUrl"`
	BusinessOwnerID               *gid.GID                `json:"businessOwnerId"`
	SecurityOwnerID               *gid.GID                `json:"securityOwnerId"`
	CreatedAt                     time.Time               `json:"createdAt"`
	UpdatedAt                     time.Time               `json:"updatedAt"`
}

func NewVendor(v *coredata.Vendor) *Vendor {
	return &Vendor{
		ID:                            v.ID,
		Name:                          v.Name,
		Category:                      v.Category,
		Description:                   v.Description,
		StatusPageURL:                 v.StatusPageURL,
		TermsOfServiceURL:             v.TermsOfServiceURL,
		PrivacyPolicyURL:              v.PrivacyPolicyURL,
		ServiceLevelAgreementURL:      v.ServiceLevelAgreementURL,
		DataProcessingAgreementURL:    v.DataProcessingAgreementURL,
		BusinessAssociateAgreementURL: v.BusinessAssociateAgreementURL,
		SubprocessorsListURL:          v.SubprocessorsListURL,
		Certifications:                v.Certifications,
		Countries:                     v.Countries,
		SecurityPageURL:               v.SecurityPageURL,
		TrustPageURL:                  v.TrustPageURL,
		HeadquarterAddress:            v.HeadquarterAddress,
		LegalName:                     v.LegalName,
		WebsiteURL:                    v.WebsiteURL,
		BusinessOwnerID:               v.BusinessOwnerID,
		SecurityOwnerID:               v.SecurityOwnerID,
		CreatedAt:                     v.CreatedAt,
		UpdatedAt:                     v.UpdatedAt,
	}
}
