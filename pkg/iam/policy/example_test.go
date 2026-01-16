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

package policy_test

import (
	"fmt"

	"go.probo.inc/probo/pkg/iam/policy"
)

func Example_definingActions() {
	// Create an action registry
	registry := policy.NewActionRegistry()

	// Register actions for the IAM service
	registry.MustRegister(policy.ActionDefinition{
		Action:      "iam:identity:get",
		Service:     "iam",
		Resource:    "identity",
		Operation:   "get",
		Description: "Get identity details",
	})

	registry.MustRegister(policy.ActionDefinition{
		Action:      "iam:identity:update",
		Service:     "iam",
		Resource:    "identity",
		Operation:   "update",
		Description: "Update identity",
	})

	// Register actions for the documents service
	registry.MustRegister(policy.ActionDefinition{
		Action:      "documents:document:read",
		Service:     "documents",
		Resource:    "document",
		Operation:   "read",
		Description: "Read a document",
	})

	registry.MustRegister(policy.ActionDefinition{
		Action:      "documents:document:write",
		Service:     "documents",
		Resource:    "document",
		Operation:   "write",
		Description: "Create or update a document",
	})

	registry.MustRegister(policy.ActionDefinition{
		Action:      "documents:document:delete",
		Service:     "documents",
		Resource:    "document",
		Operation:   "delete",
		Description: "Delete a document",
	})

	// List all actions for a service
	docActions := registry.ByService("documents")
	fmt.Printf("Documents service has %d actions\n", len(docActions))

	// Output:
	// Documents service has 3 actions
}

func Example_definingPolicies() {
	// Define a viewer policy - can read everything
	viewerPolicy := policy.NewPolicy("viewer", "Viewer Policy",
		policy.Allow("*:*:read", "*:*:list"),
	).WithDescription("Read-only access to all resources")

	// Define an admin policy - can do everything except delete organization
	adminPolicy := policy.NewPolicy("admin", "Admin Policy",
		policy.Allow("*"),
		policy.Deny("iam:organization:delete"),
	).WithDescription("Full access except organization deletion")

	// Define a self-manage policy - users can manage their own identity
	selfManagePolicy := policy.NewPolicy("self-manage", "Self Management Policy",
		policy.Allow("iam:identity:get", "iam:identity:update").
			When(policy.Equals("principal.id", "resource.id")),
	).WithDescription("Users can view and update their own identity")

	// Define a document owner policy - owners can do anything to their documents
	documentOwnerPolicy := policy.NewPolicy("doc-owner", "Document Owner Policy",
		policy.Allow("documents:document:*").
			When(policy.Equals("principal.id", "resource.owner_id")),
	).WithDescription("Document owners have full control over their documents")

	fmt.Println(viewerPolicy.Name)
	fmt.Println(adminPolicy.Name)
	fmt.Println(selfManagePolicy.Name)
	fmt.Println(documentOwnerPolicy.Name)

	// Output:
	// Viewer Policy
	// Admin Policy
	// Self Management Policy
	// Document Owner Policy
}

func Example_evaluatingPolicies() {
	evaluator := policy.NewEvaluator()

	// Define policies
	viewerPolicy := policy.NewPolicy("viewer", "Viewer",
		policy.Allow("*:*:read", "*:*:list"),
	)

	adminPolicy := policy.NewPolicy("admin", "Admin",
		policy.Allow("*"),
		policy.Deny("iam:organization:delete").WithSID("prevent-org-deletion"),
	)

	// Test 1: Viewer can read documents
	req1 := policy.AuthorizationRequest{
		Action: "documents:document:read",
		ConditionContext: policy.ConditionContext{
			Principal: map[string]string{"id": "user_123"},
			Resource:  map[string]string{"id": "doc_456"},
		},
	}
	result1 := evaluator.Evaluate(req1, []*policy.Policy{viewerPolicy})
	fmt.Printf("Viewer read document: %s\n", result1.Decision)

	// Test 2: Viewer cannot delete documents
	req2 := policy.AuthorizationRequest{
		Action: "documents:document:delete",
		ConditionContext: policy.ConditionContext{
			Principal: map[string]string{"id": "user_123"},
			Resource:  map[string]string{"id": "doc_456"},
		},
	}
	result2 := evaluator.Evaluate(req2, []*policy.Policy{viewerPolicy})
	fmt.Printf("Viewer delete document: %s\n", result2.Decision)

	// Test 3: Admin can delete documents
	result3 := evaluator.Evaluate(req2, []*policy.Policy{adminPolicy})
	fmt.Printf("Admin delete document: %s\n", result3.Decision)

	// Test 4: Admin cannot delete organization (explicit deny)
	req4 := policy.AuthorizationRequest{
		Action: "iam:organization:delete",
		ConditionContext: policy.ConditionContext{
			Principal: map[string]string{"id": "user_123"},
			Resource:  map[string]string{"id": "org_789"},
		},
	}
	result4 := evaluator.Evaluate(req4, []*policy.Policy{adminPolicy})
	fmt.Printf("Admin delete organization: %s\n", result4.Decision)
	fmt.Printf("Matched statement SID: %s\n", result4.MatchedStatement.SID)

	// Output:
	// Viewer read document: allow
	// Viewer delete document: no_match
	// Admin delete document: allow
	// Admin delete organization: deny
	// Matched statement SID: prevent-org-deletion
}

func Example_conditionBasedAccess() {
	evaluator := policy.NewEvaluator()

	// Policy: users can only update their own profile
	selfManagePolicy := policy.NewPolicy("self-manage", "Self Management",
		policy.Allow("iam:identity:update").
			When(policy.Equals("principal.id", "resource.id")),
	)

	// Test 1: User updating their own profile
	req1 := policy.AuthorizationRequest{
		Action: "iam:identity:update",
		ConditionContext: policy.ConditionContext{
			Principal: map[string]string{"id": "user_123"},
			Resource:  map[string]string{"id": "user_123"}, // Same as principal
		},
	}
	result1 := evaluator.Evaluate(req1, []*policy.Policy{selfManagePolicy})
	fmt.Printf("User update own profile: %s\n", result1.Decision)

	// Test 2: User trying to update someone else's profile
	req2 := policy.AuthorizationRequest{
		Action: "iam:identity:update",
		ConditionContext: policy.ConditionContext{
			Principal: map[string]string{"id": "user_123"},
			Resource:  map[string]string{"id": "user_456"}, // Different from principal
		},
	}
	result2 := evaluator.Evaluate(req2, []*policy.Policy{selfManagePolicy})
	fmt.Printf("User update other profile: %s\n", result2.Decision)

	// Output:
	// User update own profile: allow
	// User update other profile: no_match
}

