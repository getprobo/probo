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
	"testing"

	"go.probo.inc/probo/pkg/gid"
)

func TestResourcePattern_MatchesResource(t *testing.T) {
	tenantID := gid.NewTenantID()
	otherTenantID := gid.NewTenantID()
	frameworkEntityType := uint16(1001)
	organizationEntityType := uint16(1002)
	resource := gid.New(tenantID, frameworkEntityType)
	otherTenantResource := gid.New(otherTenantID, frameworkEntityType)
	otherEntityResource := gid.New(tenantID, organizationEntityType)

	tests := []struct {
		name     string
		pattern  ResourcePattern
		resource gid.GID
		want     bool
	}{
		{
			name:     "empty pattern matches all resources",
			pattern:  ResourcePattern{},
			resource: resource,
			want:     true,
		},
		{
			name: "tenant only pattern matches same tenant",
			pattern: ResourcePattern{
				TenantID: &tenantID,
			},
			resource: resource,
			want:     true,
		},
		{
			name: "tenant only pattern does not match different tenant",
			pattern: ResourcePattern{
				TenantID: &tenantID,
			},
			resource: otherTenantResource,
			want:     false,
		},
		{
			name: "entity type only pattern matches same type",
			pattern: ResourcePattern{
				EntityType: &frameworkEntityType,
			},
			resource: resource,
			want:     true,
		},
		{
			name: "entity type only pattern does not match different type",
			pattern: ResourcePattern{
				EntityType: &frameworkEntityType,
			},
			resource: otherEntityResource,
			want:     false,
		},
		{
			name: "tenant and entity type pattern matches when both match",
			pattern: ResourcePattern{
				TenantID:   &tenantID,
				EntityType: &frameworkEntityType,
			},
			resource: resource,
			want:     true,
		},
		{
			name: "tenant and entity type pattern fails when one mismatches",
			pattern: ResourcePattern{
				TenantID:   &tenantID,
				EntityType: &frameworkEntityType,
			},
			resource: otherEntityResource,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pattern.MatchesResource(tt.resource)
			if got != tt.want {
				t.Errorf("MatchesResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		{
			name: "in - matches value inside comma-separated set",
			condition: Condition{
				Operator: ConditionIn,
				Key:      "principal.organization_id",
				Values:   []string{"resource.organization_ids"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"organization_id": "org_2"},
				Resource:  map[string]string{"organization_ids": "org_1, org_2, org_3"},
			},
			want: true,
		},
		{
			name: "in - does not match comma-separated set",
			condition: Condition{
				Operator: ConditionIn,
				Key:      "principal.organization_id",
				Values:   []string{"resource.organization_ids"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"organization_id": "org_9"},
				Resource:  map[string]string{"organization_ids": "org_1,org_2"},
			},
			want: false,
		},
		{
			name: "in - skips unresolved references",
			condition: Condition{
				Operator: ConditionIn,
				Key:      "principal.organization_id",
				Values:   []string{"resource.missing_ids"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"organization_id": "org_2"},
				Resource:  map[string]string{"organization_ids": "org_1,org_2"},
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
		{
			name: "not in - comma-separated set contains value",
			condition: Condition{
				Operator: ConditionNotIn,
				Key:      "principal.organization_id",
				Values:   []string{"resource.organization_ids"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"organization_id": "org_2"},
				Resource:  map[string]string{"organization_ids": "org_1, org_2, org_3"},
			},
			want: false,
		},
		{
			name: "not in - comma-separated set does not contain value",
			condition: Condition{
				Operator: ConditionNotIn,
				Key:      "principal.organization_id",
				Values:   []string{"resource.organization_ids"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"organization_id": "org_9"},
				Resource:  map[string]string{"organization_ids": "org_1,org_2"},
			},
			want: true,
		},
		{
			name: "unknown operator returns false",
			condition: Condition{
				Operator: ConditionOperator("Unknown"),
				Key:      "principal.role",
				Values:   []string{"admin"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"role": "admin"},
			},
			want: false,
		},
		{
			name: "not in - skips unresolved references",
			condition: Condition{
				Operator: ConditionNotIn,
				Key:      "principal.organization_id",
				Values:   []string{"resource.missing_ids"},
			},
			ctx: ConditionContext{
				Principal: map[string]string{"organization_id": "org_9"},
				Resource:  map[string]string{"organization_ids": "org_1,org_2"},
			},
			want: true,
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

	t.Run("NotIn condition", func(t *testing.T) {
		c := Condition{
			Operator: ConditionNotIn,
			Key:      "principal.role",
			Values:   []string{"guest"},
		}
		if c.Operator != ConditionNotIn {
			t.Errorf("Expected ConditionNotIn, got %v", c.Operator)
		}
	})
}

func TestResolveKey(t *testing.T) {
	ctx := ConditionContext{
		Principal: map[string]string{
			"id": "principal-id",
		},
		Resource: map[string]string{
			"id": "resource-id",
		},
	}

	tests := []struct {
		name   string
		key    string
		want   string
		wantOK bool
	}{
		{
			name:   "resolves principal key",
			key:    "principal.id",
			want:   "principal-id",
			wantOK: true,
		},
		{
			name:   "resolves resource key",
			key:    "resource.id",
			want:   "resource-id",
			wantOK: true,
		},
		{
			name:   "returns false for unknown namespace",
			key:    "unknown.id",
			want:   "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveKey(tt.key, ctx)
			if ok != tt.wantOK {
				t.Fatalf("resolveKey() ok = %v, want %v", ok, tt.wantOK)
			}

			if got != tt.want {
				t.Fatalf("resolveKey() value = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveValue(t *testing.T) {
	ctx := ConditionContext{
		Principal: map[string]string{
			"id": "principal-id",
		},
		Resource: map[string]string{
			"id": "resource-id",
		},
	}

	tests := []struct {
		name   string
		value  string
		want   string
		wantOK bool
	}{
		{
			name:   "resolves principal reference",
			value:  "principal.id",
			want:   "principal-id",
			wantOK: true,
		},
		{
			name:   "resolves resource reference",
			value:  "resource.id",
			want:   "resource-id",
			wantOK: true,
		},
		{
			name:   "keeps literal values",
			value:  "literal",
			want:   "literal",
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := resolveValue(tt.value, ctx)
			if ok != tt.wantOK {
				t.Fatalf("resolveValue() ok = %v, want %v", ok, tt.wantOK)
			}

			if got != tt.want {
				t.Fatalf("resolveValue() value = %q, want %q", got, tt.want)
			}
		})
	}
}
