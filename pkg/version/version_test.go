// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package version_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/version"
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
