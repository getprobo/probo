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

package content_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cmd/skill/content"
)

func TestList_EmbeddedSkills(t *testing.T) {
	t.Parallel()

	skills, err := content.List()
	require.NoError(t, err)

	names := make([]string, 0, len(skills))
	for _, skill := range skills {
		names = append(names, skill.Name)
		assert.NotEmpty(t, skill.Description)
		assert.LessOrEqual(t, len(skill.Description), 1024)

		body, err := content.Read(skill.Name, "")
		require.NoError(t, err)
		assert.Contains(t, string(body), "name: "+skill.Name)
	}

	assert.Equal(t, []string{"prb-auth", "prb-records", "using-prb"}, names)
}

func TestRead_RejectsParentPath(t *testing.T) {
	t.Parallel()

	_, err := content.Read("using-prb", "../prb-auth/SKILL.md")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid skill path")
}

func TestInstall_WritesAndRefusesOverwrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	written, err := content.Install(dir, false)
	require.NoError(t, err)
	require.NotEmpty(t, written)

	body, err := os.ReadFile(filepath.Join(dir, "using-prb", "SKILL.md"))
	require.NoError(t, err)
	assert.Contains(t, string(body), "name: using-prb")

	_, err = content.Install(dir, false)
	require.NoError(t, err)

	skillPath := filepath.Join(dir, "using-prb", "SKILL.md")
	require.NoError(t, os.WriteFile(skillPath, []byte("local edit\n"), 0o644))

	_, err = content.Install(dir, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--force")

	_, err = content.Install(dir, true)
	require.NoError(t, err)

	body, err = os.ReadFile(skillPath)
	require.NoError(t, err)
	assert.Contains(t, string(body), "name: using-prb")
}
