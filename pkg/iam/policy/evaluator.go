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

import "go.probo.inc/probo/pkg/gid"

// Decision represents the result of a policy evaluation.
type Decision string

const (
	// DecisionAllow means access is explicitly allowed.
	DecisionAllow Decision = "allow"

	// DecisionDeny means access is explicitly denied.
	DecisionDeny Decision = "deny"

	// DecisionNoMatch means no policy statement matched (implicit deny).
	DecisionNoMatch Decision = "no_match"
)

// EvaluationResult contains the decision and context about how it was reached.
type EvaluationResult struct {
	// Decision is the final authorization decision.
	Decision Decision

	// MatchedStatement is the statement that produced the decision (if any).
	MatchedStatement *Statement

	// MatchedPolicy is the policy containing the matched statement (if any).
	MatchedPolicy *Policy
}

// IsAllowed returns true if access should be granted.
func (r EvaluationResult) IsAllowed() bool {
	return r.Decision == DecisionAllow
}

// AuthorizationRequest contains all information needed to evaluate access.
type AuthorizationRequest struct {
	// Principal is the actor requesting access.
	Principal gid.GID

	// Resource is the target resource.
	Resource gid.GID

	// Action is the operation being performed.
	Action string

	// ConditionContext provides attributes for condition evaluation.
	ConditionContext ConditionContext
}

// Evaluator evaluates policies to determine access decisions.
// Evaluation order: Explicit Deny > Explicit Allow > Implicit Deny
type Evaluator struct {
	matcher *ActionMatcher
}

// NewEvaluator creates a new policy evaluator.
func NewEvaluator() *Evaluator {
	return &Evaluator{
		matcher: NewActionMatcher(),
	}
}

// Evaluate evaluates a set of policies against an authorization request.
// Returns the decision and information about which policy/statement matched.
//
// Evaluation logic (AWS-style):
//  1. If any statement explicitly denies, return Deny
//  2. If any statement explicitly allows, return Allow
//  3. Otherwise, return NoMatch (implicit deny)
func (e *Evaluator) Evaluate(req AuthorizationRequest, policies []*Policy) EvaluationResult {
	var allowResult *EvaluationResult

	// First pass: check for explicit denies and collect allows
	for _, policy := range policies {
		for i := range policy.Statements {
			stmt := &policy.Statements[i]

			if !e.statementMatches(stmt, req) {
				continue
			}

			if stmt.Effect == EffectDeny {
				// Explicit deny - return immediately
				return EvaluationResult{
					Decision:         DecisionDeny,
					MatchedStatement: stmt,
					MatchedPolicy:    policy,
				}
			}

			if stmt.Effect == EffectAllow && allowResult == nil {
				// First matching allow - save it
				allowResult = &EvaluationResult{
					Decision:         DecisionAllow,
					MatchedStatement: stmt,
					MatchedPolicy:    policy,
				}
			}
		}
	}

	// No explicit deny found, check for allow
	if allowResult != nil {
		return *allowResult
	}

	// No matching statements - implicit deny
	return EvaluationResult{
		Decision: DecisionNoMatch,
	}
}

// statementMatches checks if a statement applies to the request.
func (e *Evaluator) statementMatches(stmt *Statement, req AuthorizationRequest) bool {
	// Check action match
	if !e.matcher.MatchesAny(stmt.Actions, req.Action) {
		return false
	}

	// Check resource match (if resources are specified)
	if len(stmt.Resources) > 0 {
		resourceMatched := false
		for _, pattern := range stmt.Resources {
			if pattern.MatchesResource(req.Resource) {
				resourceMatched = true
				break
			}
		}
		if !resourceMatched {
			return false
		}
	}

	// Check conditions (all must be satisfied)
	for _, condition := range stmt.Conditions {
		if !condition.Evaluate(req.ConditionContext) {
			return false
		}
	}

	return true
}
