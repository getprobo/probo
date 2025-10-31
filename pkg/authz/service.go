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

package authz

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/getprobo/probo/packages/emails"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"go.gearno.de/kit/pg"
)

type TenantAccessError struct {
	Message string
}

func (e *TenantAccessError) Error() string {
	return "not authorized"
}

type (
	Service struct {
		pg                      *pg.Client
		baseURL                 string
		tokenSecret             string
		invitationTokenValidity time.Duration
	}

	TenantAuthzService struct {
		pg                      *pg.Client
		baseURL                 string
		tokenSecret             string
		invitationTokenValidity time.Duration
		scope                   coredata.Scoper
	}
)

const (
	TokenTypeOrganizationInvitation = "organization_invitation"
)

func NewService(
	ctx context.Context,
	pgClient *pg.Client,
	baseURL string,
	tokenSecret string,
	invitationTokenValidity time.Duration,
) (*Service, error) {
	return &Service{
		pg:                      pgClient,
		baseURL:                 baseURL,
		tokenSecret:             tokenSecret,
		invitationTokenValidity: invitationTokenValidity,
	}, nil
}

func (s *Service) WithTenant(tenantID gid.TenantID) *TenantAuthzService {
	return &TenantAuthzService{
		pg:                      s.pg,
		baseURL:                 s.baseURL,
		tokenSecret:             s.tokenSecret,
		invitationTokenValidity: s.invitationTokenValidity,
		scope:                   coredata.NewScope(tenantID),
	}
}

func (s *Service) GetAllUserOrganizations(
	ctx context.Context,
	userID gid.GID,
) (coredata.Organizations, error) {
	organizations := coredata.Organizations{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organizations.LoadAllByUserID(ctx, conn, userID); err != nil {
				return fmt.Errorf("cannot load user organizations: %w", err)
			}

			return nil
		},
	)

	return organizations, err
}

func (s *Service) GetUserOrganizations(
	ctx context.Context,
	userID gid.GID,
	cursor *page.Cursor[coredata.OrganizationOrderField],
) (coredata.Organizations, error) {
	organizations := coredata.Organizations{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organizations.LoadByUserID(ctx, conn, coredata.NewNoScope(), userID, cursor); err != nil {
				return fmt.Errorf("cannot load user organizations: %w", err)
			}
			return nil
		},
	)

	return organizations, err
}

