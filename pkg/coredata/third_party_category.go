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

package coredata

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type ThirdPartyCategory string

const (
	ThirdPartyCategoryAnalytics                ThirdPartyCategory = "ANALYTICS"
	ThirdPartyCategoryCloudMonitoring          ThirdPartyCategory = "CLOUD_MONITORING"
	ThirdPartyCategoryCloudProvider            ThirdPartyCategory = "CLOUD_PROVIDER"
	ThirdPartyCategoryCollaboration            ThirdPartyCategory = "COLLABORATION"
	ThirdPartyCategoryCustomerSupport          ThirdPartyCategory = "CUSTOMER_SUPPORT"
	ThirdPartyCategoryDataStorageAndProcessing ThirdPartyCategory = "DATA_STORAGE_AND_PROCESSING"
	ThirdPartyCategoryDocumentManagement       ThirdPartyCategory = "DOCUMENT_MANAGEMENT"
	ThirdPartyCategoryEmployeeManagement       ThirdPartyCategory = "EMPLOYEE_MANAGEMENT"
	ThirdPartyCategoryEngineering              ThirdPartyCategory = "ENGINEERING"
	ThirdPartyCategoryFinance                  ThirdPartyCategory = "FINANCE"
	ThirdPartyCategoryIdentityProvider         ThirdPartyCategory = "IDENTITY_PROVIDER"
	ThirdPartyCategoryIT                       ThirdPartyCategory = "IT"
	ThirdPartyCategoryMarketing                ThirdPartyCategory = "MARKETING"
	ThirdPartyCategoryOfficeOperations         ThirdPartyCategory = "OFFICE_OPERATIONS"
	ThirdPartyCategoryOther                    ThirdPartyCategory = "OTHER"
	ThirdPartyCategoryPasswordManagement       ThirdPartyCategory = "PASSWORD_MANAGEMENT"
	ThirdPartyCategoryProductAndDesign         ThirdPartyCategory = "PRODUCT_AND_DESIGN"
	ThirdPartyCategoryProfessionalServices     ThirdPartyCategory = "PROFESSIONAL_SERVICES"
	ThirdPartyCategoryRecruiting               ThirdPartyCategory = "RECRUITING"
	ThirdPartyCategorySales                    ThirdPartyCategory = "SALES"
	ThirdPartyCategorySecurity                 ThirdPartyCategory = "SECURITY"
	ThirdPartyCategoryVersionControl           ThirdPartyCategory = "VERSION_CONTROL"
)

func ThirdPartyCategories() []ThirdPartyCategory {
	return []ThirdPartyCategory{
		ThirdPartyCategoryAnalytics,
		ThirdPartyCategoryCloudMonitoring,
		ThirdPartyCategoryCloudProvider,
		ThirdPartyCategoryCollaboration,
		ThirdPartyCategoryCustomerSupport,
		ThirdPartyCategoryDataStorageAndProcessing,
		ThirdPartyCategoryDocumentManagement,
		ThirdPartyCategoryEmployeeManagement,
		ThirdPartyCategoryEngineering,
		ThirdPartyCategoryFinance,
		ThirdPartyCategoryIdentityProvider,
		ThirdPartyCategoryIT,
		ThirdPartyCategoryMarketing,
		ThirdPartyCategoryOfficeOperations,
		ThirdPartyCategoryOther,
		ThirdPartyCategoryPasswordManagement,
		ThirdPartyCategoryProductAndDesign,
		ThirdPartyCategoryProfessionalServices,
		ThirdPartyCategoryRecruiting,
		ThirdPartyCategorySales,
		ThirdPartyCategorySecurity,
		ThirdPartyCategoryVersionControl,
	}
}

func (i ThirdPartyCategory) String() string {
	return string(i)
}

func (i *ThirdPartyCategory) Scan(value any) error {
	switch v := value.(type) {
	case string:
		switch v {
		case "ANALYTICS":
			*i = ThirdPartyCategoryAnalytics
		case "CLOUD_MONITORING":
			*i = ThirdPartyCategoryCloudMonitoring
		case "CLOUD_PROVIDER":
			*i = ThirdPartyCategoryCloudProvider
		case "COLLABORATION":
			*i = ThirdPartyCategoryCollaboration
		case "CUSTOMER_SUPPORT":
			*i = ThirdPartyCategoryCustomerSupport
		case "DATA_STORAGE_AND_PROCESSING":
			*i = ThirdPartyCategoryDataStorageAndProcessing
		case "DOCUMENT_MANAGEMENT":
			*i = ThirdPartyCategoryDocumentManagement
		case "EMPLOYEE_MANAGEMENT":
			*i = ThirdPartyCategoryEmployeeManagement
		case "ENGINEERING":
			*i = ThirdPartyCategoryEngineering
		case "FINANCE":
			*i = ThirdPartyCategoryFinance
		case "IDENTITY_PROVIDER":
			*i = ThirdPartyCategoryIdentityProvider
		case "IT":
			*i = ThirdPartyCategoryIT
		case "MARKETING":
			*i = ThirdPartyCategoryMarketing
		case "OFFICE_OPERATIONS":
			*i = ThirdPartyCategoryOfficeOperations
		case "OTHER":
			*i = ThirdPartyCategoryOther
		case "PASSWORD_MANAGEMENT":
			*i = ThirdPartyCategoryPasswordManagement
		case "PRODUCT_AND_DESIGN":
			*i = ThirdPartyCategoryProductAndDesign
		case "PROFESSIONAL_SERVICES":
			*i = ThirdPartyCategoryProfessionalServices
		case "RECRUITING":
			*i = ThirdPartyCategoryRecruiting
		case "SALES":
			*i = ThirdPartyCategorySales
		case "SECURITY":
			*i = ThirdPartyCategorySecurity
		case "VERSION_CONTROL":
			*i = ThirdPartyCategoryVersionControl
		default:
			return fmt.Errorf("invalid ThirdPartyCategory value: %q", v)
		}
	default:
		return fmt.Errorf("unsupported type for ThirdPartyCategory: %T", value)
	}
	return nil
}

