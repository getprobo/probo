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

type ThirdParty struct {
	ID                            gid.GID                     `json:"id"`
	Name                          string                      `json:"name"`
	Category                      coredata.ThirdPartyCategory `json:"category"`
	Description                   *string                     `json:"description"`
	StatusPageURL                 *string                     `json:"statusPageUrl"`
	TermsOfServiceURL             *string                     `json:"termsOfServiceUrl"`
	PrivacyPolicyURL              *string                     `json:"privacyPolicyUrl"`
	ServiceLevelAgreementURL      *string                     `json:"serviceLevelAgreementUrl"`
	DataProcessingAgreementURL    *string                     `json:"dataProcessingAgreementUrl"`
	BusinessAssociateAgreementURL *string                     `json:"businessAssociateAgreementUrl"`
	SubprocessorsListURL          *string                     `json:"subprocessorsListUrl"`
	Certifications                []string                    `json:"certifications"`
	Countries                     []coredata.CountryCode      `json:"countries"`
	SecurityPageURL               *string                     `json:"securityPageUrl"`
	TrustPageURL                  *string                     `json:"trustPageUrl"`
	HeadquarterAddress            *string                     `json:"headquarterAddress"`
	LegalName                     *string                     `json:"legalName"`
	WebsiteURL                    *string                     `json:"websiteUrl"`
	BusinessOwnerID               *gid.GID                    `json:"businessOwnerId"`
	SecurityOwnerID               *gid.GID                    `json:"securityOwnerId"`
	CreatedAt                     time.Time                   `json:"createdAt"`
	UpdatedAt                     time.Time                   `json:"updatedAt"`
}

func NewThirdParty(v *coredata.ThirdParty) *ThirdParty {
	return &ThirdParty{
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
