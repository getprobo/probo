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

	"go.gearno.de/x/ref"
)

func TestMin(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		min       int
		wantError bool
	}{
		{"valid int", 10, 5, false},
		{"exact min", 5, 5, false},
		{"below min", 3, 5, true},
		{"valid int pointer", ref.Ref(10), 5, false},
		{"nil pointer", (*int)(nil), 5, false}, // Skip validation
		{"non-numeric", "test", 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Min(tt.min)(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Min() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		max       int
		wantError bool
	}{
		{"valid int", 5, 10, false},
		{"exact max", 10, 10, false},
		{"above max", 15, 10, true},
		{"valid int pointer", ref.Ref(5), 10, false},
		{"nil pointer", (*int)(nil), 10, false}, // Skip validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Max(tt.max)(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Max() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRange(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		min       int
		max       int
		wantError bool
	}{
		{"in range", 5, 1, 10, false},
		{"at min", 1, 1, 10, false},
		{"at max", 10, 1, 10, false},
		{"below range", 0, 1, 10, true},
		{"above range", 11, 1, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Range(tt.min, tt.max)(tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Range() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
