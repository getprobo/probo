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

package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"

	"github.com/crewjam/saml"
)

func ParseIdpMetadata(metadataXML []byte) (string, string, *x509.Certificate, error) {
	var entityDescriptor saml.EntityDescriptor
	err := xml.Unmarshal(metadataXML, &entityDescriptor)
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot parse metadata XML: %w", err)
	}

	if len(entityDescriptor.IDPSSODescriptors) == 0 {
		return "", "", nil, fmt.Errorf("no IDPSSODescriptor found in metadata")
	}

	idpDescriptor := entityDescriptor.IDPSSODescriptors[0]

	ssoURL, err := getSsoURLFromMetadata(idpDescriptor)
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get SSO URL from metadata: %w", err)
	}

	cert, err := getCertificateFromMetadata(idpDescriptor)
	if err != nil {
		return "", "", nil, fmt.Errorf("cannot get certificate from metadata: %w", err)
	}

	return entityDescriptor.EntityID, ssoURL, cert, nil
}

func getSsoURLFromMetadata(idpDescriptor saml.IDPSSODescriptor) (string, error) {
	for _, sso := range idpDescriptor.SingleSignOnServices {
		if sso.Binding == saml.HTTPPostBinding || sso.Binding == saml.HTTPRedirectBinding {
			return sso.Location, nil
		}
	}

	if len(idpDescriptor.SingleSignOnServices) > 0 {
		return idpDescriptor.SingleSignOnServices[0].Location, nil
	}

	return "", fmt.Errorf("no SingleSignOnService found in metadata")
}

func getCertificateFromMetadata(idpDescriptor saml.IDPSSODescriptor) (*x509.Certificate, error) {
	for _, keyDescriptor := range idpDescriptor.KeyDescriptors {
		if keyDescriptor.Use == "signing" || keyDescriptor.Use == "" {
			if len(keyDescriptor.KeyInfo.X509Data.X509Certificates) > 0 {
				certData := keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data
				certDER, err := base64.StdEncoding.DecodeString(certData)
				if err != nil {
					return nil, fmt.Errorf("cannot decode certificate: %w", err)
				}

				cert, err := x509.ParseCertificate(certDER)
				if err != nil {
					return nil, fmt.Errorf("cannot parse certificate: %w", err)
				}

				return cert, nil
			}
		}
	}

	return nil, fmt.Errorf("no signing certificate found in metadata")
}
