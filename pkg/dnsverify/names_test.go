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

package dnsverify_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/dnsverify"
)

func TestEqualNames(t *testing.T) {
	t.Parallel()

	assert.True(t, dnsverify.EqualNames("trust.example.com", "trust.example.com."))
	assert.True(t, dnsverify.EqualNames("Trust.Example.COM", "trust.example.com"))
	assert.False(t, dnsverify.EqualNames("trust.example.com", "example.com"))
}

func TestCheckNames(t *testing.T) {
	t.Parallel()

	t.Run("subdomain starts at child then walks to apex", func(t *testing.T) {
		t.Parallel()

		names, err := dnsverify.CheckNames("trust.example.com")

		require.NoError(t, err)
		assert.Equal(t, []string{"trust.example.com", "example.com"}, names)
	})

	t.Run("apex stays on apex", func(t *testing.T) {
		t.Parallel()

		names, err := dnsverify.CheckNames("example.com")

		require.NoError(t, err)
		assert.Equal(t, []string{"example.com"}, names)
	})

	t.Run("nested subdomain walks each parent", func(t *testing.T) {
		t.Parallel()

		names, err := dnsverify.CheckNames("portal.trust.example.com")

		require.NoError(t, err)
		assert.Equal(
			t,
			[]string{"portal.trust.example.com", "trust.example.com", "example.com"},
			names,
		)
	})
}
