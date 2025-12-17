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

func TestEvaluator_Evaluate_AllowDecision(t *testing.T) {
	evaluator := NewEvaluator()

	policy := NewPolicy("test", "Test Policy",
		Allow("iam:identity:get", "iam:identity:update"),
	)

	tests := []struct {
		name   string
		action string
		want   Decision
	}{
		{
			name:   "allowed action - get",
			action: "iam:identity:get",
			want:   DecisionAllow,
		},
		{
			name:   "allowed action - update",
			action: "iam:identity:update",
			want:   DecisionAllow,
		},
		{
			name:   "not allowed action",
			action: "iam:identity:delete",
			want:   DecisionNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := AuthorizationRequest{
				Action: tt.action,
				ConditionContext: ConditionContext{
					Principal: map[string]string{"id": "user_123"},
					Resource:  map[string]string{"id": "res_456"},
				},
			}

			result := evaluator.Evaluate(req, []*Policy{policy})
			if result.Decision != tt.want {
				t.Errorf("Evaluate() decision = %v, want %v", result.Decision, tt.want)
			}
		})
	}
}

func TestEvaluator_Evaluate_DenyDecision(t *testing.T) {
	evaluator := NewEvaluator()

	policy := NewPolicy("test", "Test Policy",
		Allow("iam:*:*"),
		Deny("iam:organization:delete").WithSID("deny-org-delete"),
	)

	tests := []struct {
		name   string
		action string
		want   Decision
	}{
		{
			name:   "allowed by wildcard",
			action: "iam:identity:get",
			want:   DecisionAllow,
		},
		{
			name:   "denied explicitly",
			action: "iam:organization:delete",
			want:   DecisionDeny,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := AuthorizationRequest{
				Action: tt.action,
				ConditionContext: ConditionContext{
					Principal: map[string]string{"id": "user_123"},
					Resource:  map[string]string{"id": "org_789"},
				},
			}

			result := evaluator.Evaluate(req, []*Policy{policy})
			if result.Decision != tt.want {
				t.Errorf("Evaluate() decision = %v, want %v", result.Decision, tt.want)
			}
		})
	}
}

func TestEvaluator_Evaluate_DenyWinsOverAllow(t *testing.T) {
	evaluator := NewEvaluator()

	// Two policies: one allows, one denies the same action
	allowPolicy := NewPolicy("allow", "Allow Policy",
		Allow("iam:organization:delete"),
	)
	denyPolicy := NewPolicy("deny", "Deny Policy",
		Deny("iam:organization:delete"),
	)

	req := AuthorizationRequest{
		Action: "iam:organization:delete",
		ConditionContext: ConditionContext{
			Principal: map[string]string{"id": "user_123"},
			Resource:  map[string]string{"id": "org_789"},
		},
	}

	// Deny should win regardless of order
	t.Run("deny first", func(t *testing.T) {
		result := evaluator.Evaluate(req, []*Policy{denyPolicy, allowPolicy})
		if result.Decision != DecisionDeny {
			t.Errorf("Expected Deny, got %v", result.Decision)
		}
	})

	t.Run("allow first", func(t *testing.T) {
		result := evaluator.Evaluate(req, []*Policy{allowPolicy, denyPolicy})
		if result.Decision != DecisionDeny {
			t.Errorf("Expected Deny, got %v", result.Decision)
		}
	})
}

func TestEvaluator_Evaluate_WithConditions(t *testing.T) {
	evaluator := NewEvaluator()

	// Policy that only allows users to update their own identity
	selfManagePolicy := NewPolicy("self-manage", "Self Manage",
		Allow("iam:identity:update").
			When(Equals("principal.id", "resource.id")),
	)

	tests := []struct {
		name        string
		principalID string
		resourceID  string
		want        Decision
	}{
		{
			name:        "condition satisfied - same user",
			principalID: "user_123",
			resourceID:  "user_123",
			want:        DecisionAllow,
		},
		{
			name:        "condition not satisfied - different user",
			principalID: "user_123",
			resourceID:  "user_456",
			want:        DecisionNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := AuthorizationRequest{
				Action: "iam:identity:update",
				ConditionContext: ConditionContext{
					Principal: map[string]string{"id": tt.principalID},
					Resource:  map[string]string{"id": tt.resourceID},
				},
			}

			result := evaluator.Evaluate(req, []*Policy{selfManagePolicy})
			if result.Decision != tt.want {
				t.Errorf("Evaluate() decision = %v, want %v", result.Decision, tt.want)
			}
		})
	}
}

func TestEvaluator_Evaluate_MultipleConditions(t *testing.T) {
	evaluator := NewEvaluator()

	// Policy that requires both conditions to be met
	policy := NewPolicy("test", "Test",
		Allow("documents:document:update").
			When(
				Equals("principal.id", "resource.owner_id"),
				Equals("resource.status", "draft"),
			),
	)

	tests := []struct {
		name     string
		ctx      ConditionContext
		want     Decision
	}{
		{
			name: "both conditions satisfied",
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"owner_id": "user_123", "status": "draft"},
			},
			want: DecisionAllow,
		},
		{
			name: "first condition not satisfied",
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"owner_id": "user_456", "status": "draft"},
			},
			want: DecisionNoMatch,
		},
		{
			name: "second condition not satisfied",
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"owner_id": "user_123", "status": "published"},
			},
			want: DecisionNoMatch,
		},
		{
			name: "neither condition satisfied",
			ctx: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"owner_id": "user_456", "status": "published"},
			},
			want: DecisionNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := AuthorizationRequest{
				Action:           "documents:document:update",
				ConditionContext: tt.ctx,
			}

			result := evaluator.Evaluate(req, []*Policy{policy})
			if result.Decision != tt.want {
				t.Errorf("Evaluate() decision = %v, want %v", result.Decision, tt.want)
			}
		})
	}
}

