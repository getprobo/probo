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

// LEGACY ACCESS MANAGEMENT SERVICE - DEPRECATED
//
// This service implements the legacy authorization model that uses the Permissions
// map from permissions.go to check if a principal can perform an action.
//
// It is being replaced by Authorizer which uses a policy-based evaluation system.
// During migration, this service is still used for:
//   - API key authorization (intersection semantics between user and API key roles)
//   - Fallback for any unmapped legacy actions
//
// Once all actions are migrated and API key authorization is implemented in the
// new system, this service will be removed.
//
// Deprecated: Use Authorizer.Authorize() instead for new code.
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
// - principalID is the actor (Identity now; later service accounts)
// - credentialID is an optional credential (PersonalAPIKey now)
// - intersection semantics: actor must be allowed AND credential (if present) must be allowed.
//
// Entity scope:
// - Global/self-owned entities (Identity/Session/PersonalAPIKey) are authorized via ownership checks only (no global admin).
// - Organization-scoped entities are authorized via membership lookups that derive organization_id from entityID.
func (s *AccessManagementService) Authorize(ctx context.Context, principalID gid.GID, credentialID *gid.GID, entityID gid.GID, action Action) error {
	requiredRoles := GetPermissionsForAction(entityID.EntityType(), action)
	if requiredRoles == nil {
		entityModel, _ := coredata.EntityModel(entityID.EntityType())
		return NewNoPermissionsDefinedError(entityModel, action)
	}

	switch principalID.EntityType() {
	case coredata.IdentityEntityType:
		// ok
	default:
		return NewUnsupportedPrincipalTypeError(principalID.EntityType())
	}

	return s.pg.WithConn(ctx, func(conn pg.Conn) error {
		// Global/self-owned path
		switch entityID.EntityType() {
		case coredata.IdentityEntityType:
			if entityID != principalID {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			return nil

		case coredata.SessionEntityType:
			sess := &coredata.Session{}
			if err := sess.LoadByID(ctx, conn, entityID); err != nil {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			if sess.IdentityID != principalID {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			return nil

		case coredata.PersonalAPIKeyEntityType:
			key := &coredata.PersonalAPIKey{}
			if err := key.LoadByID(ctx, conn, entityID); err != nil {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			if key.IdentityID != principalID {
				return NewInsufficientPermissionsError(principalID, entityID, action)
			}
			return nil
		}

		// Organization-scoped path (derive org via joins)
		scope := coredata.NewScope(entityID.TenantID())

		actorRoleName, err := s.loadIdentityRoleForEntity(ctx, conn, scope, principalID, entityID)
		if err != nil || !requiredRoleNamesContain(actorRoleName, requiredRoles) {
			return NewInsufficientPermissionsError(principalID, entityID, action)
		}

		// Optional credential restriction (intersection)
		if credentialID != nil {
			switch credentialID.EntityType() {
			case coredata.PersonalAPIKeyEntityType:
				// Defensive check: credential must belong to actor
				apiKey := &coredata.PersonalAPIKey{}
				if err := apiKey.LoadByID(ctx, conn, *credentialID); err != nil {
					return NewInsufficientPermissionsError(principalID, entityID, action)
				}
				if apiKey.IdentityID != principalID {
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

func (s *AccessManagementService) loadIdentityRoleForEntity(
	ctx context.Context,
	conn pg.Conn,
	scope coredata.Scoper,
	identityID gid.GID,
	entityID gid.GID,
) (Role, error) {
	var m coredata.Membership
	if err := m.LoadRoleByIdentityAndEntityID(ctx, conn, scope, identityID, entityID); err != nil {
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
	// Load the API key to get the identity
	apiKey := &coredata.PersonalAPIKey{}
	if err := apiKey.LoadByID(ctx, conn, apiKeyID); err != nil {
		return "", fmt.Errorf("cannot load api key: %w", err)
	}

	// Use the Identity's membership role for authorization
	var m coredata.Membership
	if err := m.LoadRoleByIdentityAndEntityID(ctx, conn, scope, apiKey.IdentityID, entityID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return "", err
		}
		return "", err
	}
	return Role(m.Role.String()), nil
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
