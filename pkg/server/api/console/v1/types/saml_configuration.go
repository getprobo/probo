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
	"github.com/getprobo/probo/pkg/coredata"
)

func NewSAMLConfigurationWithURLs(c *coredata.SAMLConfiguration, spEntityID, spAcsURL string) *SAMLConfiguration {
	return &SAMLConfiguration{
		ID:                      c.ID,
		EmailDomain:             c.EmailDomain,
		Enabled:                 c.Enabled,
		EnforcementPolicy:       c.EnforcementPolicy,
		DomainVerified:          c.DomainVerified,
		DomainVerificationToken: c.DomainVerificationToken,
		DomainVerifiedAt:        c.DomainVerifiedAt,
		SpEntityID:              spEntityID,
		SpAcsURL:                spAcsURL,
		IdpEntityID:             c.IdPEntityID,
		IdpSsoURL:               c.IdPSsoURL,
		IdpCertificate:          c.IdPCertificate,
		IdpMetadataURL:          c.IdPMetadataURL,
		AttributeEmail:          c.AttributeEmail,
		AttributeFirstname:      c.AttributeFirstname,
		AttributeLastname:       c.AttributeLastname,
		AttributeRole:           c.AttributeRole,
		DefaultRole:             c.DefaultRole,
		AutoSignupEnabled:       c.AutoSignupEnabled,
		CreatedAt:               c.CreatedAt,
		UpdatedAt:               c.UpdatedAt,
	}
}
