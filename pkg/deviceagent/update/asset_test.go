// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
