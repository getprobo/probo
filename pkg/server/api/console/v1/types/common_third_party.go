// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/gid"
)

type CommonThirdParty struct {
	ID                         gid.GID                     `json:"id"`
	Name                       string                      `json:"name"`
	Category                   coredata.ThirdPartyCategory `json:"category"`
	WebsiteURL                 *string                     `json:"websiteUrl,omitempty"`
	HeadquarterAddress         *string                     `json:"headquarterAddress,omitempty"`
	LegalName                  *string                     `json:"legalName,omitempty"`
	PrivacyPolicyURL           *string                     `json:"privacyPolicyUrl,omitempty"`
	ServiceLevelAgreementURL   *string                     `json:"serviceLevelAgreementUrl,omitempty"`
	DataProcessingAgreementURL *string                     `json:"dataProcessingAgreementUrl,omitempty"`
	Certifications             []string                    `json:"certifications"`
	SecurityPageURL            *string                     `json:"securityPageUrl,omitempty"`
	TrustPageURL               *string                     `json:"trustPageUrl,omitempty"`
	StatusPageURL              *string                     `json:"statusPageUrl,omitempty"`
	TermsOfServiceURL          *string                     `json:"termsOfServiceUrl,omitempty"`
	Logo                       *File                       `json:"logo,omitempty"`
}

func (CommonThirdParty) IsTrackerPatternThirdPartyLink() {}

func NewCommonThirdParty(c *coredata.CommonThirdParty) *CommonThirdParty {
	party := &CommonThirdParty{
		ID:                         c.ID,
		Name:                       c.Name,
		Category:                   c.Category,
		WebsiteURL:                 c.WebsiteURL,
		HeadquarterAddress:         c.HeadquarterAddress,
		LegalName:                  c.LegalName,
		PrivacyPolicyURL:           c.PrivacyPolicyURL,
		ServiceLevelAgreementURL:   c.ServiceLevelAgreementURL,
		DataProcessingAgreementURL: c.DataProcessingAgreementURL,
		Certifications:             c.Certifications,
		SecurityPageURL:            c.SecurityPageURL,
		TrustPageURL:               c.TrustPageURL,
		StatusPageURL:              c.StatusPageURL,
		TermsOfServiceURL:          c.TermsOfServiceURL,
	}

	if c.LogoFileID != nil {
		party.Logo = &File{ID: *c.LogoFileID}
	}

	return party
}
