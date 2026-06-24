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
