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
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"go.gearno.de/kit/pg"
)

type (
	// Service handles all authorization logic including organization
	// membership and permissions. This service is completely independent
	// of authentication methods.
	Service struct {
		pg                      *pg.Client
		hostname                string
		tokenSecret             string
		invitationTokenValidity time.Duration
	}

	Role string
)

const (
	RoleOwner  Role = "OWNER"
	RoleAdmin  Role = "ADMIN"
	RoleMember Role = "MEMBER"
	RoleViewer Role = "VIEWER"
)

const (
	TokenTypeOrganizationInvitation = "organization_invitation"
)

var (
	//go:embed emails/invitation.txt.tmpl
	invitationEmailBodyData string

	invitationEmailBodyTemplate = template.Must(template.New("invitation").Parse(invitationEmailBodyData))
	invitationEmailSubject      = "Invitation to join organization"
)

func NewService(
	ctx context.Context,
	pgClient *pg.Client,
	hostname string,
	tokenSecret string,
	invitationTokenValidity time.Duration,
) (*Service, error) {
	return &Service{
		pg:                      pgClient,
		hostname:                hostname,
		tokenSecret:             tokenSecret,
		invitationTokenValidity: invitationTokenValidity,
	}, nil
}

func (s *Service) GetAllUserOrganizations(
	ctx context.Context,
	userID gid.GID,
) ([]*coredata.Organization, error) {
	var organizations []*coredata.Organization

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		var organizationList coredata.Organizations
		if err := organizationList.LoadAllByUserID(ctx, conn, userID); err != nil {
			return fmt.Errorf("failed to load user organizations: %w", err)
		}

		organizations = organizationList

		return nil
	})

	return organizations, err
}

func (s *Service) GetUserOrganizations(
	ctx context.Context,
	userID gid.GID,
	cursor *page.Cursor[coredata.OrganizationOrderField],
) ([]*coredata.Organization, error) {
	var organizations coredata.Organizations

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		if err := organizations.LoadByUserID(ctx, conn, userID, cursor); err != nil {
			return fmt.Errorf("failed to load user organizations: %w", err)
		}
		return nil
	})

	return organizations, err
}

