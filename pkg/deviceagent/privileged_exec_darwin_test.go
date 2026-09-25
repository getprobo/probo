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

//go:build darwin

package deviceagent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultExecutablePath(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/Library/Probo/probo-agent", DefaultExecutablePath())
	assert.Equal(t, "/usr/local/bin/probo-agent", LegacyExecutablePath())
}

func TestIsCanonicalExecutable(t *testing.T) {
	t.Parallel()

	assert.True(t, IsCanonicalExecutable("/Library/Probo/probo-agent"))
	assert.True(t, IsCanonicalExecutable("/Library/Probo/../Probo/probo-agent"))
	assert.False(t, IsCanonicalExecutable("/usr/local/bin/probo-agent"))
	assert.False(t, IsCanonicalExecutable("/tmp/probo-agent"))
}

func TestShouldMigrateExecutable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		exe          string
		plistProgram string
		want         bool
	}{
		{
			name: "already canonical",
			exe:  "/Library/Probo/probo-agent",
			want: false,
		},
		{
			name: "legacy path",
			exe:  "/usr/local/bin/probo-agent",
			want: true,
		},
		{
			name:         "legacy path after plist retarget",
			exe:          "/usr/local/bin/probo-agent",
			plistProgram: "/Library/Probo/probo-agent",
			want:         false,
		},
		{
			name:         "matches existing plist",
			exe:          "/tmp/probo-agent",
			plistProgram: "/tmp/probo-agent",
			want:         true,
		},
		{
			name:         "unrelated path with no plist",
			exe:          "/tmp/probo-agent",
			plistProgram: "",
			want:         false,
		},
		{
			name:         "unrelated path with other plist",
			exe:          "/tmp/other",
			plistProgram: "/Library/Probo/probo-agent",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, ShouldMigrateExecutable(tt.exe, tt.plistProgram))
			},
		)
	}
}

func TestValidatePrivilegedNode_RejectsSymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "probo-agent")

	require.NoError(t, os.WriteFile(target, []byte("x"), 0o755))
	require.NoError(t, os.Symlink(target, link))

	err := validatePrivilegedFile(link)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "symbolic link")
}

func TestEnsurePrivilegedExecutable_CopiesAndLocks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	src := filepath.Join(t.TempDir(), "src")
	dstDir := filepath.Join(root, "Probo")
	dst := filepath.Join(dstDir, "probo-agent")

	require.NoError(t, os.WriteFile(src, []byte("agent"), 0o700))

	got, err := ensurePrivilegedExecutable(src, dst)
	require.NoError(t, err)
	assert.Equal(t, dst, got)

	data, err := os.ReadFile(dst)
	require.NoError(t, err)
	assert.Equal(t, []byte("agent"), data)

	fi, err := os.Lstat(dst)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o755), fi.Mode().Perm())
	assert.True(t, fi.Mode().IsRegular())

	dirInfo, err := os.Lstat(dstDir)
	require.NoError(t, err)
	assert.True(t, dirInfo.IsDir())
	assert.Equal(t, os.FileMode(0o755), dirInfo.Mode().Perm())
}

func TestEnsurePrivilegedExecutable_RefusesSymlinkDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	src := filepath.Join(t.TempDir(), "src")
	realDir := filepath.Join(root, "real")
	dstDir := filepath.Join(root, "Probo")
	dst := filepath.Join(dstDir, "probo-agent")

	require.NoError(t, os.WriteFile(src, []byte("agent"), 0o755))
	require.NoError(t, os.Mkdir(realDir, 0o755))
	require.NoError(t, os.Symlink(realDir, dstDir))

	_, err := ensurePrivilegedExecutable(src, dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "symbolic link")
}
