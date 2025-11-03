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

package validator

import (
	"testing"

	"go.probo.inc/probo/pkg/gid"
)

// CustomStringType simulates coredata.AssetType
type CustomStringType string

func (c CustomStringType) String() string {
	return string(c)
}

func TestOptional_WithGIDPointer(t *testing.T) {
	tenantID := gid.NewTenantID()

	tests := []struct {
		name        string
		value       *gid.GID
		expectError bool
	}{
		{
			name:        "nil pointer - should skip validation",
			value:       nil,
			expectError: false,
		},
		{
			name: "valid GID pointer",
			value: func() *gid.GID {
				g := gid.New(tenantID, 100)
				return &g
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Check(tt.value, "owner_id", GID(100))

			if tt.expectError {
				if !v.HasErrors() {
					t.Error("expected error but got none")
				}
			} else {
				if v.HasErrors() {
					t.Errorf("unexpected error: %v", v.Error())
				}
			}
		})
	}
}

func TestOptional_WithCustomTypePointer(t *testing.T) {
	tests := []struct {
		name        string
		value       *CustomStringType
		expectError bool
	}{
		{
			name:        "nil pointer - should skip validation",
			value:       nil,
			expectError: false,
		},
		{
			name: "valid custom type pointer",
			value: func() *CustomStringType {
				v := CustomStringType("VALID")
				return &v
			}(),
			expectError: false,
		},
		{
			name: "invalid custom type pointer",
			value: func() *CustomStringType {
				v := CustomStringType("INVALID")
				return &v
			}(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Check(tt.value, "asset_type", OneOf("VALID", "ANOTHER"))

			if tt.expectError {
				if !v.HasErrors() {
					t.Error("expected error but got none")
				}
			} else {
				if v.HasErrors() {
					t.Errorf("unexpected error: %v", v.Error())
				}
			}
		})
	}
}
