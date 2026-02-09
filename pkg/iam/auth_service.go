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
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/packages/emails"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/statelesstoken"
	"go.probo.inc/probo/pkg/validator"
)

type (
	AuthService struct {
		*Service
	}

	ResetPasswordRequest struct {
		Token    string
		Password string
	}

	ChangePasswordRequest struct {
		CurrentPassword string
		NewPassword     string
	}

	CreateIdentityFromInvitationRequest struct {
		InvitationToken string
		Password        string
		FullName        string
	}

	LoadOrCreateIdentityRequest struct {
		Email    mail.Addr
		FullName string
	}

	CreateIdentityWithPasswordRequest struct {
		Email    mail.Addr
		Password string
		FullName string
	}

	SendMagicLinkRequest struct {
		Email          mail.Addr
		URLPath        string
		OrganizationID gid.GID
		// If users tries to connect to compliance page, we must brand the emails accordingly
		CompliancePageID *gid.GID
	}

	PasswordResetData struct {
		Email mail.Addr `json:"email"`
	}

	MagicLinkData struct {
		Email mail.Addr `json:"email"`
	}
)

const (
	TokenTypeOrganizationInvitation = "organization_invitation"
	TokenTypePasswordReset          = "password_reset"
	TokenTypeMagicLink              = "magic_link"
)

func NewAuthService(svc *Service) *AuthService {
	return &AuthService{Service: svc}
}

func (req CreateIdentityFromInvitationRequest) Validate() error {
	v := validator.New()

	v.Check(req.InvitationToken, "invitationToken", validator.NotEmpty())
	v.Check(req.FullName, "fullName", validator.NotEmpty(), validator.MinLen(1), validator.MaxLen(255))
	v.Check(req.Password, "password", PasswordValidator())

	return v.Error()
}

func (req ResetPasswordRequest) Validate() error {
	v := validator.New()
	v.Check(req.Token, "token", validator.NotEmpty())
	v.Check(req.Password, "password", PasswordValidator())
	return v.Error()
}

func (req ChangePasswordRequest) Validate() error {
	v := validator.New()

	// We cannot use PasswordValidator here because legacy password may not be aligned with the current password
	// policy, therefore we at least enforce a maximum length to mitigate DDoS attacks.
	v.Check(req.CurrentPassword, "currentPassword", validator.NotEmpty(), validator.MaxLen(255))

	v.Check(req.NewPassword, "newPassword", PasswordValidator())
	return v.Error()
}

func (req LoadOrCreateIdentityRequest) Validate() error {
	v := validator.New()

	v.Check(req.FullName, "fullName", validator.NotEmpty(), validator.MinLen(1), validator.MaxLen(255))

	return v.Error()
}

func (req CreateIdentityWithPasswordRequest) Validate() error {
	v := validator.New()

	v.Check(req.FullName, "fullName", validator.NotEmpty(), validator.MinLen(1), validator.MaxLen(255))
	v.Check(req.Password, "password", PasswordValidator())

	return v.Error()
}

func (s *AuthService) CreateIdentityFromInvitation(
	ctx context.Context,
	req *CreateIdentityFromInvitationRequest,
) (*coredata.Identity, *coredata.Session, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid request: %w", err)
	}

	payload, err := statelesstoken.ValidateToken[InvitationTokenData](s.tokenSecret, TokenTypeOrganizationInvitation, req.InvitationToken)
	if err != nil {
		return nil, nil, NewInvalidTokenError()
	}

	var (
		scope      = coredata.NewScopeFromObjectID(payload.Data.InvitationID)
		invitation = &coredata.Invitation{}
		identity   = &coredata.Identity{}
		session    = &coredata.Session{}
		now        = time.Now()
	)

	hashedPassword, err := s.hp.HashPassword([]byte(req.Password))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot hash password: %w", err)
	}

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			err := invitation.LoadByID(ctx, tx, scope, payload.Data.InvitationID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewInvitationNotFoundError(payload.Data.InvitationID)
				}

				return fmt.Errorf("cannot load invitation: %w", err)
			}

			if invitation.AcceptedAt != nil {
				return NewInvitationAlreadyAcceptedError(payload.Data.InvitationID)
			}

			if invitation.ExpiresAt.Before(now) {
				return NewInvitationExpiredError(payload.Data.InvitationID)
			}

			identity = &coredata.Identity{
				ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
				EmailAddress:         invitation.Email,
				FullName:             invitation.FullName,
				HashedPassword:       hashedPassword,
				EmailAddressVerified: true,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			err = identity.Insert(ctx, tx)
			if err != nil {
				if err == coredata.ErrResourceAlreadyExists {
					return NewIdentityAlreadyExistsError(invitation.Email)
				}

				return fmt.Errorf("cannot insert identity: %w", err)
			}

			session = coredata.NewRootSession(identity.ID, coredata.AuthMethodPassword, s.sessionDuration)
			err = session.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return identity, session, nil
}

