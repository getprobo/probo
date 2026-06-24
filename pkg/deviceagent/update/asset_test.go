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

func TestLayoutFor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		goos, goarch string
		archive      string
		dir          string
		binary       string
		isZip        bool
	}{
		{"linux", "amd64", "probo-agent_Linux_x86_64.tar.gz", "probo-agent_Linux_x86_64", "probo-agent", false},
		{"linux", "arm64", "probo-agent_Linux_arm64.tar.gz", "probo-agent_Linux_arm64", "probo-agent", false},
		{"darwin", "amd64", "probo-agent_Darwin_x86_64.tar.gz", "probo-agent_Darwin_x86_64", "probo-agent", false},
		{"darwin", "arm64", "probo-agent_Darwin_arm64.tar.gz", "probo-agent_Darwin_arm64", "probo-agent", false},
		{"windows", "amd64", "probo-agent_Windows_x86_64.zip", "probo-agent_Windows_x86_64", "probo-agent.exe", true},
		{"windows", "arm64", "probo-agent_Windows_arm64.zip", "probo-agent_Windows_arm64", "probo-agent.exe", true},
		{"freebsd", "amd64", "probo-agent_Freebsd_x86_64.tar.gz", "probo-agent_Freebsd_x86_64", "probo-agent", false},
		{"freebsd", "arm64", "probo-agent_Freebsd_arm64.tar.gz", "probo-agent_Freebsd_arm64", "probo-agent", false},
	}

	for _, tc := range cases {
		t.Run(
			tc.goos+"/"+tc.goarch,
			func(t *testing.T) {
				t.Parallel()

				layout, err := LayoutFor(tc.goos, tc.goarch)
				require.NoError(t, err)
				assert.Equal(t, tc.archive, layout.ArchiveName)
				assert.Equal(t, tc.dir, layout.ArchiveDir)
				assert.Equal(t, tc.binary, layout.BinaryName)
				assert.Equal(t, tc.isZip, layout.IsZip)
			},
		)
	}

	t.Run(
		"unsupported GOOS",
		func(t *testing.T) {
			t.Parallel()

			_, err := LayoutFor("plan9", "amd64")
			require.Error(t, err)
		},
	)
}
