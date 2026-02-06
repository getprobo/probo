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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/securetoken"
	"go.probo.inc/probo/pkg/statelesstoken"
	"go.probo.inc/probo/pkg/validator"
)

type (
	AccountService struct {
		*Service
	}

	PersonalAPIKeyTokenData struct {
		Version     int       `json:"v"`
		KeyID       gid.GID   `json:"kid"`
		PrincipalID gid.GID   `json:"pid"`
		IssuedAt    time.Time `json:"iat"`
	}

	EmailConfirmationData struct {
		IdentityID gid.GID   `json:"uid"`
		Email      mail.Addr `json:"email"`
	}
)

const (
	TokenTypeEmailConfirmation = "email_confirmation"
)

func NewAccountService(svc *Service) *AccountService {
	return &AccountService{Service: svc}
}

type ChangeEmailRequest struct {
	NewEmail mail.Addr
	Password string
}

func (req ChangeEmailRequest) Validate() error {
	v := validator.New()

	v.Check(req.Password, "password", validator.NotEmpty(), validator.MaxLen(255)) // We cannot use PasswordValidator here because legacy password may not be aligned with the current password policy, therefore we at least enforce a maximum length to mitigate DDoS attacks.

	return v.Error()
}

func (s AccountService) ChangeEmail(ctx context.Context, identityID gid.GID, req *ChangeEmailRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	confirmationToken, err := statelesstoken.NewToken(
		s.tokenSecret,
		TokenTypeEmailConfirmation,
		24*time.Hour,
		EmailConfirmationData{IdentityID: identityID, Email: req.NewEmail},
	)
	if err != nil {
		return fmt.Errorf("cannot generate confirmation token: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := &coredata.Identity{}
			err := identity.LoadByID(ctx, tx, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			isPasswordMatch, err := s.hp.ComparePasswordAndHash([]byte(req.Password), identity.HashedPassword)
			if err != nil {
				return fmt.Errorf("cannot compare password: %w", err)
			}

			if !isPasswordMatch {
				return NewInvalidPasswordError("invalid password")
			}

			identity.EmailAddress = req.NewEmail
			identity.EmailAddressVerified = false
			identity.UpdatedAt = time.Now()

			err = identity.Update(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot update identity: %w", err)
			}

			emailPresenter := emails.NewPresenter(s.fm, s.bucket, s.baseURL, identity.FullName)

			subject, textBody, htmlBody, err := emailPresenter.RenderConfirmEmail(ctx, "/auth/verify-email", confirmationToken)
			if err != nil {
				return fmt.Errorf("cannot render confirmation email: %w", err)
			}

			confirmationEmail := coredata.NewEmail(
				identity.FullName,
				identity.EmailAddress,
				subject,
				textBody,
				htmlBody,
			)

			err = confirmationEmail.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert confirmation email: %w", err)
			}

			return nil
		},
	)
}

