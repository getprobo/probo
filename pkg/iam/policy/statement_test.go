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
	"testing"
)

func TestCondition_Evaluate_Equals(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		ctx       ConditionContext
		want      bool
	}{
		{
			name: "equals - match literal value",
			condition: Condition{
				Operator: ConditionEquals,
				Key:      "principal.id",
				Values:   []string{"user_123"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
			},
			want: true,
		},
		{
			name: "equals - no match literal value",
			condition: Condition{
				Operator: ConditionEquals,
				Key:      "principal.id",
				Values:   []string{"user_456"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
			},
			want: false,
		},
		{
			name: "equals - match any of multiple values",
			condition: Condition{
				Operator: ConditionEquals,
				Key:      "principal.id",
				Values:   []string{"user_123", "user_456", "user_789"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_456"},
			},
			want: true,
		},
		{
			name: "equals - match resource reference",
			condition: Condition{
				Operator: ConditionEquals,
				Key:      "principal.id",
				Values:   []string{"resource.id"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"id": "user_123"},
			},
			want: true,
		},
		{
			name: "equals - no match resource reference",
			condition: Condition{
				Operator: ConditionEquals,
				Key:      "principal.id",
				Values:   []string{"resource.id"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"id": "user_456"},
			},
			want: false,
		},
		{
			name: "equals - match resource.identity_id reference",
			condition: Condition{
				Operator: ConditionEquals,
				Key:      "principal.id",
				Values:   []string{"resource.identity_id"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"identity_id": "user_123"},
			},
			want: true,
		},
		{
			name: "equals - match principal.email to resource.email reference",
			condition: Condition{
				Operator: ConditionEquals,
				Key:      "principal.email",
				Values:   []string{"resource.email"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"email": "user@example.com"},
				Resource:  map[string]string{"email": "user@example.com"},
			},
			want: true,
		},
		{
			name: "equals - key not found",
			condition: Condition{
				Operator: ConditionEquals,
				Key:      "principal.unknown",
				Values:   []string{"value"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.condition.Evaluate(tt.ctx)
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCondition_Evaluate_NotEquals(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		ctx       ConditionContext
		want      bool
	}{
		{
			name: "not equals - different values",
			condition: Condition{
				Operator: ConditionNotEquals,
				Key:      "principal.id",
				Values:   []string{"user_456"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
			},
			want: true,
		},
		{
			name: "not equals - same value",
			condition: Condition{
				Operator: ConditionNotEquals,
				Key:      "principal.id",
				Values:   []string{"user_123"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
			},
			want: false,
		},
		{
			name: "not equals - one of multiple values matches",
			condition: Condition{
				Operator: ConditionNotEquals,
				Key:      "principal.id",
				Values:   []string{"user_123", "user_456"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.condition.Evaluate(tt.ctx)
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCondition_Evaluate_In(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		ctx       ConditionContext
		want      bool
	}{
		{
			name: "in - value in list",
			condition: Condition{
				Operator: ConditionIn,
				Key:      "principal.role",
				Values:   []string{"admin", "owner", "viewer"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"role": "admin"},
			},
			want: true,
		},
		{
			name: "in - value not in list",
			condition: Condition{
				Operator: ConditionIn,
				Key:      "principal.role",
				Values:   []string{"admin", "owner"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"role": "viewer"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.condition.Evaluate(tt.ctx)
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCondition_Evaluate_NotIn(t *testing.T) {
	tests := []struct {
		name      string
		condition Condition
		ctx       ConditionContext
		want      bool
	}{
		{
			name: "not in - value not in list",
			condition: Condition{
				Operator: ConditionNotIn,
				Key:      "principal.role",
				Values:   []string{"admin", "owner"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"role": "viewer"},
			},
			want: true,
		},
		{
			name: "not in - value in list",
			condition: Condition{
				Operator: ConditionNotIn,
				Key:      "principal.role",
				Values:   []string{"admin", "owner", "viewer"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"role": "admin"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.condition.Evaluate(tt.ctx)
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConditionHelpers(t *testing.T) {
	t.Run("Equals helper", func(t *testing.T) {
		c := Equals("principal.id", "user_123", "user_456")
		if c.Operator != ConditionEquals {
			t.Errorf("Expected ConditionEquals, got %v", c.Operator)
		}
		if c.Key != "principal.id" {
			t.Errorf("Expected principal.id, got %v", c.Key)
		}
		if len(c.Values) != 2 {
			t.Errorf("Expected 2 values, got %d", len(c.Values))
		}
	})

	t.Run("NotEquals helper", func(t *testing.T) {
		c := NotEquals("principal.id", "user_123")
		if c.Operator != ConditionNotEquals {
			t.Errorf("Expected ConditionNotEquals, got %v", c.Operator)
		}
	})

	t.Run("In helper", func(t *testing.T) {
		c := In("principal.role", "admin", "owner")
		if c.Operator != ConditionIn {
			t.Errorf("Expected ConditionIn, got %v", c.Operator)
		}
	})

	t.Run("NotIn helper", func(t *testing.T) {
		c := NotIn("principal.role", "guest")
		if c.Operator != ConditionNotIn {
			t.Errorf("Expected ConditionNotIn, got %v", c.Operator)
		}
	})
}
