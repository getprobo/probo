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
)

// AssetType simulates coredata.AssetType
type AssetType string

const (
	AssetTypePhysical AssetType = "PHYSICAL"
	AssetTypeVirtual  AssetType = "VIRTUAL"
)

func (at AssetType) String() string {
	return string(at)
}

func TestOneOf_CustomStringType(t *testing.T) {
	tests := []struct {
		name        string
		value       any
		allowed     []string
		expectError bool
	}{
		{
			name:        "valid custom type - physical",
			value:       AssetTypePhysical,
			allowed:     []string{"PHYSICAL", "VIRTUAL"},
			expectError: false,
		},
		{
			name:        "valid custom type - virtual",
			value:       AssetTypeVirtual,
			allowed:     []string{"PHYSICAL", "VIRTUAL"},
			expectError: false,
		},
		{
			name:        "invalid custom type",
			value:       AssetType("INVALID"),
			allowed:     []string{"PHYSICAL", "VIRTUAL"},
			expectError: true,
		},
		{
			name:        "custom type not in allowed list",
			value:       AssetTypePhysical,
			allowed:     []string{"VIRTUAL"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Check(tt.value, "asset_type", OneOfSlice(tt.allowed))

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
