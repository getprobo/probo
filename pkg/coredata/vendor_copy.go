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

type vendorCopy struct {
	vendors  Vendors
	scope    Scoper
	position int
}

func (c *vendorCopy) Next() bool {
	return c.position < len(c.vendors)
}

func (c *vendorCopy) Values() ([]interface{}, error) {
	if c.position >= len(c.vendors) {
		return nil, nil
	}

	vendor := c.vendors[c.position]
	c.position++

	return []any{
		c.scope.GetTenantID(),
		vendor.ID,
		vendor.OrganizationID,
		vendor.Name,
		vendor.Description,
		vendor.Category,
		vendor.HeadquarterAddress,
		vendor.LegalName,
		vendor.WebsiteURL,
		vendor.PrivacyPolicyURL,
		vendor.ServiceLevelAgreementURL,
		vendor.DataProcessingAgreementURL,
		vendor.BusinessAssociateAgreementURL,
		vendor.SubprocessorsListURL,
		vendor.Certifications,
		vendor.BusinessOwnerID,
		vendor.SecurityOwnerID,
		vendor.StatusPageURL,
		vendor.TermsOfServiceURL,
		vendor.SecurityPageURL,
		vendor.TrustPageURL,
		vendor.ShowOnTrustCenter,
		vendor.SnapshotID,
		vendor.OriginalID,
		vendor.CreatedAt,
		vendor.UpdatedAt,
	}, nil
}

func (c *vendorCopy) Err() error {
	return nil
}
