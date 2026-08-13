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

package deviceagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseWindowsVerOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "english version banner",
			input: "Microsoft Windows [Version 10.0.26200.8875]",
			want:  "10.0.26200.8875",
		},
		{
			name:  "french version banner",
			input: "Microsoft Windows [version 10.0.26200.8875]",
			want:  "10.0.26200.8875",
		},
		{
			name:  "spanish version banner",
			input: "Microsoft Windows [Versión 10.0.26200.8875]",
			want:  "10.0.26200.8875",
		},
		{
			name:  "italian version banner",
			input: "Microsoft Windows [versione 10.0.26200.8875]",
			want:  "10.0.26200.8875",
		},
		{
			name:  "japanese version banner",
			input: "Microsoft Windows [バージョン 10.0.26200.8875]",
			want:  "10.0.26200.8875",
		},
		{
			name:  "russian version banner",
			input: "Microsoft Windows [Версия 10.0.26200.8875]",
			want:  "10.0.26200.8875",
		},
		{
			name:  "chinese version banner",
			input: "Microsoft Windows [版本 10.0.26200.8875]",
			want:  "10.0.26200.8875",
		},
		{
			name:  "surrounding whitespace",
			input: "  \nMicrosoft Windows [Version 10.0.26200.8875]\r\n  ",
			want:  "10.0.26200.8875",
		},
		{
			name:  "already short version",
			input: "10.0.26200.8875",
			want:  "10.0.26200.8875",
		},
		{
			name:  "empty brackets fall back to original",
			input: "Microsoft Windows []",
			want:  "Microsoft Windows []",
		},
		{
			name:  "version prefix without number falls back",
			input: "Microsoft Windows [Version]",
			want:  "Microsoft Windows [Version]",
		},
		{
			name:  "unparseable text",
			input: "not a windows ver banner",
			want:  "not a windows ver banner",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, parseWindowsVerOutput(tt.input))
			},
		)
	}
}
