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
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
)

type (
	Service struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
		baseURL       string
		certificate   *x509.Certificate
		privateKey    *rsa.PrivateKey
		logger        *log.Logger
	}

	UserInfo struct {
		Email          string
		FullName       string
		Role           *coredata.MembershipRole
		SAMLSubject    string
		OrganizationID gid.GID
		SAMLConfigID   gid.GID
	}
)

func NewService(
	pg *pg.Client,
	encryptionKey cipher.EncryptionKey,
	baseURL string,
	certificate *x509.Certificate,
	privateKey *rsa.PrivateKey,
	logger *log.Logger,
) (*Service, error) {
	return &Service{
		pg:            pg,
		encryptionKey: encryptionKey,
		baseURL:       baseURL,
		certificate:   certificate,
		privateKey:    privateKey,
		logger:        logger,
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	gc := NewGarbageCollector(s.pg, DefaultGarbageCollectionInterval, s.logger)

	gcCtx, stopGC := context.WithCancel(ctx)
	defer stopGC()

	errCh := make(chan error, 1)
	go func() {
		errCh <- gc.Run(gcCtx)
	}()

	select {
	case <-ctx.Done():
		stopGC()
		<-errCh
		return ctx.Err()
	case err := <-errCh:
		if err != nil {
			s.logger.ErrorCtx(ctx, "saml garbage collector failed", log.Error(err))
			return err
		}

		return nil
	}
}

func (s *Service) GenerateSpMetadata() ([]byte, error) {
	sp := s.baseServiceProvider()
	return xml.MarshalIndent(sp.Metadata(), "", "  ")
}

func (s *Service) InitiateLogin(
	ctx context.Context,
	configID gid.GID,
	continuePath string,
) (*url.URL, error) {
	var (
		now           = time.Now()
		requestExpiry = now.Add(10 * time.Minute)
		redirect      *url.URL
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			config := &coredata.SAMLConfiguration{}
			err := config.LoadByID(ctx, tx, coredata.NewNoScope(), configID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSAMLConfigurationNotFoundError(configID)
				}

				return fmt.Errorf("cannot load SAML configuration: %w", err)
			}

			if config.EnforcementPolicy == coredata.SAMLEnforcementPolicyOff {
				return NewSAMLDisabledError()
			}

			sp, err := s.serviceProvider(config)
			if err != nil {
				return fmt.Errorf("cannot build service provider: %w", err)
			}

			req, err := sp.MakeAuthenticationRequest(config.IdPSsoURL, saml.HTTPRedirectBinding, saml.HTTPPostBinding)
			if err != nil {
				return fmt.Errorf("cannot create authentication request: %w", err)
			}

			samlRequest := coredata.SAMLRequest{
				ID:             req.ID,
				OrganizationID: config.OrganizationID,
				CreatedAt:      now,
				ExpiresAt:      requestExpiry,
			}

			if err := samlRequest.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert SAML request: %w", err)
			}

			relayState := config.ID.String() + url.QueryEscape(continuePath)

			redirect, err = req.Redirect(relayState, sp)
			if err != nil {
				return fmt.Errorf("cannot generate redirect URL: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return redirect, nil
}

func (s *Service) HandleAssertion(
	ctx context.Context,
	samlResponse string,
	configID gid.GID,
) (*coredata.Identity, *coredata.Membership, error) {
	var (
		now        = time.Now()
		identity   = &coredata.Identity{}
		membership = &coredata.Membership{}
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			config := &coredata.SAMLConfiguration{}

			err := config.LoadByID(ctx, tx, coredata.NewNoScope(), configID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewSAMLConfigurationNotFoundError(configID)
				}

				return fmt.Errorf("cannot load SAML configuration: %w", err)
			}

			if config.EnforcementPolicy == coredata.SAMLEnforcementPolicyOff {
				return NewSAMLDisabledError()
			}

			sp, err := s.serviceProvider(config)
			if err != nil {
				return fmt.Errorf("cannot create service provider: %w", err)
			}

			possibleRequestIDs, err := coredata.LoadValidRequestIDsForOrganization(ctx, tx, config.OrganizationID, now)
			if err != nil {
				return fmt.Errorf("cannot load valid request IDs: %w", err)
			}

			decodedResponse, err := base64.StdEncoding.DecodeString(samlResponse)
			if err != nil {
				return fmt.Errorf("cannot decode SAML response: %w", err)
			}

			assertion, err := sp.ParseXMLResponse(decodedResponse, possibleRequestIDs, sp.AcsURL)
			if err != nil {
				return fmt.Errorf("cannot parse SAML response: %w", err)
			}

			err = s.validateAssertion(assertion, config, now)
			if err != nil {
				return NewInvalidAssertionError(assertion.ID, err)
			}

			expiresAt := now.Add(24 * time.Hour)
			if assertion.Conditions.NotOnOrAfter.IsZero() {
				expiresAt = assertion.Conditions.NotOnOrAfter
			}

			samlAssertion := coredata.SAMLAssertion{
				ID:             assertion.ID,
				OrganizationID: config.OrganizationID,
				UsedAt:         now,
				ExpiresAt:      expiresAt,
			}

			err = samlAssertion.Insert(ctx, tx)
			if err != nil {
				if err == coredata.ErrResourceAlreadyExists {
					return NewReplayAttackDetectedError(samlAssertion.ID)
				}

				return fmt.Errorf("cannot insert SAML assertion: %w", err)
			}

			email, fullname, role, err := extractUserAttributes(assertion, config)
			if err != nil {
				return fmt.Errorf("cannot extract user attributes: %w", err)
			}

			if !strings.EqualFold(email.Domain(), config.EmailDomain) {
				return NewEmailDomainMismatchError(email, config.EmailDomain)
			}

			err = identity.LoadByEmail(ctx, tx, email)
			if err == coredata.ErrResourceNotFound && !config.AutoSignupEnabled {
				return NewSAMLAutoSignupDisabledError(config.ID)
			} else if err == coredata.ErrResourceNotFound && config.AutoSignupEnabled {
				*identity = coredata.Identity{
					ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
					EmailAddress:         email,
					SAMLSubject:          &assertion.Subject.NameID.Value,
					FullName:             fullname,
					HashedPassword:       nil,
					EmailAddressVerified: true,
					CreatedAt:            now,
					UpdatedAt:            now,
				}

				err := identity.Insert(ctx, tx)
				if err != nil {
					return fmt.Errorf("cannot insert identity: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			} else {
				identity.EmailAddress = email
				identity.FullName = fullname

				// Identity can exist (e.g. provisioned via SCIM) but not have a SAML subject
				if identity.SAMLSubject == nil {
					identity.SAMLSubject = &assertion.Subject.NameID.Value
				}

				identity.EmailAddressVerified = true
				identity.UpdatedAt = now

				err = identity.Update(ctx, tx)
				if err != nil {
					return fmt.Errorf("cannot update identity: %w", err)
				}
			}

			scope := coredata.NewScopeFromObjectID(config.OrganizationID)

			err = membership.LoadByIdentityAndOrg(ctx, tx, scope, identity.ID, config.OrganizationID)
			if err != nil && err != coredata.ErrResourceNotFound {
				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.ID != gid.Nil && membership.State == coredata.MembershipStateInactive {
				return NewMembershipInactiveError(membership.ID)
			}

			isMember := membership.ID != gid.Nil
			if !isMember {
				membership = &coredata.Membership{
					ID:             gid.New(config.ID.TenantID(), coredata.MembershipEntityType),
					IdentityID:     identity.ID,
					OrganizationID: config.OrganizationID,
					Role:           coredata.MembershipRoleEmployee,
					Source:         coredata.MembershipSourceSAML,
					State:          coredata.MembershipStateActive,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				err = membership.Insert(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot insert membership: %w", err)
				}

				membershipProfile := &coredata.MembershipProfile{
					ID:             gid.New(membership.ID.TenantID(), coredata.MembershipProfileEntityType),
					IdentityID:     identity.ID,
					OrganizationID: config.OrganizationID,
					MembershipID:   membership.ID,
					FullName:       fullname,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				err = membershipProfile.Insert(ctx, tx)
				if err != nil {
					return fmt.Errorf("cannot insert membership profile: %w", err)
				}

				// Expire all pending invitations for email in organization
				invitations := &coredata.Invitations{}
				onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
				if err := invitations.ExpireByEmailAndOrganization(
					ctx,
					tx,
					coredata.NewScopeFromObjectID(config.OrganizationID),
					email,
					config.OrganizationID,
					onlyPending,
				); err != nil {
					return fmt.Errorf("cannot expire pending invitations by email: %w", err)
				}
			}

			if membership.Source != coredata.MembershipSourceSCIM {
				needsUpdate := false

				if role != nil {
					membership.Role = *role
					membership.UpdatedAt = now
					needsUpdate = true
				}

				if membership.Source == coredata.MembershipSourceManual {
					membership.Source = coredata.MembershipSourceSAML
					membership.UpdatedAt = now
					needsUpdate = true
				}

				if needsUpdate {
					err = membership.Update(ctx, tx, scope)
					if err != nil {
						return fmt.Errorf("cannot update membership: %w", err)
					}
				}

				memberProfile := &coredata.MembershipProfile{}
				err = memberProfile.LoadByMembershipID(ctx, tx, scope, membership.ID)
				if err != nil {
					return fmt.Errorf("cannot load membership profile: %w", err)
				}

				memberProfile.FullName = fullname
				memberProfile.UpdatedAt = now
				err = memberProfile.Update(ctx, tx, scope)
				if err != nil {
					return fmt.Errorf("cannot update membership profile: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return identity, membership, nil
}

func (s *Service) validateAssertion(assertion *saml.Assertion, config *coredata.SAMLConfiguration, now time.Time) error {
	const clockSkewTolerance = 5 * time.Minute

	if assertion.ID == "" {
		return errors.New("assertion ID is required")
	}

	if assertion.Subject == nil || assertion.Subject.NameID == nil {
		return fmt.Errorf("subject or NameID missing")
	}

	if assertion.Issuer.Value != config.IdPEntityID {
		return fmt.Errorf("assertion issuer %q does not match expected issuer %q",
			assertion.Issuer.Value, config.IdPEntityID)
	}

	if assertion.Conditions == nil {
		return errors.New("assertion conditions are required")
	}

	if assertion.Conditions.NotOnOrAfter.IsZero() {
		return errors.New("assertion NotOnOrAfter condition is required")
	}

	if !assertion.Conditions.NotBefore.IsZero() {
		if now.Add(clockSkewTolerance).Before(assertion.Conditions.NotBefore) {
			return fmt.Errorf("assertion not yet valid (NotBefore: %v, now: %v)",
				assertion.Conditions.NotBefore, now)
		}
	}

	if now.Add(-clockSkewTolerance).After(assertion.Conditions.NotOnOrAfter) {
		return fmt.Errorf("assertion expired (NotOnOrAfter: %v, now: %v)",
			assertion.Conditions.NotOnOrAfter, now)
	}

	if len(assertion.Conditions.AudienceRestrictions) == 0 {
		return errors.New("assertion audience restriction is required")
	}

	expectedAudience := baseurl.MustParse(s.baseURL).WithPath("/api/connect/v1/saml/2.0/metadata").MustString()

	audienceValid := false
	for _, restriction := range assertion.Conditions.AudienceRestrictions {
		if restriction.Audience.Value == expectedAudience {
			audienceValid = true
			break
		}
	}

	if !audienceValid {
		return fmt.Errorf("assertion audience %q does not match expected %q",
			assertion.Conditions.AudienceRestrictions, expectedAudience)
	}

	return nil
}

func (s *Service) baseServiceProvider() *saml.ServiceProvider {
	baseURL := baseurl.MustParse(s.baseURL)
	metadataURL := baseURL.WithPath("/api/connect/v1/saml/2.0/metadata").URL()
	acsURL := baseURL.WithPath("/api/connect/v1/saml/2.0/consume").URL()

	return &saml.ServiceProvider{
		EntityID:          metadataURL.String(),
		Key:               s.privateKey,
		Certificate:       s.certificate,
		MetadataURL:       metadataURL,
		AcsURL:            acsURL,
		SloURL:            acsURL,
		AuthnNameIDFormat: saml.EmailAddressNameIDFormat,
		AllowIDPInitiated: true,
	}
}
