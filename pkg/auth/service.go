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
	"net/url"
	"time"

	"github.com/getprobo/probo/packages/emails"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/crypto/passwdhash"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	// Service handles ONLY user authentication and management
	// No organization-related logic - that belongs to authz service
	Service struct {
		pg                      *pg.Client
		encryptionKey           cipher.EncryptionKey
		hp                      *passwdhash.Profile
		hostname                string
		baseURL                 string
		tokenSecret             string
		disableSignup           bool
		invitationTokenValidity time.Duration
	}

	// TenantAuthService handles tenant-scoped authentication operations
	TenantAuthService struct {
		pg            *pg.Client
		encryptionKey cipher.EncryptionKey
		hp            *passwdhash.Profile
		hostname      string
		baseURL       string
		tokenSecret   string
		scope         coredata.Scoper
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
	encryptionKey cipher.EncryptionKey,
	hp *passwdhash.Profile,
	tokenSecret string,
	hostname string,
	baseURL string,
	disableSignup bool,
	invitationTokenValidity time.Duration,
) (*Service, error) {
	return &Service{
		pg:                      pgClient,
		encryptionKey:           encryptionKey,
		hp:                      hp,
		hostname:                hostname,
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
		hostname:      s.hostname,
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
				s.hostname,
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

			confirmationUrl := url.URL{
				Scheme: "https",
				Host:   s.hostname,
				Path:   "/auth/confirm-email",
				RawQuery: url.Values{
					"token": []string{confirmationToken},
				}.Encode(),
			}

			subject, textBody, htmlBody, err := emails.RenderConfirmEmail(
				s.hostname,
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

func (s Service) CreateOrGetSAMLUser(
	ctx context.Context,
	emailAddress string,
	fullName string,
	samlSubject string,
) (*coredata.User, error) {
	if _, err := mail.ParseAddress(emailAddress); err != nil {
		return nil, &ErrInvalidEmail{emailAddress}
	}

	if fullName == "" {
		return nil, &ErrInvalidFullName{fullName}
	}

	if samlSubject == "" {
		return nil, fmt.Errorf("SAML subject cannot be empty")
	}

	var user coredata.User
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

				return nil
			}

			user = coredata.User{
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

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s Service) CreateSessionForUser(
	ctx context.Context,
	userID gid.GID,
	sessionDuration time.Duration,
) (*coredata.Session, error) {
	now := time.Now()

	session := &coredata.Session{
		ID:        gid.New(gid.NilTenant, coredata.SessionEntityType),
		UserID:    userID,
		Data:      coredata.SessionData{},
		ExpiredAt: now.Add(sessionDuration),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := session.Insert(ctx, tx); err != nil {
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
			// Load user by email (all users are global now)
			if err := user.LoadByEmail(ctx, tx, emailAddress); err != nil {
				var errUserNotFound *coredata.ErrUserNotFound
				if errors.As(err, &errUserNotFound) {
					return &ErrInvalidCredentials{"invalid email or password"}
				}
				return fmt.Errorf("cannot load user by email: %w", err)
			}

			// Verify password
			match, err := s.hp.ComparePasswordAndHash([]byte(password), user.HashedPassword)
			if err != nil {
				return fmt.Errorf("cannot verify password: %w", err)
			}
			if !match {
				return &ErrInvalidCredentials{"invalid email or password"}
			}

			// Create new session with password authentication flag set
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

			return nil
		},
	)

	if err != nil {
		return nil, nil, err
	}

	return session, user, nil
}

func (s Service) SignInWithExistingSession(
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
					ExpiredAt: now.Add(24 * time.Hour * 7),
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

// IsTenantUser removed - all users are now global (no tenant distinction)

func (s Service) GetUserAuthMethod(
	ctx context.Context,
	scope coredata.Scoper,
	userID gid.GID,
	organizationID gid.GID,
	session *coredata.Session,
) (coredata.UserAuthMethod, error) {
	// Load the user to check their email and SAML subject
	user := &coredata.User{}
	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return user.LoadByID(ctx, conn, userID)
	})
	if err != nil {
		return "", fmt.Errorf("cannot load user: %w", err)
	}

	// If user doesn't have a SAML subject, they only use password auth
	if user.SAMLSubject == nil || *user.SAMLSubject == "" {
		return coredata.UserAuthMethodPassword, nil
	}

	// User has SAML subject - check if there's SAML config for this org + user's domain
	// Extract domain from user email
	emailParts := []byte(user.EmailAddress)
	atIndex := -1
	for i, b := range emailParts {
		if b == '@' {
			atIndex = i
			break
		}
	}
	if atIndex == -1 {
		return coredata.UserAuthMethodPassword, nil
	}
	domain := string(emailParts[atIndex+1:])

	// Check if SAML is configured for this org + domain
	var samlConfig coredata.SAMLConfiguration
	orgScope := coredata.NewScope(organizationID.TenantID())
	err = s.pg.WithConn(ctx, func(conn pg.Conn) error {
		err := samlConfig.LoadByOrganizationIDAndEmailDomain(ctx, conn, orgScope, organizationID, domain)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil // No SAML config for this org+domain
			}
			return err
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("cannot check SAML configuration: %w", err)
	}

	// If SAML config exists for this org+domain, user enrolled via SAML
	if samlConfig.ID != (gid.GID{}) {
		return coredata.UserAuthMethodSAML, nil
	}

	// No SAML config for this org, user uses password
	return coredata.UserAuthMethodPassword, nil
}

// Organization Access Control

type (
	// ErrSAMLAuthRequired indicates user must authenticate via SAML to access org
	ErrSAMLAuthRequired struct {
		ConfigID       gid.GID
		OrganizationID gid.GID
		RedirectURL    string // SAML IdP login URL
	}

	// ErrPasswordAuthRequired indicates user must authenticate with password to access org
	ErrPasswordAuthRequired struct {
		OrganizationID gid.GID
		RedirectURL    string // Password login page URL
	}
)

func (e ErrSAMLAuthRequired) Error() string {
	return "SAML authentication required for this organization"
}

func (e ErrPasswordAuthRequired) Error() string {
	return "password authentication required for this organization"
}

// CheckOrganizationAccess determines if a user can access an organization
// based on SAML configuration and session authentication state
func (s Service) CheckOrganizationAccess(
	ctx context.Context,
	user *coredata.User,
	organizationID gid.GID,
	session *coredata.Session,
) error {
	// Extract domain from user email
	emailParts := []byte(user.EmailAddress)
	atIndex := -1
	for i, b := range emailParts {
		if b == '@' {
			atIndex = i
			break
		}
	}
	if atIndex == -1 {
		return fmt.Errorf("invalid email address format")
	}
	domain := string(emailParts[atIndex+1:])

	// Find SAML configuration for this organization and domain
	var samlConfig coredata.SAMLConfiguration
	scope := coredata.NewScope(organizationID.TenantID())
	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		err := samlConfig.LoadByOrganizationIDAndEmailDomain(ctx, conn, scope, organizationID, domain)
		if err != nil {
			// If no SAML config found for this organization and domain, that's okay - not an error
			// Just means this organization doesn't have SAML configured for this domain
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("cannot check SAML configuration: %w", err)
	}

	// Check if SAML is configured and enabled for this domain and organization
	if samlConfig.ID != (gid.GID{}) && samlConfig.Enabled && samlConfig.DomainVerified {
		// SAML config exists for this org - check enforcement policy
		if samlConfig.EnforcementPolicy == coredata.SAMLEnforcementPolicyRequired {
			// SAML is REQUIRED - check if user has SAML-authenticated for this org
			authInfo, hasSAMLAuth := session.Data.SAMLAuthenticatedOrgs[organizationID.String()]
			if !hasSAMLAuth {
				// Build SAML login URL
				samlLoginURL := fmt.Sprintf("%s/auth/saml/login/%s", s.baseURL, samlConfig.ID)
				return ErrSAMLAuthRequired{
					ConfigID:       samlConfig.ID,
					OrganizationID: organizationID,
					RedirectURL:    samlLoginURL,
				}
			}

			// Optional: Check if SAML auth is still recent (not too old)
			// For now, we trust the session lifetime
			_ = authInfo
		} else {
			// SAML is OPTIONAL or OFF - allow either password OR SAML auth for this specific org
			hasSAMLAuth := false
			if _, ok := session.Data.SAMLAuthenticatedOrgs[organizationID.String()]; ok {
				hasSAMLAuth = true
			}

			if !session.Data.PasswordAuthenticated && !hasSAMLAuth {
				// User needs to authenticate - offer SAML as option
				samlLoginURL := fmt.Sprintf("%s/auth/saml/login/%s", s.baseURL, samlConfig.ID)
				return ErrSAMLAuthRequired{
					ConfigID:       samlConfig.ID,
					OrganizationID: organizationID,
					RedirectURL:    samlLoginURL,
				}
			}
		}
	} else {
		// No SAML configuration for this org+domain combination
		// Require password authentication for password-only organizations
		if !session.Data.PasswordAuthenticated {
			// User hasn't authenticated with password - require password authentication
			loginURL := fmt.Sprintf("%s/authentication/login?method=password", s.baseURL)
			return ErrPasswordAuthRequired{
				OrganizationID: organizationID,
				RedirectURL:    loginURL,
			}
		}
	}

	return nil // Access granted
}

// InitiateDomainVerification creates a SAML configuration with unverified domain and generates verification token
func (s Service) InitiateDomainVerification(
	ctx context.Context,
	tenantID gid.TenantID,
	organizationID gid.GID,
	emailDomain string,
) (*coredata.SAMLConfiguration, error) {
	token, err := GenerateDomainVerificationToken()
	if err != nil {
		return nil, fmt.Errorf("cannot generate verification token: %w", err)
	}

	var config *coredata.SAMLConfiguration

	err = s.pg.WithTx(ctx, func(tx pg.Conn) error {
		now := time.Now()
		scope := coredata.NewScope(tenantID)

		config = &coredata.SAMLConfiguration{
			ID:                      gid.New(tenantID, coredata.SAMLConfigurationEntityType),
			OrganizationID:          organizationID,
			EmailDomain:             emailDomain,
			Enabled:                 false,
			EnforcementPolicy:       coredata.SAMLEnforcementPolicyOff,
			DomainVerified:          false,
			DomainVerificationToken: &token,
			// Default IdP values (placeholders until configured)
			IdPEntityID:    "not-configured",
			IdPSsoURL:      "not-configured",
			IdPCertificate: "not-configured",
			// Default attribute mappings
			AttributeEmail:     "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
			AttributeFirstname: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
			AttributeLastname:  "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
			AttributeRole:      "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/role",
			AutoSignupEnabled:  false,
			CreatedAt:          now,
			UpdatedAt:          now,
		}

		if err := config.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert SAML configuration: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return config, nil
}

// VerifyDomain checks DNS TXT record and marks domain as verified if found
func (s Service) VerifyDomain(
	ctx context.Context,
	tenantID gid.TenantID,
	configID gid.GID,
) (*coredata.SAMLConfiguration, bool, error) {
	var config *coredata.SAMLConfiguration
	var verified bool

	err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
		scope := coredata.NewScope(tenantID)

		// Load config
		config = &coredata.SAMLConfiguration{}
		if err := config.LoadByID(ctx, tx, scope, configID); err != nil {
			return fmt.Errorf("cannot load SAML configuration: %w", err)
		}

		if config.DomainVerificationToken == nil {
			return fmt.Errorf("no verification token found for this configuration")
		}

		if config.DomainVerified {
			verified = true
			return nil // Already verified
		}

		// Check DNS TXT record
		isVerified, err := VerifyDomainOwnership(ctx, config.EmailDomain, *config.DomainVerificationToken)
		if err != nil {
			return fmt.Errorf("cannot verify domain ownership: %w", err)
		}

		verified = isVerified

		if isVerified {
			now := time.Now()
			config.DomainVerified = true
			config.DomainVerifiedAt = &now
			config.UpdatedAt = now

			if err := config.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update SAML configuration: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, false, err
	}

	return config, verified, nil
}

// Domain Verification Methods

// GenerateDomainVerificationToken generates a random 32-character hex token for domain verification
func GenerateDomainVerificationToken() (string, error) {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("cannot generate domain verification token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GetDomainVerificationRecord returns the DNS TXT record string that should be added to the domain
func GetDomainVerificationRecord(token string) string {
	return fmt.Sprintf("probo-verification=%s", token)
}

// VerifyDomainOwnership performs DNS lookup to verify domain ownership via TXT record
func VerifyDomainOwnership(ctx context.Context, domain, expectedToken string) (bool, error) {
	// Use net package for DNS TXT record lookup
	var txtRecords []string
	var err error

	// Create a DNS resolver with timeout from context
	resolver := &net.Resolver{
		PreferGo: true,
	}

	txtRecords, err = resolver.LookupTXT(ctx, domain)
	if err != nil {
		// DNS lookup errors are expected if the domain doesn't exist or has no TXT records
		// We return false (not verified) but not an error, as this is a normal case
		return false, nil
	}

	// Check if any TXT record matches our verification token
	expectedRecord := GetDomainVerificationRecord(expectedToken)
	for _, record := range txtRecords {
		if record == expectedRecord {
			return true, nil
		}
	}

	// Token not found in DNS records
	return false, nil
}
