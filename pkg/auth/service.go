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
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"time"

	"github.com/getprobo/probo/packages/emails"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/passwdhash"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"go.gearno.de/kit/pg"
)

type (
	// Service handles ONLY user authentication and management
	// No organization-related logic - that belongs to authz service
	Service struct {
		pg                      *pg.Client
		hp                      *passwdhash.Profile
		hostname                string
		tokenSecret             string
		disableSignup           bool
		invitationTokenValidity time.Duration
	}

	ErrInvalidCredentials struct {
		message string
	}

	ErrInvalidEmail struct {
		email string
	}

	ErrInvalidPassword struct {
		minLength int
		maxLength int
	}

	ErrInvalidFullName struct {
		fullName string
	}

	ErrUserAlreadyExists struct {
		message string
	}

	ErrSessionNotFound struct {
		message string
	}

	ErrSessionExpired struct {
		message string
	}

	ErrInvalidTokenType struct {
		message string
	}

	ErrSignupDisabled struct{}

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
)

const (
	TokenTypeEmailConfirmation = "email_confirmation"
	TokenTypePasswordReset     = "password_reset"
)

func (e ErrInvalidCredentials) Error() string {
	return e.message
}

func (e ErrUserAlreadyExists) Error() string {
	return e.message
}

func (e ErrSessionNotFound) Error() string {
	return e.message
}

func (e ErrSessionExpired) Error() string {
	return e.message
}

func (e ErrInvalidEmail) Error() string {
	return fmt.Sprintf("invalid email: %s", e.email)
}

func (e ErrInvalidPassword) Error() string {
	return fmt.Sprintf("invalid password: the length must be between %d and %d characters", e.minLength, e.maxLength)
}

func (e ErrInvalidFullName) Error() string {
	return fmt.Sprintf("invalid full name: %s", e.fullName)
}

func (e ErrInvalidTokenType) Error() string {
	return e.message
}

func (e ErrSignupDisabled) Error() string {
	return "signup is disabled, contact the owner of the Probo instance"
}

func NewService(
	ctx context.Context,
	pgClient *pg.Client,
	hp *passwdhash.Profile,
	tokenSecret string,
	hostname string,
	disableSignup bool,
	invitationTokenValidity time.Duration,
) (*Service, error) {
	return &Service{
		pg:                      pgClient,
		hp:                      hp,
		hostname:                hostname,
		tokenSecret:             tokenSecret,
		disableSignup:           disableSignup,
		invitationTokenValidity: invitationTokenValidity,
	}, nil
}

func (s Service) ForgetPassword(
	ctx context.Context,
	email string,
) error {
	// Always generate a new token to avoid timing attacks and leaking information
	// about existing emails
	passwordResetToken, err := statelesstoken.NewToken(
		s.tokenSecret,
		TokenTypePasswordReset,
		1*time.Hour,
		PasswordResetData{Email: email},
	)
	if err != nil {
		return fmt.Errorf("cannot generate password reset token: %w", err)
	}

	resetPasswordUrl := url.URL{
		Scheme: "https",
		Host:   s.hostname,
		Path:   "/auth/reset-password",
		RawQuery: url.Values{
			"token": []string{passwordResetToken},
		}.Encode(),
	}

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			user := &coredata.User{}
			if err := user.LoadByEmail(ctx, conn, email); err != nil {
				var errUserNotFound *coredata.ErrUserNotFound
				if errors.As(err, &errUserNotFound) {
					return nil // Don't leak information about non-existent users
				}
				return fmt.Errorf("cannot load user: %w", err)
			}

			subject, textBody, htmlBody, err := emails.RenderPasswordReset(
				user.FullName,
				resetPasswordUrl.String(),
			)
			if err != nil {
				return fmt.Errorf("cannot render password reset email: %w", err)
			}

			passwordResetEmail := coredata.NewEmail(
				user.FullName,
				email,
				subject,
				textBody,
				htmlBody,
			)
			if err := passwordResetEmail.Insert(ctx, conn); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)
}

