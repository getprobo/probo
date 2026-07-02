// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
