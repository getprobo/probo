// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package version_test

import (
	"strings"
	"testing"

	"go.probo.inc/probo/pkg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBuildInfo(t *testing.T) {
	info := version.GetBuildInfo()

	// These should always be present
	assert.NotEmpty(t, info.GoVersion, "Go version should not be empty")
	assert.True(t, strings.HasPrefix(info.GoVersion, "go"), "Go version should start with 'go'")

	// Version should have a default value
	assert.NotEmpty(t, info.Version, "Version should not be empty")

	// Commit and BuildDate may be "unknown" in test environment
	assert.NotEmpty(t, info.Commit, "Commit should not be empty")
	assert.NotEmpty(t, info.BuildDate, "BuildDate should not be empty")
}

func TestUserAgent(t *testing.T) {
	tests := []struct {
		name      string
		component string
		checks    []string
	}{
		{
			name:      "ACME client user agent",
			component: "acme-client",
			checks: []string{
				"Probo/",
				"acme-client",
				"Go/go",
			},
		},
		{
			name:      "Probod user agent",
			component: "probod",
			checks: []string{
				"Probo/",
				"probod",
				"Go/go",
			},
		},
		{
			name:      "Custom component",
			component: "test-component",
			checks: []string{
				"Probo/",
				"test-component",
				"Go/go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ua := version.UserAgent(tt.component)
			require.NotEmpty(t, ua, "User agent should not be empty")

			for _, check := range tt.checks {
				assert.Contains(t, ua, check, "User agent should contain %s", check)
			}

			// Check format structure
			// Should be like: Probo/version (component; ...) Go/version
			assert.True(t, strings.HasPrefix(ua, "Probo/"), "User agent should start with Probo/")
			assert.Contains(t, ua, "(", "User agent should have opening parenthesis")
			assert.Contains(t, ua, ")", "User agent should have closing parenthesis")
		})
	}
}

func TestVersion(t *testing.T) {
	v := version.Version()
	assert.NotEmpty(t, v, "Version should not be empty")
}

func TestFullVersion(t *testing.T) {
	fv := version.FullVersion()
	require.NotEmpty(t, fv, "Full version should not be empty")

	// Check that it contains expected fields
	assert.Contains(t, fv, "Version:", "Full version should contain Version field")
	assert.Contains(t, fv, "Go:", "Full version should contain Go field")

	// Check it's multiline
	lines := strings.Split(fv, "\n")
	assert.GreaterOrEqual(t, len(lines), 2, "Full version should have at least 2 lines")
}

func TestUserAgentConsistency(t *testing.T) {
	// Test that multiple calls return consistent results
	ua1 := version.UserAgent("test")
	ua2 := version.UserAgent("test")
	assert.Equal(t, ua1, ua2, "User agent should be consistent across calls")

	// Test that different components produce different user agents
	ua3 := version.UserAgent("different")
	assert.NotEqual(t, ua1, ua3, "Different components should produce different user agents")

	// But they should have the same version info
	parts1 := strings.SplitN(ua1, " ", 2)
	parts3 := strings.SplitN(ua3, " ", 2)
	assert.Equal(t, parts1[0], parts3[0], "Version part should be the same")
}

func BenchmarkUserAgent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = version.UserAgent("benchmark")
	}
}

func BenchmarkGetBuildInfo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = version.GetBuildInfo()
	}
}
