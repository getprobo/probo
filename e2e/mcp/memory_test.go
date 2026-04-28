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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_Memory(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var getResult struct {
		Memory struct {
			OrganizationID string  `json:"organization_id"`
			Product        *string `json:"product"`
			Architecture   *string `json:"architecture"`
			Team           *string `json:"team"`
			Processes      *string `json:"processes"`
			Customers      *string `json:"customers"`
		} `json:"memory"`
	}
	mc.CallToolInto("getMemory", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	assert.Equal(t, orgID, getResult.Memory.OrganizationID)

	var updateResult struct {
		Memory struct {
			OrganizationID string  `json:"organization_id"`
			Product        *string `json:"product"`
			Architecture   *string `json:"architecture"`
		} `json:"memory"`
	}
	mc.CallToolInto("updateMemory", map[string]any{
		"organization_id": orgID,
		"product":         "We build compliance software.",
		"architecture":    "Monolith deployed on AWS ECS.",
	}, &updateResult)
	assert.Equal(t, orgID, updateResult.Memory.OrganizationID)
	require.NotNil(t, updateResult.Memory.Product)
	assert.Equal(t, "We build compliance software.", *updateResult.Memory.Product)
	require.NotNil(t, updateResult.Memory.Architecture)
	assert.Equal(t, "Monolith deployed on AWS ECS.", *updateResult.Memory.Architecture)
}