func (s AccountService) VerifyEmail(ctx context.Context, token string) error {
	payload, err := statelesstoken.ValidateToken[EmailConfirmationData](s.tokenSecret, TokenTypeEmailConfirmation, token)
	if err != nil {
		return NewInvalidTokenError()
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := &coredata.Identity{}
			err := identity.LoadByID(ctx, tx, payload.Data.IdentityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewIdentityNotFoundError(payload.Data.IdentityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			if identity.EmailAddress != payload.Data.Email {
				return NewEmailVerificationMismatchError()
			}

			if identity.EmailAddressVerified {
				return NewEmailAlreadyVerifiedError()
			}

			identity.EmailAddressVerified = true
			identity.UpdatedAt = time.Now()

			err = identity.Update(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot update identity: %w", err)
			}

			return nil
		},
	)
}

func (s *AccountService) AcceptInvitation(
	ctx context.Context,
	identityID gid.GID,
	invitationID gid.GID,
) (*coredata.Invitation, *coredata.Membership, error) {
	var (
		now        = time.Now()
		membership = &coredata.Membership{}
		invitation = &coredata.Invitation{}
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := coredata.Identity{}

			if err := identity.LoadByID(ctx, tx, identityID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			if err := invitation.LoadByID(ctx, tx, coredata.NewNoScope(), invitationID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewInvitationNotFoundError(invitationID)
				}

				return fmt.Errorf("cannot load invitation: %w", err)
			}

			if invitation.Email != identity.EmailAddress {
				return NewInvitationNotFoundError(invitationID)
			}

			if invitation.AcceptedAt != nil {
				return NewInvitationAlreadyAcceptedError(invitationID)
			}

			if invitation.ExpiresAt.Before(now) {
				return NewInvitationExpiredError(invitationID)
			}

			tenantID := invitation.OrganizationID.TenantID()
			scope := coredata.NewScope(invitation.OrganizationID.TenantID())

			existingMembership := &coredata.Membership{}
			if err := existingMembership.LoadByIdentityAndOrg(
				ctx,
				tx,
				scope,
				identityID,
				invitation.OrganizationID,
			); err != nil && err != coredata.ErrResourceNotFound {
				return fmt.Errorf("cannot load existing membership: %w", err)
			}

			if existingMembership.ID != gid.Nil && existingMembership.State == coredata.MembershipStateInactive {
				existingMembership.State = coredata.MembershipStateActive
				existingMembership.Role = invitation.Role
				existingMembership.UpdatedAt = now

				if err := existingMembership.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot reactivate membership: %w", err)
				}

				membership = existingMembership
			} else {
				membership = &coredata.Membership{
					ID:             gid.New(tenantID, coredata.MembershipEntityType),
					IdentityID:     identityID,
					OrganizationID: invitation.OrganizationID,
					Role:           invitation.Role,
					Source:         coredata.MembershipSourceManual,
					State:          coredata.MembershipStateActive,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := membership.Insert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot create membership: %w", err)
				}

				profile := &coredata.MembershipProfile{
					ID:             gid.New(tenantID, coredata.MembershipProfileEntityType),
					IdentityID:     identity.ID,
					OrganizationID: invitation.OrganizationID,
					MembershipID:   membership.ID,
					FullName:       identity.FullName,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := profile.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert profile: %w", err)
				}
			}

			invitation.AcceptedAt = &now
			if err := invitation.Update(ctx, tx, scope); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewInvitationNotFoundError(invitationID)
				}

				return fmt.Errorf("cannot update invitation: %w", err)
			}

			// Expire other pending invitations for email in organization
			invitations := &coredata.Invitations{}
			onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
			if err := invitations.ExpireByEmailAndOrganization(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(invitation.OrganizationID),
				invitation.Email,
				invitation.OrganizationID,
				onlyPending,
			); err != nil {
				return fmt.Errorf("cannot expire pending invitations by email: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return invitation, membership, nil
}

func (s *AccountService) ListPendingInvitations(
	ctx context.Context,
	identityID gid.GID,
	cursor *page.Cursor[coredata.InvitationOrderField],
) (*page.Page[*coredata.Invitation, coredata.InvitationOrderField], error) {
	var invitations coredata.Invitations

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			identity := coredata.Identity{}
			err := identity.LoadByID(ctx, conn, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})

			err = invitations.LoadByIdentityID(ctx, conn, coredata.NewNoScope(), identity.EmailAddress, cursor, onlyPending)
			if err != nil {
				return fmt.Errorf("cannot load invitations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(invitations, cursor), nil
}

func (s *AccountService) CountPendingInvitations(
	ctx context.Context,
	identityID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			identity := coredata.Identity{}
			err := identity.LoadByID(ctx, conn, identityID)
			if err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			invitations := coredata.Invitations{}
			onlyPending := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})

			count, err = invitations.CountByEmail(ctx, conn, identity.EmailAddress, onlyPending)
			if err != nil {
				return fmt.Errorf("cannot count pending invitations: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *AccountService) ListMemberships(
	ctx context.Context,
	identityID gid.GID,
	cursor *page.Cursor[coredata.MembershipOrderField],
) (*page.Page[*coredata.Membership, coredata.MembershipOrderField], error) {
	var memberships coredata.Memberships

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := memberships.LoadByIdentityID(ctx, conn, coredata.NewNoScope(), identityID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load memberships: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(memberships, cursor), nil
}

func (s *AccountService) CountMemberships(
	ctx context.Context,
	identityID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			memberships := coredata.Memberships{}
			count, err = memberships.CountByIdentityID(ctx, conn, identityID)
			if err != nil {
				return fmt.Errorf("cannot count memberships: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s AccountService) ChangePassword(ctx context.Context, identityID gid.GID, req *ChangePasswordRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := &coredata.Identity{}
			err := identity.LoadByID(ctx, tx, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			isLegacyPasswordMatch, err := s.hp.ComparePasswordAndHash([]byte(req.CurrentPassword), identity.HashedPassword)
			if err != nil {
				return fmt.Errorf("cannot compare legacy password: %w", err)
			}

			if !isLegacyPasswordMatch {
				return NewInvalidPasswordError("invalid current password")
			}

			newPasswordHash, err := s.hp.HashPassword([]byte(req.NewPassword))
			if err != nil {
				return fmt.Errorf("cannot hash new password: %w", err)
			}

			identity.HashedPassword = newPasswordHash
			identity.UpdatedAt = time.Now()

			err = identity.Update(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot update identity: %w", err)
			}

			// TODO: email to notify identity that their password has been changed

			return nil
		},
	)
}

func (s AccountService) CountSessions(ctx context.Context, identityID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			sessions := coredata.Sessions{}
			count, err = sessions.CountByIdentityID(ctx, conn, identityID)
			if err != nil {
				return fmt.Errorf("cannot count sessions: %w", err)
			}

			return nil
		},
	)

	return count, err
}

func (s AccountService) ListSessions(
	ctx context.Context,
	identityID gid.GID,
	cursor *page.Cursor[coredata.SessionOrderField],
) (*page.Page[*coredata.Session, coredata.SessionOrderField], error) {
	var sessions coredata.Sessions

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := sessions.LoadByIdentityID(ctx, conn, identityID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load sessions: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(sessions, cursor), nil
}

func (s AccountService) GetIdentity(ctx context.Context, identityID gid.GID) (*coredata.Identity, error) {
	identity := &coredata.Identity{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := identity.LoadByID(ctx, conn, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return identity, nil
}

func (s AccountService) ListPersonalAPIKeys(
	ctx context.Context,
	identityID gid.GID,
	cursor *page.Cursor[coredata.PersonalAPIKeyOrderField],
) (*page.Page[*coredata.PersonalAPIKey, coredata.PersonalAPIKeyOrderField], error) {
	var personalAccessTokens coredata.PersonalAPIKeys

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := personalAccessTokens.LoadByIdentityID(ctx, conn, identityID)
			if err != nil {
				return fmt.Errorf("cannot load personal access tokens: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(personalAccessTokens, cursor), nil
}

func (s AccountService) CountPersonalAPIKeys(ctx context.Context, identityID gid.GID) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			personalAccessTokens := coredata.PersonalAPIKeys{}
			count, err = personalAccessTokens.CountByIdentityID(ctx, conn, identityID)
			if err != nil {
				return fmt.Errorf("cannot count personal access tokens: %w", err)
			}

			return nil
		},
	)

	return count, err
}
func (s *AccountService) RevealPersonalAPIKeyToken(
	ctx context.Context,
	identityID gid.GID,
	personalAPIKeyID gid.GID,
) (string, error) {
	var token string

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) (err error) {
			personalAPIKey := &coredata.PersonalAPIKey{}
			if err := personalAPIKey.LoadByID(ctx, tx, personalAPIKeyID); err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewPersonalAPIKeyNotFoundError(personalAPIKeyID)
				}
				return fmt.Errorf("cannot load personal api key: %w", err)
			}

			if personalAPIKey.IdentityID != identityID {
				return NewPersonalAPIKeyNotFoundError(personalAPIKeyID)
			}

			token, err = securetoken.Sign(
				personalAPIKey.ID.String(),
				s.tokenSecret,
			)
			if err != nil {
				return fmt.Errorf("cannot generate personal api key token: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s AccountService) GetIdentityForMembership(ctx context.Context, membershipID gid.GID) (*coredata.Identity, error) {
	var (
		scope    = coredata.NewScopeFromObjectID(membershipID)
		identity = &coredata.Identity{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			membership := &coredata.Membership{}
			err := membership.LoadByID(ctx, conn, scope, membershipID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(membershipID)
				}

				return fmt.Errorf("cannot load membership: %w", err)
			}

			err = identity.LoadByID(ctx, conn, membership.IdentityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewIdentityNotFoundError(membership.IdentityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return identity, nil
}

func (s *AccountService) CreatePersonalAPIKey(
	ctx context.Context,
	identityID gid.GID,
	name string,
	expiresAt time.Time,
) (*coredata.PersonalAPIKey, string, error) {
	var (
		personalAPIKey *coredata.PersonalAPIKey
		token          string
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) (err error) {
			now := time.Now()

			personalAPIKey = &coredata.PersonalAPIKey{
				ID:         gid.New(gid.NilTenant, coredata.PersonalAPIKeyEntityType),
				IdentityID: identityID,
				Name:       name,
				ExpiresAt:  expiresAt,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			if err := personalAPIKey.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert personal api key: %w", err)
			}

			token, err = securetoken.Sign(
				personalAPIKey.ID.String(),
				s.tokenSecret,
			)
			if err != nil {
				return fmt.Errorf("cannot generate personal api key token: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, "", err
	}

	return personalAPIKey, token, nil
}

func (s *AccountService) DeletePersonalAPIKey(
	ctx context.Context,
	identityID gid.GID,
	personalAPIKeyID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			personalAPIKey := &coredata.PersonalAPIKey{}
			err := personalAPIKey.LoadByID(ctx, tx, personalAPIKeyID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewPersonalAPIKeyNotFoundError(personalAPIKeyID)
				}

				return fmt.Errorf("cannot load personal api key: %w", err)
			}

			if personalAPIKey.IdentityID != identityID {
				return NewPersonalAPIKeyNotFoundError(personalAPIKeyID)
			}

			err = personalAPIKey.Delete(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot delete personal api key: %w", err)
			}

			return nil
		},
	)
}

func (s AccountService) ListOrganizations(ctx context.Context, identityID gid.GID) ([]*coredata.Organization, error) {
	var organizations coredata.Organizations
	orderBy := page.OrderBy[coredata.OrganizationOrderField]{
		Field:     coredata.OrganizationOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	}
	cursor := page.NewCursor(1000, nil, page.Head, orderBy)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := organizations.LoadByIdentityID(ctx, conn, coredata.NewNoScope(), identityID, cursor)
			if err != nil {
				return fmt.Errorf("cannot load organizations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organizations, nil
}

func (s AccountService) GetMembershipForOrganization(
	ctx context.Context,
	identityID gid.GID,
	organizationID gid.GID,
) (*coredata.Membership, error) {
	membership := &coredata.Membership{}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := &coredata.Identity{}

			if err := identity.LoadByID(ctx, tx, identityID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load identity %q: %w", identityID, err)
			}

			if err := membership.LoadByIdentityInOrganization(ctx, tx, identityID, organizationID); err != nil {
				return fmt.Errorf("cannot load membership: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return membership, nil
}

func (s AccountService) GetProfileForMembership(ctx context.Context, membershipID gid.GID) (*coredata.MembershipProfile, error) {
	var (
		scope   = coredata.NewScopeFromObjectID(membershipID)
		profile = &coredata.MembershipProfile{}
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			membership := &coredata.Membership{}
			err := membership.LoadByID(ctx, conn, scope, membershipID)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewMembershipNotFoundError(membershipID)
				}

				return fmt.Errorf("cannot load membership: %w", err)
			}

			err = profile.LoadByMembershipID(ctx, conn, scope, membershipID)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewProfileNotFoundError(membershipID)
				}

				return fmt.Errorf("cannot load membership profile: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (s AccountService) ListSAMLConfigurationsForEmail(
	ctx context.Context,
	email mail.Addr,
) (coredata.SAMLConfigurations, error) {
	samlConfigurations := coredata.SAMLConfigurations{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := samlConfigurations.LoadVerifiedByEmailDomain(ctx, conn, email.Domain())
			if err != nil {
				return fmt.Errorf("cannot load saml configurations: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return samlConfigurations, nil
}

func (s AccountService) CountSAMLConfigurationsForEmail(
	ctx context.Context,
	email mail.Addr,
) (int, error) {
	var (
		count              int
		samlConfigurations coredata.SAMLConfigurations
	)

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			count, err = samlConfigurations.CountVerifiedByEmailDomain(ctx, conn, email.Domain())
			if err != nil {
				return fmt.Errorf("cannot count saml configurations: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}
