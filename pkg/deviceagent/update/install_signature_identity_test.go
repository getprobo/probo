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

package update

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCodeSigningIdentity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		out  string
		want codeSigningIdentity
	}{
		{
			name: "developer id signed binary",
			out: "" +
				"Executable=/usr/local/bin/probo-agent\n" +
				"Identifier=com.probo.agent\n" +
				"Format=Mach-O thin (arm64)\n" +
				"TeamIdentifier=ABCD123456\n",
			want: codeSigningIdentity{
				Team:       "ABCD123456",
				Identifier: "com.probo.agent",
			},
		},
		{
			name: "team identifier not set",
			out: "" +
				"Identifier=probo-agent\n" +
				"Signature=adhoc\n" +
				"TeamIdentifier=not set\n",
			want: codeSigningIdentity{
				Team:       "",
				Identifier: "probo-agent",
			},
		},
		{
			name: "unsigned output has empty fields",
			out:  "code object is not signed at all\n",
			want: codeSigningIdentity{},
		},
		{
			name: "empty identifier value",
			out: "" +
				"Identifier=\n" +
				"TeamIdentifier=ABCD123456\n",
			want: codeSigningIdentity{
				Team:       "ABCD123456",
				Identifier: "",
			},
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				got := parseCodeSigningIdentity(tc.out)
				assert.Equal(t, tc.want, got)
			},
		)
	}
}

func TestCodeSigningIdentitiesCompatible(t *testing.T) {
	t.Parallel()

	stable := codeSigningIdentity{
		Team:       "ABCD123456",
		Identifier: "com.probo.agent",
	}

	cases := []struct {
		name      string
		current   codeSigningIdentity
		candidate codeSigningIdentity
		wantErr   string
	}{
		{
			name:    "allows unsigned to signed upgrade",
			current: codeSigningIdentity{},
			candidate: codeSigningIdentity{
				Team:       "ABCD123456",
				Identifier: "com.probo.agent",
			},
		},
		{
			name:      "allows matching team and identifier",
			current:   stable,
			candidate: stable,
		},
		{
			name:    "refuses signed to unsigned downgrade",
			current: stable,
			candidate: codeSigningIdentity{
				Identifier: "com.probo.agent",
			},
			wantErr: "refusing signature downgrade",
		},
		{
			name:    "refuses team mismatch",
			current: stable,
			candidate: codeSigningIdentity{
				Team:       "OTHERTEAM1",
				Identifier: "com.probo.agent",
			},
			wantErr: "does not match current Team ID",
		},
		{
			name:    "refuses same team with different identifier",
			current: stable,
			candidate: codeSigningIdentity{
				Team:       "ABCD123456",
				Identifier: "com.probo.agent.other",
			},
			wantErr: "does not match current Identifier",
		},
	}

	for _, tc := range cases {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				err := codeSigningIdentitiesCompatible(tc.current, tc.candidate)
				if tc.wantErr == "" {
					assert.NoError(t, err)
					return
				}

				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			},
		)
	}
}

func TestIsUnsignedCodesignOutput(t *testing.T) {
	t.Parallel()

	assert.True(t, isUnsignedCodesignOutput("code object is not signed at all"))
	assert.True(t, isUnsignedCodesignOutput("TeamIdentifier=not set\n"))
	assert.False(t, isUnsignedCodesignOutput("TeamIdentifier=ABCD123456\n"))
}
