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

package attachment

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/gid"
)

func TestParseID(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.TenantID{}, 25).String()
	path := Path(fileID)

	tests := []struct {
		name string
		raw  string
		ok   bool
	}{
		{name: "path", raw: path, ok: true},
		{name: "https", raw: "https://files.example.com" + path, ok: true},
		{name: "http", raw: "http://files.example.com" + path, ok: true},
		{name: "file scheme", raw: "file://" + path, ok: false},
		{name: "javascript scheme", raw: "javascript:" + path, ok: false},
		{name: "protocol relative", raw: "//evil.example" + path, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := ParseID(tt.raw)
			assert.Equal(t, tt.ok, ok)

			if tt.ok {
				assert.Equal(t, fileID, got)

				return
			}

			assert.Empty(t, got)
		})
	}
}