func (s AuthService) ResetPassword(
	ctx context.Context,
	req *ResetPasswordRequest,
) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	payload, err := statelesstoken.ValidateToken[PasswordResetData](s.tokenSecret, TokenTypePasswordReset, req.Token)
	if err != nil {
		return NewInvalidTokenError()
	}

	hashedPassword, err := s.hp.HashPassword([]byte(req.Password))
	if err != nil {
		return fmt.Errorf("cannot hash password: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := &coredata.Identity{}
			err := identity.LoadByEmail(ctx, tx, payload.Data.Email)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return nil // Don't leak information about non-existent identities
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			identity.HashedPassword = hashedPassword
			identity.UpdatedAt = time.Now()

			err = identity.Update(ctx, tx)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return nil // Don't leak information about non-existent identities
				}

				return fmt.Errorf("cannot update identity: %w", err)
			}

			return nil
		},
	)
}

func (s AuthService) SendPasswordResetInstructionByEmail(
	ctx context.Context,
	email mail.Addr,
) error {
	token, err := statelesstoken.NewToken(
		s.tokenSecret,
		TokenTypePasswordReset,
		s.passwordResetTokenValidity,
		PasswordResetData{Email: email},
	)
	if err != nil {
		return fmt.Errorf("cannot generate password reset token: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity := &coredata.Identity{}
			if err := identity.LoadByEmail(ctx, tx, email); err != nil {
				if err == coredata.ErrResourceNotFound {
					return nil // Don't leak information about non-existent identities
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			emailPresenter := emails.NewPresenter(s.fm, s.bucket, s.baseURL, identity.FullName)

			subject, textBody, htmlBody, err := emailPresenter.RenderPasswordReset(
				ctx,
				"/auth/reset-password",
				token,
			)
			if err != nil {
				return fmt.Errorf("cannot render password reset email: %w", err)
			}

			passwordResetEmail := coredata.NewEmail(
				identity.FullName,
				identity.EmailAddress,
				subject,
				textBody,
				htmlBody,
			)

			err = passwordResetEmail.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)
}

func (s AuthService) LoadOrCreateIdentity(
	ctx context.Context,
	req *LoadOrCreateIdentityRequest,
) (*coredata.Identity, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	var (
		identity *coredata.Identity
		now      = time.Now()
	)

	if err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
		identity = &coredata.Identity{}

		if err := identity.LoadByEmail(ctx, tx, req.Email); err != nil {
			if !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			identity = &coredata.Identity{
				ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
				EmailAddress:         req.Email,
				FullName:             req.FullName,
				HashedPassword:       nil,
				EmailAddressVerified: false,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			err = identity.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert identity: %w", err)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return identity, nil
}

func (s AuthService) CreateIdentityWithPassword(
	ctx context.Context,
	req *CreateIdentityWithPasswordRequest,
) (*coredata.Identity, *coredata.Session, error) {
	if s.disableSignup { // TODO Rename this one to disableSignup
		return nil, nil, NewErrSignupDisabled()
	}

	if err := req.Validate(); err != nil {
		return nil, nil, fmt.Errorf("invalid request: %w", err)
	}

	hashedPassword, err := s.hp.HashPassword([]byte(req.Password))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot hash password: %w", err)
	}

	var (
		now = time.Now()

		identity = &coredata.Identity{
			ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
			EmailAddress:         req.Email,
			FullName:             req.FullName,
			HashedPassword:       hashedPassword,
			EmailAddressVerified: false,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		session = coredata.NewRootSession(identity.ID, coredata.AuthMethodPassword, 24*time.Hour*7)
	)

	confirmationToken, err := statelesstoken.NewToken(
		s.tokenSecret,
		TokenTypeEmailConfirmation,
		24*time.Hour,
		EmailConfirmationData{IdentityID: identity.ID, Email: identity.EmailAddress},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate confirmation token: %w", err)
	}

	emailPresenter := emails.NewPresenter(s.fm, s.bucket, s.baseURL, req.FullName)

	subject, textBody, htmlBody, err := emailPresenter.RenderConfirmEmail(ctx, "/auth/verify-email", confirmationToken)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot render confirmation email: %w", err)
	}

	confirmationEmail := coredata.NewEmail(
		req.FullName,
		identity.EmailAddress,
		subject,
		textBody,
		htmlBody,
	)

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			err := identity.Insert(ctx, tx)
			if err != nil {
				if err == coredata.ErrResourceAlreadyExists {
					return NewIdentityAlreadyExistsError(identity.EmailAddress)
				}

				return fmt.Errorf("cannot insert identity: %w", err)
			}

			if err := confirmationEmail.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			if err := session.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			return nil
		},
	)

	return identity, session, err
}

func (s AuthService) OpenSessionWithSAML(ctx context.Context, identityID gid.GID) (*coredata.Session, error) {
	session := &coredata.Session{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) (err error) {
			session = coredata.NewRootSession(identityID, coredata.AuthMethodSAML, s.sessionDuration)
			err = session.Insert(ctx, conn)
			if err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s AuthService) CheckCredentials(
	ctx context.Context,
	email mail.Addr,
	password string,
) (*coredata.Identity, error) {
	v := validator.New()
	v.Check(password, "password", PasswordValidator())

	err := v.Error()
	if err != nil {
		return nil, NewInvalidPasswordError("invalid password")
	}

	identity := &coredata.Identity{}

	err = s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			err := identity.LoadByEmail(ctx, conn, email)
			if err != nil {
				// Do not leak information about non-existent identities
				if err != coredata.ErrResourceNotFound {
					return fmt.Errorf("cannot load identity by email: %w", err)
				}
			}

			// Perform a password comparison even when the identity does not exist to mitigate timing attacks
			// and prevent revealing account existence.
			if identity.ID == gid.Nil {
				_, _ = s.hp.ComparePasswordAndHash([]byte(password+"qwertyuiop1234567890"), []byte("qwertyuiop1234567890"))
				return NewInvalidCredentialsError("invalid email or password")
			}

			isPasswordMatch, err := s.hp.ComparePasswordAndHash([]byte(password), identity.HashedPassword)
			if err != nil {
				return fmt.Errorf("cannot verify password: %w", err)
			}

			if !isPasswordMatch {
				return NewInvalidCredentialsError("invalid email or password")
			}

			return nil
		},
	)

	return identity, err
}

func (s AuthService) OpenSessionWithPassword(ctx context.Context, identityID gid.GID) (*coredata.Session, error) {
	session := &coredata.Session{}

	err := s.pg.WithTx(
		ctx,
		func(conn pg.Conn) (err error) {
			session = coredata.NewRootSession(identityID, coredata.AuthMethodPassword, s.sessionDuration)
			err = session.Insert(ctx, conn)
			if err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s AuthService) SendMagicLink(ctx context.Context, req *SendMagicLinkRequest) error {
	tokenString, err := statelesstoken.NewToken(
		s.tokenSecret,
		TokenTypeMagicLink,
		s.magicLinkTokenValidity,
		MagicLinkData{
			Email: req.Email,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot generate magic link token: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			hashedToken := HashToken(tokenString)
			token := &coredata.Token{
				ID:          gid.New(gid.NilTenant, coredata.TokenEntityType),
				HashedValue: hashedToken,
				CreatedAt:   time.Now(),
			}
			if err := token.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert token: %w", err)
			}

			fullName := req.Email.Username()
			identity := &coredata.Identity{}
			organization := &coredata.Organization{}

			if err := identity.LoadByEmail(ctx, tx, req.Email); err == nil {
				if identity.FullName != "" {
					fullName = identity.FullName
				}
			} else {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load identity: %w", err)
				}
			}

			if err := organization.LoadByID(ctx, tx, coredata.NewNoScope(), req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			emailPresenterCfg := emails.DefaultPresenterConfig(s.bucket, s.baseURL)
			if req.CompliancePageID != nil {
				var err error
				emailPresenterCfg, err = s.CompliancePageService.EmailPresenterConfig(ctx, *req.CompliancePageID)
				if err != nil {
					return fmt.Errorf("cannot get compliance page email presenter config: %w", err)
				}
			}

			emailPresenter := emails.NewPresenterFromConfig(s.fm, emailPresenterCfg, fullName)

			subject, textBody, htmlBody, err := emailPresenter.RenderMagicLink(
				ctx,
				req.URLPath,
				tokenString,
				s.magicLinkTokenValidity,
				organization.Name,
			)
			if err != nil {
				return fmt.Errorf("cannot render magic link email: %w", err)
			}

			magicLinkEmail := coredata.NewEmail(
				fullName,
				req.Email,
				subject,
				textBody,
				htmlBody,
			)

			if err := magicLinkEmail.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)
}

func (s AuthService) OpenSessionWithMagicLink(ctx context.Context, tokenString string) (*coredata.Identity, *coredata.Session, error) {
	var (
		now      = time.Now()
		identity = &coredata.Identity{}
		session  = &coredata.Session{}
	)

	payload, err := statelesstoken.ValidateToken[MagicLinkData](s.tokenSecret, TokenTypeMagicLink, tokenString)
	if err != nil {
		return nil, nil, NewInvalidTokenError()
	}

	if err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			hashedValue := HashToken(tokenString)
			token := &coredata.Token{}

			if err := s.pg.WithConn(
				ctx,
				func(conn pg.Conn) error {
					if err := token.LoadByHashedValueForUpdate(ctx, conn, hashedValue); err != nil {
						if errors.Is(err, coredata.ErrResourceNotFound) {
							return NewInvalidTokenError()
						}

						return fmt.Errorf("cannot load token by hashed value: %w", err)
					}

					return nil
				},
			); err != nil {
				return fmt.Errorf("cannot load token: %w", err)
			}

			err := identity.LoadByEmail(ctx, tx, payload.Data.Email)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					identity = &coredata.Identity{
						ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
						EmailAddress:         payload.Data.Email,
						EmailAddressVerified: true,
						CreatedAt:            now,
						UpdatedAt:            now,
					}

					if err := identity.Insert(ctx, tx); err != nil {
						return fmt.Errorf("cannot create identity: %w", err)
					}
				} else {
					return fmt.Errorf("cannot load identity by email: %w", err)
				}
			}

			session = coredata.NewRootSession(identity.ID, coredata.AuthMethodMagicLink, s.sessionDuration)
			err = session.Insert(ctx, tx)
			if err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			if err := token.Delete(ctx, tx); err != nil {
				return fmt.Errorf("cannot delete token: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	return identity, session, nil
}

func (s *AuthService) UpdateIdentity(ctx context.Context, identityID gid.GID, fullName string) (*coredata.Identity, error) {
	identity := &coredata.Identity{}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := identity.LoadByID(ctx, tx, identityID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot load identity: %w", err)
			}

			identity.FullName = fullName
			identity.UpdatedAt = time.Now()

			if err := identity.Update(ctx, tx); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewIdentityNotFoundError(identityID)
				}

				return fmt.Errorf("cannot update identity: %w", err)
			}

			return nil
		},
	)

	return identity, err
}

func HashToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))
	return hash[:]
}
