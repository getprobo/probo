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
