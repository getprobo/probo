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
	"go.probo.inc/probo/pkg/probo"
)

func NewVendorRiskAssessment(v *coredata.VendorRiskAssessment) *VendorRiskAssessment {
	return &VendorRiskAssessment{
		ID:              v.ID,
		OrganizationID:  v.OrganizationID,
		VendorID:        v.VendorID,
		ExpiresAt:       v.ExpiresAt,
		DataSensitivity: v.DataSensitivity,
		BusinessImpact:  v.BusinessImpact,
		Notes:           v.Notes,
		SnapshotID:      v.SnapshotID,
		CreatedAt:       v.CreatedAt,
		UpdatedAt:       v.UpdatedAt,
	}
}

func NewListVendorRiskAssessmentsOutput(p *page.Page[*coredata.VendorRiskAssessment, coredata.VendorRiskAssessmentOrderField]) ListVendorRiskAssessmentsOutput {
	assessments := make([]*VendorRiskAssessment, 0, len(p.Data))
	for _, v := range p.Data {
		assessments = append(assessments, NewVendorRiskAssessment(v))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListVendorRiskAssessmentsOutput{
		NextCursor:            nextCursor,
		VendorRiskAssessments: assessments,
	}
}

func NewAddVendorRiskAssessmentOutput(v *coredata.VendorRiskAssessment) AddVendorRiskAssessmentOutput {
	return AddVendorRiskAssessmentOutput{
		VendorRiskAssessment: NewVendorRiskAssessment(v),
	}
}

func NewVendor(v *coredata.Vendor) *Vendor {
	countries := make([]string, len(v.Countries))
	for i, c := range v.Countries {
		countries[i] = string(c)
	}

	return &Vendor{
		ID:                            v.ID,
		OrganizationID:                v.OrganizationID,
		Name:                          v.Name,
		Description:                   v.Description,
		Category:                      VendorCategory(v.Category),
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
		CreatedAt:                     v.CreatedAt,
		UpdatedAt:                     v.UpdatedAt,
	}
}

func NewListVendorsOutput(vendorPage *page.Page[*coredata.Vendor, coredata.VendorOrderField]) ListVendorsOutput {
	vendors := make([]*Vendor, 0, len(vendorPage.Data))
	for _, v := range vendorPage.Data {
		vendors = append(vendors, NewVendor(v))
	}

	var nextCursor *page.CursorKey
	if len(vendorPage.Data) > 0 {
		cursorKey := vendorPage.Data[len(vendorPage.Data)-1].CursorKey(vendorPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListVendorsOutput{
		NextCursor: nextCursor,
		Vendors:    vendors,
	}
}

func NewAddVendorOutput(v *coredata.Vendor) AddVendorOutput {
	return AddVendorOutput{
		Vendor: NewVendor(v),
	}
}

func NewUpdateVendorOutput(v *coredata.Vendor) UpdateVendorOutput {
	return UpdateVendorOutput{
		Vendor: NewVendor(v),
	}
}

func NewVendorContact(vc *coredata.VendorContact) *VendorContact {
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

	return &VendorContact{
		ID:        vc.ID,
		VendorID:  vc.VendorID,
		FullName:  fullName,
		Email:     email,
		Phone:     phone,
		Role:      role,
		CreatedAt: vc.CreatedAt,
		UpdatedAt: vc.UpdatedAt,
	}
}

func NewListVendorContactsOutput(p *page.Page[*coredata.VendorContact, coredata.VendorContactOrderField]) ListVendorContactsOutput {
	contacts := make([]*VendorContact, 0, len(p.Data))
	for _, vc := range p.Data {
		contacts = append(contacts, NewVendorContact(vc))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListVendorContactsOutput{
		NextCursor:     nextCursor,
		VendorContacts: contacts,
	}
}

func NewVendorService(vs *coredata.VendorService) *VendorService {
	var description string
	if vs.Description != nil {
		description = *vs.Description
	}

	return &VendorService{
		ID:          vs.ID,
		VendorID:    vs.VendorID,
		Name:        vs.Name,
		Description: description,
		CreatedAt:   vs.CreatedAt,
		UpdatedAt:   vs.UpdatedAt,
	}
}

func NewListVendorServicesOutput(p *page.Page[*coredata.VendorService, coredata.VendorServiceOrderField]) ListVendorServicesOutput {
	services := make([]*VendorService, 0, len(p.Data))
	for _, vs := range p.Data {
		services = append(services, NewVendorService(vs))
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		cursorKey := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListVendorServicesOutput{
		NextCursor:     nextCursor,
		VendorServices: services,
	}
}

func NewVendorSubprocessors(sps []probo.Subprocessor) []*VendorSubprocessor {
	result := make([]*VendorSubprocessor, len(sps))
	for i, sp := range sps {
		result[i] = &VendorSubprocessor{
			Name:    sp.Name,
			Country: sp.Country,
			Purpose: sp.Purpose,
		}
	}
	return result
}

func NewAssessVendorOutput(result *probo.AssessVendorResult) AssessVendorOutput {
	return AssessVendorOutput{
		Vendor:        NewVendor(result.Vendor),
		Report:        result.Report,
		Subprocessors: NewVendorSubprocessors(result.Subprocessors),
	}
}
