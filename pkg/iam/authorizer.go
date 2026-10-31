// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package iam

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
	"go.probo.inc/probo/pkg/iam/policy"
)

// AuthorizationAttributer is implemented by entities that can provide
// authorization attributes for multiple resources in one query.
type AuthorizationAttributer interface {
	AuthorizationAttributes(
		ctx context.Context,
		conn pg.Querier,
		resourceIDs []gid.GID,
	) (policy.AttributesByID, error)
}

// AuthorizeParams contains the parameters for an authorization request.
type AuthorizeParams struct {
	Principal           gid.GID
	Resource            gid.GID
	Session             *gid.GID
	Action              string
	ResourceAttributes  policy.Attributes
	DryRun              bool
	SkipAssumptionCheck bool
}

// AuthorizeBatchParams contains the parameters for a batch authorization request.
type AuthorizeBatchParams struct {
	Principal           gid.GID
	Session             *gid.GID
	Action              string
	Resources           []gid.GID
	ResourceAttributes  policy.Attributes
	DryRun              bool
	SkipAssumptionCheck bool
}

// MultiAuthorizeItem contains one authorization request in a multi-authorization batch.
// ResourceAttributes are merged on top of the resource attributes loaded
// for the resource before the policy is evaluated.
type MultiAuthorizeItem struct {
	Resource            gid.GID
	Action              string
	ResourceAttributes  policy.Attributes
	DryRun              bool
	SkipAssumptionCheck bool
}

// AuthorizeMultiParams contains the parameters for a multi-authorization request.
type AuthorizeMultiParams struct {
	Principal gid.GID
	Session   *gid.GID
	Items     []MultiAuthorizeItem
}

// Authorizer evaluates authorization requests against registered policies.
type Authorizer struct {
	pg            *pg.Client
	evaluator     *policy.Evaluator
	policySet     *PolicySet
	scopeRegistry *oauth2scope.Registry
	logger        *log.Logger
}

// NewAuthorizer creates a new Authorizer instance.
func NewAuthorizer(pgClient *pg.Client, logger *log.Logger, scopeRegistry *oauth2scope.Registry) *Authorizer {
	return &Authorizer{
		pg:            pgClient,
		evaluator:     policy.NewEvaluator(),
		policySet:     NewPolicySet(),
		scopeRegistry: scopeRegistry,
		logger:        logger,
	}
}

// RegisterPolicySet merges the given policy set into the authorizer.
func (a *Authorizer) RegisterPolicySet(ps *PolicySet) {
	a.policySet.Merge(ps)
}

func (a *Authorizer) checkOAuth2Scope(
	ctx context.Context,
	principal gid.GID,
	action Action,
) error {
	accessToken, ok := oauth2.AccessTokenFromContext(ctx)
	if !ok {
		return nil
	}

	if a.scopeRegistry == nil || !a.scopeRegistry.Allows(accessToken.Scopes, action) {
		var scopes []coredata.OAuth2Scope
		if a.scopeRegistry != nil {
			scopes = a.scopeRegistry.ScopesForAction(action)
		}

		return NewInsufficientOAuth2ScopeError(principal, scopes...)
	}

	return nil
}

// Authorize checks if the principal is allowed to perform the action on the resource.
func (a *Authorizer) Authorize(ctx context.Context, params AuthorizeParams) (*coredata.Scope, error) {
	return a.AuthorizeBatch(
		ctx,
		AuthorizeBatchParams{
			Principal:           params.Principal,
			Session:             params.Session,
			Action:              params.Action,
			Resources:           []gid.GID{params.Resource},
			ResourceAttributes:  params.ResourceAttributes,
			DryRun:              params.DryRun,
			SkipAssumptionCheck: params.SkipAssumptionCheck,
		},
	)
}