func TestEvaluator_Evaluate_WildcardActions(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name    string
		policy  *Policy
		action  string
		want    Decision
	}{
		{
			name:   "full wildcard allows everything",
			policy: NewPolicy("test", "Test", Allow("*")),
			action: "any:action:here",
			want:   DecisionAllow,
		},
		{
			name:   "service wildcard",
			policy: NewPolicy("test", "Test", Allow("iam:*:*")),
			action: "iam:identity:get",
			want:   DecisionAllow,
		},
		{
			name:   "service wildcard no match",
			policy: NewPolicy("test", "Test", Allow("iam:*:*")),
			action: "documents:document:read",
			want:   DecisionNoMatch,
		},
		{
			name:   "operation wildcard",
			policy: NewPolicy("test", "Test", Allow("iam:identity:*")),
			action: "iam:identity:delete",
			want:   DecisionAllow,
		},
		{
			name:   "read operations only",
			policy: NewPolicy("test", "Test", Allow("*:*:get", "*:*:list")),
			action: "iam:identity:get",
			want:   DecisionAllow,
		},
		{
			name:   "read operations only - write denied",
			policy: NewPolicy("test", "Test", Allow("*:*:get", "*:*:list")),
			action: "iam:identity:update",
			want:   DecisionNoMatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := AuthorizationRequest{
				Action: tt.action,
				ConditionContext: ConditionContext{
					Principal: map[string]string{"id": "user_123"},
					Resource:  map[string]string{"id": "res_456"},
				},
			}

			result := evaluator.Evaluate(req, []*Policy{tt.policy})
			if result.Decision != tt.want {
				t.Errorf("Evaluate() decision = %v, want %v", result.Decision, tt.want)
			}
		})
	}
}

func TestEvaluator_Evaluate_EmptyPolicies(t *testing.T) {
	evaluator := NewEvaluator()

	req := AuthorizationRequest{
		Action: "iam:identity:get",
		ConditionContext: ConditionContext{
			Principal: map[string]string{"id": "user_123"},
			Resource:  map[string]string{"id": "res_456"},
		},
	}

	result := evaluator.Evaluate(req, []*Policy{})
	if result.Decision != DecisionNoMatch {
		t.Errorf("Expected NoMatch for empty policies, got %v", result.Decision)
	}
}

func TestEvaluator_Evaluate_NilPolicies(t *testing.T) {
	evaluator := NewEvaluator()

	req := AuthorizationRequest{
		Action: "iam:identity:get",
		ConditionContext: ConditionContext{
			Principal: map[string]string{"id": "user_123"},
			Resource:  map[string]string{"id": "res_456"},
		},
	}

	result := evaluator.Evaluate(req, nil)
	if result.Decision != DecisionNoMatch {
		t.Errorf("Expected NoMatch for nil policies, got %v", result.Decision)
	}
}

func TestEvaluator_Evaluate_MatchedStatementAndPolicy(t *testing.T) {
	evaluator := NewEvaluator()

	policy := NewPolicy("test-policy", "Test Policy",
		Allow("iam:identity:get").WithSID("allow-get"),
		Deny("iam:identity:delete").WithSID("deny-delete"),
	)

	t.Run("matched allow statement", func(t *testing.T) {
		req := AuthorizationRequest{
			Action: "iam:identity:get",
			ConditionContext: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"id": "res_456"},
			},
		}

		result := evaluator.Evaluate(req, []*Policy{policy})
		if result.MatchedStatement == nil {
			t.Fatal("Expected matched statement")
		}
		if result.MatchedStatement.SID != "allow-get" {
			t.Errorf("Expected SID 'allow-get', got %q", result.MatchedStatement.SID)
		}
		if result.MatchedPolicy == nil {
			t.Fatal("Expected matched policy")
		}
		if result.MatchedPolicy.ID != "test-policy" {
			t.Errorf("Expected policy ID 'test-policy', got %q", result.MatchedPolicy.ID)
		}
	})

	t.Run("matched deny statement", func(t *testing.T) {
		req := AuthorizationRequest{
			Action: "iam:identity:delete",
			ConditionContext: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"id": "res_456"},
			},
		}

		result := evaluator.Evaluate(req, []*Policy{policy})
		if result.MatchedStatement == nil {
			t.Fatal("Expected matched statement")
		}
		if result.MatchedStatement.SID != "deny-delete" {
			t.Errorf("Expected SID 'deny-delete', got %q", result.MatchedStatement.SID)
		}
	})

	t.Run("no match - no statement or policy", func(t *testing.T) {
		req := AuthorizationRequest{
			Action: "iam:identity:update",
			ConditionContext: ConditionContext{
				Principal: map[string]string{"id": "user_123"},
				Resource:  map[string]string{"id": "res_456"},
			},
		}

		result := evaluator.Evaluate(req, []*Policy{policy})
		if result.MatchedStatement != nil {
			t.Error("Expected no matched statement")
		}
		if result.MatchedPolicy != nil {
			t.Error("Expected no matched policy")
		}
	})
}

func TestEvaluationResult_IsAllowed(t *testing.T) {
	tests := []struct {
		decision Decision
		want     bool
	}{
		{DecisionAllow, true},
		{DecisionDeny, false},
		{DecisionNoMatch, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.decision), func(t *testing.T) {
			result := EvaluationResult{Decision: tt.decision}
			if result.IsAllowed() != tt.want {
				t.Errorf("IsAllowed() = %v, want %v", result.IsAllowed(), tt.want)
			}
		})
	}
}

