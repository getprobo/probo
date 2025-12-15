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
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type (
	SAMLService struct {
		pg              *pg.Client
		encryptionKey   cipher.EncryptionKey
		baseURL         string
		sessionDuration time.Duration
		cookieName      string
		cookieSecret    string
		certificate     *x509.Certificate
		privateKey      *rsa.PrivateKey
		logger          *log.Logger
	}

	ErrSPCertificateNotConfigured struct{}

	ErrSAMLConfigurationNotFound struct {
		OrganizationID gid.GID
	}

	ErrSAMLDisabled struct {
		OrganizationID gid.GID
	}

	ErrInvalidIdPCertificate struct {
		Err error
	}

	ErrInvalidURL struct {
		Field string
		URL   string
		Err   error
	}

	ErrCannotCreateServiceProvider struct {
		Err error
	}

	ErrCannotCreateAuthRequest struct {
		Err error
	}

	ErrCannotGenerateRedirectURL struct {
		Err error
	}

	ErrCannotParseSAMLResponse struct {
		Err error
	}

	ErrCannotValidateAssertion struct {
		Err error
	}

	ErrCannotExtractUserAttributes struct {
		Err error
	}

	ErrCannotMapRole struct {
		Err error
	}

	ErrReplayAttackDetected struct {
		AssertionID string
		Err         error
	}
)

func (e ErrSPCertificateNotConfigured) Error() string {
	return "SP certificate and private key are not configured"
}

func (e ErrSAMLConfigurationNotFound) Error() string {
	return fmt.Sprintf("SAML configuration not found for organization %s", e.OrganizationID)
}

func (e ErrSAMLDisabled) Error() string {
	return fmt.Sprintf("SAML is disabled for organization %s", e.OrganizationID)
}

func (e ErrInvalidIdPCertificate) Error() string {
	return fmt.Sprintf("cannot parse IdP certificate: %v", e.Err)
}

func (e ErrInvalidURL) Error() string {
	return fmt.Sprintf("cannot parse %s URL %q: %v", e.Field, e.URL, e.Err)
}

func (e ErrCannotCreateServiceProvider) Error() string {
	return fmt.Sprintf("cannot create service provider: %v", e.Err)
}

func (e ErrCannotCreateAuthRequest) Error() string {
	return fmt.Sprintf("cannot create AuthnRequest: %v", e.Err)
}

func (e ErrCannotGenerateRedirectURL) Error() string {
	return fmt.Sprintf("cannot generate redirect URL: %v", e.Err)
}

func (e ErrCannotParseSAMLResponse) Error() string {
	return fmt.Sprintf("cannot parse SAML response: %v", e.Err)
}

func (e ErrCannotValidateAssertion) Error() string {
	return fmt.Sprintf("cannot validate assertion: %v", e.Err)
}

func (e ErrCannotExtractUserAttributes) Error() string {
	return fmt.Sprintf("cannot extract user attributes: %v", e.Err)
}

func (e ErrCannotMapRole) Error() string {
	return fmt.Sprintf("cannot map role: %v", e.Err)
}

func (e ErrReplayAttackDetected) Error() string {
	return fmt.Sprintf("replay attack detected for assertion %s: %v", e.AssertionID, e.Err)
}

func NewSAMLService(
	pg *pg.Client,
	encryptionKey cipher.EncryptionKey,
	baseURL string,
	sessionDuration time.Duration,
	cookieName string,
	cookieSecret string,
	certificatePEM string,
	privateKeyPEM string,
	logger *log.Logger,
) (*SAMLService, error) {
	var certificate *x509.Certificate
	var privateKey *rsa.PrivateKey

	if certificatePEM != "" {
		block, _ := pem.Decode([]byte(certificatePEM))
		if block == nil || block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("invalid certificate PEM format")
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("cannot parse certificate: %w", err)
		}
		certificate = cert
	}

	if privateKeyPEM != "" {
		block, _ := pem.Decode([]byte(privateKeyPEM))
		if block == nil {
			return nil, fmt.Errorf("invalid private key PEM format")
		}

		var key *rsa.PrivateKey
		var err error
		switch block.Type {
		case "RSA PRIVATE KEY":
			key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("cannot parse PKCS1 private key: %w", err)
			}
		case "PRIVATE KEY":
			parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("cannot parse PKCS8 private key: %w", err)
			}
			var ok bool
			key, ok = parsedKey.(*rsa.PrivateKey)
			if !ok {
				return nil, fmt.Errorf("private key is not RSA")
			}
		default:
			return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
		}
		privateKey = key
	}

	return &SAMLService{
		pg:              pg,
		encryptionKey:   encryptionKey,
		baseURL:         baseURL,
		sessionDuration: sessionDuration,
		cookieName:      cookieName,
		cookieSecret:    cookieSecret,
		certificate:     certificate,
		privateKey:      privateKey,
		logger:          logger,
	}, nil
}

