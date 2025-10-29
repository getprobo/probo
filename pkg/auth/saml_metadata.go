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

package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"math/big"
	"time"

	"github.com/crewjam/saml"
)

func GenerateServiceProviderMetadata(
	entityID string,
	acsURL string,
	spCert *x509.Certificate,
) ([]byte, error) {
	certData := base64.StdEncoding.EncodeToString(spCert.Raw)

	trueVal := true

	metadata := &saml.EntityDescriptor{
		EntityID: entityID,
		SPSSODescriptors: []saml.SPSSODescriptor{
			{
				SSODescriptor: saml.SSODescriptor{
					RoleDescriptor: saml.RoleDescriptor{
						ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
						KeyDescriptors: []saml.KeyDescriptor{
							{
								Use: "signing",
								KeyInfo: saml.KeyInfo{
									X509Data: saml.X509Data{
										X509Certificates: []saml.X509Certificate{
											{Data: certData},
										},
									},
								},
							},
							{
								Use: "encryption",
								KeyInfo: saml.KeyInfo{
									X509Data: saml.X509Data{
										X509Certificates: []saml.X509Certificate{
											{Data: certData},
										},
									},
								},
							},
						},
					},
				},
				AuthnRequestsSigned:  &trueVal,
				WantAssertionsSigned: &trueVal,
				AssertionConsumerServices: []saml.IndexedEndpoint{
					{
						Binding:  saml.HTTPPostBinding,
						Location: acsURL,
						Index:    0,
					},
				},
			},
		},
	}

	xmlBytes, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SP metadata to XML: %w", err)
	}

	return xmlBytes, nil
}

func ParseIdPCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block from IdP certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse X.509 certificate: %w", err)
	}

	return cert, nil
}

type IdPMetadata struct {
	EntityID    string
	SsoURL      string
	Certificate string
	MetadataURL *string
}

func ParseIdPMetadata(metadataXML string) (*IdPMetadata, error) {
	var entityDescriptor saml.EntityDescriptor
	if err := xml.Unmarshal([]byte(metadataXML), &entityDescriptor); err != nil {
		return nil, fmt.Errorf("failed to parse IdP metadata XML: %w", err)
	}

	if len(entityDescriptor.IDPSSODescriptors) == 0 {
		return nil, fmt.Errorf("no IDPSSODescriptor found in metadata")
	}

	idpDescriptor := entityDescriptor.IDPSSODescriptors[0]

	var ssoURL string
	for _, sso := range idpDescriptor.SingleSignOnServices {
		if sso.Binding == saml.HTTPPostBinding || sso.Binding == saml.HTTPRedirectBinding {
			ssoURL = sso.Location
			break
		}
	}
	if ssoURL == "" && len(idpDescriptor.SingleSignOnServices) > 0 {
		ssoURL = idpDescriptor.SingleSignOnServices[0].Location
	}
	if ssoURL == "" {
		return nil, fmt.Errorf("no SingleSignOnService found in metadata")
	}

	var certPEM string
	for _, keyDescriptor := range idpDescriptor.KeyDescriptors {
		if keyDescriptor.Use == "signing" || keyDescriptor.Use == "" {
			if len(keyDescriptor.KeyInfo.X509Data.X509Certificates) > 0 {
				certData := keyDescriptor.KeyInfo.X509Data.X509Certificates[0].Data
				certDER, err := base64.StdEncoding.DecodeString(certData)
				if err != nil {
					return nil, fmt.Errorf("failed to decode certificate: %w", err)
				}
				certPEM = string(pem.EncodeToMemory(&pem.Block{
					Type:  "CERTIFICATE",
					Bytes: certDER,
				}))
				break
			}
		}
	}
	if certPEM == "" {
		return nil, fmt.Errorf("no signing certificate found in metadata")
	}

	return &IdPMetadata{
		EntityID:    entityDescriptor.EntityID,
		SsoURL:      ssoURL,
		Certificate: certPEM,
	}, nil
}

func GenerateSelfSignedCertificate(entityID string) (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   entityID,
			Organization: []string{"Probo"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse created certificate: %w", err)
	}

	return cert, privateKey, nil
}
