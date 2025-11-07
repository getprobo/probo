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

	"github.com/google/jsonschema-go/jsonschema"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	VendorOrderBy OrderBy[coredata.VendorOrderField]

	VendorFilter struct {
		SnapshotID *gid.GID `json:"snapshot_id"`
	}

	ListVendorsInput struct {
		OrganizationID gid.GID         `json:"organization_id"`
		Filter         *VendorFilter   `json:"filter"`
		OrderBy        *VendorOrderBy  `json:"order_field"`
		Cursor         *page.CursorKey `json:"cursor"`
		Size           *int            `json:"size"`
	}

	ListVendorsOutput struct {
		NextCursor *string  `json:"next_cursor"`
		Vendors    []Vendor `json:"vendors"`
	}

	AddVendorInput struct {
		OrganizationID                gid.GID                  `json:"organization_id"`
		Name                          string                   `json:"name"`
		Description                   *string                  `json:"description"`
		HeadquarterAddress            *string                  `json:"headquarter_address"`
		LegalName                     *string                  `json:"legal_name"`
		WebsiteURL                    *string                  `json:"website_url"`
		Category                      *coredata.VendorCategory `json:"category"`
		PrivacyPolicyURL              *string                  `json:"privacy_policy_url"`
		ServiceLevelAgreementURL      *string                  `json:"service_level_agreement_url"`
		DataProcessingAgreementURL    *string                  `json:"data_processing_agreement_url"`
		BusinessAssociateAgreementURL *string                  `json:"business_associate_agreement_url"`
		SubprocessorsListURL          *string                  `json:"subprocessors_list_url"`
		Certifications                []string                 `json:"certifications"`
		Countries                     []coredata.CountryCode   `json:"countries"`
		SecurityPageURL               *string                  `json:"security_page_url"`
		TrustPageURL                  *string                  `json:"trust_page_url"`
		TermsOfServiceURL             *string                  `json:"terms_of_service_url"`
		StatusPageURL                 *string                  `json:"status_page_url"`
		BusinessOwnerID               *gid.GID                 `json:"business_owner_id"`
		SecurityOwnerID               *gid.GID                 `json:"security_owner_id"`
	}

	AddVendorOutput struct {
		Vendor Vendor `json:"vendor" jsonschema:"the created vendor"`
	}

	Vendor struct {
		ID                            gid.GID                 `json:"id"`
		OrganizationID                gid.GID                 `json:"organization_id"`
		Name                          string                  `json:"name"`
		Description                   *string                 `json:"description"`
		Category                      coredata.VendorCategory `json:"category"`
		HeadquarterAddress            *string                 `json:"headquarter_address"`
		LegalName                     *string                 `json:"legal_name"`
		WebsiteURL                    *string                 `json:"website_url"`
		PrivacyPolicyURL              *string                 `json:"privacy_policy_url"`
		ServiceLevelAgreementURL      *string                 `json:"service_level_agreement_url"`
		DataProcessingAgreementURL    *string                 `json:"data_processing_agreement_url"`
		BusinessAssociateAgreementURL *string                 `json:"business_associate_agreement_url"`
		SubprocessorsListURL          *string                 `json:"subprocessors_list_url"`
		Certifications                []string                `json:"certifications"`
		Countries                     []coredata.CountryCode  `json:"countries"`
		BusinessOwnerID               *gid.GID                `json:"business_owner_id,omitempty"`
		SecurityOwnerID               *gid.GID                `json:"security_owner_id,omitempty"`
		StatusPageURL                 *string                 `json:"status_page_url,omitempty"`
		TermsOfServiceURL             *string                 `json:"terms_of_service_url,omitempty"`
		SecurityPageURL               *string                 `json:"security_page_url,omitempty"`
		TrustPageURL                  *string                 `json:"trust_page_url,omitempty"`
		ShowOnTrustCenter             bool                    `json:"show_on_trust_center,omitempty"`
		SnapshotID                    *gid.GID                `json:"snapshot_id,omitempty"`
		SourceID                      *gid.GID                `json:"source_id,omitempty"`
		CreatedAt                     time.Time               `json:"created_at"`
		UpdatedAt                     time.Time               `json:"updated_at"`
	}
)

