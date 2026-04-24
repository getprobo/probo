// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
)

// AuthorizationAttributer is implemented by entities that provide attributes
// for policy condition evaluation.
type AuthorizationAttributer interface {
	AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error)
}

// AuthorizeParams contains the parameters for an authorization request.
type AuthorizeParams struct {
	Principal           gid.GID
	Resource            gid.GID
	Session             *gid.GID
	Action              string
	ResourceAttributes  map[string]string
	DryRun              bool
	SkipAssumptionCheck bool
}

// Authorizer evaluates authorization requests against registered policies.
type Authorizer struct {
	pg        *pg.Client
	evaluator *policy.Evaluator
	policySet *PolicySet
	logger    *log.Logger
}

// NewAuthorizer creates a new Authorizer instance.
func NewAuthorizer(pgClient *pg.Client, logger *log.Logger) *Authorizer {
	return &Authorizer{
		pg:        pgClient,
		evaluator: policy.NewEvaluator(),
		policySet: NewPolicySet(),
		logger:    logger,
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

	return a.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error { return a.authorize(ctx, tx, params) })
}

func (a *Authorizer) authorize(ctx context.Context, tx pg.Tx, params AuthorizeParams) error {
	resourceAttrs, err := a.buildResourceAttributes(ctx, tx, params)
	if err != nil {
		return fmt.Errorf("cannot build resource attributes: %w", err)
	}

	resourceOrgID := resourceAttrs["organization_id"]

	// Find role for resource's organization
	membership, err := a.loadMembership(ctx, tx, params.Principal, resourceOrgID)
	if err != nil {
		return fmt.Errorf("cannot load memberships for principal: %w", err)
	}

	// Check whether the viewer is currently assuming the org of the accessed resource
	if membership != nil && params.Session != nil && !params.SkipAssumptionCheck {
		if _, err := a.getActiveChildSessionForMembership(
			ctx,
			tx,
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

	principalAttrs, err := a.buildPrincipalAttributes(ctx, tx, params.Principal, scopedPrincipalAttrs)
	if err != nil {
		return fmt.Errorf("cannot build principal attributes: %w", err)
	}

	if params.Session != nil {
		principalAttrs["session_id"] = params.Session.String()
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
		a.recordAuditLog(ctx, tx, params, resourceAttrs)
		return nil
	}

	return NewInsufficientPermissionsError(params.Principal, params.Resource, params.Action)
}

func (a *Authorizer) loadMembership(
	ctx context.Context,
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Querier,
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

// resourceTypeFromAction extracts the resource type name from an action
// string. For example, "core:third-party:create" returns "ThirdParty" and
// "core:webhook-subscription:delete" returns "WebhookSubscription".
func resourceTypeFromAction(action string) string {
	parts := strings.Split(action, ":")
	if len(parts) < 3 {
		return "Unknown"
	}

	segments := strings.Split(parts[1], "-")
	for i, s := range segments {
		if len(s) > 0 {
			segments[i] = strings.ToUpper(s[:1]) + s[1:]
		}
	}

	return strings.Join(segments, "")
}

func (a *Authorizer) recordAuditLog(
	ctx context.Context,
	tx pg.Tx,
	params AuthorizeParams,
	resourceAttrs map[string]string,
) {
	if params.DryRun {
		return
	}

	orgIDStr := resourceAttrs["organization_id"]
	if orgIDStr == "" {
		return
	}

	orgID, err := gid.ParseGID(orgIDStr)
	if err != nil {
		a.logger.ErrorCtx(
			ctx,
			"cannot parse organization id for audit log",
			log.Error(err),
		)
		return
	}

	var actorType coredata.AuditLogActorType
	if params.Session != nil {
		actorType = coredata.AuditLogActorTypeUser
	} else {
		actorType = coredata.AuditLogActorTypeAPIKey
	}

	resourceType := resourceTypeFromAction(params.Action)

	metadata, err := json.Marshal(map[string]any{})
	if err != nil {
		a.logger.ErrorCtx(
			ctx,
			"cannot marshal audit log metadata",
			log.Error(err),
		)
		return
	}

	entry := &coredata.AuditLogEntry{
		ID:             gid.New(orgID.TenantID(), coredata.AuditLogEntryEntityType),
		OrganizationID: orgID,
		ActorID:        params.Principal,
		ActorType:      actorType,
		Action:         params.Action,
		ResourceType:   resourceType,
		ResourceID:     params.Resource,
		Metadata:       metadata,
		CreatedAt:      time.Now(),
	}

	scope := coredata.NewScope(orgID.TenantID())

	if err := entry.Insert(ctx, tx, scope); err != nil {
		a.logger.ErrorCtx(
			ctx,
			"cannot insert audit log entry",
			log.Error(err),
			log.String("action", params.Action),
			log.String("resource_id", params.Resource.String()),
		)
	}
}
