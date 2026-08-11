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

package console_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/apps/console"
)

func TestContentSecurityPolicy_SubstitutesOrigins(t *testing.T) {
	t.Parallel()

	policy, err := console.ContentSecurityPolicy(
		"https://app.example.com",
		"https://probod.s3.eu-west-1.amazonaws.com",
	)
	require.NoError(t, err)
	require.NotEmpty(t, policy)

	assert.NotContains(t, policy, "{{")
	assert.NotContains(t, policy, "AppOrigin")
	assert.NotContains(t, policy, "FileStorageOrigin")
	assert.Contains(t, policy, "default-src 'self'")
	assert.Contains(t, policy, "frame-ancestors 'none'")
	assert.Contains(t, policy, "https://fonts.googleapis.com")
	assert.Contains(t, policy, "https://fonts.gstatic.com")
	assert.Contains(
		t,
		policy,
		"img-src 'self' data: https://app.example.com https://probod.s3.eu-west-1.amazonaws.com https://www.google.com https://*.gstatic.com",
	)
	assert.Contains(
		t,
		policy,
		"connect-src 'self' https://app.example.com https://probod.s3.eu-west-1.amazonaws.com",
	)

	for part := range strings.FieldsSeq(policy) {
		assert.NotEqual(t, "https:", strings.TrimSuffix(part, ";"))
	}

	assert.NotContains(t, policy, "\n")
	assert.NotContains(t, policy, "\r")
	assert.False(t, strings.HasPrefix(policy, " "))
	assert.False(t, strings.HasSuffix(policy, " "))
}

func TestContentSecurityPolicy_RejectsAppOriginInjection(t *testing.T) {
	t.Parallel()

	_, err := console.ContentSecurityPolicy(
		"https://evil.com;frame-ancestors",
		"https://probod.s3.eu-west-1.amazonaws.com",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CSP app origin")
}

func TestContentSecurityPolicy_RejectsFileStorageOriginInjection(t *testing.T) {
	t.Parallel()

	_, err := console.ContentSecurityPolicy(
		"https://app.example.com",
		"https://evil.com;frame-ancestors",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid CSP file storage origin")
}
