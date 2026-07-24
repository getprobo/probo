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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrollmentMarker(t *testing.T) {
	t.Parallel()

	t.Run(
		"marker tracks enrollment state",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			runDir := EnrollmentRunDir(dir)

			enrolled, err := IsEnrolled(runDir)
			require.NoError(t, err)
			assert.False(t, enrolled)

			require.NoError(t, MarkEnrolled(runDir))
			enrolled, err = IsEnrolled(runDir)
			require.NoError(t, err)
			assert.True(t, enrolled)

			markerPath := enrollmentMarkerPath(runDir)
			assert.NotEqual(t, filepath.Join(dir, EnrollmentMarkerName), markerPath)

			info, err := os.Stat(markerPath)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(EnrollmentMarkerMode), info.Mode().Perm())

			info, err = os.Stat(runDir)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(EnrollmentRunDirMode), info.Mode().Perm())

			require.NoError(t, ClearEnrollmentMarker(runDir))
			enrolled, err = IsEnrolled(runDir)
			require.NoError(t, err)
			assert.False(t, enrolled)
		},
	)

	t.Run(
		"clear is idempotent",
		func(t *testing.T) {
			t.Parallel()

			runDir := EnrollmentRunDir(t.TempDir())

			require.NoError(t, ClearEnrollmentMarker(runDir))
			enrolled, err := IsEnrolled(runDir)
			require.NoError(t, err)
			assert.False(t, enrolled)
		},
	)

	t.Run(
		"default config dir resolves run dir",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, DefaultEnrollmentRunDir(), EnrollmentRunDir(""))
		},
	)

	t.Run(
		"non-ENOENT stat errors are not treated as unenrolled",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			runDir := EnrollmentRunDir(dir)
			require.NoError(t, os.MkdirAll(filepath.Dir(runDir), 0o755))
			require.NoError(t, os.WriteFile(runDir, []byte("x"), 0o644))

			enrolled, err := IsEnrolled(runDir)
			assert.False(t, enrolled)
			require.Error(t, err)
		},
	)
}
