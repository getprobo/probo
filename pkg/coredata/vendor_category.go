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

package coredata

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type VendorCategory string

const (
	VendorCategoryAnalytics                VendorCategory = "ANALYTICS"
	VendorCategoryCloudMonitoring          VendorCategory = "CLOUD_MONITORING"
	VendorCategoryCloudProvider            VendorCategory = "CLOUD_PROVIDER"
	VendorCategoryCollaboration            VendorCategory = "COLLABORATION"
	VendorCategoryCustomerSupport          VendorCategory = "CUSTOMER_SUPPORT"
	VendorCategoryDataStorageAndProcessing VendorCategory = "DATA_STORAGE_AND_PROCESSING"
	VendorCategoryDocumentManagement       VendorCategory = "DOCUMENT_MANAGEMENT"
	VendorCategoryEmployeeManagement       VendorCategory = "EMPLOYEE_MANAGEMENT"
	VendorCategoryEngineering              VendorCategory = "ENGINEERING"
	VendorCategoryFinance                  VendorCategory = "FINANCE"
	VendorCategoryIdentityProvider         VendorCategory = "IDENTITY_PROVIDER"
	VendorCategoryIT                       VendorCategory = "IT"
	VendorCategoryMarketing                VendorCategory = "MARKETING"
	VendorCategoryOfficeOperations         VendorCategory = "OFFICE_OPERATIONS"
	VendorCategoryOther                    VendorCategory = "OTHER"
	VendorCategoryPasswordManagement       VendorCategory = "PASSWORD_MANAGEMENT"
	VendorCategoryProductAndDesign         VendorCategory = "PRODUCT_AND_DESIGN"
	VendorCategoryProfessionalServices     VendorCategory = "PROFESSIONAL_SERVICES"
	VendorCategoryRecruiting               VendorCategory = "RECRUITING"
	VendorCategorySales                    VendorCategory = "SALES"
	VendorCategorySecurity                 VendorCategory = "SECURITY"
	VendorCategoryVersionControl           VendorCategory = "VERSION_CONTROL"
)

func VendorCategories() []VendorCategory {
	return []VendorCategory{
		VendorCategoryAnalytics,
		VendorCategoryCloudMonitoring,
		VendorCategoryCloudProvider,
		VendorCategoryCollaboration,
		VendorCategoryCustomerSupport,
		VendorCategoryDataStorageAndProcessing,
		VendorCategoryDocumentManagement,
		VendorCategoryEmployeeManagement,
		VendorCategoryEngineering,
		VendorCategoryFinance,
		VendorCategoryIdentityProvider,
		VendorCategoryIT,
		VendorCategoryMarketing,
		VendorCategoryOfficeOperations,
		VendorCategoryOther,
		VendorCategoryPasswordManagement,
		VendorCategoryProductAndDesign,
		VendorCategoryProfessionalServices,
		VendorCategoryRecruiting,
		VendorCategorySales,
		VendorCategorySecurity,
		VendorCategoryVersionControl,
	}
}

func (i VendorCategory) String() string {
	return string(i)
}

func (i *VendorCategory) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		switch v {
		case "ANALYTICS":
			*i = VendorCategoryAnalytics
		case "CLOUD_MONITORING":
			*i = VendorCategoryCloudMonitoring
		case "CLOUD_PROVIDER":
			*i = VendorCategoryCloudProvider
		case "COLLABORATION":
			*i = VendorCategoryCollaboration
		case "CUSTOMER_SUPPORT":
			*i = VendorCategoryCustomerSupport
		case "DATA_STORAGE_AND_PROCESSING":
			*i = VendorCategoryDataStorageAndProcessing
		case "DOCUMENT_MANAGEMENT":
			*i = VendorCategoryDocumentManagement
		case "EMPLOYEE_MANAGEMENT":
			*i = VendorCategoryEmployeeManagement
		case "ENGINEERING":
			*i = VendorCategoryEngineering
		case "FINANCE":
			*i = VendorCategoryFinance
		case "IDENTITY_PROVIDER":
			*i = VendorCategoryIdentityProvider
		case "IT":
			*i = VendorCategoryIT
		case "MARKETING":
			*i = VendorCategoryMarketing
		case "OFFICE_OPERATIONS":
			*i = VendorCategoryOfficeOperations
		case "OTHER":
			*i = VendorCategoryOther
		case "PASSWORD_MANAGEMENT":
			*i = VendorCategoryPasswordManagement
		case "PRODUCT_AND_DESIGN":
			*i = VendorCategoryProductAndDesign
		case "PROFESSIONAL_SERVICES":
			*i = VendorCategoryProfessionalServices
		case "RECRUITING":
			*i = VendorCategoryRecruiting
		case "SALES":
			*i = VendorCategorySales
		case "SECURITY":
			*i = VendorCategorySecurity
		case "VERSION_CONTROL":
			*i = VendorCategoryVersionControl
		default:
			return fmt.Errorf("invalid VendorCategory value: %q", v)
		}
	default:
		return fmt.Errorf("unsupported type for VendorCategory: %T", value)
	}
	return nil
}