// AuthorizeBatch checks whether the principal is allowed to perform the action
// on all provided resources.
func (a *Authorizer) AuthorizeBatch(ctx context.Context, params AuthorizeBatchParams) (*coredata.Scope, error) {
	if params.Principal.EntityType() != coredata.IdentityEntityType {
		return nil, NewUnsupportedPrincipalTypeError(params.Principal.EntityType())
	}

	if len(params.Resources) == 0 {
		return nil, NewEmptyResourceBatchError(params.Action)
	}

	expectedEntityType := params.Resources[0].EntityType()
	items := make([]MultiAuthorizeItem, 0, len(params.Resources))

	for _, resourceID := range params.Resources {
		if resourceID.EntityType() != expectedEntityType {
			entityTypes := make([]uint16, 0, len(params.Resources))
			for _, r := range params.Resources {
				entityTypes = append(entityTypes, r.EntityType())
			}

			return nil, NewMixedEntityTypeBatchError(
				params.Action,
				uniqueSortedEntityTypes(entityTypes),
			)
		}

		items = append(
			items,
			MultiAuthorizeItem{
				Resource:            resourceID,
				Action:              params.Action,
				DryRun:              params.DryRun,
				SkipAssumptionCheck: params.SkipAssumptionCheck,
			},
		)
	}

	var (
		scope  *coredata.Scope
		denial error
	)

	if err := a.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			authorizedScope, decision, err := a.authorizeMulti(
				ctx,
				tx,
				AuthorizeMultiParams{
					Principal: params.Principal,
					Session:   params.Session,
					Items:     items,
				},
				params.ResourceAttributes,
			)
			if err != nil {
				return err
			}

			scope = authorizedScope
			denial = decision

			return nil
		},
	); err != nil {
		return nil, err
	}

	if denial != nil {
		return nil, denial
	}

	return scope, nil
}

// AuthorizeMulti evaluates each item independently and returns one decision
// per item: a nil entry means allowed, a non-nil entry carries the iam
// error explaining the denial (typically *ErrInsufficientPermissions).
//
// Audit log entries for non-dry-run items are written in a single bulk
// insert, including denials. Use AuthorizeBatch instead when callers want
// all-or-nothing semantics for a homogeneous batch.
func (a *Authorizer) AuthorizeMulti(
	ctx context.Context,
	params AuthorizeMultiParams,
) (*coredata.Scope, []error, error) {
	if params.Principal.EntityType() != coredata.IdentityEntityType {
		return nil, nil, NewUnsupportedPrincipalTypeError(params.Principal.EntityType())
	}

	if len(params.Items) == 0 {
		return nil, nil, NewEmptyResourceBatchError("")
	}

	var (
		scope     *coredata.Scope
		decisions []error
	)

	if err := a.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			s, itemAttrs, d, err := a.evaluateMultiInTx(ctx, tx, params, nil)
			if err != nil {
				return err
			}

			if err := a.insertAuditEntries(ctx, tx, s, params, itemAttrs, d, false); err != nil {
				return err
			}

			scope = s
			decisions = d

			return nil
		},
	); err != nil {
		return nil, nil, err
	}

	return scope, decisions, nil
}

func (a *Authorizer) authorizeMulti(
	ctx context.Context,
	tx pg.Tx,
	params AuthorizeMultiParams,
	extraResourceAttributes policy.Attributes,
) (*coredata.Scope, error, error) {
	scope, itemAttrs, decisions, err := a.evaluateMultiInTx(
		ctx,
		tx,
		params,
		extraResourceAttributes,
	)
	if err != nil {
		return nil, nil, err
	}

	var denial error

	for _, decision := range decisions {
		if decision != nil {
			denial = decision
			break
		}
	}

	if err := a.insertAuditEntries(ctx, tx, scope, params, itemAttrs, decisions, denial != nil); err != nil {
		return nil, nil, err
	}

	if denial != nil {
		return nil, denial, nil
	}

	return scope, nil, nil
}