var (
	ListVendorsInputSchema = &jsonschema.Schema{
		Type:     "object",
		Required: []string{"organizationID"},
		Properties: map[string]*jsonschema.Schema{
			"organizationID": {Type: "string"},
			"filter": {
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"snapshotID": {Type: "string"},
				},
			},
			"orderBy": {
				Types: []string{"object", "null"},
				Properties: map[string]*jsonschema.Schema{
					"field":     {Type: "string", Enum: []any{"CREATED_AT"}},
					"direction": OrderByDirectionSchema,
				},
			},
			"cursor": {Types: []string{"string", "null"}},
			"size":   {Types: []string{"integer", "null"}},
		},
	}

	OrderByDirectionSchema = &jsonschema.Schema{
		Type: "string",
		Enum: []any{"ASC", "DESC"},
	}

	NullableStringSchema = &jsonschema.Schema{
		Types: []string{"string", "null"},
	}

	VendorCategorySchema = &jsonschema.Schema{
		Type: "string",
		Enum: []any{
			"ANALYTICS",
			"CLOUD_MONITORING",
			"CLOUD_PROVIDER",
			"COLLABORATION",
			"CUSTOMER_SUPPORT",
			"DATA_STORAGE_AND_PROCESSING",
			"DOCUMENT_MANAGEMENT",
			"EMPLOYEE_MANAGEMENT",
			"ENGINEERING",
			"FINANCE",
			"IDENTITY_PROVIDER",
			"IT",
			"MARKETING",
			"OFFICE_OPERATIONS",
			"OTHER",
			"PASSWORD_MANAGEMENT",
			"PRODUCT_AND_DESIGN",
			"PROFESSIONAL_SERVICES",
			"RECRUITING",
			"SALES",
			"SECURITY",
			"VERSION_CONTROL",
		},
	}

	VendorSchema = &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"id":                               {Type: "string"},
			"name":                             {Type: "string"},
			"organization_id":                  {Type: "string"},
			"description":                      NullableStringSchema,
			"category":                         VendorCategorySchema,
			"headquarter_address":              NullableStringSchema,
			"legalName":                        NullableStringSchema,
			"website_url":                      NullableStringSchema,
			"privacy_policy_url":               NullableStringSchema,
			"service_level_agreement_url":      NullableStringSchema,
			"data_processing_agreement_url":    NullableStringSchema,
			"business_associate_agreement_url": NullableStringSchema,
			"subprocessors_list_url":           NullableStringSchema,
			"certifications":                   {Types: []string{"array", "null"}, Items: &jsonschema.Schema{Type: "string"}},
			"countries":                        {Types: []string{"array", "null"}, Items: &jsonschema.Schema{Type: "string", Enum: []any{"US", "CA", "GB", "DE", "FR", "IT", "ES", "NL", "BE", "CH", "AT", "SE", "NO", "DK", "FI", "EE", "LT", "LV", "PL", "CZ", "SK", "HU", "RO", "BG", "HR", "SI", "ME", "AL", "MK", "BA", "XK", "XA", "XZ"}}},
			"business_owner_id":                NullableStringSchema,
			"security_owner_id":                NullableStringSchema,
			"status_page_url":                  NullableStringSchema,
			"terms_of_service_url":             NullableStringSchema,
			"security_page_url":                NullableStringSchema,
			"trust_page_url":                   NullableStringSchema,
			"show_on_trust_center":             {Type: "boolean"},
			"snapshot_id":                      NullableStringSchema,
			"source_id":                        NullableStringSchema,
			"created_at":                       {Type: "string"},
			"updated_at":                       {Type: "string"},
		},
	}

	ListVendorsOutputSchema = &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"next_cursor": {Types: []string{"string", "null"}},
			"vendors": {
				Type:  "array",
				Items: VendorSchema,
			},
		},
	}

	AddVendorInputSchema = &jsonschema.Schema{
		Type:     "object",
		Required: []string{"organization_id", "name"},
		Properties: map[string]*jsonschema.Schema{
			"organization_id":                  {Type: "string"},
			"name":                             {Type: "string"},
			"description":                      NullableStringSchema,
			"headquarter_address":              NullableStringSchema,
			"legal_name":                       NullableStringSchema,
			"website_url":                      NullableStringSchema,
			"category":                         VendorCategorySchema,
			"privacy_policy_url":               NullableStringSchema,
			"service_level_agreement_url":      NullableStringSchema,
			"data_processing_agreement_url":    NullableStringSchema,
			"business_associate_agreement_url": NullableStringSchema,
			"subprocessors_list_url":           NullableStringSchema,
			"certifications":                   {Types: []string{"array", "null"}, Items: &jsonschema.Schema{Type: "string"}},
			"countries":                        {Types: []string{"array", "null"}, Items: &jsonschema.Schema{Type: "string", Enum: []any{"US", "CA", "GB", "DE", "FR", "IT", "ES", "NL", "BE", "CH", "AT", "SE", "NO", "DK", "FI", "EE", "LT", "LV", "PL", "CZ", "SK", "HU", "RO", "BG", "HR", "SI", "ME", "AL", "MK", "BA", "XK", "XA", "XZ"}}},
			"business_owner_id":                NullableStringSchema,
			"security_owner_id":                NullableStringSchema,
			"status_page_url":                  NullableStringSchema,
			"terms_of_service_url":             NullableStringSchema,
			"security_page_url":                NullableStringSchema,
			"trust_page_url":                   NullableStringSchema,
			"show_on_trust_center":             {Type: "boolean"},
			"snapshot_id":                      NullableStringSchema,
			"source_id":                        NullableStringSchema,
		},
	}

	AddVendorOutputSchema = &jsonschema.Schema{
		Type: "object",
		Properties: map[string]*jsonschema.Schema{
			"vendor": VendorSchema,
		},
	}
)

