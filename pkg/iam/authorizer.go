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
	"errors"
	"fmt"
	"maps"
	"slices"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

// Authorizer handles authorization using the policy engine.
type Authorizer struct {
	pg        *pg.Client
	evaluator *policy.Evaluator
	policySet *PolicySet
}

// NewAuthorizer creates a new authorizer.
// Services register their policies by calling RegisterPolicySet.
//
// Example:
//
//	authorizer := iam.NewAuthorizer(pgClient)
//	authorizer.RegisterPolicySet(iam.IAMPolicySet())
//	authorizer.RegisterPolicySet(probo.ProboPolicySet())
func NewAuthorizer(pgClient *pg.Client) *Authorizer {
	return &Authorizer{
		pg:        pgClient,
		evaluator: policy.NewEvaluator(),
		policySet: NewPolicySet(),
	}
}

// RegisterPolicySet merges policies from another service into this authorizer.
// Services call this method to register their policies.
func (a *Authorizer) RegisterPolicySet(policySet *PolicySet) {
	a.policySet.Merge(policySet)
}

// AuthorizeParams contains all parameters for an authorization check.
type AuthorizeParams struct {
	// Principal is the user requesting access.
	Principal gid.GID

	// Resource is the target resource.
	Resource gid.GID

	// Action is the operation being performed (e.g., "iam:organization:get").
	Action string

	// ResourceAttributes provides additional context for condition evaluation.
	// Keys like "user_id", "owner_id" are used for self-management checks.
	ResourceAttributes map[string]string
}

func (a *Authorizer) GetPermissionsForMembership(ctx context.Context, identityID gid.GID, membershipID gid.GID) (map[string]map[Action]bool, error) {
	var (
		scope      = coredata.NewScopeFromObjectID(membershipID)
		membership = &coredata.Membership{}
	)

	err := a.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			err := membership.LoadByID(ctx, conn, scope, membershipID)
			if err != nil {
				if err == coredata.ErrResourceNotFound {
					return NewMembershipNotFoundError(membershipID)
				}

				return fmt.Errorf("cannot load membership: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	permissions := make(map[string]map[Action]bool)

	for entityType, actions := range Permissions {
		entityTypeName, ok := coredata.EntityModel(entityType)
		if !ok {
			continue
		}

		if permissions[entityTypeName] == nil {
			permissions[entityTypeName] = make(map[Action]bool)
		}

		for action, allowedRoles := range actions {
			if slices.Contains(allowedRoles, Role(membership.Role)) {
				permissions[entityTypeName][action] = true
			}
		}
	}

	return permissions, nil
}

// Authorize checks if the principal can perform the action on the resource.
// It combines self-management policies with role-based policies.
func (a *Authorizer) Authorize(ctx context.Context, params AuthorizeParams) error {
	// Validate principal type
	if params.Principal.EntityType() != coredata.IdentityEntityType {
		return NewUnsupportedPrincipalTypeError(params.Principal.EntityType())
	}

	// Build policies to evaluate
	policies := a.buildPolicies(ctx, params)

	// Build condition context
	conditionCtx := policy.ConditionContext{
		Principal: map[string]string{
			"id": params.Principal.String(),
		},
		Resource: map[string]string{
			"id": params.Resource.String(),
		},
	}

	maps.Copy(conditionCtx.Resource, params.ResourceAttributes)

	// Evaluate
	req := policy.AuthorizationRequest{
		Principal:        params.Principal,
		Resource:         params.Resource,
		Action:           params.Action,
		ConditionContext: conditionCtx,
	}

	result := a.evaluator.Evaluate(req, policies)

	if result.IsAllowed() {
		return nil
	}

	if result.Decision == policy.DecisionDeny {
		return NewInsufficientPermissionsError(params.Principal, params.Resource, params.Action)
	}

	// No match = implicit deny
	return NewInsufficientPermissionsError(params.Principal, params.Resource, params.Action)
}

// buildPolicies constructs the list of policies to evaluate.
// This includes self-management policies and role-based policies.
func (a *Authorizer) buildPolicies(ctx context.Context, params AuthorizeParams) []*policy.Policy {
	// Start with self-management policies
	policies := make([]*policy.Policy, len(a.policySet.SelfManagePolicies))
	copy(policies, a.policySet.SelfManagePolicies)

	// For organization-scoped resources, add role-based policies
	if params.Resource.TenantID() != gid.NilTenant {
		rolePolicies := a.loadRolePolicies(ctx, params.Principal, params.Resource)
		policies = append(policies, rolePolicies...)
	}

	return policies
}

// loadRolePolicies loads the role-based policies for a user in an organization.
func (a *Authorizer) loadRolePolicies(ctx context.Context, principalID gid.GID, resourceID gid.GID) []*policy.Policy {
	var roleName string

	err := a.pg.WithConn(ctx, func(conn pg.Conn) error {
		scope := coredata.NewScope(resourceID.TenantID())

		var m coredata.Membership
		if err := m.LoadRoleByIdentityAndEntityID(ctx, conn, scope, principalID, resourceID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil // No membership = no role-based policies
			}
			return err
		}

		roleName = m.Role.String()
		return nil
	})

	if err != nil || roleName == "" {
		// On error or no role, return empty policies (fail closed)
		return nil
	}

	// Get policies for the user's role
	return a.policySet.RolePolicies[roleName]
}