// evaluateMultiInTx evaluates every item in params against the loaded
// policies and resource attributes. It returns the shared scope, the merged
// per-item resource attributes (used for both evaluation and audit log
// building), and a parallel slice of per-item decisions (nil = allowed).
// Callers decide which decisions to persist to the audit log.
func (a *Authorizer) evaluateMultiInTx(
	ctx context.Context,
	tx pg.Tx,
	params AuthorizeMultiParams,
	extraResourceAttributes policy.Attributes,
) (*coredata.Scope, []policy.Attributes, []error, error) {
	uniqueResourceIDs := make([]gid.GID, 0, len(params.Items))
	seenResourceIDs := make(map[gid.GID]struct{}, len(params.Items))
	requiresAssumptionCheck := false

	for _, item := range params.Items {
		if !item.SkipAssumptionCheck {
			requiresAssumptionCheck = true
		}

		if _, ok := seenResourceIDs[item.Resource]; ok {
			continue
		}

		seenResourceIDs[item.Resource] = struct{}{}
		uniqueResourceIDs = append(uniqueResourceIDs, item.Resource)
	}

	resourceAttrsByResourceID, err := a.buildResourceAttributesBatch(
		ctx,
		tx,
		uniqueResourceIDs,
		extraResourceAttributes,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot build resource attributes batch: %w", err)
	}

	actionForErrors := params.Items[0].Action

	resourceOrgID := resourceAttrsByResourceID[uniqueResourceIDs[0]]["organization_id"]
	for _, resourceID := range uniqueResourceIDs[1:] {
		if resourceAttrsByResourceID[resourceID]["organization_id"] != resourceOrgID {
			orgIDs := make([]string, 0, len(uniqueResourceIDs))
			for _, id := range uniqueResourceIDs {
				orgIDs = append(orgIDs, resourceAttrsByResourceID[id]["organization_id"])
			}

			return nil, nil, nil, NewMixedOrganizationBatchError(
				actionForErrors,
				uniqueSortedStrings(orgIDs),
			)
		}
	}

	membership, err := a.loadMembership(ctx, tx, params.Principal, resourceOrgID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot load memberships for principal: %w", err)
	}

	// The assumption check is a property of (principal, membership, session),
	// so it only runs once even though SkipAssumptionCheck is per-item.
	// On failure, ErrAssumptionRequired is recorded only against items that
	// did not opt out.
	var assumptionErr error

	if requiresAssumptionCheck {
		err := a.checkAssumption(
			ctx,
			tx,
			params.Principal,
			params.Session,
			membership,
			false,
		)
		if err != nil {
			if _, ok := errors.AsType[*ErrAssumptionRequired](err); !ok {
				return nil, nil, nil, err
			}

			assumptionErr = err
		}
	}

	var role string
	if membership != nil {
		role = membership.Role.String()
	}

	policies := a.buildPoliciesForRole(role)

	var scopedPrincipalAttrs policy.Attributes
	if membership != nil && role != "" {
		scopedPrincipalAttrs = policy.Attributes{
			"organization_id": membership.OrganizationID.String(),
			"role":            membership.Role.String(),
		}

		if policiesReferenceProfileID(policies, params) {
			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByIdentityIDAndOrganizationID(
				ctx,
				tx,
				coredata.NewScopeFromObjectID(membership.OrganizationID),
				params.Principal,
				membership.OrganizationID,
			); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return nil, nil, nil, fmt.Errorf("cannot load membership profile: %w", err)
				}
			} else {
				scopedPrincipalAttrs["profile_id"] = profile.ID.String()
			}
		}
	}

	principalAttrs, err := a.buildPrincipalAttributes(ctx, tx, params.Principal, scopedPrincipalAttrs)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot build principal attributes: %w", err)
	}

	if params.Session != nil {
		principalAttrs["session_id"] = params.Session.String()
	}

	decisions := make([]error, len(params.Items))
	itemAttrs := make([]policy.Attributes, len(params.Items))

	for i, item := range params.Items {
		resAttrs := resourceAttrsByResourceID[item.Resource]
		if len(item.ResourceAttributes) > 0 {
			merged := maps.Clone(resAttrs)
			maps.Copy(merged, item.ResourceAttributes)
			resAttrs = merged
		}

		itemAttrs[i] = resAttrs

		if assumptionErr != nil && !item.SkipAssumptionCheck {
			decisions[i] = assumptionErr
			a.logDecision(
				ctx,
				DecisionRecord{
					Effect:     effectError,
					Action:     item.Action,
					ResourceID: item.Resource,
					Principal:  params.Principal,
					Reason:     assumptionErr.Error(),
				},
			)

			continue
		}

		if err := a.checkOAuth2Scope(ctx, params.Principal, item.Action); err != nil {
			decisions[i] = err
			a.logDecision(
				ctx,
				DecisionRecord{
					Effect:     effectError,
					Action:     item.Action,
					ResourceID: item.Resource,
					Principal:  params.Principal,
					Reason:     err.Error(),
				},
			)

			continue
		}

		req := policy.AuthorizationRequest{
			Principal: params.Principal,
			Resource:  item.Resource,
			Action:    item.Action,
			ConditionContext: policy.ConditionContext{
				Principal: principalAttrs,
				Resource:  resAttrs,
			},
		}

		startedAt := time.Now()
		result := a.evaluator.Evaluate(req, policies)

		a.logDecision(
			ctx,
			newDecisionRecord(
				result,
				params.Principal,
				item.Resource,
				item.Action,
				role,
				time.Since(startedAt),
			),
		)

		if !result.IsAllowed() {
			decisions[i] = NewInsufficientPermissionsError(params.Principal, item.Resource, item.Action)
		}
	}

	scope := coredata.NewScope(uniqueResourceIDs[0].TenantID(), params.Principal)

	if resourceOrgID != "" {
		orgID, _ := gid.ParseGID(resourceOrgID)
		scope = coredata.NewScope(orgID.TenantID(), params.Principal)
	}

	return scope, itemAttrs, decisions, nil
}

