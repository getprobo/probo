// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
