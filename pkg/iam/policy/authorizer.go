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

package policy

import (
	"errors"
	"fmt"

	"go.probo.inc/probo/pkg/gid"
)

var (
	// ErrAccessDenied is returned when access is explicitly denied.
	ErrAccessDenied = errors.New("access denied")

	// ErrNoMatchingPolicy is returned when no policy grants access (implicit deny).
	ErrNoMatchingPolicy = errors.New("no matching policy")
)

// AccessDeniedError provides detailed information about why access was denied.
type AccessDeniedError struct {
	Principal gid.GID
	Resource  gid.GID
	Action    string
	Reason    string
	Statement *Statement // The statement that denied access (if explicit deny)
}

func (e *AccessDeniedError) Error() string {
	if e.Statement != nil && e.Statement.SID != "" {
		return fmt.Sprintf("access denied: principal %s cannot perform %s on %s (denied by %s)",
			e.Principal, e.Action, e.Resource, e.Statement.SID)
	}
	return fmt.Sprintf("access denied: principal %s cannot perform %s on %s: %s",
		e.Principal, e.Action, e.Resource, e.Reason)
}

func (e *AccessDeniedError) Unwrap() error {
	return ErrAccessDenied
}

// Authorizer evaluates policies to authorize actions.
type Authorizer struct {
	evaluator *Evaluator
	registry  *ActionRegistry
}

// NewAuthorizer creates a new authorizer with the given action registry.
func NewAuthorizer(registry *ActionRegistry) *Authorizer {
	return &Authorizer{
		evaluator: NewEvaluator(),
		registry:  registry,
	}
}

// AuthorizeParams contains all parameters for an authorization check.
type AuthorizeParams struct {
	// Principal is the actor requesting access.
	Principal gid.GID

	// Resource is the target resource.
	Resource gid.GID

	// Action is the operation being performed.
	Action string

	// Policies are the policies to evaluate (typically role-based + self-manage).
	Policies []*Policy

	// ResourceAttributes provides additional attributes about the resource
	// for condition evaluation (e.g., owner_id, tenant_id).
	ResourceAttributes map[string]string
}

// Authorize checks if the action is allowed based on the provided policies.
// Returns nil if allowed, or an error describing why access was denied.
func (a *Authorizer) Authorize(params AuthorizeParams) error {
	// Validate action exists in registry (optional - can be disabled for flexibility)
	if a.registry != nil && !a.registry.Exists(Action(params.Action)) {
		return &AccessDeniedError{
			Principal: params.Principal,
			Resource:  params.Resource,
			Action:    params.Action,
			Reason:    "unknown action",
		}
	}

	// Build condition context
	conditionCtx := ConditionContext{
		Principal: map[string]string{
			"id": params.Principal.String(),
		},
		Resource: map[string]string{
			"id": params.Resource.String(),
		},
	}

	// Add resource attributes to context
	for k, v := range params.ResourceAttributes {
		conditionCtx.Resource[k] = v
	}

	// Build authorization request
	req := AuthorizationRequest{
		Principal:        params.Principal,
		Resource:         params.Resource,
		Action:           params.Action,
		ConditionContext: conditionCtx,
	}

	// Evaluate policies
	result := a.evaluator.Evaluate(req, params.Policies)

	switch result.Decision {
	case DecisionAllow:
		return nil

	case DecisionDeny:
		return &AccessDeniedError{
			Principal: params.Principal,
			Resource:  params.Resource,
			Action:    params.Action,
			Reason:    "explicitly denied",
			Statement: result.MatchedStatement,
		}

	case DecisionNoMatch:
		return &AccessDeniedError{
			Principal: params.Principal,
			Resource:  params.Resource,
			Action:    params.Action,
			Reason:    "no policy allows this action",
		}

	default:
		return &AccessDeniedError{
			Principal: params.Principal,
			Resource:  params.Resource,
			Action:    params.Action,
			Reason:    "unexpected evaluation result",
		}
	}
}

// IsAllowed is a convenience method that returns true if access is allowed.
func (a *Authorizer) IsAllowed(params AuthorizeParams) bool {
	return a.Authorize(params) == nil
}
