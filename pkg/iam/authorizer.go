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

type (
	AuthorizationAttributes     = map[string]string
	AuthorizationAttributesByID = map[gid.GID]AuthorizationAttributes

	AuthorizationAttributer interface {
		AuthorizationAttributes(ctx context.Context, conn pg.Conn) (AuthorizationAttributes, error)
	}

	BulkAuthorizationAttributer interface {
		AuthorizationAttributes(ctx context.Context, conn pg.Conn) (AuthorizationAttributesByID, error)
	}

	AuthorizeParams struct {
		Principal          gid.GID
		Resource           gid.GID
		Session            *gid.GID
		Action             string
		ResourceAttributes AuthorizationAttributes
	}

	BulkAuthorizeParams struct {
		Principal gid.GID
		Resources []gid.GID
		Session   *gid.GID
		Action    string
	}

	Authorizer struct {
		pg        *pg.Client
		evaluator *policy.Evaluator
		policySet *PolicySet
	}

	evaluationContext struct {
		principalAttrs AuthorizationAttributes
		policies       []*policy.Policy
	}
)

func NewAuthorizer(pgClient *pg.Client) *Authorizer {
	return &Authorizer{
		pg:        pgClient,
		evaluator: policy.NewEvaluator(),
		policySet: NewPolicySet(),
	}
}

func (a *Authorizer) RegisterPolicySet(ps *PolicySet) {
	a.policySet.Merge(ps)
}

func (a *Authorizer) Authorize(ctx context.Context, params AuthorizeParams) error {
	if params.Principal.EntityType() != coredata.IdentityEntityType {
		return NewUnsupportedPrincipalTypeError(params.Principal.EntityType())
	}

	return a.pg.WithConn(ctx, func(conn pg.Conn) error { return a.authorize(ctx, conn, params) })
}

func (a *Authorizer) BulkAuthorize(ctx context.Context, params BulkAuthorizeParams) error {
	if params.Principal.EntityType() != coredata.IdentityEntityType {
		return NewUnsupportedPrincipalTypeError(params.Principal.EntityType())
	}

	return a.pg.WithConn(ctx, func(conn pg.Conn) error {
		return a.bulkAuthorize(ctx, conn, params)
	})
}

func (a *Authorizer) authorize(ctx context.Context, conn pg.Conn, params AuthorizeParams) error {
	resourceAttrs, err := a.buildResourceAttributes(ctx, conn, params)
	if err != nil {
		return fmt.Errorf("cannot build resource attributes: %w", err)
	}

	evalCtx, err := a.buildEvaluationContext(
		ctx,
		conn,
		params.Principal,
		params.Session,
		resourceAttrs["organization_id"],
	)
	if err != nil {
		return err
	}

	req := policy.AuthorizationRequest{
		Principal: params.Principal,
		Resource:  params.Resource,
		Action:    params.Action,
		ConditionContext: policy.ConditionContext{
			Principal: evalCtx.principalAttrs,
			Resource:  resourceAttrs,
		},
	}

	if a.evaluator.Evaluate(req, evalCtx.policies).IsAllowed() {
		return nil
	}

	return NewInsufficientPermissionsError(params.Principal, params.Resource, params.Action)
}

func (a *Authorizer) bulkAuthorize(ctx context.Context, conn pg.Conn, params BulkAuthorizeParams) error {
	if len(params.Resources) == 0 {
		return nil
	}

	entity, ok := coredata.NewEntitiesFromIDs(params.Resources)
	if !ok {
		return fmt.Errorf("unsupported or mixed resource types for bulk authorization")
	}

	collection, ok := entity.(BulkAuthorizationAttributer)
	if !ok {
		return fmt.Errorf("resource type %d does not implement BulkAuthorizationAttributer", params.Resources[0].EntityType())
	}

	allAttrs, err := collection.AuthorizationAttributes(ctx, conn)
	if err != nil {
		return fmt.Errorf("cannot load bulk resource attributes: %w", err)
	}

	orgID := allAttrs[params.Resources[0]]["organization_id"]
	for id, attrs := range allAttrs {
		if attrs["organization_id"] != orgID {
			return fmt.Errorf("cannot bulk authorize resources from different organizations")
		}
		attrs["id"] = id.String()
	}

	evalCtx, err := a.buildEvaluationContext(
		ctx,
		conn,
		params.Principal,
		params.Session,
		orgID,
	)
	if err != nil {
		return err
	}

	for _, resourceID := range params.Resources {
		req := policy.AuthorizationRequest{
			Principal: params.Principal,
			Resource:  resourceID,
			Action:    params.Action,
			ConditionContext: policy.ConditionContext{
				Principal: evalCtx.principalAttrs,
				Resource:  allAttrs[resourceID],
			},
		}

		if !a.evaluator.Evaluate(req, evalCtx.policies).IsAllowed() {
			return NewInsufficientPermissionsError(params.Principal, resourceID, params.Action)
		}
	}

	return nil
}

func (a *Authorizer) buildEvaluationContext(
	ctx context.Context,
	conn pg.Conn,
	principalID gid.GID,
	session *gid.GID,
	resourceOrgID string,
) (*evaluationContext, error) {
	membership, err := a.loadMembership(ctx, conn, principalID, resourceOrgID)
	if err != nil {
		return nil, fmt.Errorf("cannot load memberships for principal: %w", err)
	}

	if membership != nil && session != nil {
		if _, err := a.getActiveChildSessionForMembership(
			ctx,
			conn,
			*session,
			membership.ID,
		); err != nil {
			var errSessionNotFound *ErrSessionNotFound
			var errSessionExpired *ErrSessionExpired

			if errors.As(err, &errSessionNotFound) || errors.As(err, &errSessionExpired) {
				return nil, NewAssumptionRequiredError(principalID, membership.ID)
			}

			return nil, fmt.Errorf("cannot get active child session for membership: %w", err)
		}
	}

	var role string
	if membership != nil {
		role = membership.Role.String()
	}

	var scopedPrincipalAttrs AuthorizationAttributes
	if membership != nil && role != "" {
		scopedPrincipalAttrs = AuthorizationAttributes{
			"organization_id": membership.OrganizationID.String(),
			"role":            membership.Role.String(),
		}
	}

	principalAttrs, err := a.buildPrincipalAttributes(ctx, conn, principalID, scopedPrincipalAttrs)
	if err != nil {
		return nil, fmt.Errorf("cannot build principal attributes: %w", err)
	}

	return &evaluationContext{
		principalAttrs: principalAttrs,
		policies:       a.buildPoliciesForRole(role),
	}, nil
}

func (a *Authorizer) loadMembership(
	ctx context.Context,
	conn pg.Conn,
	principalID gid.GID,
	resourceOrgID string,
) (*coredata.Membership, error) {
	if resourceOrgID == "" {
		return nil, nil
	}

	orgID, err := gid.ParseGID(resourceOrgID)
	if err != nil {
		return nil, fmt.Errorf("cannot parse gid: %w", err)
	}

	membership := &coredata.Membership{}
	if err := membership.LoadActiveByIdentityIDAndOrganizationID(ctx, conn, principalID, orgID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load active membership: %w", err)
	}

	return membership, nil
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
	defaultAttrs AuthorizationAttributes,
) (AuthorizationAttributes, error) {
	attrs := AuthorizationAttributes{
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
) (AuthorizationAttributes, error) {
	attrs := AuthorizationAttributes{
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
