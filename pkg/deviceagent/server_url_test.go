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
	"github.com/stretchr/testify/require"
)

func TestNormalizeServerURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "adds https to bare hostnames", input: "eu.probo.com", want: EUConsoleURL},
		{name: "trims trailing slash", input: "https://us.probo.com/", want: USConsoleURL},
		{name: "rejects whitespace-only input", input: "   ", wantErr: true},
		{name: "rejects paths", input: "https://probo.example.com/workspace", wantErr: true},
		{name: "accepts uppercase scheme", input: "HTTPS://eu.probo.com/", want: EUConsoleURL},
		{name: "lowercases mixed-case hostname", input: "HTTPS://US.Probo.Com/", want: USConsoleURL},
		{name: "lowercases bare mixed-case hostname", input: "EU.probo.com", want: EUConsoleURL},
		{name: "lowercases hostname with port", input: "http://LocalHost:3000", want: "http://localhost:3000"},
		{name: "rejects user credentials", input: "https://user@eu.probo.com", wantErr: true},
		{name: "rejects query parameters", input: "https://eu.probo.com?foo=bar", wantErr: true},
		{name: "rejects fragments", input: "https://eu.probo.com#fragment", wantErr: true},
		{name: "rejects port-only host", input: "https://:443", wantErr: true},
		{name: "rejects bare port", input: ":443", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeServerURL(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