func (s Service) SignUp(
	ctx context.Context,
	emailAddress string,
	password string,
	fullName string,
) (*coredata.User, *coredata.Session, error) {
	if s.disableSignup {
		return nil, nil, &ErrSignupDisabled{}
	}

	if _, err := mail.ParseAddress(emailAddress); err != nil {
		return nil, nil, &ErrInvalidEmail{emailAddress}
	}

	if len(password) < 8 || len(password) > 128 {
		return nil, nil, &ErrInvalidPassword{minLength: 8, maxLength: 128}
	}

	if fullName == "" {
		return nil, nil, &ErrInvalidFullName{fullName}
	}

	hashedPassword, err := s.hp.HashPassword([]byte(password))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot hash password: %w", err)
	}

	now := time.Now()
	user := &coredata.User{
		ID:                   gid.New(gid.NilTenant, coredata.UserEntityType),
		EmailAddress:         emailAddress,
		HashedPassword:       hashedPassword,
		EmailAddressVerified: false,
		FullName:             fullName,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	session := &coredata.Session{
		ID:        gid.New(gid.NilTenant, coredata.SessionEntityType),
		UserID:    user.ID,
		Data:      coredata.SessionData{},
		ExpiredAt: now.Add(24 * time.Hour * 7), // 7 days,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := user.Insert(ctx, tx); err != nil {
				var errUserAlreadyExists *coredata.ErrUserAlreadyExists
				if errors.As(err, &errUserAlreadyExists) {
					return &ErrUserAlreadyExists{errUserAlreadyExists.Error()}
				}
				return fmt.Errorf("cannot insert user: %w", err)
			}

			confirmationToken, err := statelesstoken.NewToken(
				s.tokenSecret,
				TokenTypeEmailConfirmation,
				24*time.Hour,
				EmailConfirmationData{UserID: user.ID, Email: user.EmailAddress},
			)
			if err != nil {
				return fmt.Errorf("cannot generate confirmation token: %w", err)
			}

			confirmationUrl := url.URL{
				Scheme: "https",
				Host:   s.hostname,
				Path:   "/auth/confirm-email",
				RawQuery: url.Values{
					"token": []string{confirmationToken},
				}.Encode(),
			}

			subject, textBody, htmlBody, err := emails.RenderConfirmEmail(
				user.FullName,
				confirmationUrl.String(),
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

			if err := confirmationEmail.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			if err := session.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return user, session, nil
}

func (s Service) SignIn(
	ctx context.Context,
	emailAddress string,
	password string,
) (*coredata.Session, *coredata.User, error) {
	if _, err := mail.ParseAddress(emailAddress); err != nil {
		return nil, nil, &ErrInvalidCredentials{"invalid email or password"}
	}

	user := &coredata.User{}
	session := &coredata.Session{}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := user.LoadByEmail(ctx, tx, emailAddress); err != nil {
				var errUserNotFound *coredata.ErrUserNotFound
				if errors.As(err, &errUserNotFound) {
					return &ErrInvalidCredentials{"invalid email or password"}
				}
				return fmt.Errorf("cannot load user by email: %w", err)
			}

			match, err := s.hp.ComparePasswordAndHash([]byte(password), user.HashedPassword)
			if err != nil {
				return fmt.Errorf("cannot verify password: %w", err)
			}
			if !match {
				return &ErrInvalidCredentials{"invalid email or password"}
			}

			now := time.Now()
			session = &coredata.Session{
				ID:        gid.New(gid.NilTenant, coredata.SessionEntityType),
				UserID:    user.ID,
				Data:      coredata.SessionData{},
				ExpiredAt: now.Add(24 * time.Hour * 7), // 7 days
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := session.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return session, user, nil
}

func (s Service) SignOut(ctx context.Context, sessionID gid.GID) error {
	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			session := &coredata.Session{}
			if err := session.LoadByID(ctx, conn, sessionID); err != nil {
				return &ErrSessionNotFound{"session not found"}
			}

			if err := coredata.DeleteSession(ctx, conn, sessionID); err != nil {
				return fmt.Errorf("cannot delete session: %w", err)
			}

			return nil
		},
	)
}

