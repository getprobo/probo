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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewThirdPartyRiskAssessment(v *coredata.ThirdPartyRiskAssessment) *ThirdPartyRiskAssessment {
	return &ThirdPartyRiskAssessment{
		ID:              v.ID,
		OrganizationID:  v.OrganizationID,
		ThirdPartyID:    v.ThirdPartyID,
		ExpiresAt:       v.ExpiresAt,
		DataSensitivity: v.DataSensitivity,
		BusinessImpact:  v.BusinessImpact,
		Notes:           v.Notes,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

func NewListThirdPartyRiskAssessmentsOutput(p *page.Page[*coredata.ThirdPartyRiskAssessment, coredata.ThirdPartyRiskAssessmentOrderField]) ListThirdPartyRiskAssessmentsOutput {
	assessments := make([]*ThirdPartyRiskAssessment, 0, len(p.Data))
	for _, v := range p.Data {
		assessments = append(assessments, NewThirdPartyRiskAssessment(v))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListThirdPartyRiskAssessmentsOutput{
		NextCursor:                nextCursor,
		ThirdPartyRiskAssessments: assessments,
	}
}

func NewAddThirdPartyRiskAssessmentOutput(v *coredata.ThirdPartyRiskAssessment) AddThirdPartyRiskAssessmentOutput {
	return AddThirdPartyRiskAssessmentOutput{
		ThirdPartyRiskAssessment: NewThirdPartyRiskAssessment(v),
	}
}

func NewThirdParty(v *coredata.ThirdParty) *ThirdParty {
	countries := make([]string, len(v.Countries))
	for i, c := range v.Countries {
		countries[i] = string(c)
	}

	return &ThirdParty{
		ID:                            v.ID,
		OrganizationID:                v.OrganizationID,
		Name:                          v.Name,
		Description:                   v.Description,
		Category:                      ThirdPartyCategory(v.Category),
		HeadquarterAddress:            v.HeadquarterAddress,
		LegalName:                     v.LegalName,
		WebsiteURL:                    v.WebsiteURL,
		PrivacyPolicyURL:              v.PrivacyPolicyURL,
		ServiceLevelAgreementURL:      v.ServiceLevelAgreementURL,
		DataProcessingAgreementURL:    v.DataProcessingAgreementURL,
		BusinessAssociateAgreementURL: v.BusinessAssociateAgreementURL,
		SubprocessorsListURL:          v.SubprocessorsListURL,
		Certifications:                v.Certifications,
		Countries:                     countries,
		BusinessOwnerID:               v.BusinessOwnerID,
		SecurityOwnerID:               v.SecurityOwnerID,
		StatusPageURL:                 v.StatusPageURL,
		TermsOfServiceURL:             v.TermsOfServiceURL,
		SecurityPageURL:               v.SecurityPageURL,
		TrustPageURL:                  v.TrustPageURL,
		Level:                         v.Level,
		CreatedAt:                     v.CreatedAt,
		UpdatedAt:                     v.UpdatedAt,
	}
}

func NewListThirdPartiesOutput(thirdPartyPage *page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField]) ListThirdPartiesOutput {
	thirdParties := make([]*ThirdParty, 0, len(thirdPartyPage.Data))
	for _, v := range thirdPartyPage.Data {
		thirdParties = append(thirdParties, NewThirdParty(v))
	}

	var nextCursor *page.CursorKey

	if len(thirdPartyPage.Data) > 0 {
		cursorKey := thirdPartyPage.Data[len(thirdPartyPage.Data)-1].CursorKey(thirdPartyPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListThirdPartiesOutput{
		NextCursor:   nextCursor,
		ThirdParties: thirdParties,
	}
}

func NewListChildThirdPartiesOutput(thirdPartyPage *page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField]) ListChildThirdPartiesOutput {
	thirdParties := make([]*ThirdParty, 0, len(thirdPartyPage.Data))
	for _, v := range thirdPartyPage.Data {
		thirdParties = append(thirdParties, NewThirdParty(v))
	}

	var nextCursor *page.CursorKey

	if len(thirdPartyPage.Data) > 0 {
		cursorKey := thirdPartyPage.Data[len(thirdPartyPage.Data)-1].CursorKey(thirdPartyPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListChildThirdPartiesOutput{
		NextCursor:   nextCursor,
		ThirdParties: thirdParties,
	}
}

func NewAddThirdPartyOutput(v *coredata.ThirdParty) AddThirdPartyOutput {
	return AddThirdPartyOutput{
		ThirdParty: NewThirdParty(v),
	}
}

func NewUpdateThirdPartyOutput(v *coredata.ThirdParty) UpdateThirdPartyOutput {
	return UpdateThirdPartyOutput{
		ThirdParty: NewThirdParty(v),
	}
}

func NewThirdPartyContact(vc *coredata.ThirdPartyContact) *ThirdPartyContact {
	var fullName string
	if vc.FullName != nil {
		fullName = *vc.FullName
	}

	var email string
	if vc.Email != nil {
		email = vc.Email.String()
	}

	var phone string
	if vc.Phone != nil {
		phone = *vc.Phone
	}

	var role string
	if vc.Role != nil {
		role = *vc.Role
	}

	return &ThirdPartyContact{
		ID:           vc.ID,
		ThirdPartyID: vc.ThirdPartyID,
		FullName:     fullName,
		Email:        email,
		Phone:        phone,
		Role:         role,
		CreatedAt:    vc.CreatedAt,
		UpdatedAt:    vc.UpdatedAt,
	}
}

func NewListThirdPartyContactsOutput(p *page.Page[*coredata.ThirdPartyContact, coredata.ThirdPartyContactOrderField]) ListThirdPartyContactsOutput {
	contacts := make([]*ThirdPartyContact, 0, len(p.Data))
	for _, vc := range p.Data {
		contacts = append(contacts, NewThirdPartyContact(vc))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListThirdPartyContactsOutput{
		NextCursor:         nextCursor,
		ThirdPartyContacts: contacts,
	}
}

func NewThirdPartyService(vs *coredata.ThirdPartyService) *ThirdPartyService {
	var description string
	if vs.Description != nil {
		description = *vs.Description
	}

	return &ThirdPartyService{
		ID:           vs.ID,
		ThirdPartyID: vs.ThirdPartyID,
		Name:         vs.Name,
		Description:  description,
		CreatedAt:    vs.CreatedAt,
		UpdatedAt:    vs.UpdatedAt,
	}
}

func NewListThirdPartyServicesOutput(p *page.Page[*coredata.ThirdPartyService, coredata.ThirdPartyServiceOrderField]) ListThirdPartyServicesOutput {
	services := make([]*ThirdPartyService, 0, len(p.Data))
	for _, vs := range p.Data {
		services = append(services, NewThirdPartyService(vs))
	}

	var nextCursor *page.CursorKey

	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListThirdPartyServicesOutput{
		NextCursor:         nextCursor,
		ThirdPartyServices: services,
	}
}