func (i ThirdPartyCategory) Value() (driver.Value, error) {
	return i.String(), nil
}

func (i ThirdPartyCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *ThirdPartyCategory) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "ANALYTICS":
		*i = ThirdPartyCategoryAnalytics
	case "CLOUD_MONITORING":
		*i = ThirdPartyCategoryCloudMonitoring
	case "CLOUD_PROVIDER":
		*i = ThirdPartyCategoryCloudProvider
	case "COLLABORATION":
		*i = ThirdPartyCategoryCollaboration
	case "CUSTOMER_SUPPORT":
		*i = ThirdPartyCategoryCustomerSupport
	case "DATA_STORAGE_AND_PROCESSING":
		*i = ThirdPartyCategoryDataStorageAndProcessing
	case "DOCUMENT_MANAGEMENT":
		*i = ThirdPartyCategoryDocumentManagement
	case "EMPLOYEE_MANAGEMENT":
		*i = ThirdPartyCategoryEmployeeManagement
	case "ENGINEERING":
		*i = ThirdPartyCategoryEngineering
	case "FINANCE":
		*i = ThirdPartyCategoryFinance
	case "IDENTITY_PROVIDER":
		*i = ThirdPartyCategoryIdentityProvider
	case "IT":
		*i = ThirdPartyCategoryIT
	case "MARKETING":
		*i = ThirdPartyCategoryMarketing
	case "OFFICE_OPERATIONS":
		*i = ThirdPartyCategoryOfficeOperations
	case "OTHER":
		*i = ThirdPartyCategoryOther
	case "PASSWORD_MANAGEMENT":
		*i = ThirdPartyCategoryPasswordManagement
	case "PRODUCT_AND_DESIGN":
		*i = ThirdPartyCategoryProductAndDesign
	case "PROFESSIONAL_SERVICES":
		*i = ThirdPartyCategoryProfessionalServices
	case "RECRUITING":
		*i = ThirdPartyCategoryRecruiting
	case "SALES":
		*i = ThirdPartyCategorySales
	case "SECURITY":
		*i = ThirdPartyCategorySecurity
	case "VERSION_CONTROL":
		*i = ThirdPartyCategoryVersionControl
	default:
		return fmt.Errorf("invalid ThirdPartyCategory value: %q", s)
	}
	return nil
}

func (i *ThirdPartyCategory) UnmarshalText(text []byte) error {
	s := string(text)

	switch s {
	case "ANALYTICS":
		*i = ThirdPartyCategoryAnalytics
	case "CLOUD_MONITORING":
		*i = ThirdPartyCategoryCloudMonitoring
	case "CLOUD_PROVIDER":
		*i = ThirdPartyCategoryCloudProvider
	case "COLLABORATION":
		*i = ThirdPartyCategoryCollaboration
	case "CUSTOMER_SUPPORT":
		*i = ThirdPartyCategoryCustomerSupport
	case "DATA_STORAGE_AND_PROCESSING":
		*i = ThirdPartyCategoryDataStorageAndProcessing
	case "DOCUMENT_MANAGEMENT":
		*i = ThirdPartyCategoryDocumentManagement
	case "EMPLOYEE_MANAGEMENT":
		*i = ThirdPartyCategoryEmployeeManagement
	case "ENGINEERING":
		*i = ThirdPartyCategoryEngineering
	case "FINANCE":
		*i = ThirdPartyCategoryFinance
	case "IDENTITY_PROVIDER":
		*i = ThirdPartyCategoryIdentityProvider
	case "IT":
		*i = ThirdPartyCategoryIT
	case "MARKETING":
		*i = ThirdPartyCategoryMarketing
	case "OFFICE_OPERATIONS":
		*i = ThirdPartyCategoryOfficeOperations
	case "OTHER":
		*i = ThirdPartyCategoryOther
	case "PASSWORD_MANAGEMENT":
		*i = ThirdPartyCategoryPasswordManagement
	case "PRODUCT_AND_DESIGN":
		*i = ThirdPartyCategoryProductAndDesign
	case "PROFESSIONAL_SERVICES":
		*i = ThirdPartyCategoryProfessionalServices
	case "RECRUITING":
		*i = ThirdPartyCategoryRecruiting
	case "SALES":
		*i = ThirdPartyCategorySales
	case "SECURITY":
		*i = ThirdPartyCategorySecurity
	case "VERSION_CONTROL":
		*i = ThirdPartyCategoryVersionControl
	default:
		return fmt.Errorf("invalid ThirdPartyCategory value: %q", s)
	}
	return nil
}
