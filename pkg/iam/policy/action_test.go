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

func TestActionRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		def     ActionDefinition
		wantErr bool
	}{
		{
			name: "valid action",
			def: ActionDefinition{
				Action:      "iam:identity:get",
				Service:     "iam",
				Resource:    "identity",
				Operation:   "get",
				Description: "Get identity",
			},
			wantErr: false,
		},
		{
			name: "invalid format - missing parts",
			def: ActionDefinition{
				Action:  "iam:identity",
				Service: "iam",
			},
			wantErr: true,
		},
		{
			name: "invalid format - empty part",
			def: ActionDefinition{
				Action:  "iam::get",
				Service: "iam",
			},
			wantErr: true,
		},
		{
			name: "invalid format - too many parts",
			def: ActionDefinition{
				Action:  "iam:identity:get:extra",
				Service: "iam",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewActionRegistry()
			err := r.Register(tt.def)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestActionRegistry_DuplicateRegistration(t *testing.T) {
	r := NewActionRegistry()

	def := ActionDefinition{
		Action:      "iam:identity:get",
		Service:     "iam",
		Resource:    "identity",
		Operation:   "get",
		Description: "Get identity",
	}

	// First registration should succeed
	if err := r.Register(def); err != nil {
		t.Fatalf("First registration failed: %v", err)
	}

	// Second registration should fail
	if err := r.Register(def); err == nil {
		t.Error("Expected error for duplicate registration, got nil")
	}
}

func TestActionRegistry_Get(t *testing.T) {
	r := NewActionRegistry()

	def := ActionDefinition{
		Action:      "iam:identity:get",
		Service:     "iam",
		Resource:    "identity",
		Operation:   "get",
		Description: "Get identity",
	}
	r.MustRegister(def)

	// Get existing action
	got, ok := r.Get("iam:identity:get")
	if !ok {
		t.Error("Expected to find action")
	}
	if got.Action != def.Action {
		t.Errorf("Got action %v, want %v", got.Action, def.Action)
	}

	// Get non-existing action
	_, ok = r.Get("iam:identity:delete")
	if ok {
		t.Error("Expected not to find action")
	}
}

func TestActionRegistry_Exists(t *testing.T) {
	r := NewActionRegistry()

	r.MustRegister(ActionDefinition{
		Action:    "iam:identity:get",
		Service:   "iam",
		Resource:  "identity",
		Operation: "get",
	})

	if !r.Exists("iam:identity:get") {
		t.Error("Expected action to exist")
	}

	if r.Exists("iam:identity:delete") {
		t.Error("Expected action not to exist")
	}
}

func TestActionRegistry_ByService(t *testing.T) {
	r := NewActionRegistry()

	r.MustRegister(ActionDefinition{Action: "iam:identity:get", Service: "iam", Resource: "identity", Operation: "get"})
	r.MustRegister(ActionDefinition{Action: "iam:identity:update", Service: "iam", Resource: "identity", Operation: "update"})
	r.MustRegister(ActionDefinition{Action: "documents:document:read", Service: "documents", Resource: "document", Operation: "read"})

	iamActions := r.ByService("iam")
	if len(iamActions) != 2 {
		t.Errorf("Expected 2 IAM actions, got %d", len(iamActions))
	}

	docActions := r.ByService("documents")
	if len(docActions) != 1 {
		t.Errorf("Expected 1 documents action, got %d", len(docActions))
	}

	unknownActions := r.ByService("unknown")
	if len(unknownActions) != 0 {
		t.Errorf("Expected 0 unknown actions, got %d", len(unknownActions))
	}
}

func TestParseAction(t *testing.T) {
	tests := []struct {
		action    Action
		wantSvc   string
		wantRes   string
		wantOp    string
		wantErr   bool
	}{
		{
			action:  "iam:identity:get",
			wantSvc: "iam",
			wantRes: "identity",
			wantOp:  "get",
			wantErr: false,
		},
		{
			action:  "documents:document:read",
			wantSvc: "documents",
			wantRes: "document",
			wantOp:  "read",
			wantErr: false,
		},
		{
			action:  "invalid",
			wantErr: true,
		},
		{
			action:  "invalid:action",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			svc, res, op, err := ParseAction(tt.action)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if svc != tt.wantSvc {
					t.Errorf("service = %v, want %v", svc, tt.wantSvc)
				}
				if res != tt.wantRes {
					t.Errorf("resource = %v, want %v", res, tt.wantRes)
				}
				if op != tt.wantOp {
					t.Errorf("operation = %v, want %v", op, tt.wantOp)
				}
			}
		})
	}
}