func (i VendorCategory) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i VendorCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *VendorCategory) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "ANALYTICS":
		*i = VendorCategoryAnalytics
	case "CLOUD_MONITORING":
		*i = VendorCategoryCloudMonitoring
	case "CLOUD_PROVIDER":
		*i = VendorCategoryCloudProvider
	case "COLLABORATION":
		*i = VendorCategoryCollaboration
	case "CUSTOMER_SUPPORT":
		*i = VendorCategoryCustomerSupport
	case "DATA_STORAGE_AND_PROCESSING":
		*i = VendorCategoryDataStorageAndProcessing
	case "DOCUMENT_MANAGEMENT":
		*i = VendorCategoryDocumentManagement
	case "EMPLOYEE_MANAGEMENT":
		*i = VendorCategoryEmployeeManagement
	case "ENGINEERING":
		*i = VendorCategoryEngineering
	case "FINANCE":
		*i = VendorCategoryFinance
	case "IDENTITY_PROVIDER":
		*i = VendorCategoryIdentityProvider
	case "IT":
		*i = VendorCategoryIT
	case "MARKETING":
		*i = VendorCategoryMarketing
	case "OFFICE_OPERATIONS":
		*i = VendorCategoryOfficeOperations
	case "OTHER":
		*i = VendorCategoryOther
	case "PASSWORD_MANAGEMENT":
		*i = VendorCategoryPasswordManagement
	case "PRODUCT_AND_DESIGN":
		*i = VendorCategoryProductAndDesign
	case "PROFESSIONAL_SERVICES":
		*i = VendorCategoryProfessionalServices
	case "RECRUITING":
		*i = VendorCategoryRecruiting
	case "SALES":
		*i = VendorCategorySales
	case "SECURITY":
		*i = VendorCategorySecurity
	case "VERSION_CONTROL":
		*i = VendorCategoryVersionControl
	default:
		return fmt.Errorf("invalid VendorCategory value: %q", s)
	}
	return nil
}

func (i *VendorCategory) UnmarshalText(text []byte) error {
	s := string(text)

	switch s {
	case "ANALYTICS":
		*i = VendorCategoryAnalytics
	case "CLOUD_MONITORING":
		*i = VendorCategoryCloudMonitoring
	case "CLOUD_PROVIDER":
		*i = VendorCategoryCloudProvider
	case "COLLABORATION":
		*i = VendorCategoryCollaboration
	case "CUSTOMER_SUPPORT":
		*i = VendorCategoryCustomerSupport
	case "DATA_STORAGE_AND_PROCESSING":
		*i = VendorCategoryDataStorageAndProcessing
	case "DOCUMENT_MANAGEMENT":
		*i = VendorCategoryDocumentManagement
	case "EMPLOYEE_MANAGEMENT":
		*i = VendorCategoryEmployeeManagement
	case "ENGINEERING":
		*i = VendorCategoryEngineering
	case "FINANCE":
		*i = VendorCategoryFinance
	case "IDENTITY_PROVIDER":
		*i = VendorCategoryIdentityProvider
	case "IT":
		*i = VendorCategoryIT
	case "MARKETING":
		*i = VendorCategoryMarketing
	case "OFFICE_OPERATIONS":
		*i = VendorCategoryOfficeOperations
	case "OTHER":
		*i = VendorCategoryOther
	case "PASSWORD_MANAGEMENT":
		*i = VendorCategoryPasswordManagement
	case "PRODUCT_AND_DESIGN":
		*i = VendorCategoryProductAndDesign
	case "PROFESSIONAL_SERVICES":
		*i = VendorCategoryProfessionalServices
	case "RECRUITING":
		*i = VendorCategoryRecruiting
	case "SALES":
		*i = VendorCategorySales
	case "SECURITY":
		*i = VendorCategorySecurity
	case "VERSION_CONTROL":
		*i = VendorCategoryVersionControl
	default:
		return fmt.Errorf("invalid VendorCategory value: %q", s)
	}
	return nil
}
