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

package policy

import (
	"strings"

	"go.probo.inc/probo/pkg/gid"
)

// Effect represents whether a statement allows or denies access.
type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

// Statement represents a single permission rule within a policy.
// A statement specifies what actions are allowed or denied on what resources,
// with optional conditions for attribute-based access control.
type Statement struct {
	// SID is an optional identifier for the statement (useful for debugging).
	SID string

	// Effect specifies whether this statement allows or denies access.
	Effect Effect

	// Actions is the list of actions this statement applies to.
	// Supports wildcards: "documents:*", "*:*:read", "*"
	Actions []string

	// Resources defines which resources this statement applies to.
	// If empty, applies to all resources.
	Resources []ResourcePattern

	// Conditions are optional attribute-based constraints.
	// All conditions must be satisfied for the statement to apply.
	Conditions []Condition
}

// ResourcePattern defines a pattern for matching resources.
// Nil fields act as wildcards (match any value).
type ResourcePattern struct {
	// TenantID restricts to a specific tenant. Nil matches any tenant.
	TenantID *gid.TenantID

	// EntityType restricts to a specific entity type. Nil matches any type.
	EntityType *uint16
}

// MatchesResource checks if the pattern matches a given resource GID.
func (p ResourcePattern) MatchesResource(resource gid.GID) bool {
	if p.TenantID != nil && *p.TenantID != resource.TenantID() {
		return false
	}

	if p.EntityType != nil && *p.EntityType != resource.EntityType() {
		return false
	}

	return true
}

// Condition represents an attribute-based access control constraint.
// Example: principal.id == resource.owner_id
type Condition struct {
	// Operator is the comparison operator.
	Operator ConditionOperator

	// Key is the attribute path to check (e.g., "principal.id", "resource.owner_id").
	Key string

	// Values are the values to compare against.
	Values []string
}

// ConditionOperator defines how to compare condition values.
type ConditionOperator string

const (
	// ConditionEquals checks if the key value equals any of the specified values.
	ConditionEquals ConditionOperator = "Equals"

	// ConditionNotEquals checks if the key value does not equal any of the specified values.
	ConditionNotEquals ConditionOperator = "NotEquals"

	// ConditionIn checks if the key value is in the list of values.
	ConditionIn ConditionOperator = "In"

	// ConditionNotIn checks if the key value is not in the list of values.
	ConditionNotIn ConditionOperator = "NotIn"
)

type (
	// Attributes is a flat key/value bag consumed by policy condition
	// evaluation (e.g. "organization_id", "role", "id").
	Attributes = map[string]string

	// AttributesByID groups Attributes by resource id, as returned by
	// batch attribute loaders.
	AttributesByID = map[gid.GID]Attributes
)

// ConditionContext provides attribute values for condition evaluation.
type ConditionContext struct {
	Principal Attributes
	Resource  Attributes
}

// Evaluate checks if the condition is satisfied given the context.
func (c Condition) Evaluate(ctx ConditionContext) bool {
	value, ok := resolveKey(c.Key, ctx)
	if !ok {
		return false
	}

	switch c.Operator {
	case ConditionEquals:
		for _, v := range c.Values {
			resolved, ok := resolveValue(v, ctx)
			if ok && value == resolved {
				return true
			}
		}

		return false

	case ConditionNotEquals:
		for _, v := range c.Values {
			resolved, ok := resolveValue(v, ctx)
			if ok && value == resolved {
				return false
			}
		}

		return true

	case ConditionIn:
		for _, v := range c.Values {
			resolved, ok := resolveValue(v, ctx)
			if !ok {
				continue
			}

			// Support a comma-separated "set" value, e.g.
			// principal.organization_ids = "org_1,org_2"
			if strings.Contains(resolved, ",") {
				for item := range strings.SplitSeq(resolved, ",") {
					if value == strings.TrimSpace(item) {
						return true
					}
				}

				continue
			}

			if value == resolved {
				return true
			}
		}

		return false

	case ConditionNotIn:
		for _, v := range c.Values {
			resolved, ok := resolveValue(v, ctx)
			if !ok {
				continue
			}

			if strings.Contains(resolved, ",") {
				for item := range strings.SplitSeq(resolved, ",") {
					if value == strings.TrimSpace(item) {
						return false
					}
				}

				continue
			}

			if value == resolved {
				return false
			}
		}

		return true

	default:
		return false
	}
}

// resolveKey extracts a value from the context based on a key path.
// Key format: "principal.id", "resource.owner_id", etc.
func resolveKey(key string, ctx ConditionContext) (string, bool) {
	if len(key) > 10 && key[:10] == "principal." {
		attrKey := key[10:]
		val, ok := ctx.Principal[attrKey]

		return val, ok
	}

	if len(key) > 9 && key[:9] == "resource." {
		attrKey := key[9:]
		val, ok := ctx.Resource[attrKey]

		return val, ok
	}

	return "", false
}

// resolveValue returns either a context reference (e.g. "principal.id")
// resolved against ctx, or the value itself when it is a literal.
func resolveValue(value string, ctx ConditionContext) (string, bool) {
	if len(value) > 10 && value[:10] == "principal." {
		return resolveKey(value, ctx)
	}

	if len(value) > 9 && value[:9] == "resource." {
		return resolveKey(value, ctx)
	}

	return value, true
}
