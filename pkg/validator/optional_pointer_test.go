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
				if v.Error() == nil {
					t.Error("expected error but got none")
				}
			} else {
				if v.Error() != nil {
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
			v.Check(tt.value, "asset_type", OneOfSlice([]string{"VALID", "ANOTHER"}))

			if tt.expectError {
				if v.Error() == nil {
					t.Error("expected error but got none")
				}
			} else {
				if v.Error() != nil {
					t.Errorf("unexpected error: %v", v.Error())
				}
			}
		})
	}
}