func NewVendor(v *coredata.Vendor) Vendor {
	return Vendor{
		Name:                          v.Name,
		ID:                            v.ID,
		OrganizationID:                v.OrganizationID,
		Description:                   v.Description,
		Category:                      v.Category,
		HeadquarterAddress:            v.HeadquarterAddress,
		LegalName:                     v.LegalName,
		WebsiteURL:                    v.WebsiteURL,
		PrivacyPolicyURL:              v.PrivacyPolicyURL,
		ServiceLevelAgreementURL:      v.ServiceLevelAgreementURL,
		DataProcessingAgreementURL:    v.DataProcessingAgreementURL,
		BusinessAssociateAgreementURL: v.BusinessAssociateAgreementURL,
		SubprocessorsListURL:          v.SubprocessorsListURL,
		Certifications:                v.Certifications,
		Countries:                     v.Countries,
		BusinessOwnerID:               v.BusinessOwnerID,
		SecurityOwnerID:               v.SecurityOwnerID,
		StatusPageURL:                 v.StatusPageURL,
		TermsOfServiceURL:             v.TermsOfServiceURL,
		SecurityPageURL:               v.SecurityPageURL,
		TrustPageURL:                  v.TrustPageURL,
		ShowOnTrustCenter:             v.ShowOnTrustCenter,
		SnapshotID:                    v.SnapshotID,
		SourceID:                      v.SourceID,
		CreatedAt:                     v.CreatedAt,
		UpdatedAt:                     v.UpdatedAt,
	}
}

func NewListVendorsOutput(vendorPage *page.Page[*coredata.Vendor, coredata.VendorOrderField]) ListVendorsOutput {
	vendors := make([]Vendor, 0, len(vendorPage.Data))
	for _, v := range vendorPage.Data {
		vendors = append(vendors, NewVendor(v))
	}

	var nextCursor *string
	if len(vendorPage.Data) > 0 {
		cursorKey := vendorPage.Data[len(vendorPage.Data)-1].CursorKey(vendorPage.Cursor.OrderBy.Field).String()
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
