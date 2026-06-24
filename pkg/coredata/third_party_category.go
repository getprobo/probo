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

package coredata

import (
	"encoding"
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

var (
	_ fmt.Stringer             = ThirdPartyCategory("")
	_ encoding.TextMarshaler   = ThirdPartyCategory("")
	_ encoding.TextUnmarshaler = (*ThirdPartyCategory)(nil)
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

func (v ThirdPartyCategory) IsValid() bool {
	switch v {
	case
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
		ThirdPartyCategoryVersionControl:
		return true
	}

	return false
}

func (v ThirdPartyCategory) String() string {
	return string(v)
}

func (v ThirdPartyCategory) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *ThirdPartyCategory) UnmarshalText(text []byte) error {
	val := ThirdPartyCategory(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid ThirdPartyCategory value: %q", string(text))
	}

	*v = val

	return nil
}