func (s Service) GetSession(ctx context.Context, sessionID gid.GID) (*coredata.Session, error) {
	session := &coredata.Session{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := session.LoadByID(ctx, conn, sessionID); err != nil {
				return &ErrSessionNotFound{"session not found"}
			}

			if time.Now().After(session.ExpiredAt) {
				// Clean up expired session
				_ = coredata.DeleteSession(ctx, conn, sessionID)
				return &ErrSessionExpired{"session expired"}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s Service) GetUserByID(ctx context.Context, userID gid.GID) (*coredata.User, error) {
	user := &coredata.User{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := user.LoadByID(ctx, conn, userID); err != nil {
				return fmt.Errorf("cannot load user by ID: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s Service) GetUserByEmail(ctx context.Context, email string) (*coredata.User, error) {
	user := &coredata.User{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := user.LoadByEmail(ctx, conn, email); err != nil {
				return fmt.Errorf("cannot load user by email: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s Service) GetUserBySession(ctx context.Context, sessionID gid.GID) (*coredata.User, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return s.GetUserByID(ctx, session.UserID)
}

func (s Service) UpdateSession(ctx context.Context, sessionID gid.GID) (*coredata.Session, error) {
	session := &coredata.Session{}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := session.LoadByID(ctx, tx, sessionID); err != nil {
				return &ErrSessionNotFound{"session not found"}
			}

			if time.Now().After(session.ExpiredAt) {
				return &ErrSessionExpired{"session expired"}
			}

			now := time.Now()
			session.ExpiredAt = now.Add(24 * time.Hour * 7) // Extend by 7 days
			session.UpdatedAt = now
			if err := session.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s Service) ConfirmEmail(ctx context.Context, tokenString string) error {
	payload, err := statelesstoken.ValidateToken[EmailConfirmationData](
		s.tokenSecret,
		TokenTypeEmailConfirmation,
		tokenString,
	)
	if err != nil {
		return &ErrInvalidTokenType{"invalid confirmation token"}
	}
	emailConfirmationData := payload.Data

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			user := &coredata.User{}
			if err := user.LoadByID(ctx, tx, emailConfirmationData.UserID); err != nil {
				return fmt.Errorf("cannot load user: %w", err)
			}

			if user.EmailAddressVerified {
				return nil
			}

			if err := user.UpdateEmailVerification(ctx, tx, true); err != nil {
				return fmt.Errorf("cannot update user email verification: %w", err)
			}

			return nil
		},
	)
}

func (s Service) ResetPassword(ctx context.Context, tokenString string, newPassword string) error {
	payload, err := statelesstoken.ValidateToken[PasswordResetData](
		s.tokenSecret,
		TokenTypePasswordReset,
		tokenString,
	)
	if err != nil {
		return &ErrInvalidTokenType{"invalid reset token"}
	}
	passwordResetData := payload.Data

	if len(newPassword) < 8 || len(newPassword) > 128 {
		return &ErrInvalidPassword{minLength: 8, maxLength: 128}
	}

	hashedPassword, err := s.hp.HashPassword([]byte(newPassword))
	if err != nil {
		return fmt.Errorf("cannot hash password: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			user := &coredata.User{}
			if err := user.LoadByEmail(ctx, tx, passwordResetData.Email); err != nil {
				var errUserNotFound *coredata.ErrUserNotFound
				if errors.As(err, &errUserNotFound) {
					return nil // Don't leak information about non-existent users
				}
				return fmt.Errorf("cannot load user: %w", err)
			}

			if err := user.UpdatePassword(ctx, tx, hashedPassword); err != nil {
				return fmt.Errorf("cannot update password: %w", err)
			}

			return nil
		},
	)
}

func (s Service) SignupFromInvitation(
	ctx context.Context,
	token string,
	password string,
	fullName string,
) (*coredata.User, *coredata.Session, error) {
	payload, err := statelesstoken.ValidateToken[coredata.InvitationData](
		s.tokenSecret,
		"organization_invitation",
		token,
	)
	if err != nil {
		return nil, nil, &ErrInvalidTokenType{"invalid invitation token"}
	}
	invitationData := payload.Data

	if len(password) < 8 || len(password) > 128 {
		return nil, nil, &ErrInvalidPassword{minLength: 8, maxLength: 128}
	}

	if _, err := mail.ParseAddress(invitationData.Email); err != nil {
		return nil, nil, &ErrInvalidEmail{invitationData.Email}
	}

	if fullName == "" {
		fullName = invitationData.FullName
	}

	if fullName == "" {
		return nil, nil, &ErrInvalidFullName{fullName}
	}

	hashedPassword, err := s.hp.HashPassword([]byte(password))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot hash password: %w", err)
	}

	var user *coredata.User
	var session *coredata.Session

	scope := coredata.NewScope(invitationData.InvitationID.TenantID())
	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			invitation := &coredata.Invitation{}
			if err := invitation.LoadByID(ctx, tx, scope, invitationData.InvitationID); err != nil {
				var errInvitationNotFound *coredata.ErrInvitationNotFound
				if errors.As(err, &errInvitationNotFound) {
					return fmt.Errorf("invitation was deleted or no longer exists")
				}
				return fmt.Errorf("cannot load invitation: %w", err)
			}

			if invitation.AcceptedAt != nil {
				return fmt.Errorf("invitation already accepted")
			}

			if time.Now().After(invitation.ExpiresAt) {
				return fmt.Errorf("invitation expired")
			}

			now := time.Now()
			user = &coredata.User{
				ID:                   gid.New(gid.NilTenant, coredata.UserEntityType),
				EmailAddress:         invitationData.Email,
				HashedPassword:       hashedPassword,
				EmailAddressVerified: true,
				FullName:             fullName,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			if err := user.Insert(ctx, tx); err != nil {
				var errUserAlreadyExists *coredata.ErrUserAlreadyExists
				if errors.As(err, &errUserAlreadyExists) {
					return &ErrUserAlreadyExists{errUserAlreadyExists.Error()}
				}
				return fmt.Errorf("cannot insert user: %w", err)
			}

			session = &coredata.Session{
				ID:        gid.New(gid.NilTenant, coredata.SessionEntityType),
				UserID:    user.ID,
				Data:      coredata.SessionData{},
				ExpiredAt: now.Add(24 * time.Hour * 7),
				CreatedAt: now,
				UpdatedAt: now,
			}

			if err := session.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return user, session, nil
}
