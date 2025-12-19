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
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/statelesstoken"
	"go.probo.inc/probo/pkg/validator"
)

type (
	AccountService struct {
		*Service
	}

	UserAPIKeyTokenData struct {
		Version     int       `json:"v"`
		KeyID       gid.GID   `json:"kid"`
		PrincipalID gid.GID   `json:"pid"`
		IssuedAt    time.Time `json:"iat"`
	}

	EmailConfirmationData struct {
		UserID gid.GID   `json:"uid"`
		Email  mail.Addr `json:"email"`
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
		EmailConfirmationData{UserID: identityID, Email: req.NewEmail},
	)
	if err != nil {
		return fmt.Errorf("cannot generate confirmation token: %w", err)
	}

	base, err := baseurl.Parse(s.baseURL)
	if err != nil {
		return fmt.Errorf("cannot parse base URL: %w", err)
	}

	confirmationUrl := base.
		WithPath("/auth/confirm-email").
		WithQuery("token", confirmationToken).
		MustString()

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			user := &coredata.User{}
			err := user.LoadByID(ctx, tx, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewUserNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load user: %w", err)
			}

			isPasswordMatch, err := s.hp.ComparePasswordAndHash([]byte(req.Password), user.HashedPassword)
			if err != nil {
				return fmt.Errorf("cannot compare password: %w", err)
			}

			if !isPasswordMatch {
				return NewInvalidPasswordError("invalid password")
			}

			user.EmailAddress = req.NewEmail
			user.EmailAddressVerified = false
			user.UpdatedAt = time.Now()

			err = user.Update(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot update user: %w", err)
			}

			subject, textBody, htmlBody, err := emails.RenderConfirmEmail(
				s.baseURL,
				user.FullName,
				confirmationUrl,
			)
			if err != nil {
				return fmt.Errorf("cannot render confirmation email: %w", err)
			}

			confirmationEmail := coredata.NewEmail(
				user.FullName,
				user.EmailAddress,
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
			user := &coredata.User{}
			err := user.LoadByID(ctx, tx, payload.Data.UserID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewUserNotFoundError(payload.Data.UserID)
				}

				return fmt.Errorf("cannot load user: %w", err)
			}

			if user.EmailAddress != payload.Data.Email {
				return NewEmailVerificationMismatchError()
			}

			if user.EmailAddressVerified {
				return NewEmailAlreadyVerifiedError()
			}

			user.EmailAddressVerified = true
			user.UpdatedAt = time.Now()

			err = user.Update(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot update user: %w", err)
			}

			return nil
		},
	)
}

