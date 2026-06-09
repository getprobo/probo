// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

func TestControlMaturityLevelIsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level ControlMaturityLevel
		want  bool
	}{
		{name: "none", level: ControlMaturityLevelNone, want: true},
		{name: "initial", level: ControlMaturityLevelInitial, want: true},
		{name: "managed", level: ControlMaturityLevelManaged, want: true},
		{name: "defined", level: ControlMaturityLevelDefined, want: true},
		{name: "quantitatively managed", level: ControlMaturityLevelQuantitativelyManaged, want: true},
		{name: "optimizing", level: ControlMaturityLevelOptimizing, want: true},
		{name: "empty string", level: "", want: false},
		{name: "unknown value", level: "BOGUS", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.level.IsValid(); got != tt.want {
				t.Fatalf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestControlMaturityLevelMarshalUnmarshalText(t *testing.T) {
	t.Parallel()

	for _, level := range []ControlMaturityLevel{
		ControlMaturityLevelNone,
		ControlMaturityLevelInitial,
		ControlMaturityLevelManaged,
		ControlMaturityLevelDefined,
		ControlMaturityLevelQuantitativelyManaged,
		ControlMaturityLevelOptimizing,
	} {
		t.Run(string(level), func(t *testing.T) {
			t.Parallel()

			data, err := level.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText() returned error: %v", err)
			}

			var roundtrip ControlMaturityLevel
			if err := roundtrip.UnmarshalText(data); err != nil {
				t.Fatalf("UnmarshalText(%q) returned error: %v", string(data), err)
			}

			if roundtrip != level {
				t.Fatalf("roundtrip = %q, want %q", roundtrip, level)
			}
		})
	}

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		var l ControlMaturityLevel
		if err := l.UnmarshalText([]byte("BOGUS")); err == nil {
			t.Fatal("UnmarshalText(BOGUS) expected error")
		}
	})
}
