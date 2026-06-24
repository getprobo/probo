// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

func TestPolicy_AddStatement(t *testing.T) {
	p := NewPolicy("test", "Test")

	stmt := Allow("iam:identity:get").WithSID("allow-identity-get")
	p.AddStatement(stmt)

	if len(p.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(p.Statements))
	}

	if p.Statements[0].SID != "allow-identity-get" {
		t.Errorf("expected SID %q, got %q", "allow-identity-get", p.Statements[0].SID)
	}
}

func TestStatement_WithResources(t *testing.T) {
	tenantID := gid.NewTenantID()
	entityType := uint16(1001)

	stmt := Allow("core:framework:get").WithResources(
		ResourcePattern{
			TenantID:   &tenantID,
			EntityType: &entityType,
		},
	)

	if len(stmt.Resources) != 1 {
		t.Fatalf("expected 1 resource pattern, got %d", len(stmt.Resources))
	}

	if stmt.Resources[0].TenantID == nil || *stmt.Resources[0].TenantID != tenantID {
		t.Fatalf("expected tenant id %q in resource pattern", tenantID)
	}

	if stmt.Resources[0].EntityType == nil || *stmt.Resources[0].EntityType != entityType {
		t.Fatalf("expected entity type %d in resource pattern", entityType)
	}
}

func TestEvaluator_Evaluate_ResourcePatternFilters(t *testing.T) {
	evaluator := NewEvaluator()
	tenantID := gid.NewTenantID()
	otherTenantID := gid.NewTenantID()
	entityType := uint16(1001)
	otherEntityType := uint16(1002)

	p := NewPolicy(
		"resource-filter",
		"Resource Filter",
		Allow("core:framework:get").WithResources(
			ResourcePattern{
				TenantID:   &tenantID,
				EntityType: &entityType,
			},
		),
	)

	tests := []struct {
		name     string
		resource gid.GID
		want     Decision
	}{
		{
			name:     "matching tenant and entity type allows",
			resource: gid.New(tenantID, entityType),
			want:     DecisionAllow,
		},
		{
			name:     "different tenant does not match",
			resource: gid.New(otherTenantID, entityType),
			want:     DecisionNoMatch,
		},
		{
			name:     "different entity type does not match",
			resource: gid.New(tenantID, otherEntityType),
			want:     DecisionNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.Evaluate(
				AuthorizationRequest{
					Action:   "core:framework:get",
					Resource: tt.resource,
				},
				[]*Policy{p},
			)
			if result.Decision != tt.want {
				t.Errorf("Evaluate() decision = %v, want %v", result.Decision, tt.want)
			}
		})
	}
}
