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

package iam

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/statelesstoken"
)

type (
	ErrCreateOrganizationDisabled struct{}

	OrganizationAccessResponse struct {
		OrganizationID gid.GID
		HasAccess      bool
		Error          error
		SAMLConfig     *coredata.SAMLConfiguration
	}

	ErrInvalidCredentials struct {
		message string
	}

	ErrInvalidEmail struct {
		email string
	}

	ErrInvalidFullName struct {
		fullName string
	}

	ErrSignupDisabled struct{}

	ErrSAMLAutoSignupDisabled struct{}

	EmailConfirmationData struct {
		UserID gid.GID `json:"uid"`
		Email  string  `json:"email"`
	}

	InvitationData struct {
		OrganizationID gid.GID `json:"organization_id"`
		Email          string  `json:"email"`
		FullName       string  `json:"full_name"`
		CreatePeople   bool    `json:"create_people"`
	}
	PasswordResetData struct {
		Email string `json:"email"`
	}

	ErrSAMLAuthRequired struct {
		ConfigID       gid.GID
		OrganizationID gid.GID
		RedirectURL    string
	}

	ErrPasswordAuthRequired struct {
		OrganizationID gid.GID
		RedirectURL    string
	}

	UserAPIKeyMembershipRequest struct {
		MembershipID gid.GID
		Role         coredata.APIRole
	}

	UserAPIKeyOrganizationRequest struct {
		OrganizationID gid.GID
		Role           coredata.APIRole
	}

	UserAPIKeyWithMembershipsResponse struct {
		UserAPIKey  *coredata.UserAPIKey
		Memberships []*coredata.UserAPIKeyMembership
	}
)

const (
	TokenTypeEmailConfirmation = "email_confirmation"
	TokenTypePasswordReset     = "password_reset"
)

func (e ErrInvalidCredentials) Error() string {
	return e.message
}

func (e ErrInvalidEmail) Error() string {
	return fmt.Sprintf("invalid email: %s", e.email)
}

func (e ErrInvalidFullName) Error() string {
	return fmt.Sprintf("invalid full name: %s", e.fullName)
}

func (e ErrSignupDisabled) Error() string {
	return "signup is disabled, contact the owner of the Probo instance"
}

func (e ErrSAMLAutoSignupDisabled) Error() string {
	return "SAML auto-signup is disabled for this organization"
}

func (e ErrCreateOrganizationDisabled) Error() string {
	return "organization creation is disabled for users without existing admin or owner membership"
}

func (e ErrSAMLAuthRequired) Error() string {
	return "SAML authentication required for this organization"
}

func (e ErrPasswordAuthRequired) Error() string {
	return "password authentication required for this organization"
}

