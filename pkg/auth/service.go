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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"time"

	"github.com/getprobo/probo/packages/emails"
	"github.com/getprobo/probo/pkg/baseurl"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/crypto/passwdhash"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Service struct {
		pg                      *pg.Client
		encryptionKey           cipher.EncryptionKey
		hp                      *passwdhash.Profile
		baseURL                 string
		tokenSecret             string
		disableSignup           bool
		invitationTokenValidity time.Duration
	}

	TenantAuthService struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
		hp            *passwdhash.Profile
		baseURL       string
		tokenSecret   string
		scope         coredata.Scoper
	}

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

func (e ErrSAMLAutoSignupDisabled) Error() string {
	return "SAML auto-signup is disabled for this organization"
}

func (e ErrSAMLAuthRequired) Error() string {
	return "SAML authentication required for this organization"
}

func (e ErrPasswordAuthRequired) Error() string {
	return "password authentication required for this organization"
}

func NewService(
	ctx context.Context,
	pgClient *pg.Client,
	encryptionKey cipher.EncryptionKey,
	hp *passwdhash.Profile,
	tokenSecret string,
	baseURL string,
	disableSignup bool,
	invitationTokenValidity time.Duration,
) (*Service, error) {
	return &Service{
		pg:                      pgClient,
		encryptionKey:           encryptionKey,
		hp:                      hp,
		baseURL:                 baseURL,
		tokenSecret:             tokenSecret,
		disableSignup:           disableSignup,
		invitationTokenValidity: invitationTokenValidity,
	}, nil
}

