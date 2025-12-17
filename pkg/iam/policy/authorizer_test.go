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
	"testing"
)

func TestAuthorizer_Authorize(t *testing.T) {
	// Create a simple registry for testing
	registry := NewActionRegistry()
	registry.MustRegister(ActionDefinition{
		Action:      "test:resource:get",
		Service:     "test",
		Resource:    "resource",
		Operation:   "get",
		Description: "Get resource",
	})
	registry.MustRegister(ActionDefinition{
		Action:      "test:resource:update",
		Service:     "test",
		Resource:    "resource",
		Operation:   "update",
		Description: "Update resource",
	})
	registry.MustRegister(ActionDefinition{
		Action:      "test:resource:delete",
		Service:     "test",
		Resource:    "resource",
		Operation:   "delete",
		Description: "Delete resource",
	})
	registry.MustRegister(ActionDefinition{
		Action:      "test:other:get",
		Service:     "test",
		Resource:    "other",
		Operation:   "get",
		Description: "Get other",
	})

	authorizer := NewAuthorizer(registry)

	policies := []*Policy{
		NewPolicy("test", "Test",
			Allow("test:resource:get", "test:resource:update"),
			Deny("test:resource:delete"),
		),
	}

	tests := []struct {
		name    string
		action  string
		wantErr bool
		errType error
	}{
		{
			name:    "allowed action",
			action:  "test:resource:get",
			wantErr: false,
		},
		{
			name:    "denied action",
			action:  "test:resource:delete",
			wantErr: true,
			errType: ErrAccessDenied,
		},
		{
			name:    "no matching policy",
			action:  "test:other:get",
			wantErr: true,
			errType: ErrAccessDenied,
		},
		{
			name:    "unknown action",
			action:  "unknown:action:here",
			wantErr: true,
			errType: ErrAccessDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authorizer.Authorize(AuthorizeParams{
				Action:   tt.action,
				Policies: policies,
				ResourceAttributes: map[string]string{
					"id": "res_123",
				},
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != nil {
				if !errors.Is(err, tt.errType) {
					t.Errorf("Authorize() error type = %T, want %T", err, tt.errType)
				}
			}
		})
	}
}

func TestAuthorizer_Authorize_WithConditions(t *testing.T) {
	authorizer := NewAuthorizer(nil)

	// Self-manage policy for testing
	selfManagePolicy := NewPolicy("self-manage", "Self Manage",
		Allow("test:identity:get", "test:identity:update").
			When(Equals("principal.id", "resource.id")),
	)

	tests := []struct {
		name               string
		action             string
		resourceAttributes map[string]string
		wantErr            bool
	}{
		{
			name:   "condition not satisfied - GID won't match string",
			action: "test:identity:get",
			resourceAttributes: map[string]string{
				"id": "user_123",
			},
			wantErr: true, // Will fail because principal GID won't match string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authorizer.Authorize(AuthorizeParams{
				Action:             tt.action,
				Policies:           []*Policy{selfManagePolicy},
				ResourceAttributes: tt.resourceAttributes,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthorizer_Authorize_WithoutRegistry(t *testing.T) {
	// Authorizer without registry should not validate actions
	authorizer := NewAuthorizer(nil)

	policies := []*Policy{
		NewPolicy("test", "Test", Allow("custom:action:here")),
	}

	err := authorizer.Authorize(AuthorizeParams{
		Action:   "custom:action:here",
		Policies: policies,
	})

	if err != nil {
		t.Errorf("Expected no error for custom action without registry, got %v", err)
	}
}

func TestAuthorizer_IsAllowed(t *testing.T) {
	authorizer := NewAuthorizer(nil)

	allowPolicy := NewPolicy("test", "Test", Allow("test:resource:get"))
	denyPolicy := NewPolicy("test", "Test", Deny("test:resource:delete"))

	tests := []struct {
		name     string
		action   string
		policies []*Policy
		want     bool
	}{
		{
			name:     "allowed",
			action:   "test:resource:get",
			policies: []*Policy{allowPolicy},
			want:     true,
		},
		{
			name:     "denied",
			action:   "test:resource:delete",
			policies: []*Policy{denyPolicy},
			want:     false,
		},
		{
			name:     "no match",
			action:   "test:resource:update",
			policies: []*Policy{allowPolicy},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := authorizer.IsAllowed(AuthorizeParams{
				Action:   tt.action,
				Policies: tt.policies,
			})

			if got != tt.want {
				t.Errorf("IsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAccessDeniedError(t *testing.T) {
	t.Run("error message without statement", func(t *testing.T) {
		err := &AccessDeniedError{
			Action: "test:resource:delete",
			Reason: "no policy allows this action",
		}

		msg := err.Error()
		if msg == "" {
			t.Error("Expected non-empty error message")
		}
	})

	t.Run("error message with statement SID", func(t *testing.T) {
		err := &AccessDeniedError{
			Action: "test:resource:delete",
			Reason: "explicitly denied",
			Statement: &Statement{
				SID: "deny-delete",
			},
		}

		msg := err.Error()
		if msg == "" {
			t.Error("Expected non-empty error message")
		}
		// Should contain the SID
		if !contains(msg, "deny-delete") {
			t.Errorf("Expected error message to contain SID, got %q", msg)
		}
	})

	t.Run("unwrap returns ErrAccessDenied", func(t *testing.T) {
		err := &AccessDeniedError{
			Action: "test:resource:delete",
			Reason: "no policy",
		}

		if !errors.Is(err, ErrAccessDenied) {
			t.Error("Expected error to unwrap to ErrAccessDenied")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
