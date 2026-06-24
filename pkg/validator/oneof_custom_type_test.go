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