func (a *Authorizer) buildResourceAttributesBatch(
	ctx context.Context,
	conn pg.Querier,
	uniqueResourceIDs []gid.GID,
	extraResourceAttributes policy.Attributes,
) (policy.AttributesByID, error) {
	resourceAttrsByID, err := a.loadResourceAttributesByType(ctx, conn, uniqueResourceIDs)
	if err != nil {
		return nil, err
	}

	resourceAttrsByResourceID := make(policy.AttributesByID, len(uniqueResourceIDs))
	for _, resourceID := range uniqueResourceIDs {
		resourceAttrs := resourceAttrsByID[resourceID]

		attrs := policy.Attributes{
			"id": resourceID.String(),
		}
		maps.Copy(attrs, resourceAttrs)

		if extraResourceAttributes != nil {
			maps.Copy(attrs, extraResourceAttributes)
		}

		resourceAttrsByResourceID[resourceID] = attrs
	}

	return resourceAttrsByResourceID, nil
}

func (a *Authorizer) loadResourceAttributesByType(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	resourceIDsByEntityType := make(map[uint16][]gid.GID)
	orderedEntityTypes := make([]uint16, 0, len(resourceIDs))

	for _, resourceID := range resourceIDs {
		entityType := resourceID.EntityType()

		if _, ok := resourceIDsByEntityType[entityType]; !ok {
			orderedEntityTypes = append(orderedEntityTypes, entityType)
		}

		resourceIDsByEntityType[entityType] = append(resourceIDsByEntityType[entityType], resourceID)
	}

	resourceAttrsByID := make(policy.AttributesByID, len(resourceIDs))

	for _, entityType := range orderedEntityTypes {
		groupResourceIDs := resourceIDsByEntityType[entityType]

		entity, ok := coredata.NewEntityFromID(groupResourceIDs[0])
		if !ok {
			return nil, fmt.Errorf("unsupported resource type: %d", groupResourceIDs[0].EntityType())
		}

		attributer, ok := entity.(AuthorizationAttributer)
		if !ok {
			return nil, NewBatchAuthorizationUnsupportedResourceTypeError(entityType)
		}

		groupAttrsByID, err := attributer.AuthorizationAttributes(ctx, conn, groupResourceIDs)
		if err != nil {
			return nil, fmt.Errorf(
				"cannot load batched resource attributes for entity type %d: %w",
				entityType,
				err,
			)
		}

		for _, resourceID := range groupResourceIDs {
			resourceAttrs, ok := groupAttrsByID[resourceID]
			if !ok {
				return nil, coredata.ErrResourceNotFound
			}

			resourceAttrsByID[resourceID] = resourceAttrs
		}
	}

	return resourceAttrsByID, nil
}

