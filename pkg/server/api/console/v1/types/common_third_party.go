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
