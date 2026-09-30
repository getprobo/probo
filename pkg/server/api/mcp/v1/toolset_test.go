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

package mcp_v1

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInDocumentToolset(t *testing.T) {
	t.Parallel()

	assert.True(t, InDocumentToolset("readDocument"))
	assert.True(t, InDocumentToolset("listDocumentVersions"))
	assert.True(t, InDocumentToolset("publishDocument"))
	assert.True(t, InDocumentToolset("signDocument"))
	assert.True(t, InDocumentToolset("listUsers"))
	assert.False(t, InDocumentToolset("deleteCompliancePortalDocument"))
	assert.False(t, InDocumentToolset("listRisks"))
	assert.False(t, InDocumentToolset("publishCookieBannerVersion"))
}

func TestToolsetFromRequest(t *testing.T) {
	t.Parallel()

	t.Run("query selects documents", func(t *testing.T) {
		t.Parallel()

		req := &http.Request{URL: &url.URL{Path: "/", RawQuery: "toolset=documents"}}
		name, err := toolsetFromRequest(req)
		require.NoError(t, err)
		assert.Equal(t, toolsetDocuments, name)
	})

	t.Run("path selects documents", func(t *testing.T) {
		t.Parallel()

		req := &http.Request{URL: &url.URL{Path: "/toolsets/documents"}}
		name, err := toolsetFromRequest(req)
		require.NoError(t, err)
		assert.Equal(t, toolsetDocuments, name)
	})

	t.Run("unknown toolset", func(t *testing.T) {
		t.Parallel()

		req := &http.Request{URL: &url.URL{Path: "/", RawQuery: "toolset=risks"}}
		_, err := toolsetFromRequest(req)
		require.Error(t, err)
		assert.EqualError(t, err, "unknown mcp toolset")
	})

	t.Run("conflicting selectors", func(t *testing.T) {
		t.Parallel()

		req := &http.Request{URL: &url.URL{Path: "/toolsets/documents", RawQuery: "toolset=other"}}
		_, err := toolsetFromRequest(req)
		require.Error(t, err)
		assert.EqualError(t, err, "conflicting mcp toolset")
	})
}