func (s *Service) WithTenant(tenantID gid.TenantID) *TenantAuthService {
	return &TenantAuthService{
		pg:            s.pg,
		encryptionKey: s.encryptionKey,
		hp:            s.hp,
		baseURL:       s.baseURL,
		tokenSecret:   s.tokenSecret,
		scope:         coredata.NewScope(tenantID),
	}
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

	base, err := baseurl.Parse(s.baseURL)
	if err != nil {
		return fmt.Errorf("cannot parse base URL: %w", err)
	}

	resetPasswordUrl, err := base.
		WithPath("/auth/reset-password").
		WithQuery("token", passwordResetToken).
		String()
	if err != nil {
		return fmt.Errorf("cannot build reset password URL: %w", err)
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
				s.baseURL,
				user.FullName,
				resetPasswordUrl,
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

	emailAddress2, err := mail.ParseAddress(emailAddress)
	if err != nil {
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
		EmailAddress:         emailAddress2.Address,
		HashedPassword:       hashedPassword,
		EmailAddressVerified: false,
		FullName:             fullName,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	session := &coredata.Session{
		ID:     gid.New(gid.NilTenant, coredata.SessionEntityType),
		UserID: user.ID,
		Data: coredata.SessionData{
			PasswordAuthenticated: true,
			SAMLAuthenticatedOrgs: make(map[string]coredata.SAMLAuthInfo),
		},
		ExpiredAt: now.Add(24 * time.Hour * 7), // 7 days,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := user.Insert(ctx, tx, coredata.NewNoScope()); err != nil {
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

			base, err := baseurl.Parse(s.baseURL)
			if err != nil {
				return fmt.Errorf("cannot parse base URL: %w", err)
			}

			confirmationUrl, err := base.
				WithPath("/auth/confirm-email").
				WithQuery("token", confirmationToken).
				String()
			if err != nil {
				return fmt.Errorf("cannot build confirmation URL: %w", err)
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

func (s Service) ProvisionSAMLUser(
	ctx context.Context,
	samlConfigID gid.GID,
	organizationID gid.GID,
	emailAddress string,
	fullName string,
	samlSubject string,
	existingSession *coredata.Session,
	sessionDuration time.Duration,
) (*coredata.Session, *coredata.User, error) {
	if _, err := mail.ParseAddress(emailAddress); err != nil {
		return nil, nil, &ErrInvalidEmail{emailAddress}
	}

	if fullName == "" {
		return nil, nil, &ErrInvalidFullName{fullName}
	}

	if samlSubject == "" {
		return nil, nil, fmt.Errorf("SAML subject cannot be empty")
	}

	user := &coredata.User{}
	session := &coredata.Session{}
	now := time.Now()

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := user.LoadByEmail(ctx, tx, emailAddress); err == nil {
				user.SAMLSubject = &samlSubject
				user.FullName = fullName
				user.EmailAddressVerified = true
				user.UpdatedAt = now

				if err := user.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update user: %w", err)
				}
			} else {
				var samlConfig coredata.SAMLConfiguration
				if err := samlConfig.LoadByID(ctx, tx, coredata.NewNoScope(), samlConfigID); err != nil {
					return fmt.Errorf("cannot load SAML configuration: %w", err)
				}

				if !samlConfig.AutoSignupEnabled {
					return &ErrSAMLAutoSignupDisabled{}
				}

				*user = coredata.User{
					ID:                   gid.New(gid.NilTenant, coredata.UserEntityType),
					EmailAddress:         emailAddress,
					HashedPassword:       nil,
					EmailAddressVerified: true,
					FullName:             fullName,
					SAMLSubject:          &samlSubject,
					CreatedAt:            now,
					UpdatedAt:            now,
				}

				if err := user.Insert(ctx, tx, coredata.NewNoScope()); err != nil {
					return fmt.Errorf("cannot insert SAML user: %w", err)
				}
			}

			if existingSession != nil && existingSession.UserID == user.ID {
				if err := session.LoadByID(ctx, tx, existingSession.ID); err != nil {
					return fmt.Errorf("cannot load session: %w", err)
				}

				if session.Data.SAMLAuthenticatedOrgs == nil {
					session.Data.SAMLAuthenticatedOrgs = make(map[string]coredata.SAMLAuthInfo)
				}
				session.Data.SAMLAuthenticatedOrgs[organizationID.String()] = coredata.SAMLAuthInfo{
					AuthenticatedAt: now,
					SAMLConfigID:    samlConfigID,
					SAMLSubject:     samlSubject,
				}
				session.UpdatedAt = now

				if err := session.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update session: %w", err)
				}
			} else {
				*session = coredata.Session{
					ID:     gid.New(gid.NilTenant, coredata.SessionEntityType),
					UserID: user.ID,
					Data: coredata.SessionData{
						SAMLAuthenticatedOrgs: map[string]coredata.SAMLAuthInfo{
							organizationID.String(): {
								AuthenticatedAt: now,
								SAMLConfigID:    samlConfigID,
								SAMLSubject:     samlSubject,
							},
						},
					},
					ExpiredAt: now.Add(sessionDuration),
					CreatedAt: now,
					UpdatedAt: now,
				}

				if err := session.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert session: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return session, user, nil
}

func (s Service) SignIn(
	ctx context.Context,
	emailAddress string,
	password string,
	existingSession *coredata.Session,
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

			if existingSession != nil && existingSession.UserID == user.ID {
				session = &coredata.Session{}
				if err := session.LoadByID(ctx, tx, existingSession.ID); err != nil {
					return fmt.Errorf("cannot load session: %w", err)
				}

				session.Data.PasswordAuthenticated = true
				if session.Data.SAMLAuthenticatedOrgs == nil {
					session.Data.SAMLAuthenticatedOrgs = make(map[string]coredata.SAMLAuthInfo)
				}
				session.UpdatedAt = time.Now()

				if err := session.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update session: %w", err)
				}
			} else {
				now := time.Now()
				session = &coredata.Session{
					ID:     gid.New(gid.NilTenant, coredata.SessionEntityType),
					UserID: user.ID,
					Data: coredata.SessionData{
						PasswordAuthenticated: true,
						SAMLAuthenticatedOrgs: make(map[string]coredata.SAMLAuthInfo),
					},
					ExpiredAt: now.Add(24 * time.Hour * 7), // 7 days
					CreatedAt: now,
					UpdatedAt: now,
				}

				if err := session.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert session: %w", err)
				}
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
			session.ExpiredAt = now.Add(24 * time.Hour * 7)
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

func (s Service) UpdateSessionData(ctx context.Context, sessionID gid.GID, data coredata.SessionData) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			session := &coredata.Session{}
			if err := session.LoadByID(ctx, tx, sessionID); err != nil {
				return &ErrSessionNotFound{"session not found"}
			}

			if time.Now().After(session.ExpiredAt) {
				return &ErrSessionExpired{"session expired"}
			}

			session.Data = data
			session.UpdatedAt = time.Now()
			if err := session.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update session: %w", err)
			}

			return nil
		},
	)
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

	if len(password) < 8 || len(password) > 128 {
		return nil, nil, &ErrInvalidPassword{minLength: 8, maxLength: 128}
	}

	if _, err := mail.ParseAddress(payload.Data.Email); err != nil {
		return nil, nil, &ErrInvalidEmail{payload.Data.Email}
	}

	if fullName == "" {
		fullName = payload.Data.FullName
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

	scope := coredata.NewScope(payload.Data.InvitationID.TenantID())
	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			invitation := &coredata.Invitation{}
			if err := invitation.LoadByID(ctx, tx, scope, payload.Data.InvitationID); err != nil {
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
				EmailAddress:         payload.Data.Email,
				HashedPassword:       hashedPassword,
				EmailAddressVerified: true,
				FullName:             fullName,
				CreatedAt:            now,
				UpdatedAt:            now,
			}

			if err := user.Insert(ctx, tx, coredata.NewNoScope()); err != nil {
				var errUserAlreadyExists *coredata.ErrUserAlreadyExists
				if errors.As(err, &errUserAlreadyExists) {
					return &ErrUserAlreadyExists{errUserAlreadyExists.Error()}
				}

				return fmt.Errorf("cannot insert user: %w", err)
			}

			session = &coredata.Session{
				ID:     gid.New(gid.NilTenant, coredata.SessionEntityType),
				UserID: user.ID,
				Data: coredata.SessionData{
					PasswordAuthenticated: true,
					SAMLAuthenticatedOrgs: make(map[string]coredata.SAMLAuthInfo),
				},
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

func (s *TenantAuthService) GetUserAuthMethod(
	ctx context.Context,
	userID gid.GID,
	organizationID gid.GID,
	session *coredata.Session,
) (coredata.UserAuthMethod, error) {
	var authMethod coredata.UserAuthMethod

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			user := &coredata.User{}
			if err := user.LoadByID(ctx, conn, userID); err != nil {
				return fmt.Errorf("cannot load user: %w", err)
			}

			// If user doesn't have a SAML subject, they only use password auth
			if user.SAMLSubject == nil || *user.SAMLSubject == "" {
				authMethod = coredata.UserAuthMethodPassword
				return nil
			}

			// User has SAML subject - check if there's SAML config for this org + user's domain
			domain := extractDomain(user.EmailAddress)
			if domain == "" {
				authMethod = coredata.UserAuthMethodPassword
				return nil
			}

			// Check if SAML is configured for this org + domain
			var samlConfig coredata.SAMLConfiguration
			err := samlConfig.LoadByOrganizationIDAndEmailDomain(ctx, conn, s.scope, organizationID, domain)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					authMethod = coredata.UserAuthMethodPassword
					return nil
				}
				return fmt.Errorf("cannot check SAML configuration: %w", err)
			}

			authMethod = coredata.UserAuthMethodSAML
			return nil
		},
	)

	return authMethod, err
}

