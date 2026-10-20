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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_UpdateLinearSettings_WhenNotConnected(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	ownerMC := testutil.NewMCPClient(t, owner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)
	orgID := owner.GetOrganizationID().String()

	msg := ownerMC.CallToolExpectToolError("updateLinearSettings", map[string]any{
		"organization_id": orgID,
		"default_team_id": "team-id",
		"issue_import":    "ACTIVE",
	})
	assert.Contains(t, msg, "linear connector is not connected")

	denied := viewerMC.CallToolExpectToolError("updateLinearSettings", map[string]any{
		"organization_id": orgID,
		"issue_import":    "INACTIVE",
	})
	assert.Contains(t, denied, "permission denied")
}