func (s *SAMLService) GetEntityID() string {
	return fmt.Sprintf("%s/connect/saml/metadata", s.baseURL)
}

func (s *SAMLService) GetAcsURL() string {
	return fmt.Sprintf("%s/connect/saml/consume", s.baseURL)
}

func (s *SAMLService) GetServiceProvider(
	ctx context.Context,
	config *coredata.SAMLConfiguration,
) (*saml.ServiceProvider, error) {
	if s.certificate == nil || s.privateKey == nil {
		return nil, ErrSPCertificateNotConfigured{}
	}

	idpCert, err := ParseIdPCertificate(config.IdPCertificate)
	if err != nil {
		return nil, ErrInvalidIdPCertificate{Err: err}
	}

	acsURL, err := url.Parse(s.GetAcsURL())
	if err != nil {
		return nil, ErrInvalidURL{Field: "ACS", URL: s.GetAcsURL(), Err: err}
	}

	metadataURL, err := url.Parse(s.GetEntityID())
	if err != nil {
		return nil, ErrInvalidURL{Field: "Metadata", URL: s.GetEntityID(), Err: err}
	}

	idpSSOURL, err := url.Parse(config.IdPSsoURL)
	if err != nil {
		return nil, ErrInvalidURL{Field: "IdP SSO", URL: config.IdPSsoURL, Err: err}
	}

	sp := &saml.ServiceProvider{
		EntityID:          s.GetEntityID(),
		Key:               s.privateKey,
		Certificate:       s.certificate,
		MetadataURL:       *metadataURL,
		AcsURL:            *acsURL,
		SloURL:            *acsURL,
		AllowIDPInitiated: true,
		IDPMetadata: &saml.EntityDescriptor{
			EntityID: config.IdPEntityID,
			IDPSSODescriptors: []saml.IDPSSODescriptor{
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
												{Data: base64.StdEncoding.EncodeToString(idpCert.Raw)},
											},
										},
									},
								},
							},
						},
					},
					SingleSignOnServices: []saml.Endpoint{
						{
							Binding:  saml.HTTPRedirectBinding,
							Location: idpSSOURL.String(),
						},
					},
				},
			},
		},
	}

	return sp, nil
}

func (s *SAMLService) InitiateSAMLLogin(
	ctx context.Context,
	organizationID gid.GID,
	tenantID gid.TenantID,
	emailDomain string,
) (string, error) {
	var config coredata.SAMLConfiguration
	scope := coredata.NewScope(tenantID)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return config.LoadByOrganizationIDAndEmailDomain(ctx, conn, scope, organizationID, emailDomain)
		},
	)
	if err != nil {
		return "", ErrSAMLConfigurationNotFound{OrganizationID: organizationID}
	}

	if !config.Enabled {
		return "", ErrSAMLDisabled{OrganizationID: organizationID}
	}

	sp, err := s.GetServiceProvider(ctx, &config)
	if err != nil {
		return "", ErrCannotCreateServiceProvider{Err: err}
	}

	authReq, err := sp.MakeAuthenticationRequest(
		config.IdPSsoURL,
		saml.HTTPRedirectBinding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		return "", ErrCannotCreateAuthRequest{Err: err}
	}

	now := time.Now()
	requestExpiry := now.Add(10 * time.Minute)

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			samlRequest := coredata.SAMLRequest{
				ID:             authReq.ID,
				OrganizationID: organizationID,
				CreatedAt:      now,
				ExpiresAt:      requestExpiry,
			}

			if err := samlRequest.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert SAML request: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	redirectURL, err := authReq.Redirect(config.ID.String(), sp)
	if err != nil {
		return "", ErrCannotGenerateRedirectURL{Err: err}
	}

	return redirectURL.String(), nil
}

type SAMLUserInfo struct {
	Email          mail.Addr
	FullName       string
	Role           *coredata.MembershipRole
	SAMLSubject    string
	OrganizationID gid.GID
	SAMLConfigID   gid.GID
}

