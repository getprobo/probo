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
	"fmt"
	"strings"
)

// Action represents a permission action in the format "service:resource:operation"
// Examples: "iam:identity:get", "documents:document:write", "risks:risk:delete"
type Action string

// ActionDefinition provides metadata about an action for documentation and validation.
type ActionDefinition struct {
	Action      Action
	Service     string // e.g., "iam", "documents", "risks"
	Resource    string // e.g., "identity", "document", "risk"
	Operation   string // e.g., "get", "list", "create", "update", "delete"
	Description string
}

// ActionRegistry holds all registered actions and provides lookup/validation.
type ActionRegistry struct {
	actions map[Action]ActionDefinition
}

// NewActionRegistry creates a new empty action registry.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[Action]ActionDefinition),
	}
}

// Register adds an action definition to the registry.
// Returns an error if the action is already registered.
func (r *ActionRegistry) Register(def ActionDefinition) error {
	if _, exists := r.actions[def.Action]; exists {
		return fmt.Errorf("action %q already registered", def.Action)
	}

	// Validate action format
	if err := validateActionFormat(def.Action); err != nil {
		return fmt.Errorf("invalid action format: %w", err)
	}

	r.actions[def.Action] = def
	return nil
}

// MustRegister is like Register but panics on error.
// Useful for setting up registries in application startup.
func (r *ActionRegistry) MustRegister(def ActionDefinition) {
	if err := r.Register(def); err != nil {
		panic(err)
	}
}

// Get returns the definition for an action, or false if not found.
func (r *ActionRegistry) Get(action Action) (ActionDefinition, bool) {
	def, ok := r.actions[action]
	return def, ok
}

// Exists checks if an action is registered.
func (r *ActionRegistry) Exists(action Action) bool {
	_, ok := r.actions[action]
	return ok
}

// All returns all registered action definitions.
func (r *ActionRegistry) All() []ActionDefinition {
	result := make([]ActionDefinition, 0, len(r.actions))
	for _, def := range r.actions {
		result = append(result, def)
	}
	return result
}

// ByService returns all actions for a given service.
func (r *ActionRegistry) ByService(service string) []ActionDefinition {
	var result []ActionDefinition
	for _, def := range r.actions {
		if def.Service == service {
			result = append(result, def)
		}
	}
	return result
}

// validateActionFormat ensures action follows "service:resource:operation" format.
func validateActionFormat(action Action) error {
	parts := strings.Split(string(action), ":")
	if len(parts) != 3 {
		return fmt.Errorf("action must have format 'service:resource:operation', got %q", action)
	}

	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("action part %d is empty in %q", i, action)
		}
	}

	return nil
}

// ParseAction extracts service, resource, and operation from an action string.
func ParseAction(action Action) (service, resource, operation string, err error) {
	parts := strings.Split(string(action), ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid action format: %q", action)
	}
	return parts[0], parts[1], parts[2], nil
}
