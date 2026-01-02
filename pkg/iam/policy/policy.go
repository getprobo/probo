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

// Policy represents a collection of statements that define permissions.
// Policies can be attached to roles or directly to principals.
type Policy struct {
	// ID is the unique identifier for the policy.
	ID string

	// Name is a human-readable name for the policy.
	Name string

	// Description explains what the policy is for.
	Description string

	// Statements are the permission rules in this policy.
	Statements []Statement
}

// NewPolicy creates a new policy with the given name and statements.
func NewPolicy(id, name string, statements ...Statement) *Policy {
	return &Policy{
		ID:         id,
		Name:       name,
		Statements: statements,
	}
}

// WithDescription sets the description and returns the policy for chaining.
func (p *Policy) WithDescription(desc string) *Policy {
	p.Description = desc
	return p
}

// AddStatement adds a statement to the policy.
func (p *Policy) AddStatement(stmt Statement) {
	p.Statements = append(p.Statements, stmt)
}

// Allow is a helper to create an allow statement.
func Allow(actions ...string) Statement {
	return Statement{
		Effect:  EffectAllow,
		Actions: actions,
	}
}

// Deny is a helper to create a deny statement.
func Deny(actions ...string) Statement {
	return Statement{
		Effect:  EffectDeny,
		Actions: actions,
	}
}

// WithSID sets the statement ID and returns the statement for chaining.
func (s Statement) WithSID(sid string) Statement {
	s.SID = sid
	return s
}

// WithResources sets the resource patterns and returns the statement for chaining.
func (s Statement) WithResources(resources ...ResourcePattern) Statement {
	s.Resources = resources
	return s
}

// WithConditions sets the conditions and returns the statement for chaining.
func (s Statement) WithConditions(conditions ...Condition) Statement {
	s.Conditions = conditions
	return s
}

// When is an alias for WithConditions for more readable policy definitions.
func (s Statement) When(conditions ...Condition) Statement {
	return s.WithConditions(conditions...)
}

// Equals creates an Equals condition.
func Equals(key string, values ...string) Condition {
	return Condition{
		Operator: ConditionEquals,
		Key:      key,
		Values:   values,
	}
}

// NotEquals creates a NotEquals condition.
func NotEquals(key string, values ...string) Condition {
	return Condition{
		Operator: ConditionNotEquals,
		Key:      key,
		Values:   values,
	}
}

// In creates an In condition.
func In(key string, values ...string) Condition {
	return Condition{
		Operator: ConditionIn,
		Key:      key,
		Values:   values,
	}
}

// NotIn creates a NotIn condition.
func NotIn(key string, values ...string) Condition {
	return Condition{
		Operator: ConditionNotIn,
		Key:      key,
		Values:   values,
	}
}

// ForEntityType creates a resource pattern for a specific entity type.
func ForEntityType(entityType uint16) ResourcePattern {
	return ResourcePattern{
		EntityType: &entityType,
	}
}