func (s *Service) AcceptInvitation(
	ctx context.Context,
	token string,
	userID gid.GID,
) error {
	payload, err := statelesstoken.ValidateToken[coredata.InvitationData](
		s.tokenSecret,
		TokenTypeOrganizationInvitation,
		token,
	)

	if err != nil {
		return fmt.Errorf("invalid invitation token: %w", err)
	}

	invitationData := payload.Data
	scope := coredata.NewScope(invitationData.InvitationID.TenantID())

	return s.pg.WithTx(
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
			membershipID := gid.New(scope.GetTenantID(), coredata.MembershipEntityType)

			membership := &coredata.Membership{
				ID:             membershipID,
				UserID:         userID,
				OrganizationID: invitation.OrganizationID,
				Role:           invitation.Role,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := membership.Create(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot add user to organization: %w", err)
			}

			invitation.AcceptedAt = &now
			if err := invitation.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot mark invitation as accepted: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) AcceptInvitationByID(
	ctx context.Context,
	invitationID gid.GID,
	userID gid.GID,
) (*coredata.Invitation, error) {
	var acceptedInvitation *coredata.Invitation
	scope := coredata.NewScope(invitationID.TenantID())
	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			invitation := &coredata.Invitation{}
			if err := invitation.LoadByID(ctx, tx, scope, invitationID); err != nil {
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

			user := &coredata.User{}
			if err := user.LoadByID(ctx, tx, userID); err != nil {
				return fmt.Errorf("cannot load user: %w", err)
			}

			if invitation.Email != user.EmailAddress {
				return fmt.Errorf("invitation email does not match user email")
			}

			now := time.Now()
			membershipID := gid.New(scope.GetTenantID(), coredata.MembershipEntityType)

			membership := &coredata.Membership{
				ID:             membershipID,
				UserID:         userID,
				OrganizationID: invitation.OrganizationID,
				Role:           invitation.Role,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			if err := membership.Create(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot add user to organization: %w", err)
			}

			invitation.AcceptedAt = &now
			if err := invitation.Update(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot mark invitation as accepted: %w", err)
			}

			acceptedInvitation = invitation
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return acceptedInvitation, nil
}

func (s *Service) GetUserInvitations(
	ctx context.Context,
	email string,
	cursor *page.Cursor[coredata.InvitationOrderField],
	filter *coredata.InvitationFilter,
) (*page.Page[*coredata.Invitation, coredata.InvitationOrderField], error) {
	var invitations coredata.Invitations

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := invitations.LoadByEmail(ctx, conn, coredata.NewNoScope(), email, cursor, filter); err != nil {
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

type UserInvitation struct {
	ID             gid.GID
	Email          string
	FullName       string
	Role           coredata.Role
	ExpiresAt      time.Time
	AcceptedAt     *time.Time
	CreatedAt      time.Time
	OrganizationID gid.GID
	Organization   OrganizationSummary
}

type OrganizationSummary struct {
	ID   gid.GID
	Name string
}

func (s *Service) GetUserPendingInvitations(
	ctx context.Context,
	email string,
) ([]*UserInvitation, error) {
	userInvitations := []*UserInvitation{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			cursor := page.NewCursor(
				1000,
				nil,
				page.Head,
				page.OrderBy[coredata.InvitationOrderField]{
					Field:     coredata.InvitationOrderFieldCreatedAt,
					Direction: page.OrderDirectionDesc,
				},
			)
			filter := coredata.NewInvitationFilter([]coredata.InvitationStatus{coredata.InvitationStatusPending})
			invitations := coredata.Invitations{}

			if err := invitations.LoadByEmail(ctx, conn, coredata.NewNoScope(), email, cursor, filter); err != nil {
				return fmt.Errorf("cannot load invitations: %w", err)
			}

			organizationIDs := []gid.GID{}
			for _, invitation := range invitations {
				organizationIDs = append(organizationIDs, invitation.OrganizationID)
			}

			organizations := coredata.Organizations{}
			if err := organizations.BatchLoadByID(ctx, conn, coredata.NewNoScope(), organizationIDs); err != nil {
				return fmt.Errorf("cannot load organizations: %w", err)
			}

			for _, invitation := range invitations {
				userInvitation := &UserInvitation{
					ID:             invitation.ID,
					Email:          invitation.Email,
					FullName:       invitation.FullName,
					Role:           invitation.Role,
					ExpiresAt:      invitation.ExpiresAt,
					AcceptedAt:     invitation.AcceptedAt,
					CreatedAt:      invitation.CreatedAt,
					OrganizationID: invitation.OrganizationID,
				}

				for _, org := range organizations {
					if org.ID == invitation.OrganizationID {
						userInvitation.Organization = OrganizationSummary{
							ID:   org.ID,
							Name: org.Name,
						}
					}
				}

				userInvitations = append(userInvitations, userInvitation)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return userInvitations, nil
}

func (s *TenantAuthzService) GetOrganizationByInvitationID(
	ctx context.Context,
	invitationID gid.GID,
) (*coredata.Organization, error) {
	var organization coredata.Organization
	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var invitation coredata.Invitation
			if err := invitation.LoadByID(ctx, conn, s.scope, invitationID); err != nil {
				return fmt.Errorf("cannot load invitation: %w", err)
			}

			if err := organization.LoadByID(ctx, conn, s.scope, invitation.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &organization, nil
}

func (s *TenantAuthzService) AddUserToOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
	role coredata.Role,
) error {
	now := time.Now()
	membershipID := gid.New(s.scope.GetTenantID(), coredata.MembershipEntityType)

	membership := &coredata.Membership{
		ID:             membershipID,
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := membership.Create(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot add user to organization: %w", err)
			}
			return nil
		},
	)
}

func (s *TenantAuthzService) GetInvitationsByOrganizationID(
	ctx context.Context,
	orgID gid.GID,
	cursor *page.Cursor[coredata.InvitationOrderField],
	filter *coredata.InvitationFilter,
) (*page.Page[*coredata.Invitation, coredata.InvitationOrderField], error) {
	var invitations coredata.Invitations

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := invitations.LoadByOrganizationID(ctx, conn, s.scope, orgID, cursor, filter); err != nil {
				return fmt.Errorf("cannot load organization invitations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return page.NewPage(invitations, cursor), nil
}

func (s *TenantAuthzService) CountOrganizationInvitations(
	ctx context.Context,
	orgID gid.GID,
	filter *coredata.InvitationFilter,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			var invitations coredata.Invitations

			count, err = invitations.CountByOrganizationID(ctx, conn, s.scope, orgID, filter)
			if err != nil {
				return fmt.Errorf("cannot count organization invitations: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *TenantAuthzService) GetInvitationByID(
	ctx context.Context,
	invitationID gid.GID,
) (*coredata.Invitation, error) {
	invitation := &coredata.Invitation{}
	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := invitation.LoadByID(ctx, conn, s.scope, invitationID); err != nil {
				return fmt.Errorf("cannot load invitation: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *TenantAuthzService) DeleteInvitation(
	ctx context.Context,
	invitationID gid.GID,
) error {
	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			invitation := &coredata.Invitation{}
			if err := invitation.LoadByID(ctx, conn, s.scope, invitationID); err != nil {
				return fmt.Errorf("cannot load invitation: %w", err)
			}

			if err := invitation.Delete(ctx, conn, s.scope); err != nil {
				return fmt.Errorf("cannot delete invitation: %w", err)
			}

			return nil
		},
	)
}

func (s *TenantAuthzService) GetMembershipsByOrganizationID(
	ctx context.Context,
	orgID gid.GID,
	cursor *page.Cursor[coredata.MembershipOrderField],
) (*page.Page[*coredata.Membership, coredata.MembershipOrderField], error) {
	var memberships coredata.Memberships

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := memberships.LoadByOrganizationID(ctx, conn, s.scope, orgID, cursor); err != nil {
				return fmt.Errorf("cannot load organization memberships: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(memberships, cursor), nil
}

func (s *TenantAuthzService) CountOrganizationMemberships(
	ctx context.Context,
	orgID gid.GID,
) (int, error) {
	var count int
	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var memberships coredata.Memberships
			var err error
			count, err = memberships.CountByOrganizationID(ctx, conn, s.scope, orgID)
			return err
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count memberships: %w", err)
	}

	return count, nil
}

func (s *TenantAuthzService) CountOrganizationUsers(
	ctx context.Context,
	orgID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			var users coredata.Users

			count, err = users.CountByOrganizationID(ctx, conn, s.scope, orgID)
			if err != nil {
				return fmt.Errorf("cannot count organization users: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count users: %w", err)
	}

	return count, nil
}

func (s *TenantAuthzService) CanUserAccessOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
) (bool, error) {
	membership := &coredata.Membership{}

	haveAccess := false

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := membership.LoadByUserAndOrg(ctx, conn, s.scope, userID, orgID); err != nil {
				if _, ok := err.(coredata.ErrMembershipNotFound); ok {
					return nil // Not an error, just no access
				}
				return fmt.Errorf("cannot check organization access: %w", err)
			}
			haveAccess = true
			return nil
		},
	)

	if err != nil {
		return false, err
	}

	return haveAccess, nil
}

func (s *TenantAuthzService) GetUserRoleInOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
) (coredata.Role, error) {
	membership := &coredata.Membership{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := membership.LoadByUserAndOrg(ctx, conn, s.scope, userID, orgID); err != nil {
				return fmt.Errorf("cannot get user role: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return "", err
	}

	return membership.Role, nil
}

func (s *TenantAuthzService) RemoveMemberFromOrganization(
	ctx context.Context,
	orgID gid.GID,
	memberID gid.GID,
) error {
	membership := &coredata.Membership{}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := membership.LoadByID(ctx, tx, s.scope, memberID); err != nil {
				return fmt.Errorf("cannot load membership: %w", err)
			}

			if membership.OrganizationID != orgID {
				return fmt.Errorf("membership does not belong to organization")
			}

			if err := membership.Delete(ctx, tx, s.scope); err != nil {
				return fmt.Errorf("cannot delete membership: %w", err)
			}

			return nil
		},
	)
}

func (s *TenantAuthzService) UpdateUserRole(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
	newRole coredata.Role,
) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			membership := &coredata.Membership{}
			if err := membership.LoadByUserAndOrg(ctx, tx, s.scope, userID, orgID); err != nil {
				return fmt.Errorf("cannot find membership: %w", err)
			}

			membership.Role = newRole
			membership.UpdatedAt = time.Now()

			if err := membership.Update(ctx, tx, s.scope); err != nil {
				return fmt.Errorf("cannot update user role: %w", err)
			}

			return nil
		},
	)
}

func (s *TenantAuthzService) InviteUserToOrganization(
	ctx context.Context,
	organizationID gid.GID,
	emailAddress string,
	fullName string,
	role coredata.Role,
) (*coredata.Invitation, error) {
	var invitation *coredata.Invitation

	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			user := &coredata.User{}
			userExists := true
			if err := user.LoadByEmail(ctx, tx, emailAddress); err != nil {
				var userNotFound *coredata.ErrUserNotFound
				if errors.As(err, &userNotFound) {
					userExists = false
				} else {
					return fmt.Errorf("cannot check if user exists: %w", err)
				}
			}

			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, tx, s.scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			invitationID := gid.New(s.scope.GetTenantID(), coredata.InvitationEntityType)
			now := time.Now()
			invitation = &coredata.Invitation{
				ID:             invitationID,
				OrganizationID: organizationID,
				Email:          emailAddress,
				FullName:       fullName,
				Role:           role,
				ExpiresAt:      now.Add(s.invitationTokenValidity),
				CreatedAt:      now,
			}

			var err error
			var invitationURL string
			var recipientName string

			if userExists {
				recipientName = user.FullName
				invitationURL = s.baseURL + "/"
			} else {
				recipientName = fullName
				invitationData := coredata.InvitationData{
					InvitationID:   invitationID,
					OrganizationID: organizationID,
					Email:          emailAddress,
					FullName:       fullName,
					Role:           role,
				}

				invitationToken, err := statelesstoken.NewToken(
					s.tokenSecret,
					TokenTypeOrganizationInvitation,
					s.invitationTokenValidity,
					invitationData,
				)
				if err != nil {
					return fmt.Errorf("cannot generate invitation token: %w", err)
				}

				invitationURL = fmt.Sprintf("%s/auth/signup-from-invitation?token=%s&fullName=%s", s.baseURL, invitationToken, url.QueryEscape(fullName))
			}

			subject, textBody, htmlBody, err := emails.RenderInvitation(
				s.baseURL,
				recipientName,
				organization.Name,
				invitationURL,
			)
			if err != nil {
				return fmt.Errorf("cannot render invitation email: %w", err)
			}

			email := coredata.NewEmail(
				fullName,
				emailAddress,
				subject,
				textBody,
				htmlBody,
			)

			if err := email.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			if err := invitation.Create(ctx, tx, s.scope); err != nil {
				return fmt.Errorf("cannot create invitation: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

func (s *TenantAuthzService) EnsureSAMLMembership(
	ctx context.Context,
	userID gid.GID,
	organizationID gid.GID,
	role coredata.Role,
) error {
	now := time.Now()

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			var membership coredata.Membership

			err := membership.LoadByUserAndOrg(ctx, tx, s.scope, userID, organizationID)
			if err != nil {
				if _, ok := err.(coredata.ErrMembershipNotFound); !ok {
					return fmt.Errorf("cannot load membership: %w", err)
				}

				membershipID := gid.New(s.scope.GetTenantID(), coredata.MembershipEntityType)
				membership = coredata.Membership{
					ID:             membershipID,
					UserID:         userID,
					OrganizationID: organizationID,
					Role:           role,
					CreatedAt:      now,
					UpdatedAt:      now,
				}

				if err := membership.Create(ctx, tx, s.scope); err != nil {
					return fmt.Errorf("cannot create membership: %w", err)
				}

				return nil
			}

			if membership.Role != role {
				membership.Role = role
				membership.UpdatedAt = now

				if err := membership.Update(ctx, tx, s.scope); err != nil {
					return fmt.Errorf("cannot update membership role: %w", err)
				}
			}

			return nil
		},
	)
}

// This is a placeholder for future permission system
func (s *TenantAuthzService) HasPermission(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
	resource string,
	action string,
) (bool, error) {
	// For now, just check if user is a member
	// In the future, this will check specific permissions based on role
	return s.CanUserAccessOrganization(ctx, userID, orgID)
}