func (s *SAMLService) loadConfigFromRelayState(
	ctx context.Context,
	relayStateValue string,
) (*coredata.SAMLConfiguration, *coredata.Organization, error) {
	if relayStateValue == "" {
		return nil, nil, fmt.Errorf("RelayState is required and must contain SAML config ID")
	}

	samlConfigID, err := gid.ParseGID(relayStateValue)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid SAML config ID in RelayState: %w", err)
	}

	var config coredata.SAMLConfiguration
	var org coredata.Organization

	err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := config.LoadByID(ctx, conn, coredata.NewNoScope(), samlConfigID); err != nil {
				return fmt.Errorf("cannot load SAML configuration: %w", err)
			}

			if err := org.LoadByID(ctx, conn, coredata.NewNoScope(), config.OrganizationID); err != nil {
				return fmt.Errorf("organization not found: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return &config, &org, nil
}

func (s *SAMLService) HandleSAMLAssertion(
	ctx context.Context,
	req *http.Request,
) (*SAMLUserInfo, error) {
	samlResponseEncoded := req.FormValue("SAMLResponse")
	if samlResponseEncoded == "" {
		return nil, fmt.Errorf("missing SAMLResponse in request")
	}

	relayStateValue := req.FormValue("RelayState")
	config, org, err := s.loadConfigFromRelayState(ctx, relayStateValue)
	if err != nil {
		return nil, err
	}

	if !config.Enabled {
		return nil, ErrSAMLDisabled{OrganizationID: config.OrganizationID}
	}

	sp, err := s.GetServiceProvider(ctx, config)
	if err != nil {
		return nil, ErrCannotCreateServiceProvider{Err: err}
	}

	if req.URL.Scheme == "" {
		req.URL.Scheme = "https"
	}
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}

	now := time.Now()

	var possibleRequestIDs []string
	err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			requestIDs, err := coredata.LoadValidRequestIDsForOrganization(ctx, conn, config.OrganizationID, now)
			if err != nil {
				return err
			}
			possibleRequestIDs = requestIDs

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load valid request IDs: %w", err)
	}

	assertion, err := sp.ParseResponse(req, possibleRequestIDs)
	if err != nil {
		return nil, fmt.Errorf("cannot parse SAML response: %w", err)
	}

	if err := ValidateAssertion(assertion, s.GetEntityID(), now); err != nil {
		return nil, ErrCannotValidateAssertion{Err: err}
	}

	if assertion.ID != "" {
		var expiresAt time.Time
		if assertion.Conditions != nil && !assertion.Conditions.NotOnOrAfter.IsZero() {
			expiresAt = assertion.Conditions.NotOnOrAfter
		} else {
			expiresAt = now.Add(24 * time.Hour)
		}

		scope := coredata.NewScope(org.TenantID)
		err = s.pg.WithTx(
			ctx,
			func(tx pg.Conn) error {
				if err := PreventReplayAttack(ctx, tx, scope, assertion.ID, config.OrganizationID, expiresAt); err != nil {
					return fmt.Errorf("cannot prevent replay attack: %w", err)
				}

				return nil
			},
		)
		if err != nil {
			var replayAttackErr *coredata.ErrAssertionAlreadyUsed
			if errors.As(err, &replayAttackErr) {
				return nil, ErrReplayAttackDetected{AssertionID: assertion.ID, Err: replayAttackErr}
			}

			return nil, fmt.Errorf("cannot prevent replay attack: %w", err)
		}
	}

	email, fullname, samlRole, err := ExtractUserAttributes(
		assertion,
		config.AttributeEmail,
		config.AttributeFirstname,
		config.AttributeLastname,
		config.AttributeRole,
	)
	if err != nil {
		return nil, ErrCannotExtractUserAttributes{Err: err}
	}

	if !strings.EqualFold(email.Domain(), config.EmailDomain) {
		return nil, fmt.Errorf("email domain mismatch: assertion contains email with domain %s but SAML config is for domain %s", email.Domain(), config.EmailDomain)
	}

	systemRole := MapSAMLRoleToSystemRole(samlRole)

	samlSubject := ""
	if assertion.Subject != nil && assertion.Subject.NameID != nil {
		samlSubject = assertion.Subject.NameID.Value
	}

	return &SAMLUserInfo{
		Email:          email,
		FullName:       fullname,
		Role:           systemRole,
		SAMLSubject:    samlSubject,
		OrganizationID: config.OrganizationID,
		SAMLConfigID:   config.ID,
	}, nil
}

func (s *SAMLService) GetMetadataURL(organizationID gid.GID) string {
	return fmt.Sprintf("%s/connect/saml/metadata/%s", s.baseURL, organizationID)
}

func (s *SAMLService) GenerateMetadata() ([]byte, error) {
	if s.certificate == nil {
		return nil, ErrSPCertificateNotConfigured{}
	}

	return GenerateServiceProviderMetadata(
		s.GetEntityID(),
		s.GetAcsURL(),
		s.certificate,
	)
}
