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
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

// AuthorizationAttributer is implemented by entities that provide attributes
// for policy condition evaluation.
type AuthorizationAttributer interface {
	AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error)
}

// AuthorizeParams contains the parameters for an authorization request.
type AuthorizeParams struct {
	Principal          gid.GID
	Resource           gid.GID
	Session            *gid.GID
	Action             string
	ResourceAttributes map[string]string
}

// Authorizer evaluates authorization requests against registered policies.
type Authorizer struct {
	pg        *pg.Client
	evaluator *policy.Evaluator
	policySet *PolicySet
}

// NewAuthorizer creates a new Authorizer instance.
func NewAuthorizer(pgClient *pg.Client) *Authorizer {
	return &Authorizer{
		pg:        pgClient,
		evaluator: policy.NewEvaluator(),
		policySet: NewPolicySet(),
	}
}

// RegisterPolicySet merges the given policy set into the authorizer.
func (a *Authorizer) RegisterPolicySet(ps *PolicySet) {
	a.policySet.Merge(ps)
}

// Authorize checks if the principal is allowed to perform the action on the resource.
func (a *Authorizer) Authorize(ctx context.Context, params AuthorizeParams) error {
	if params.Principal.EntityType() != coredata.IdentityEntityType {
		return NewUnsupportedPrincipalTypeError(params.Principal.EntityType())
	}

	return a.pg.WithConn(ctx, func(conn pg.Conn) error { return a.authorize(ctx, conn, params) })
}

func (a *Authorizer) authorize(ctx context.Context, conn pg.Conn, params AuthorizeParams) error {
	resourceAttrs, err := a.buildResourceAttributes(ctx, conn, params)
	if err != nil {
		return fmt.Errorf("cannot build resource attributes: %w", err)
	}

	resourceOrgID := resourceAttrs["organization_id"]

	// Find role for resource's organization
	memberships, err := a.loadMemberships(ctx, conn, params.Principal)
	if err != nil {
		return fmt.Errorf("cannot load memberships for principal: %w", err)
	}
	membership := findMembershipForOrg(memberships, resourceOrgID)

	// Check whether the viewer is currently assuming the org of the accessed resource
	if membership != nil && params.Session != nil {
		if _, err := a.getActiveChildSessionForMembership(
			ctx,
			conn,
			*params.Session,
			membership.ID,
		); err != nil {
			var errSessionNotFound *ErrSessionNotFound
			var errSessionExpired *ErrSessionExpired

			if errors.As(err, &errSessionNotFound) || errors.As(err, &errSessionExpired) {
				return NewAssumptionRequiredError(params.Principal, membership.ID)
			}

			return fmt.Errorf("cannot get active child session for membership: %w", err)
		}
	}

	var role string
	if membership != nil {
		role = membership.Role.String()
	}

	// Only set principal.organization_id if they have a role in this org
	var scopedPrincipalAttrs map[string]string
	if membership != nil && role != "" {
		scopedPrincipalAttrs = map[string]string{
			"organization_id": membership.OrganizationID.String(),
			"role":            membership.Role.String(),
		}
	}

	principalAttrs, err := a.buildPrincipalAttributes(ctx, conn, params.Principal, scopedPrincipalAttrs)
	if err != nil {
		return fmt.Errorf("cannot build principal attributes: %w", err)
	}

	policies := a.buildPoliciesForRole(role)

	req := policy.AuthorizationRequest{
		Principal: params.Principal,
		Resource:  params.Resource,
		Action:    params.Action,
		ConditionContext: policy.ConditionContext{
			Principal: principalAttrs,
			Resource:  resourceAttrs,
		},
	}

	if a.evaluator.Evaluate(req, policies).IsAllowed() {
		return nil
	}

	return NewInsufficientPermissionsError(params.Principal, params.Resource, params.Action)
}

func (a *Authorizer) loadMemberships(ctx context.Context, conn pg.Conn, principalID gid.GID) (coredata.Memberships, error) {
	var memberships coredata.Memberships
	if err := memberships.LoadAllByIdentityID(ctx, conn, principalID); err != nil {
		return nil, fmt.Errorf("cannot load memberships: %w", err)
	}
	return memberships, nil
}

func (a *Authorizer) getActiveChildSessionForMembership(
	ctx context.Context,
	conn pg.Conn,
	rootSessionID gid.GID,
	membershipID gid.GID,
) (*coredata.Session, error) {
	childSession := &coredata.Session{}

	if err := childSession.LoadByRootSessionIDAndMembershipID(ctx, conn, rootSessionID, membershipID); err != nil {
		if err == coredata.ErrResourceNotFound {
			return nil, NewSessionNotFoundError(gid.Nil)
		}

		return nil, fmt.Errorf("cannot load child session: %w", err)
	}

	if childSession.ExpireReason != nil || time.Now().After(childSession.ExpiredAt) {
		return nil, NewSessionExpiredError(childSession.ID)
	}

	return childSession, nil
}

func (a *Authorizer) buildPrincipalAttributes(
	ctx context.Context,
	conn pg.Conn,
	principalID gid.GID,
	defaultAttrs map[string]string,
) (map[string]string, error) {
	attrs := map[string]string{
		"id": principalID.String(),
	}
	maps.Copy(attrs, defaultAttrs)

	if entity, ok := coredata.NewEntityFromID(principalID); ok {
		if attributer, ok := entity.(AuthorizationAttributer); ok {
			entityAttrs, err := attributer.AuthorizationAttributes(ctx, conn)
			if err != nil {
				return nil, fmt.Errorf("cannot load principal attributes: %w", err)
			}
			maps.Copy(attrs, entityAttrs)
		}
	}

	return attrs, nil
}

func (a *Authorizer) buildResourceAttributes(
	ctx context.Context,
	conn pg.Conn,
	params AuthorizeParams,
) (map[string]string, error) {
	attrs := map[string]string{
		"id": params.Resource.String(),
	}

	entity, ok := coredata.NewEntityFromID(params.Resource)
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %d", params.Resource.EntityType())
	}

	attributer, ok := entity.(AuthorizationAttributer)
	if !ok {
		return nil, fmt.Errorf("resource %d does not implement AuthorizationAttributer", params.Resource.EntityType())
	}

	entityAttrs, err := attributer.AuthorizationAttributes(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("cannot load resource attributes: %w", err)
	}
	maps.Copy(attrs, entityAttrs)

	if params.ResourceAttributes != nil {
		maps.Copy(attrs, params.ResourceAttributes)
	}

	return attrs, nil
}

func (a *Authorizer) buildPoliciesForRole(role string) []*policy.Policy {
	policies := append([]*policy.Policy{}, a.policySet.IdentityScopedPolicies...)

	if role != "" {
		policies = append(policies, a.policySet.RolePolicies[role]...)
	}

	return policies
}

func findMembershipForOrg(memberships coredata.Memberships, orgID string) *coredata.Membership {
	for _, m := range memberships {
		if m.OrganizationID.String() == orgID && m.State == coredata.MembershipStateActive {
			return m
		}
	}

	return nil
}
