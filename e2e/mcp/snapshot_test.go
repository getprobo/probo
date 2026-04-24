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

func TestMCP_Snapshot(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create a thirdParty so the snapshot has data
	factory.CreateThirdParty(owner)

	// Take snapshot
	var takeResult struct {
		Snapshot struct {
			ID string `json:"id"`
		} `json:"snapshot"`
	}
	mc.CallToolInto("takeSnapshot", map[string]any{
		"organizationId": orgID,
		"name":           factory.SafeName("Snapshot"),
		"snapshotsType":  "THIRD_PARTIES",
	}, &takeResult)
	require.NotEmpty(t, takeResult.Snapshot.ID)

	// Get
	var getResult struct {
		Snapshot struct {
			ID string `json:"id"`
		} `json:"snapshot"`
	}
	mc.CallToolInto("getSnapshot", map[string]any{
		"id": takeResult.Snapshot.ID,
	}, &getResult)
	assert.Equal(t, takeResult.Snapshot.ID, getResult.Snapshot.ID)

	// List
	var listResult struct {
		Snapshots []struct {
			ID string `json:"id"`
		} `json:"snapshots"`
	}
	mc.CallToolInto("listSnapshots", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Snapshots)
}
