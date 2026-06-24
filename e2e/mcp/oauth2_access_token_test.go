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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_OAuth2AccessToken_ListOrganizations(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	token := owner.CreateOAuth2AccessToken("e2e-mcp-oauth-token", []string{"v1:iam:read"})

	mc := testutil.NewMCPClientWithAccessToken(t, owner, token)

	var result struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
	}
	mc.CallToolInto("listOrganizations", map[string]any{}, &result)

	require.NotEmpty(t, result.Organizations)
}

func TestMCP_OAuth2AccessToken_ScopeEnforcement(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	token := owner.CreateOAuth2AccessToken("e2e-mcp-oauth-scope", []string{"v1:org:read"})

	mc := testutil.NewMCPClientWithAccessToken(t, owner, token)

	msg := mc.CallToolExpectToolError("listThirdParties", map[string]any{
		"organizationId": orgID,
	})
	assert.Equal(t, "insufficient scope", msg)
}
