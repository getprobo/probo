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

package usrmgr

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/passwdhash"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/managederror"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
)

type (
	Service struct {
		pg                      *pg.Client
		hp                      *passwdhash.Profile
		hostname                string
		tokenSecret             string
		disableSignup           bool
		invitationTokenValidity time.Duration
	}

	EmailConfirmationData struct {
		UserID gid.GID `json:"uid"`
		Email  string  `json:"email"`
	}

	InvitationData struct {
		OrganizationID gid.GID `json:"organization_id"`
		Email          string  `json:"email"`
		FullName       string  `json:"full_name"`
	}

	PasswordResetData struct {
		Email string `json:"email"`
	}
)

// Token types
const (
	TokenTypeEmailConfirmation      = "email_confirmation"
	TokenTypeOrganizationInvitation = "organization_invitation"
	TokenTypePasswordReset          = "password_reset"
)

var (
	signupEmailSubject  = "Confirm your email address"
	signupEmailTemplate = `
	Thanks joining Probo!
	Please confirm your email address by clicking the link below[1]

	[1] %s
	`

	invitationEmailSubject  = "Join Probo"
	invitationEmailTemplate = `
	You have been invited to join Probo!
	Please click the link below to sign up[1]

	[1] %s
	`

	passwordResetEmailSubject  = "Reset your password"
	passwordResetEmailTemplate = `
	You have requested a password reset for your Probo account.
	Please click the link below to reset your password[1]

	If you did not request this password reset, please ignore this email.

	[1] %s
	`
)

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

	user := &coredata.User{}

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := user.LoadByEmail(ctx, tx, email); err != nil {
				var errUserNotFound *coredata.ErrUserNotFound

				if errors.As(err, &errUserNotFound) {
					// We don't want to leak information about existing emails
					// Return success even if the email doesn't exist
					return nil
				}
				return fmt.Errorf("cannot load user by %q email: %w", email, err)
			}

			resetPasswordEmail := coredata.NewEmail(
				user.FullName,
				user.EmailAddress,
				passwordResetEmailSubject,
				fmt.Sprintf(passwordResetEmailTemplate, resetPasswordUrl.String()),
			)

			if err := resetPasswordEmail.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) SignUp(
	ctx context.Context,
	email, password, fullName string,
) (*coredata.User, *coredata.Session, error) {
	if s.disableSignup {
		return nil, nil, managederror.NewOperationNotAllowedError("signup is disabled")
	}

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, nil, managederror.NewInvalidInputError("invalid email")
	}

	if len(password) < 8 || len(password) > 128 {
		return nil, nil, managederror.NewInvalidInputError("the length of the password must be between 8 and 128 characters")
	}

	if fullName == "" {
		return nil, nil, managederror.NewInvalidInputError("invalid full name")
	}

	hashedPassword, err := s.hp.HashPassword([]byte(password))
	if err != nil {
		return nil, nil, fmt.Errorf("cannot hash password: %w", err)
	}

	now := time.Now()
	user := &coredata.User{
		ID:             gid.New(gid.NilTenant, coredata.UserEntityType),
		EmailAddress:   email,
		HashedPassword: hashedPassword,
		FullName:       fullName,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	session := &coredata.Session{
		ID:        gid.New(gid.NilTenant, coredata.SessionEntityType),
		UserID:    user.ID,
		ExpiredAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}

	confirmationToken, err := statelesstoken.NewToken(
		s.tokenSecret,
		TokenTypeEmailConfirmation,
		1*time.Hour,
		EmailConfirmationData{UserID: user.ID, Email: user.EmailAddress},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot generate confirmation token: %w", err)
	}

	confirmationEmailUrl := url.URL{
		Scheme: "https",
		Host:   s.hostname,
		Path:   "/auth/confirm-email",
		RawQuery: url.Values{
			"token": []string{confirmationToken},
		}.Encode(),
	}

	confirmationEmail := coredata.NewEmail(
		user.FullName,
		user.EmailAddress,
		signupEmailSubject,
		fmt.Sprintf(signupEmailTemplate, confirmationEmailUrl.String()),
	)

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := user.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert user: %w", err)
			}

			if err := session.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert session: %w", err)
			}

			if err := confirmationEmail.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
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
	email, password string,
) (*coredata.User, *coredata.Session, error) {
	now := time.Now()
	user := &coredata.User{}
	session := &coredata.Session{
		ID:        gid.New(gid.NilTenant, coredata.SessionEntityType),
		UserID:    gid.Nil,
		ExpiredAt: now.Add(24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if len(password) < 8 || len(password) > 128 {
		return nil, nil, managederror.NewInvalidInputError("the length of the password must be between 8 and 128 characters")
	}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := user.LoadByEmail(ctx, tx, email); err != nil {
				_, _ = s.hp.ComparePasswordAndHash([]byte("this-compare-should-never-succeed"), []byte("it-just-to-prevent-timing-attack"))

				var errUserNotFound *coredata.ErrUserNotFound

				if errors.As(err, &errUserNotFound) {
					return managederror.NewInvalidInputError("invalid email or password")
				}

				return fmt.Errorf("cannot load user by email: %w", err)
			}

			ok, err := s.hp.ComparePasswordAndHash([]byte(password), user.HashedPassword)
			if err != nil {
				return fmt.Errorf("cannot compare password: %w", err)
			}

			if !ok {
				return managederror.NewInvalidInputError("invalid email or password")
			}

			session.UserID = user.ID

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

func (s Service) SignOut(
	ctx context.Context,
	sessionID gid.GID,
) error {
	return s.pg.WithConn(
		ctx,
		func(tx pg.Conn) error {
			err := coredata.DeleteSession(ctx, tx, sessionID)
			if err != nil {
				return fmt.Errorf("cannot delete session: %w", err)
			}

			return nil
		},
	)
}

func (s Service) GetSession(
	ctx context.Context,
	sessionID gid.GID,
) (*coredata.Session, error) {
	session := &coredata.Session{}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := session.LoadByID(ctx, tx, sessionID); err != nil {
				return managederror.NewNotFoundError("session")
			}

			if time.Now().After(session.ExpiredAt) {
				if err := coredata.DeleteSession(ctx, tx, sessionID); err != nil {
					return fmt.Errorf("cannot delete expired session: %w", err)
				}
				return managederror.NewSessionExpiredError()
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s Service) GetUserByID(
	ctx context.Context,
	userID gid.GID,
) (*coredata.User, error) {
	user := &coredata.User{}

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := user.LoadByID(ctx, tx, userID); err != nil {
				return fmt.Errorf("user not found: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s Service) GetUserBySession(
	ctx context.Context,
	sessionID gid.GID,
) (*coredata.User, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return s.GetUserByID(ctx, session.UserID)
}

func (s Service) ListOrganizationsForUserID(
	ctx context.Context,
	userID gid.GID,
) (coredata.Organizations, error) {

	uos := coredata.UserOrganizations{}
	organizations := []*coredata.Organization{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := uos.ForUserID(ctx, conn, userID); err != nil {
				return fmt.Errorf("cannot list user organizations: %w", err)
			}

			for _, uo := range uos {
				scope := coredata.NewScope(uo.OrganizationID.TenantID())
				organization := &coredata.Organization{}
				if err := organization.LoadByID(ctx, conn, scope, uo.OrganizationID); err != nil {
					return fmt.Errorf("cannot load organization by id: %w", err)
				}
				organizations = append(organizations, organization)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organizations, nil
}

// Tenant id scope is not applied in this functions because we want to access all user's organizations.
func (s Service) ListOrganizationsForUserIDPaginated(
	ctx context.Context,
	userID gid.GID,
	cursor *page.Cursor[coredata.OrganizationOrderField],
) (coredata.Organizations, error) {
	organizations := coredata.Organizations{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := organizations.ListForUserID(ctx, conn, userID, cursor)
			if err != nil {
				return fmt.Errorf("cannot list user organizations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return organizations, nil
}

func (s Service) ListTenantsForUserID(
	ctx context.Context,
	userID gid.GID,
) ([]gid.TenantID, error) {

	uos := coredata.UserOrganizations{}

	err := s.pg.WithConn(
		ctx,
		func(tx pg.Conn) error {
			return uos.ForUserID(ctx, tx, userID)
		},
	)

	if err != nil {
		return nil, err
	}

	tenantIDs := make([]gid.TenantID, len(uos))
	for _, uo := range uos {
		tenantIDs = append(tenantIDs, uo.OrganizationID.TenantID())
	}

	return tenantIDs, nil
}

func (s Service) EnrollUserInOrganization(
	ctx context.Context,
	userID gid.GID,
	organizationID gid.GID,
) error {

	uo := coredata.UserOrganization{
		UserID:         userID,
		OrganizationID: organizationID,
		CreatedAt:      time.Now(),
	}

	return s.pg.WithConn(
		ctx,
		func(tx pg.Conn) error {
			return uo.Insert(ctx, tx)
		},
	)
}

func (s Service) UpdateSession(
	ctx context.Context,
	session *coredata.Session,
) error {
	session.UpdatedAt = time.Now()
	session.ExpiredAt = time.Now().Add(24 * time.Hour)

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			return session.Update(ctx, tx)
		},
	)
}

func (s Service) ConfirmEmail(ctx context.Context, tokenString string) error {
	token, err := statelesstoken.ValidateToken[EmailConfirmationData](
		s.tokenSecret,
		TokenTypeEmailConfirmation,
		tokenString,
	)
	if err != nil {
		return fmt.Errorf("cannot validate email confirmation token: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			user := &coredata.User{}

			if err := user.LoadByID(ctx, tx, token.Data.UserID); err != nil {
				return fmt.Errorf("user not found: %w", err)
			}

			if user.EmailAddress != token.Data.Email {
				return fmt.Errorf("token email does not match user email")
			}

			if err := user.UpdateEmailVerification(ctx, tx, true); err != nil {
				return fmt.Errorf("cannot update user email verification: %w", err)
			}

			return nil
		},
	)
}

func (s Service) ListUsersForTenant(
	ctx context.Context,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.UserOrderField],
) (*page.Page[*coredata.User, coredata.UserOrderField], error) {
	users := coredata.Users{}

	err := s.pg.WithConn(
		ctx,
		func(tx pg.Conn) error {
			return users.LoadByOrganizationID(ctx, tx, organizationID, cursor)
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(users, cursor), nil
}

func (s Service) InviteUser(
	ctx context.Context,
	organizationID gid.GID,
	fullName string,
	emailAddress string,
) error {
	if _, err := mail.ParseAddress(emailAddress); err != nil {
		return managederror.NewInvalidInputError("invalid email")
	}
	if fullName == "" {
		return managederror.NewInvalidInputError("invalid full name")
	}

	var userExists bool
	err := s.pg.WithConn(
		ctx,
		func(tx pg.Conn) error {
			user := &coredata.User{}

			if err := user.LoadByEmail(ctx, tx, emailAddress); err != nil {
				var errUserNotFound *coredata.ErrUserNotFound

				if errors.As(err, &errUserNotFound) {
					userExists = false
					return nil
				}

				return fmt.Errorf("cannot load user by email: %w", err)
			}

			userExists = true
			uo := coredata.UserOrganization{
				UserID:         user.ID,
				OrganizationID: organizationID,
				CreatedAt:      time.Now(),
			}

			if err := uo.Insert(ctx, tx); err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					return managederror.NewAlreadyExistsError("user")
				}
				return fmt.Errorf("cannot insert user organization: %w", err)
			}

			people := &coredata.People{}
			scope := coredata.NewScope(organizationID.TenantID())
			if err := people.LoadByEmail(ctx, tx, scope, emailAddress); err != nil {
				var errPeopleNotFound *coredata.ErrPeopleNotFound

				if errors.As(err, &errPeopleNotFound) {
					people = &coredata.People{
						ID:                       gid.New(organizationID.TenantID(), coredata.PeopleEntityType),
						OrganizationID:           organizationID,
						UserID:                   &user.ID,
						FullName:                 fullName,
						PrimaryEmailAddress:      emailAddress,
						Kind:                     coredata.PeopleKindContractor,
						AdditionalEmailAddresses: []string{},
						CreatedAt:                time.Now(),
						UpdatedAt:                time.Now(),
					}

					if err := people.Insert(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot insert people: %w", err)
					}

					return nil
				}

				return fmt.Errorf("cannot load people by email: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	if userExists {
		return nil
	}

	confirmationToken, err := statelesstoken.NewToken(
		s.tokenSecret,
		TokenTypeOrganizationInvitation,
		s.invitationTokenValidity,
		InvitationData{OrganizationID: organizationID, Email: emailAddress, FullName: fullName},
	)
	if err != nil {
		return fmt.Errorf("cannot generate confirmation token: %w", err)
	}

	confirmationInvitationUrl := url.URL{
		Scheme: "https",
		Host:   s.hostname,
		Path:   "/auth/confirm-invitation",
		RawQuery: url.Values{
			"token": []string{confirmationToken},
		}.Encode(),
	}

	confirmationEmail := coredata.NewEmail(
		fullName,
		emailAddress,
		invitationEmailSubject,
		fmt.Sprintf(invitationEmailTemplate, confirmationInvitationUrl.String()),
	)

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := confirmationEmail.Insert(ctx, conn); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)
}

func (s Service) ConfirmInvitation(ctx context.Context, tokenString string, password string) (*coredata.User, error) {
	token, err := statelesstoken.ValidateToken[InvitationData](
		s.tokenSecret,
		TokenTypeOrganizationInvitation,
		tokenString,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot validate organization invitation token: %w", err)
	}

	if len(password) < 8 || len(password) > 128 {
		return nil, managederror.NewInvalidInputError("the length of the password must be between 8 and 128 characters")
	}

	now := time.Now()

	hashedPassword, err := s.hp.HashPassword([]byte(password))
	if err != nil {
		return nil, fmt.Errorf("cannot hash password: %w", err)
	}

	user := &coredata.User{}

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {

			if err := user.LoadByEmail(ctx, tx, token.Data.Email); err != nil {
				var errUserNotFound *coredata.ErrUserNotFound

				if errors.As(err, &errUserNotFound) {
					user = &coredata.User{
						ID:                   gid.New(gid.NilTenant, coredata.UserEntityType),
						EmailAddress:         token.Data.Email,
						HashedPassword:       hashedPassword,
						EmailAddressVerified: true,
						FullName:             token.Data.FullName,
						CreatedAt:            now,
						UpdatedAt:            now,
					}

					if err := user.Insert(ctx, tx); err != nil {
						return fmt.Errorf("cannot insert user: %w", err)
					}
				}
			}

			uo := coredata.UserOrganization{
				UserID:         user.ID,
				OrganizationID: token.Data.OrganizationID,
				CreatedAt:      now,
			}

			if err := uo.Insert(ctx, tx); err != nil {
				var pgErr *pgconn.PgError
				if errors.As(err, &pgErr) && pgErr.Code == "23505" {
					return managederror.NewAlreadyExistsError("user")
				}
				return fmt.Errorf("cannot insert user organization: %w", err)
			}

			people := &coredata.People{}
			scope := coredata.NewScope(token.Data.OrganizationID.TenantID())

			if err := people.LoadByEmail(ctx, tx, scope, token.Data.Email); err != nil {
				var errPeopleNotFound *coredata.ErrPeopleNotFound

				if errors.As(err, &errPeopleNotFound) {
					peopleID := gid.New(token.Data.OrganizationID.TenantID(), coredata.PeopleEntityType)
					people = &coredata.People{
						ID:                       peopleID,
						OrganizationID:           token.Data.OrganizationID,
						UserID:                   &user.ID,
						FullName:                 token.Data.FullName,
						PrimaryEmailAddress:      token.Data.Email,
						Kind:                     coredata.PeopleKindEmployee,
						AdditionalEmailAddresses: []string{},
						CreatedAt:                now,
						UpdatedAt:                now,
					}

					if err := people.Insert(ctx, tx, scope); err != nil {
						return fmt.Errorf("cannot insert people: %w", err)
					}
				} else {
					return fmt.Errorf("cannot load people by email: %w", err)
				}
			} else {
				people.UserID = &user.ID
				people.FullName = token.Data.FullName
				people.UpdatedAt = now

				if err := people.Update(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot update people: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s Service) RemoveUser(ctx context.Context, organizationID gid.GID, userID gid.GID) error {
	return s.pg.WithConn(
		ctx,
		func(tx pg.Conn) error {
			uo := coredata.UserOrganization{
				UserID:         userID,
				OrganizationID: organizationID,
			}

			if err := uo.Delete(ctx, tx); err != nil {
				return fmt.Errorf("cannot delete user organization: %w", err)
			}

			return nil
		},
	)
}

func (s Service) ResetPassword(ctx context.Context, tokenString string, newPassword string) error {
	token, err := statelesstoken.ValidateToken[PasswordResetData](
		s.tokenSecret,
		TokenTypePasswordReset,
		tokenString,
	)
	if err != nil {
		return fmt.Errorf("cannot validate password reset token: %w", err)
	}

	if len(newPassword) < 8 || len(newPassword) > 128 {
		return managederror.NewInvalidInputError("the length of the password must be between 8 and 128 characters")
	}

	hashedPassword, err := s.hp.HashPassword([]byte(newPassword))
	if err != nil {
		return fmt.Errorf("cannot hash password: %w", err)
	}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			user := &coredata.User{}

			if err := user.LoadByEmail(ctx, tx, token.Data.Email); err != nil {
				var errUserNotFound *coredata.ErrUserNotFound

				if errors.As(err, &errUserNotFound) {
					return fmt.Errorf("user not found: %w", err)
				}

				return fmt.Errorf("cannot load user by email: %w", err)
			}

			if err := user.UpdatePassword(ctx, tx, hashedPassword); err != nil {
				return fmt.Errorf("cannot update user password: %w", err)
			}

			return nil
		},
	)
}
