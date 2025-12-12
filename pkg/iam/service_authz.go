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
	"slices"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type TenantAccessError struct {
	Message string
}

func (e *TenantAccessError) Error() string {
	return "not authorized"
}

type PermissionDeniedError struct {
	Message string
}

func (e *PermissionDeniedError) Error() string {
	return e.Message
}

const (
	TokenTypeOrganizationInvitation = "organization_invitation"
)

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

func (s *Service) GetUserOrganizationsWithRole(
	ctx context.Context,
	userID gid.GID,
	role coredata.MembershipRole,
) (coredata.Organizations, error) {
	organizations := coredata.Organizations{}

	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := organizations.LoadAllByUserIDWithRole(ctx, conn, userID, role); err != nil {
				return fmt.Errorf("cannot load user organizations with role: %w", err)
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

func (s *TenantService) Authorize(
	ctx context.Context,
	user *coredata.User,
	apiKey *coredata.UserAPIKey,
	entityGID gid.GID,
	action Action,
) error {
	requiredRoles := GetPermissionsForAction(entityGID.EntityType(), action)
	if requiredRoles == nil {
		entityModel, _ := coredata.EntityModel(entityGID.EntityType())
		return &PermissionDeniedError{
			Message: fmt.Sprintf("no permissions defined for action %s on entity %s", action, entityModel),
		}
	}

	role, err := s.GetUserOrAPIKeyRole(ctx, user, apiKey, entityGID)
	if err != nil {
		return fmt.Errorf("cannot get user or API key role: %w", err)
	}

	if !slices.Contains(requiredRoles, role) {
		return &PermissionDeniedError{
			Message: fmt.Sprintf("role %s not authorized for action %s, requires one of %v", role, action, requiredRoles),
		}
	}

	return nil
}

func (s *TenantService) CanAssignRole(
	ctx context.Context,
	user *coredata.User,
	apiKey *coredata.UserAPIKey,
	entityGID gid.GID,
	targetRole coredata.MembershipRole,
) error {
	currentRole, err := s.GetUserOrAPIKeyRole(ctx, user, apiKey, entityGID)
	if err != nil {
		return fmt.Errorf("cannot get user or API key role: %w", err)
	}

	if currentRole == RoleOwner || currentRole == RoleFull {
		return nil
	}

	if currentRole == RoleAdmin {
		if targetRole == coredata.MembershipRoleOwner {
			return &PermissionDeniedError{Message: "admin users cannot assign owner role"}
		}
		return nil
	}

	return &PermissionDeniedError{Message: fmt.Sprintf("role %s cannot assign roles", currentRole)}
}

func (s *TenantService) GetUserOrAPIKeyRole(
	ctx context.Context,
	user *coredata.User,
	apiKey *coredata.UserAPIKey,
	entityGID gid.GID,
) (Role, error) {
	var role Role
	err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if user != nil {
				membership := &coredata.Membership{}
				if err := membership.LoadRoleByUserAndEntityID(ctx, conn, s.scope, user.ID, entityGID); err != nil {
					return fmt.Errorf("cannot get user role: %w", err)
				}
				role = Role(membership.Role.String())
				return nil
			}

			if apiKey != nil {
				apiKeyMembership := &coredata.UserAPIKeyMembership{}
				if err := apiKeyMembership.LoadRoleByAPIKeyAndEntityID(ctx, conn, s.scope, apiKey.ID, entityGID); err != nil {
					return fmt.Errorf("cannot get API key role: %w", err)
				}
				role = Role(apiKeyMembership.Role.String())
				return nil
			}

			return fmt.Errorf("no user or API key provided")
		},
	)
	if err != nil {
		return "", err
	}

	return role, nil
}
