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

// LEGACY: This is the legacy access management service that is used to authorize actions on entities.
// It is deprecated and will be removed in the future.
// Use the Authorizer instead.
package iam

import (
	"context"
	"errors"
	"fmt"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	AccessManagementService struct {
		*Service
	}
)

func NewAccessManagementService(svc *Service) *AccessManagementService {
	return &AccessManagementService{Service: svc}
}

// Authorize implements Model 2 authorization:
// - principalID is the actor (User now; later service accounts)
// - credentialID is an optional credential (UserAPIKey now)
// - intersection semantics: actor must be allowed AND credential (if present) must be allowed.
//
// Entity scope:
// - Global/self-owned entities (User/Session/UserAPIKey) are authorized via ownership checks only (no global admin).
// - Organization-scoped entities are authorized via membership lookups that derive organization_id from entityID.
func (s *AccessManagementService) Authorize(ctx context.Context, principalID gid.GID, credentialID *gid.GID, entityID gid.GID, action Action) error {
	requiredRoles := GetPermissionsForAction(entityID.EntityType(), action)
	if requiredRoles == nil {
		entityModel, _ := coredata.EntityModel(entityID.EntityType())
		return NewNoPermissionsDefinedError(entityModel, action)
	}

	switch principalID.EntityType() {
	case coredata.UserEntityType:
		// ok
	default:
		return NewUnsupportedPrincipalTypeError(principalID.EntityType())
	}

	return s.pg.WithConn(ctx, func(conn pg.Conn) error {
		// Global/self-owned path
		switch entityID.EntityType() {
		case coredata.UserEntityType:
			if entityID != principalID {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			return nil

		case coredata.SessionEntityType:
			sess := &coredata.Session{}
			if err := sess.LoadByID(ctx, conn, entityID); err != nil {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			if sess.UserID != principalID {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			return nil

		case coredata.UserAPIKeyEntityType:
			key := &coredata.UserAPIKey{}
			if err := key.LoadByID(ctx, conn, entityID); err != nil {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			if key.UserID != principalID {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			return nil
		}

		// Organization-scoped path (derive org via joins)
		scope := coredata.NewScope(entityID.TenantID())

		actorRoleName, err := s.loadUserRoleForEntity(ctx, conn, scope, principalID, entityID)
		if err != nil || !requiredRoleNamesContain(actorRoleName, requiredRoles) {
			return NewInsufficientPermissionsError(principalID, entityID, action)
		}

		// Optional credential restriction (intersection)
		if credentialID != nil {
			switch credentialID.EntityType() {
			case coredata.UserAPIKeyEntityType:
				// Defensive check: credential must belong to actor
				apiKey := &coredata.UserAPIKey{}
				if err := apiKey.LoadByID(ctx, conn, *credentialID); err != nil {
					return NewInsufficientPermissionsError(principalID, entityID, action)
				}
				if apiKey.UserID != principalID {
					return NewInsufficientPermissionsError(principalID, entityID, action)
				}

				keyRoleName, err := s.loadAPIKeyRoleForEntity(ctx, conn, scope, *credentialID, entityID)
				if err != nil || !requiredRoleNamesContain(keyRoleName, requiredRoles) {
					return NewInsufficientPermissionsError(principalID, entityID, action)
				}
			default:
				return NewUnsupportedPrincipalTypeError(credentialID.EntityType())
			}
		}

		return nil
	})
}

func (s *AccessManagementService) loadUserRoleForEntity(
	ctx context.Context,
	conn pg.Conn,
	scope coredata.Scoper,
	userID gid.GID,
	entityID gid.GID,
) (Role, error) {
	var m coredata.Membership
	if err := m.LoadRoleByUserAndEntityID(ctx, conn, scope, userID, entityID); err != nil {
		// Do not leak existence details
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return "", err
		}
		return "", err
	}
	return Role(m.Role.String()), nil
}

func (s *AccessManagementService) loadAPIKeyRoleForEntity(
	ctx context.Context,
	conn pg.Conn,
	scope coredata.Scoper,
	apiKeyID gid.GID,
	entityID gid.GID,
) (Role, error) {
	var akm coredata.UserAPIKeyMembership
	if err := akm.LoadRoleByAPIKeyAndEntityID(ctx, conn, scope, apiKeyID, entityID); err != nil {
		return "", err
	}

	// Strict API key semantics: FULL only matches RoleFull explicitly.
	switch akm.Role {
	case coredata.APIRoleFull:
		return RoleFull, nil
	default:
		return "", fmt.Errorf("unsupported api key role: %s", akm.Role)
	}
}

// requiredRoleNamesContain is a temporary evaluator for the current in-code permissions registry
// (`Permissions` in `permissions.go`). In the future this becomes policy-document evaluation
// where the role name resolves to policy statements.
func requiredRoleNamesContain(roleName Role, required []Role) bool {
	for _, r := range required {
		if r == roleName {
			return true
		}
	}
	return false
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