func (a *Authorizer) checkAssumption(
	ctx context.Context,
	tx pg.Tx,
	principalID gid.GID,
	sessionID *gid.GID,
	membership *coredata.Membership,
	skipAssumptionCheck bool,
) error {
	if membership == nil || sessionID == nil || skipAssumptionCheck {
		return nil
	}

	if _, err := a.getActiveChildSessionForMembership(
		ctx,
		tx,
		*sessionID,
		membership.ID,
	); err != nil {
		if _, ok := errors.AsType[*ErrSessionNotFound](err); ok {
			return NewAssumptionRequiredError(principalID, membership.ID)
		}

		if _, ok := errors.AsType[*ErrSessionExpired](err); ok {
			return NewAssumptionRequiredError(principalID, membership.ID)
		}

		return fmt.Errorf("cannot get active child session for membership: %w", err)
	}

	return nil
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
	defaultAttrs policy.Attributes,
) (policy.Attributes, error) {
	attrs := policy.Attributes{
		"id": principalID.String(),
	}
	maps.Copy(attrs, defaultAttrs)

	if entity, ok := coredata.NewEntityFromID(principalID); ok {
		attributer, ok := entity.(AuthorizationAttributer)
		if !ok {
			return nil, NewBatchAuthorizationUnsupportedResourceTypeError(principalID.EntityType())
		}

		entityAttrsByID, err := attributer.AuthorizationAttributes(ctx, conn, []gid.GID{principalID})
		if err != nil {
			return nil, fmt.Errorf("cannot load principal attributes: %w", err)
		}

		entityAttrs, ok := entityAttrsByID[principalID]
		if !ok {
			return nil, coredata.ErrResourceNotFound
		}

		maps.Copy(attrs, entityAttrs)
	}

	return attrs, nil
}

func policiesReferenceProfileID(policies []*policy.Policy, params AuthorizeMultiParams) bool {
	matcher := policy.ActionMatcher{}

	for _, item := range policies {
		for _, statement := range item.Statements {
			applies := false

			for _, request := range params.Items {
				for _, pattern := range statement.Actions {
					if matcher.Matches(pattern, request.Action) {
						applies = true
					}
				}
			}

			if !applies {
				continue
			}

			for _, condition := range statement.Conditions {
				if strings.Contains(condition.Key, "profile_id") {
					return true
				}

				for _, value := range condition.Values {
					if strings.Contains(value, "profile_id") {
						return true
					}
				}
			}
		}
	}

	return false
}

func (a *Authorizer) buildPoliciesForRole(role string) []*policy.Policy {
	policies := append([]*policy.Policy{}, a.policySet.IdentityScopedPolicies...)

	if role != "" {
		policies = append(policies, a.policySet.RolePolicies[role]...)
	}

	return policies
}

func uniqueSortedStrings(values []string) []string {
	set := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))

	for _, value := range values {
		if _, ok := set[value]; ok {
			continue
		}

		set[value] = struct{}{}
		unique = append(unique, value)
	}

	slices.Sort(unique)

	return unique
}

func uniqueSortedEntityTypes(values []uint16) []uint16 {
	set := make(map[uint16]struct{}, len(values))
	unique := make([]uint16, 0, len(values))

	for _, value := range values {
		if _, ok := set[value]; ok {
			continue
		}

		set[value] = struct{}{}
		unique = append(unique, value)
	}

	slices.Sort(unique)

	return unique
}

// resourceTypeFromAction extracts the PascalCase resource type from an
// action string, e.g. "core:webhook-subscription:delete" -> "WebhookSubscription".
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

type actorSnapshot struct {
	email *string
	name  *string
}