func (s *Service) GetAllOrganizationInvitations(
	ctx context.Context,
	orgID gid.GID,
	cursor *page.Cursor[coredata.InvitationOrderField],
) (*page.Page[*coredata.Invitation, coredata.InvitationOrderField], error) {
	var invitations coredata.Invitations

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := invitations.LoadByOrganizationID(ctx, conn, orgID, cursor); err != nil {
				return fmt.Errorf("failed to load organization invitations: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(invitations, cursor), nil
}

func (s *Service) CountOrganizationInvitations(
	ctx context.Context,
	orgID gid.GID,
) (int, error) {
	var count int
	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var invitations coredata.Invitations
			var err error
			count, err = invitations.CountByOrganizationID(ctx, conn, orgID)
			return err
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to count invitations: %w", err)
	}

	return count, nil
}

func (s *Service) DeleteInvitation(
	ctx context.Context,
	invitationID gid.GID,
) error {
	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			invitation := &coredata.Invitation{}
			if err := invitation.LoadByID(ctx, conn, invitationID); err != nil {
				return fmt.Errorf("failed to load invitation: %w", err)
			}

			if err := invitation.Delete(ctx, conn); err != nil {
				return fmt.Errorf("failed to delete invitation: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) GetAllOrganizationMemberships(
	ctx context.Context,
	orgID gid.GID,
	cursor *page.Cursor[coredata.MembershipOrderField],
) (*page.Page[*coredata.Membership, coredata.MembershipOrderField], error) {
	var memberships coredata.Memberships

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := memberships.LoadByOrganizationID(ctx, conn, orgID, cursor); err != nil {
				return fmt.Errorf("failed to load organization memberships: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(memberships, cursor), nil
}

func (s *Service) CountOrganizationMemberships(
	ctx context.Context,
	orgID gid.GID,
) (int, error) {
	var count int
	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			var memberships coredata.Memberships
			var err error
			count, err = memberships.CountByOrganizationID(ctx, conn, orgID)
			return err
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to count memberships: %w", err)
	}

	return count, nil
}

func (s *Service) CanUserAccessOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
) (bool, error) {
	membership := &coredata.Membership{}

	haveAccess := false

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := membership.LoadByUserAndOrg(ctx, conn, userID, orgID); err != nil {
				if _, ok := err.(coredata.ErrMembershipNotFound); ok {
					return nil // Not an error, just no access
				}
				return fmt.Errorf("failed to check organization access: %w", err)
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

func (s *Service) GetUserRoleInOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
) (string, error) {
	membership := &coredata.Membership{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := membership.LoadByUserAndOrg(ctx, conn, userID, orgID); err != nil {
				return fmt.Errorf("failed to get user role: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return "", err
	}

	return membership.Role, nil
}

func (s *Service) RemoveMemberFromOrganization(
	ctx context.Context,
	orgID gid.GID,
	memberID gid.GID,
) error {
	membership := &coredata.Membership{}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := membership.LoadByID(ctx, tx, memberID); err != nil {
				return fmt.Errorf("failed to load membership: %w", err)
			}

			if membership.OrganizationID != orgID {
				return fmt.Errorf("membership does not belong to organization")
			}

			if err := membership.Delete(ctx, tx); err != nil {
				return fmt.Errorf("failed to delete membership: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) AddUserToOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
	role string,
) error {
	membership := &coredata.Membership{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := membership.Create(ctx, conn); err != nil {
				return fmt.Errorf("failed to add user to organization: %w", err)
			}
			return nil
		},
	)
}

func (s *Service) UpdateUserRole(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
	newRole string,
) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			membership := &coredata.Membership{}
			if err := membership.LoadByUserAndOrg(ctx, tx, userID, orgID); err != nil {
				return fmt.Errorf("failed to find membership: %w", err)
			}

			membership.Role = newRole
			membership.UpdatedAt = time.Now()

			if err := membership.Update(ctx, tx); err != nil {
				return fmt.Errorf("failed to update user role: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) InviteUserToOrganization(
	ctx context.Context,
	organizationID gid.GID,
	emailAddress string,
	fullName string,
	role string,
) (*coredata.Invitation, error) {
	var invitation *coredata.Invitation

	err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
		user := &coredata.User{}
		userExists := true
		if err := user.LoadByEmail(ctx, tx, emailAddress); err != nil {
			var userNotFound *coredata.ErrUserNotFound
			if errors.As(err, &userNotFound) {
				userExists = false
			} else {
				return fmt.Errorf("failed to check if user exists: %w", err)
			}
		}

		organization := &coredata.Organization{}
		scope := coredata.NewScope(organizationID.TenantID())
		if err := organization.LoadByID(ctx, tx, scope, organizationID); err != nil {
			return fmt.Errorf("failed to load organization: %w", err)
		}

		invitationID := gid.New(organizationID.TenantID(), coredata.InvitationEntityType)
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

		if userExists {
			membership := &coredata.Membership{
				UserID:         user.ID,
				OrganizationID: organizationID,
				Role:           role,
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			if err := membership.Create(ctx, tx); err != nil {
				return fmt.Errorf("failed to add user to organization: %w", err)
			}

			invitation.AcceptedAt = &now
		} else {
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
				return fmt.Errorf("failed to generate invitation token: %w", err)
			}

			body := bytes.NewBuffer(nil)
			err = invitationEmailBodyTemplate.Execute(
				body,
				map[string]string{
					"FullName":         fullName,
					"OrganizationName": organization.Name,
					"InvitationURL":    fmt.Sprintf("https://%s/auth/confirm-invitation?token=%s", s.hostname, invitationToken),
				},
			)
			if err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			email := coredata.NewEmail(
				fullName,
				emailAddress,
				invitationEmailSubject,
				body.String(),
			)

			if err := email.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}
		}

		if err := invitation.Create(ctx, tx); err != nil {
			return fmt.Errorf("cannot create invitation: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return invitation, nil
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

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			invitation := &coredata.Invitation{}
			if err := invitation.LoadByID(ctx, tx, invitationData.InvitationID); err != nil {
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

			membership := &coredata.Membership{
				UserID:         userID,
				OrganizationID: invitation.OrganizationID,
				Role:           invitation.Role,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			if err := membership.Create(ctx, tx); err != nil {
				return fmt.Errorf("failed to add user to organization: %w", err)
			}

			now := time.Now()
			invitation.AcceptedAt = &now
			if err := invitation.Update(ctx, tx); err != nil {
				return fmt.Errorf("failed to mark invitation as accepted: %w", err)
			}

			return nil
		},
	)
}

// This is a placeholder for future permission system
func (s *Service) HasPermission(
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

func (s *Service) ListUserInvitations(
	ctx context.Context,
	email string,
) ([]*coredata.Invitation, error) {
	var invitations coredata.Invitations

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := invitations.LoadByEmail(ctx, conn, email); err != nil {
				return fmt.Errorf("failed to load invitations: %w", err)
			}
			return nil
		},
	)

	return invitations, err
}