func (s *TenantService) InitiateDomainVerification(
	ctx context.Context,
	organizationID gid.GID,
	emailDomain string,
) (*coredata.SAMLConfiguration, error) {
	token, err := generateDomainVerificationToken()
	if err != nil {
		return nil, fmt.Errorf("cannot generate verification token: %w", err)
	}

	var config *coredata.SAMLConfiguration

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			now := time.Now()

			config = &coredata.SAMLConfiguration{
				ID:                      gid.New(s.scope.GetTenantID(), coredata.SAMLConfigurationEntityType),
				OrganizationID:          organizationID,
				EmailDomain:             emailDomain,
				Enabled:                 false,
				EnforcementPolicy:       coredata.SAMLEnforcementPolicyOff,
				DomainVerified:          false,
				DomainVerificationToken: &token,
				IdPEntityID:             "",
				IdPSsoURL:               "",
				IdPCertificate:          "",
				AttributeEmail:          "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
				AttributeFirstname:      "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
				AttributeLastname:       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
				AttributeRole:           "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
				AutoSignupEnabled:       false,
				CreatedAt:               now,
				UpdatedAt:               now,
			}

			if err := config.Insert(ctx, tx, s.scope); err != nil {
				return fmt.Errorf("cannot insert SAML configuration: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return config, nil
}

func (s *TenantService) VerifyDomain(
	ctx context.Context,
	configID gid.GID,
) (*coredata.SAMLConfiguration, bool, error) {
	var config *coredata.SAMLConfiguration
	var verified bool

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			config = &coredata.SAMLConfiguration{}
			if err := config.LoadByID(ctx, tx, s.scope, configID); err != nil {
				return fmt.Errorf("cannot load SAML configuration: %w", err)
			}

			if config.DomainVerificationToken == nil {
				return fmt.Errorf("no verification token found for this configuration")
			}

			if config.DomainVerified {
				verified = true
				return nil
			}

			isVerified, err := verifyDomainOwnership(ctx, config.EmailDomain, *config.DomainVerificationToken)
			if err != nil {
				return fmt.Errorf("cannot verify domain ownership: %w", err)
			}

			verified = isVerified

			if isVerified {
				now := time.Now()
				config.DomainVerified = true
				config.DomainVerifiedAt = &now
				config.UpdatedAt = now

				if err := config.Update(ctx, tx, s.scope); err != nil {
					return fmt.Errorf("cannot update SAML configuration: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, false, err
	}

	return config, verified, nil
}

func generateDomainVerificationToken() (string, error) {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("cannot generate domain verification token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func GetDomainVerificationRecord(token string) string {
	return fmt.Sprintf("probo-verification=%s", token)
}

func verifyDomainOwnership(ctx context.Context, domain, expectedToken string) (bool, error) {
	var txtRecords []string
	var err error

	resolver := &net.Resolver{PreferGo: true}

	txtRecords, err = resolver.LookupTXT(ctx, domain)
	if err != nil {
		return false, nil
	}

	expectedRecord := GetDomainVerificationRecord(expectedToken)
	for _, record := range txtRecords {
		if record == expectedRecord {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) ValidateAndBuildUserAPIKeyMemberships(
	ctx context.Context,
	userID gid.GID,
	organizationRequests []UserAPIKeyOrganizationRequest,
) ([]UserAPIKeyMembershipRequest, error) {
	memberships := make([]UserAPIKeyMembershipRequest, 0, len(organizationRequests))

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			for _, org := range organizationRequests {
				tenantID := org.OrganizationID.TenantID()

				var membership coredata.Membership
				if err := membership.LoadByUserAndOrg(ctx, conn, coredata.NewScope(tenantID), userID, org.OrganizationID); err != nil {
					return fmt.Errorf("you do not have access to organization %s", org.OrganizationID)
				}

				var role coredata.APIRole
				switch org.Role {
				case coredata.APIRoleFull:
					role = coredata.APIRoleFull
				default:
					return fmt.Errorf("invalid role: %s", org.Role)
				}

				memberships = append(memberships, UserAPIKeyMembershipRequest{
					MembershipID: membership.ID,
					Role:         role,
				})
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return memberships, nil
}

func (s *Service) GetUserAPIKey(
	ctx context.Context,
	userAPIKeyID gid.GID,
	userID gid.GID,
) (string, error) {
	var token string

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			userAPIKey := &coredata.UserAPIKey{}
			if err := userAPIKey.LoadByID(ctx, conn, userAPIKeyID); err != nil {
				return fmt.Errorf("cannot load user api key: %w", err)
			}

			if userAPIKey.UserID != userID {
				return fmt.Errorf("user api key does not belong to user")
			}

			tokenData := UserAPIKeyTokenData{
				ID:        userAPIKey.ID,
				CreatedAt: userAPIKey.CreatedAt,
			}

			generatedToken, err := statelesstoken.NewDeterministicToken(
				s.tokenSecret,
				TokenTypeAPIKey,
				userAPIKey.ExpiresAt,
				userAPIKey.CreatedAt,
				tokenData,
			)
			if err != nil {
				return fmt.Errorf("cannot generate user api key token: %w", err)
			}

			token = generatedToken
			return nil
		},
	)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) UpdateUserAPIKeyMemberships(
	ctx context.Context,
	userAPIKeyID gid.GID,
	userID gid.GID,
	memberships []UserAPIKeyMembershipRequest,
) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			userAPIKey := &coredata.UserAPIKey{}
			if err := userAPIKey.LoadByID(ctx, tx, userAPIKeyID); err != nil {
				return fmt.Errorf("cannot load user api key: %w", err)
			}

			if userAPIKey.UserID != userID {
				return fmt.Errorf("user api key does not belong to user")
			}

			if err := coredata.DeleteAllUserAPIKeyMembershipsByUserAPIKeyID(ctx, tx, userAPIKeyID); err != nil {
				return fmt.Errorf("cannot delete existing memberships: %w", err)
			}

			now := time.Now()
			for _, membership := range memberships {
				scope := coredata.NewScope(membership.MembershipID.TenantID())

				var m coredata.Membership
				if err := m.LoadByID(ctx, tx, scope, membership.MembershipID); err != nil {
					return fmt.Errorf("cannot load membership: %w", err)
				}

				userAPIKeyMembership := &coredata.UserAPIKeyMembership{
					ID:             gid.New(membership.MembershipID.TenantID(), coredata.UserAPIKeyMembershipEntityType),
					UserAPIKeyID:   userAPIKey.ID,
					MembershipID:   membership.MembershipID,
					Role:           membership.Role,
					OrganizationID: m.OrganizationID,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := userAPIKeyMembership.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot insert user api key membership: %w", err)
				}
			}

			return nil
		},
	)
}

func (s *Service) AddAPIKeyMembershipToOrganization(
	ctx context.Context,
	tenantID gid.TenantID,
	userAPIKeyID gid.GID,
	membershipID gid.GID,
	organizationID gid.GID,
	role coredata.APIRole,
) error {
	scope := coredata.NewScope(tenantID)
	now := time.Now()

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			userAPIKeyMembership := &coredata.UserAPIKeyMembership{
				ID:             gid.New(tenantID, coredata.UserAPIKeyMembershipEntityType),
				UserAPIKeyID:   userAPIKeyID,
				MembershipID:   membershipID,
				Role:           role,
				OrganizationID: organizationID,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := userAPIKeyMembership.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert user api key membership: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) UpdateUserAPIKeyName(
	ctx context.Context,
	userAPIKeyID gid.GID,
	userID gid.GID,
	name string,
) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			userAPIKey := &coredata.UserAPIKey{}
			if err := userAPIKey.LoadByID(ctx, tx, userAPIKeyID); err != nil {
				return fmt.Errorf("cannot load user api key: %w", err)
			}

			if userAPIKey.UserID != userID {
				return fmt.Errorf("user api key does not belong to user")
			}

			userAPIKey.Name = name
			userAPIKey.UpdatedAt = time.Now()

			if err := userAPIKey.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update user api key: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) ValidateUserAPIKey(
	ctx context.Context,
	token string,
) (*coredata.User, *coredata.UserAPIKey, error) {
	payload, err := statelesstoken.ValidateToken[UserAPIKeyTokenData](
		s.tokenSecret,
		TokenTypeAPIKey,
		token,
	)
	if err != nil {
		return nil, nil, &ErrInvalidCredentials{message: "invalid user api key"}
	}

	tokenData := payload.Data

	var user *coredata.User
	var userAPIKey *coredata.UserAPIKey

	err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			userAPIKey = &coredata.UserAPIKey{}
			if err := userAPIKey.LoadByID(ctx, conn, tokenData.ID); err != nil {
				var errNotFound *coredata.ErrUserAPIKeyNotFound
				if errors.As(err, &errNotFound) {
					return &ErrInvalidCredentials{message: "invalid user api key"}
				}
				return fmt.Errorf("cannot load user api key: %w", err)
			}

			if !userAPIKey.CreatedAt.Equal(tokenData.CreatedAt) {
				return &ErrInvalidCredentials{message: "invalid user api key"}
			}

			if time.Now().After(userAPIKey.ExpiresAt) {
				return &ErrInvalidCredentials{message: "user api key expired"}
			}

			user = &coredata.User{}
			if err := user.LoadByID(ctx, conn, userAPIKey.UserID); err != nil {
				return fmt.Errorf("cannot load user: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return user, userAPIKey, nil
}
