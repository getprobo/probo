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
	"maps"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

type Authorizer struct {
	pg        *pg.Client
	evaluator *policy.Evaluator
	policySet *PolicySet
}

func NewAuthorizer(pgClient *pg.Client) *Authorizer {
	return &Authorizer{
		pg:        pgClient,
		evaluator: policy.NewEvaluator(),
		policySet: NewPolicySet(),
	}
}

func (a *Authorizer) RegisterPolicySet(policySet *PolicySet) {
	a.policySet.Merge(policySet)
}

type AuthorizeParams struct {
	Principal          gid.GID
	Resource           gid.GID
	Action             string
	ResourceAttributes map[string]string
}

func (a *Authorizer) Authorize(ctx context.Context, params AuthorizeParams) error {
	if params.Principal.EntityType() != coredata.IdentityEntityType {
		return NewUnsupportedPrincipalTypeError(params.Principal.EntityType())
	}

	policies := a.buildPolicies(ctx, params)

	// Pre-allocate Resource map with capacity for id + attributes
	resourceAttrs := make(map[string]string, 1+len(params.ResourceAttributes))
	resourceAttrs["id"] = params.Resource.String()
	maps.Copy(resourceAttrs, params.ResourceAttributes)

	conditionCtx := policy.ConditionContext{
		Principal: map[string]string{
			"id": params.Principal.String(),
		},
		Resource: resourceAttrs,
	}

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

	return NewInsufficientPermissionsError(params.Principal, params.Resource, params.Action)
}

func (a *Authorizer) buildPolicies(ctx context.Context, params AuthorizeParams) []*policy.Policy {
	selfManageCount := len(a.policySet.SelfManagePolicies)

	var rolePolicies []*policy.Policy
	if params.Resource.TenantID() != gid.NilTenant {
		rolePolicies = a.loadRolePolicies(ctx, params.Principal, params.Resource)
	}

	totalCount := selfManageCount + len(rolePolicies)
	policies := make([]*policy.Policy, selfManageCount, totalCount)
	copy(policies, a.policySet.SelfManagePolicies)
	policies = append(policies, rolePolicies...)

	return policies
}

func (a *Authorizer) loadRolePolicies(ctx context.Context, principalID gid.GID, resourceID gid.GID) []*policy.Policy {
	var role coredata.MembershipRole

	err := a.pg.WithConn(
		ctx,
		func(conn pg.Conn) (err error) {
			scope := coredata.NewScopeFromObjectID(resourceID)
			role, err = coredata.LoadRoleByIdentityAndEntityIDOnly(ctx, conn, scope, principalID, resourceID)
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil // No membership = no role-based policies
			}

			return err
		},
	)

	if err != nil || role == "" {
		return nil
	}

	return a.policySet.RolePolicies[role.String()]
}
