// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package coredata

import "testing"

func TestAccessEntryFlagScan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    AccessEntryFlag
		wantErr bool
	}{
		{name: "none string", input: "NONE", want: AccessEntryFlagNone},
		{name: "orphaned string", input: "ORPHANED", want: AccessEntryFlagOrphaned},
		{name: "inactive string", input: "INACTIVE", want: AccessEntryFlagInactive},
		{name: "excessive string", input: "EXCESSIVE", want: AccessEntryFlagExcessive},
		{name: "role_mismatch bytes", input: []byte("ROLE_MISMATCH"), want: AccessEntryFlagRoleMismatch},
		{name: "new string", input: "NEW", want: AccessEntryFlagNew},
		{name: "invalid value", input: "BOGUS", wantErr: true},
		{name: "unsupported type", input: 42, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got AccessEntryFlag
			err := got.Scan(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Scan(%v) expected error", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("Scan(%v) returned error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("Scan(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAccessEntryFlagValue(t *testing.T) {
	t.Parallel()

	got, err := AccessEntryFlagNone.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}
	if got != "NONE" {
		t.Fatalf("Value() = %q, want %q", got, "NONE")
	}
}
