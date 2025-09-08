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

	// Role represents a user's role within an organization
	Role string
)

// Role constants
const (
	RoleOwner  Role = "owner"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

// Token types
const (
	TokenTypeOrganizationInvitation = "organization_invitation"
)

var (
	// go:embed invitation.txt.tmpl
	invitationEmailBodyData string

	invitationEmailBodyTemplate = template.Must(template.New("invitation").Parse(invitationEmailBodyData))
)

// NewService creates a new authorization service
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

// GetUserOrganizations returns all organizations a user has access to
func (s *Service) GetUserOrganizations(
	ctx context.Context,
	userID gid.GID,
) ([]*coredata.OrganizationMembership, error) {
	var memberships coredata.OrganizationMemberships

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		if err := memberships.LoadByUserID(ctx, conn, userID); err != nil {
			return fmt.Errorf("failed to load user organizations: %w", err)
		}

		return nil
	})

	return memberships, err
}

// GetOrganizationMembers returns all members of an organization
func (s *Service) GetOrganizationMembers(
	ctx context.Context,
	orgID gid.GID,
	cursor *page.Cursor[coredata.MembershipOrderField],
) (*page.Page[*coredata.User, coredata.UserOrderField], error) {
	var memberships coredata.OrganizationMemberships

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := memberships.LoadByOrganizationID(ctx, conn, orgID, cursor); err != nil {
				return fmt.Errorf("failed to load organization members: %w", err)
			}

			userIDs := make([]*gid.GID, len(memberships))
			for i := range users {
				userIDs[i] = &memberships[i].UserID
			}

			users := coredata.Users{}
			if err := users.LoadByIDs(ctx, conn, userIDs); err != nil {
				return fmt.Errorf("failed to load users: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(memberships, cursor), nil
}

// CanUserAccessOrganization checks if a user can access an organization
func (s *Service) CanUserAccessOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
) (bool, error) {
	membership := &coredata.OrganizationMembership{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := membership.LoadByUserAndOrg(ctx, conn, userID, orgID); err != nil {
				if _, ok := err.(coredata.ErrMembershipNotFound); ok {
					return nil // Not an error, just no access
				}
				return fmt.Errorf("failed to check organization access: %w", err)
			}
			return nil
		},
	)

	if err != nil {
		return false, err
	}

	// If LoadByUserAndOrg didn't return ErrMembershipNotFound, user has access
	return membership.UserID != gid.Nil, nil
}

// GetUserRoleInOrganization returns the user's role in an organization
func (s *Service) GetUserRoleInOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
) (string, error) {
	membership := &coredata.OrganizationMembership{}

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

func (s *Service) RemoveUserFromOrganization(
	ctx context.Context,
	orgID gid.GID,
	userID gid.GID,
) error {
	membership := &coredata.OrganizationMembership{}

	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := membership.LoadByUserAndOrg(ctx, tx, userID, orgID); err != nil {
				return fmt.Errorf("failed to load membership: %w", err)
			}

			if err := membership.Delete(ctx, tx); err != nil {
				return fmt.Errorf("failed to delete membership: %w", err)
			}

			return nil
		},
	)
}

// AddUserToOrganization adds a user to an organization with a specific role
func (s *Service) AddUserToOrganization(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
	role string,
) error {
	membership := &coredata.OrganizationMembership{
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

// UpdateUserRole updates a user's role in an organization
func (s *Service) UpdateUserRole(
	ctx context.Context,
	userID gid.GID,
	orgID gid.GID,
	newRole string,
) error {
	return s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			membership := &coredata.OrganizationMembership{}
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

// InviteUserToOrganization creates an invitation for a user to join an organization
func (s *Service) InviteUserToOrganization(
	ctx context.Context,
	organizationID gid.GID,
	emailAddress string,
	fullName string,
	role string,
) (*coredata.Invitation, error) {
	invitationID := gid.New(organizationID.TenantID(), coredata.InvitationEntityType)

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
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	invitation := &coredata.Invitation{
		ID:             invitationID,
		OrganizationID: organizationID,
		Email:          emailAddress,
		FullName:       fullName,
		Role:           role,
		ExpiresAt:      time.Now().Add(s.invitationTokenValidity),
		CreatedAt:      time.Now(),
	}

	body := bytes.NewBuffer(nil)
	err = invitationEmailBodyTemplate.Execute(
		body,
		map[string]string{
			"FullName":         fullName,
			"OrganizationName": organizationID.String(),
			"InvitationURL":    fmt.Sprintf("https://%s/auth/accept-invitation?token=%s", s.hostname, invitationToken),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	email := coredata.NewEmail(
		fullName,
		emailAddress,
		"Invitation to join organization",
		body.String(),
	)

	err = s.pg.WithTx(
		ctx,
		func(conn pg.Conn) error {
			if err := invitation.Create(ctx, conn); err != nil {
				return fmt.Errorf("cannot create invitation: %w", err)
			}

			if err := email.Insert(ctx, conn); err != nil {
				return fmt.Errorf("cannot insert email: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return invitation, nil
}

// AcceptInvitation accepts an invitation and adds the user to the organization
func (s *Service) AcceptInvitation(
	ctx context.Context,
	token string,
	userID gid.GID,
) error {
	// Validate and decode the token
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
			// Load invitation from database using ID from token
			invitation := &coredata.Invitation{}
			if err := invitation.LoadByID(ctx, tx, invitationData.InvitationID); err != nil {
				return fmt.Errorf("cannot load invitation: %w", err)
			}

			// Check if already accepted
			if invitation.AcceptedAt != nil {
				return fmt.Errorf("invitation already accepted")
			}

			// Check if expired
			if time.Now().After(invitation.ExpiresAt) {
				return fmt.Errorf("invitation expired")
			}

			// Add user to organization
			membership := &coredata.OrganizationMembership{
				UserID:         userID,
				OrganizationID: invitation.OrganizationID,
				Role:           invitation.Role,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			if err := membership.Create(ctx, tx); err != nil {
				return fmt.Errorf("failed to add user to organization: %w", err)
			}

			// Mark invitation as accepted
			now := time.Now()
			invitation.AcceptedAt = &now
			if err := invitation.Update(ctx, tx); err != nil {
				return fmt.Errorf("failed to mark invitation as accepted: %w", err)
			}

			return nil
		},
	)
}

// HasPermission checks if a user has a specific permission in an organization
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

// ListUserInvitations returns all pending invitations for a user email
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