func (s *AccountService) AcceptInvitation(
	ctx context.Context,
	identityID gid.GID,
	invitationID gid.GID,
) (*coredata.Membership, error) {
	var (
		now        = time.Now()
		membership = &coredata.Membership{}
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			user := coredata.User{}
			invitation := coredata.Invitation{}

			err := user.LoadByID(ctx, tx, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewUserNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load user: %w", err)
			}

			err = invitation.LoadByID(ctx, tx, coredata.NewNoScope(), invitationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewInvitationNotFoundError(invitationID)
				}

				return fmt.Errorf("cannot load invitation: %w", err)
			}

			if invitation.Email != user.EmailAddress {
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

			membership = &coredata.Membership{
				ID:             gid.New(tenantID, coredata.MembershipEntityType),
				UserID:         identityID,
				OrganizationID: invitation.OrganizationID,
				Role:           invitation.Role,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			err = membership.Insert(ctx, tx, scope)
			if err != nil {
				return fmt.Errorf("cannot create membership: %w", err)
			}

			invitation.AcceptedAt = &now
			err = invitation.Update(ctx, tx, scope)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewInvitationNotFoundError(invitationID)
				}

				return fmt.Errorf("cannot update invitation: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return membership, nil
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
			identity := coredata.User{}
			err := identity.LoadByID(ctx, conn, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewUserNotFoundError(identityID)
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
			identity := coredata.User{}
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
			err := memberships.LoadByUserID(ctx, conn, coredata.NewNoScope(), identityID, cursor)
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
			count, err = memberships.CountByUserID(ctx, conn, identityID)
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
			user := &coredata.User{}
			err := user.LoadByID(ctx, tx, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewUserNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load user: %w", err)
			}

			isLegacyPasswordMatch, err := s.hp.ComparePasswordAndHash([]byte(req.CurrentPassword), user.HashedPassword)
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

			user.HashedPassword = newPasswordHash
			user.UpdatedAt = time.Now()

			err = user.Update(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot update user: %w", err)
			}

			// TODO: email to notify user that their password has been changed

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
			count, err = sessions.CountByUserID(ctx, conn, identityID)
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
			err := sessions.LoadByUserID(ctx, conn, identityID, cursor)
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

func (s AccountService) GetIdentity(ctx context.Context, identityID gid.GID) (*coredata.User, error) {
	user := &coredata.User{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := user.LoadByID(ctx, conn, identityID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewUserNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load user: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s AccountService) ListPersonalAPIKeys(
	ctx context.Context,
	identityID gid.GID,
	cursor *page.Cursor[coredata.UserAPIKeyOrderField],
) (*page.Page[*coredata.UserAPIKey, coredata.UserAPIKeyOrderField], error) {
	var personalAccessTokens coredata.UserAPIKeys

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := personalAccessTokens.LoadByUserID(ctx, conn, identityID)
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
			personalAccessTokens := coredata.UserAPIKeys{}
			count, err = personalAccessTokens.CountByUserID(ctx, conn, identityID)
			if err != nil {
				return fmt.Errorf("cannot count personal access tokens: %w", err)
			}

			return nil
		},
	)

	return count, err
}

func (s AccountService) GetIdentityForMembership(ctx context.Context, membershipID gid.GID) (*coredata.User, error) {
	var (
		scope    = coredata.NewScopeFromObjectID(membershipID)
		identity = &coredata.User{}
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

			err = identity.LoadByID(ctx, conn, membership.UserID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewUserNotFoundError(membership.UserID)
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
) (*coredata.UserAPIKey, string, error) {
	var (
		userAPIKey *coredata.UserAPIKey
		token      string
	)

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) (err error) {
			now := time.Now()

			userAPIKey = &coredata.UserAPIKey{
				ID:        gid.New(gid.NilTenant, coredata.UserAPIKeyEntityType),
				UserID:    identityID,
				Name:      name,
				ExpiresAt: expiresAt,
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := userAPIKey.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert user api key: %w", err)
			}

			token, err = statelesstoken.NewDeterministicToken(
				s.tokenSecret,
				TokenTypeAPIKey,
				userAPIKey.ExpiresAt,
				userAPIKey.CreatedAt,
				UserAPIKeyTokenData{
					Version:     2,
					KeyID:       userAPIKey.ID,
					PrincipalID: identityID,
					IssuedAt:    userAPIKey.CreatedAt,
				},
			)
			if err != nil {
				return fmt.Errorf("cannot generate user api key token: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, "", err
	}

	return userAPIKey, token, nil
}

func (s *AccountService) DeletePersonalAPIKey(
	ctx context.Context,
	identityID gid.GID,
	userAPIKeyID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			userAPIKey := &coredata.UserAPIKey{}
			err := userAPIKey.LoadByID(ctx, tx, userAPIKeyID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewUserAPIKeyNotFoundError(userAPIKeyID)
				}

				return fmt.Errorf("cannot load user api key: %w", err)
			}

			if userAPIKey.UserID != identityID {
				return NewUserAPIKeyNotFoundError(userAPIKeyID)
			}

			err = userAPIKey.Delete(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot delete user api key: %w", err)
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
			err := organizations.LoadByUserID(ctx, conn, coredata.NewNoScope(), identityID, cursor)
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

// func (s AccountService) AllAccessibleTenants(ctx context.Context, identityID gid.GID) ([]gid.TenantID, error) {
// 	var tenants []gid.TenantID

// 	err := s.pg.WithConn(
// 		ctx,
// 		func(conn pg.Conn) error {
// 			memberships := coredata.Memberships{}
// 			orderBy := page.OrderBy[coredata.MembershipOrderField]{
// 				Field:     coredata.MembershipOrderFieldCreatedAt,
// 				Direction: page.OrderDirectionDesc,
// 			}
// 			cursor := page.NewCursor(1000, nil, page.Head, orderBy)

// 			err := memberships.LoadByUserID(ctx, conn, coredata.NewNoScope(), identityID, cursor)
// 			if err != nil {
// 				return fmt.Errorf("cannot load memberships: %w", err)
// 			}

// 			for _, membership := range memberships {
// 				tenants = append(tenants, membership.ID.TenantID())
// 			}
// 			return nil
// 		},
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return tenants, nil
// }