func extractDomain(email string) string {
	atIndex := -1
	for i := 0; i < len(email); i++ {
		if email[i] == '@' {
			atIndex = i
			break
		}
	}
	if atIndex == -1 || atIndex == len(email)-1 {
		return ""
	}
	return email[atIndex+1:]
}

// CheckOrganizationAccess checks access to multiple organizations in a single database query
// Always uses batch processing to avoid N+1 query problems
func (s Service) CheckOrganizationAccess(
	ctx context.Context,
	user *coredata.User,
	organizationIDs []gid.GID,
	session *coredata.Session,
) (map[gid.GID]AccessResult, error) {
	results := make(map[gid.GID]AccessResult, len(organizationIDs))

	domain := extractDomain(user.EmailAddress)
	if domain == "" {
		// All organizations fail with invalid email
		for _, orgID := range organizationIDs {
			results[orgID] = AccessResult{
				OrganizationID: orgID,
				Allowed:        false,
				MissingAuth:    AuthMethodPassword,
				SAMLConfig:     nil,
			}
		}
		return results, nil
	}

	// Fetch all SAML configs for these organizations in a single query
	var samlConfigs map[gid.GID]*coredata.SAMLConfiguration
	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		var err error
		samlConfigs, err = coredata.LoadSAMLConfigurationsByOrganizationIDsAndEmailDomain(
			ctx,
			conn,
			organizationIDs,
			domain,
		)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("cannot load SAML configurations: %w", err)
	}

	// Apply business logic for each organization using pure function
	for _, orgID := range organizationIDs {
		requirement := OrgAuthRequirement{
			OrganizationID: orgID,
			EmailDomain:    domain,
			SAMLConfig:     samlConfigs[orgID], // May be nil
		}
		results[orgID] = requirement.Check(session.Data)
	}

	return results, nil
}

// CheckSingleOrganizationAccess is a convenience wrapper for checking access to a single organization
// It uses the batch method internally to maintain a single code path
func (s Service) CheckSingleOrganizationAccess(
	ctx context.Context,
	user *coredata.User,
	organizationID gid.GID,
	session *coredata.Session,
) error {
	results, err := s.CheckOrganizationAccess(ctx, user, []gid.GID{organizationID}, session)
	if err != nil {
		return err
	}

	result, ok := results[organizationID]
	if !ok {
		return fmt.Errorf("no access result for organization %s", organizationID)
	}

	return result.ToError(s.baseURL)
}

// BaseURL returns the base URL for the service
func (s Service) BaseURL() string {
	return s.baseURL
}

func (s Service) GetOrganizationLogoFile(
	ctx context.Context,
	user *coredata.User,
	organizationID gid.GID,
	session *coredata.Session,
) (*coredata.File, error) {
	// Check authentication requirements before allowing access to logo
	err := s.CheckSingleOrganizationAccess(ctx, user, organizationID, session)
	if err != nil {
		return nil, fmt.Errorf("access denied: %w", err)
	}

	var logoFile *coredata.File

	err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			scope := coredata.NewScope(organizationID.TenantID())

			var membership coredata.Membership
			err := membership.LoadByUserAndOrg(ctx, conn, scope, user.ID, organizationID)
			if err != nil {
				if _, ok := err.(coredata.ErrMembershipNotFound); ok {
					return fmt.Errorf("user does not have access to this organization")
				}

				return fmt.Errorf("cannot verify membership: %w", err)
			}

			var organization coredata.Organization
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if organization.LogoFileID == nil {
				return fmt.Errorf("organization has no logo")
			}

			var file coredata.File
			if err := file.LoadByID(ctx, conn, scope, *organization.LogoFileID); err != nil {
				return fmt.Errorf("cannot load logo file: %w", err)
			}

			logoFile = &file
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return logoFile, nil
}

func (s *TenantAuthService) InitiateDomainVerification(
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

func (s *TenantAuthService) VerifyDomain(
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
