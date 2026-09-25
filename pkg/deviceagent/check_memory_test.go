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
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/deviceagent/checks"
)

func TestLoadCheckMemory_MissingFile(t *testing.T) {
	t.Parallel()

	memory, err := loadCheckMemory(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, memory)
}

func TestLoadCheckMemory_CorruptFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(
		t,
		os.WriteFile(
			filepath.Join(dir, checkMemoryFileName),
			[]byte("{"),
			0o600,
		),
	)

	_, err := loadCheckMemory(dir)
	require.Error(t, err)
}

func TestRememberChecks_UnreadableMemoryIsNotReplaced(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, checkMemoryFileName)
	require.NoError(t, os.WriteFile(path, []byte("{"), 0o600))

	a := New(dir, "test", nil)
	results := a.rememberChecks(
		context.Background(),
		[]checks.Result{
			{CheckKey: "FIREWALL", Status: checks.StatusPass},
		},
	)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, []byte("{"), data)
	assert.Equal(t, checks.StatusPass, results[0].Status)
}

func TestSaveAndLoadCheckMemory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(
		t,
		saveCheckMemory(dir, checkMemory{"SCREEN_LOCK": checks.StatusFail}),
	)

	memory, err := loadCheckMemory(dir)
	require.NoError(t, err)
	assert.Equal(t, checks.StatusFail, memory["SCREEN_LOCK"])
}

func TestApplyCheckMemory(t *testing.T) {
	t.Parallel()

	t.Run(
		"stores pass and fail",
		func(t *testing.T) {
			t.Parallel()

			memory := checkMemory{}
			results := []checks.Result{
				{
					CheckKey: "SCREEN_LOCK",
					Status:   checks.StatusFail,
				},
			}

			assert.True(t, applyCheckMemory(results, memory))
			assert.Equal(t, checks.StatusFail, results[0].Status)
			assert.Equal(t, checks.StatusFail, memory["SCREEN_LOCK"])
		},
	)

	t.Run(
		"recalls last pass for a rememberable unknown",
		func(t *testing.T) {
			t.Parallel()

			memory := checkMemory{"SCREEN_LOCK": checks.StatusPass}
			results := []checks.Result{
				{
					CheckKey: "SCREEN_LOCK",
					Status:   checks.StatusUnknown,
					Evidence: map[string]any{
						"backend": "hkey_users",
						"note":    "no interactive user hives loaded",
					},
					Rememberable: true,
				},
			}

			assert.False(t, applyCheckMemory(results, memory))
			assert.Equal(t, checks.StatusPass, results[0].Status)
			assert.Equal(t, true, results[0].Evidence["screen_lock_enforced"])
			assert.Equal(t, true, results[0].Evidence["remembered"])
			assert.Equal(t, "hkey_users", results[0].Evidence["backend"])
		},
	)

	t.Run(
		"leaves rememberable unknown without memory",
		func(t *testing.T) {
			t.Parallel()

			memory := checkMemory{}
			results := []checks.Result{
				{
					CheckKey:     "SCREEN_LOCK",
					Status:       checks.StatusUnknown,
					Rememberable: true,
				},
			}

			assert.False(t, applyCheckMemory(results, memory))
			assert.Equal(t, checks.StatusUnknown, results[0].Status)
			assert.Empty(t, memory)
		},
	)

	t.Run(
		"ignores a non-rememberable unknown",
		func(t *testing.T) {
			t.Parallel()

			memory := checkMemory{"SCREEN_LOCK": checks.StatusPass}
			results := []checks.Result{
				{
					CheckKey: "SCREEN_LOCK",
					Status:   checks.StatusUnknown,
					Evidence: map[string]any{"error": "powershell failed"},
				},
			}

			assert.False(t, applyCheckMemory(results, memory))
			assert.Equal(t, checks.StatusUnknown, results[0].Status)
			assert.Nil(t, results[0].Evidence["remembered"])
			assert.Equal(t, checks.StatusPass, memory["SCREEN_LOCK"])
		},
	)

	t.Run(
		"does not store not applicable",
		func(t *testing.T) {
			t.Parallel()

			memory := checkMemory{}
			results := []checks.Result{
				{CheckKey: "SCREEN_LOCK", Status: checks.StatusNotApplicable},
			}

			assert.False(t, applyCheckMemory(results, memory))
			assert.Empty(t, memory)
		},
	)

	t.Run(
		"does not mark dirty when status is unchanged",
		func(t *testing.T) {
			t.Parallel()

			memory := checkMemory{"SCREEN_LOCK": checks.StatusFail}
			results := []checks.Result{
				{
					CheckKey: "SCREEN_LOCK",
					Status:   checks.StatusFail,
				},
			}

			assert.False(t, applyCheckMemory(results, memory))
			assert.Equal(t, checks.StatusFail, memory["SCREEN_LOCK"])
		},
	)

	t.Run(
		"marks dirty when stored status changes",
		func(t *testing.T) {
			t.Parallel()

			memory := checkMemory{"SCREEN_LOCK": checks.StatusPass}
			results := []checks.Result{
				{
					CheckKey: "SCREEN_LOCK",
					Status:   checks.StatusFail,
				},
			}

			assert.True(t, applyCheckMemory(results, memory))
			assert.Equal(t, checks.StatusFail, memory["SCREEN_LOCK"])
		},
	)
}
