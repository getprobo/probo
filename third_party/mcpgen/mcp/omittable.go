package mcp

import (
	"encoding/json"
	"fmt"
)

// Omittable represents a value that can be in one of three states:
// 1. Not set (field was not provided in JSON)
// 2. Explicitly set to null
// 3. Set to a value
//
// This is useful for distinguishing between "don't update this field" (not set)
// and "set this field to null" (explicitly null) in update operations.
//
// Example usage:
//
//	type UpdateUserInput struct {
//	    Name  Omittable[string] `json:"name,omitempty"`
//	    Email Omittable[string] `json:"email,omitempty"`
//	}
//
//	func (r *Resolver) UpdateUser(input UpdateUserInput) {
//	    if input.Name.IsSet() {
//	        if input.Name.IsNull() {
//	            // Set name to null
//	        } else {
//	            // Update name to input.Name.Value()
//	        }
//	    }
//	    // If !IsSet(), don't touch the name field
//	}
type Omittable[T any] struct {
	value *T
	isSet bool
}

func NewOmittable[T any](value T) Omittable[T] {
	return Omittable[T]{
		value: &value,
		isSet: true,
	}
}

func NewOmittableNull[T any]() Omittable[T] {
	return Omittable[T]{
		value: nil,
		isSet: true,
	}
}

// IsSet returns true if the field was provided in the input (either null or a value).
func (o Omittable[T]) IsSet() bool {
	return o.isSet
}

// IsNull returns true if the field was explicitly set to null.
// Returns false if the field was not set or has a value.
func (o Omittable[T]) IsNull() bool {
	return o.isSet && o.value == nil
}

// Value returns the value and a boolean indicating if it has a non-null value.
// If the field is not set or is null, returns the zero value and false.
func (o Omittable[T]) Value() (T, bool) {
	if o.value != nil {
		return *o.value, true
	}
	var zero T
	return zero, false
}

func (o Omittable[T]) ValueOrZero() T {
	if o.value != nil {
		return *o.value
	}
	var zero T
	return zero
}

func (o Omittable[T]) Ptr() *T {
	return o.value
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *Omittable[T]) UnmarshalJSON(data []byte) error {
	o.isSet = true

	// Handle explicit null
	if string(data) == "null" {
		o.value = nil
		return nil
	}

	// Unmarshal the actual value
	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("failed to unmarshal omittable value: %w", err)
	}

	o.value = &value
	return nil
}

// MarshalJSON implements json.Marshaler.
func (o Omittable[T]) MarshalJSON() ([]byte, error) {
	// Note: When marshaling structs with Omittable fields, use *Omittable[T]
	// if you need omitempty to work correctly. With value types, omitempty
	// doesn't work well with custom MarshalJSON.
	// For MCP use cases (unmarshaling input), this is not typically an issue.

	if !o.isSet || o.value == nil {
		return []byte("null"), nil
	}

	return json.Marshal(*o.value)
}