func (a *Authorizer) insertAuditEntries(
	ctx context.Context,
	tx pg.Tx,
	scope *coredata.Scope,
	params AuthorizeMultiParams,
	itemAttrs []policy.Attributes,
	decisions []error,
	denialsOnly bool,
) error {
	indexes := make([]int, 0, len(params.Items))

	for i, item := range params.Items {
		if denialsOnly && decisions[i] == nil {
			continue
		}

		if item.DryRun {
			continue
		}

		indexes = append(indexes, i)
	}

	if len(indexes) == 0 {
		return nil
	}

	snapshot := a.loadActorSnapshot(ctx, tx, params.Principal)
	now := time.Now()
	request, _ := AuditRequestFromContext(ctx)
	entries := make(coredata.AuditLogEntries, 0, len(indexes))

	for _, i := range indexes {
		item := params.Items[i]

		entry := a.buildAuditLogEntry(
			ctx,
			AuthorizeParams{
				Principal: params.Principal,
				Resource:  item.Resource,
				Session:   params.Session,
				Action:    item.Action,
				DryRun:    item.DryRun,
			},
			itemAttrs[i],
			decisions[i],
			snapshot,
			request,
			now,
		)
		if entry == nil {
			continue
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil
	}

	if err := entries.BulkInsert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot bulk insert audit log entries: %w", err)
	}

	return nil
}

func (a *Authorizer) loadActorSnapshot(
	ctx context.Context,
	conn pg.Querier,
	principalID gid.GID,
) actorSnapshot {
	identity := &coredata.Identity{}
	if err := identity.LoadByID(ctx, conn, principalID); err != nil {
		a.logger.ErrorCtx(
			ctx,
			"cannot load audit log actor",
			log.Error(err),
			log.String("principal_id", principalID.String()),
		)

		return actorSnapshot{}
	}

	email := identity.EmailAddress.String()
	name := identity.FullName

	return actorSnapshot{
		email: &email,
		name:  &name,
	}
}

// buildAuditLogEntry returns nil when no entry should be recorded
// (dry run or missing/invalid organization id).
func (a *Authorizer) buildAuditLogEntry(
	ctx context.Context,
	params AuthorizeParams,
	resourceAttrs policy.Attributes,
	decision error,
	snapshot actorSnapshot,
	request AuditRequest,
	now time.Time,
) *coredata.AuditLogEntry {
	if params.DryRun {
		return nil
	}

	orgIDStr := resourceAttrs["organization_id"]
	if orgIDStr == "" {
		return nil
	}

	orgID, err := gid.ParseGID(orgIDStr)
	if err != nil {
		a.logger.ErrorCtx(
			ctx,
			"cannot parse organization id for audit log",
			log.Error(err),
		)

		return nil
	}

	var actorType coredata.AuditLogActorType
	if params.Session != nil {
		actorType = coredata.AuditLogActorTypeUser
	} else {
		actorType = coredata.AuditLogActorTypeAPIKey
	}

	outcome := coredata.AuditLogOutcomeAllowed

	var denialReason *string

	if decision != nil {
		outcome = coredata.AuditLogOutcomeDenied
		reason := truncateAuditString(decision.Error(), 1024)
		denialReason = &reason
	}

	return &coredata.AuditLogEntry{
		ID:             gid.New(orgID.TenantID(), coredata.AuditLogEntryEntityType),
		OrganizationID: orgID,
		ActorID:        params.Principal,
		ActorType:      actorType,
		ActorEmail:     snapshot.email,
		ActorName:      snapshot.name,
		Action:         params.Action,
		ResourceType:   resourceTypeFromAction(params.Action),
		ResourceID:     params.Resource,
		Outcome:        outcome,
		DenialReason:   denialReason,
		Metadata:       []byte("{}"),
		OriginIP:       optionalAuditString(request.IP),
		UserAgent:      optionalAuditString(truncateAuditString(request.UserAgent, 512)),
		RequestID:      optionalAuditString(truncateAuditString(request.RequestID, 128)),
		CreatedAt:      now,
		RecordedAt:     now,
	}
}

func optionalAuditString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

func truncateAuditString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}

	cut := value[:limit]
	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}

	return cut
}
