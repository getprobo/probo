// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_User_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var createResult struct {
		User struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"user"`
	}
	mc.CallToolInto("createUser", map[string]any{
		"organizationId": orgID,
		"fullName":       "Test User",
		"emailAddress":   factory.SafeEmail(),
		"role":           "EMPLOYEE",
		"kind":           "EMPLOYEE",
	}, &createResult)
	require.NotEmpty(t, createResult.User.ID)

	// Get
	var getResult struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	mc.CallToolInto("getUser", map[string]any{
		"id": createResult.User.ID,
	}, &getResult)
	assert.Equal(t, createResult.User.ID, getResult.User.ID)

	// List
	var listResult struct {
		Users []struct {
			ID string `json:"id"`
		} `json:"users"`
	}
	mc.CallToolInto("listUsers", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Users)
}
