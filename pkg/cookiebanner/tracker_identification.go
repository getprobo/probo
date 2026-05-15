// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package cookiebanner

// TrackerIdentification is the structured output the tracker-mapping
// agent returns. The Category field uses the same values as the
// third_party_category PostgreSQL enum so auto-created CommonThirdParty
// rows get a valid category without mapping.
type TrackerIdentification struct {
	ThirdPartyName string  `json:"third_party_name" jsonschema:"Name of the company or service that sets this tracker (e.g. 'Google Analytics', 'Meta Pixel'). Empty string if truly unknown."`
	Category       string  `json:"category" jsonschema:"Third party category. One of: ANALYTICS, ADVERTISING, CLOUD_MONITORING, CLOUD_PROVIDER, COLLABORATION, CUSTOMER_SUPPORT, DATA_STORAGE_AND_PROCESSING, DOCUMENT_MANAGEMENT, EMPLOYEE_MANAGEMENT, ENGINEERING, FINANCE, IDENTITY_PROVIDER, IT, MARKETING, OFFICE_OPERATIONS, OTHER, PASSWORD_MANAGEMENT, PRODUCT_AND_DESIGN, PROFESSIONAL_SERVICES, RECRUITING, SALES, SECURITY, VERSION_CONTROL"`
	Description    string  `json:"description" jsonschema:"What this tracker does in one sentence"`
	Confidence     float64 `json:"confidence" jsonschema:"Confidence level from 0.0 to 1.0. Set below 0.5 if unsure."`
}
