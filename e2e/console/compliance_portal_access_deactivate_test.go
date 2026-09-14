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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const deactivateCompliancePortalAccessMutation = `
	mutation($input: DeactivateCompliancePortalAccessInput!) {
		deactivateCompliancePortalAccess(input: $input) {
			compliancePortalAccess {
				id
				state
			}
		}
	}
`

const activateCompliancePortalAccessMutation = `
	mutation($input: ActivateCompliancePortalAccessInput!) {
		activateCompliancePortalAccess(input: $input) {
			compliancePortalAccess {
				id
				state
			}
		}
	}
`

func deactivateCompliancePortalAccess(t *testing.T, client *testutil.Client, accessID string) string {
	t.Helper()

	var result struct {
		DeactivateCompliancePortalAccess struct {
			CompliancePortalAccess struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"compliancePortalAccess"`
		} `json:"deactivateCompliancePortalAccess"`
	}
	err := client.Execute(deactivateCompliancePortalAccessMutation, map[string]any{
		"input": map[string]any{"id": accessID},
	}, &result)
	require.NoError(t, err)

	return result.DeactivateCompliancePortalAccess.CompliancePortalAccess.State
}

func activateCompliancePortalAccess(t *testing.T, client *testutil.Client, accessID string) string {
	t.Helper()

	var result struct {
		ActivateCompliancePortalAccess struct {
			CompliancePortalAccess struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"compliancePortalAccess"`
		} `json:"activateCompliancePortalAccess"`
	}
	err := client.Execute(activateCompliancePortalAccessMutation, map[string]any{
		"input": map[string]any{"id": accessID},
	}, &result)
	require.NoError(t, err)

	return result.ActivateCompliancePortalAccess.CompliancePortalAccess.State
}

func TestCompliancePortalAccess_DeactivateThenActivate(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)

	node := createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              factory.SafeEmail(),
	})
	assert.Equal(t, "ACTIVE", node.State)

	assert.Equal(t, "DEACTIVATED", deactivateCompliancePortalAccess(t, owner, node.ID))
	assert.Equal(t, "DEACTIVATED", deactivateCompliancePortalAccess(t, owner, node.ID))
	assert.Equal(t, "ACTIVE", activateCompliancePortalAccess(t, owner, node.ID))
	assert.Equal(t, "ACTIVE", activateCompliancePortalAccess(t, owner, node.ID))
}

func TestCompliancePortalAccess_DeactivateAccessManager(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	manager := testutil.NewClientInOrg(t, testutil.RoleCompliancePortalAccessManager, owner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	compliancePortalID := compliancePortalID(t, owner)

	node := createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              factory.SafeEmail(),
	})

	assert.Equal(t, "DEACTIVATED", deactivateCompliancePortalAccess(t, manager, node.ID))
	assert.Equal(t, "ACTIVE", activateCompliancePortalAccess(t, manager, node.ID))

	err := viewer.ExecuteShouldFail(deactivateCompliancePortalAccessMutation, map[string]any{
		"input": map[string]any{"id": node.ID},
	})
	require.Error(t, err)
}

func TestCompliancePortalAccess_DeactivateTenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)
	org1CompliancePortalID := compliancePortalID(t, org1Owner)

	node := createCompliancePortalAccess(t, org1Owner, map[string]any{
		"compliancePortalId": org1CompliancePortalID,
		"email":              factory.SafeEmail(),
	})

	err := org2Owner.ExecuteShouldFail(deactivateCompliancePortalAccessMutation, map[string]any{
		"input": map[string]any{"id": node.ID},
	})
	require.Error(t, err)

	err = org2Owner.ExecuteShouldFail(activateCompliancePortalAccessMutation, map[string]any{
		"input": map[string]any{"id": node.ID},
	})
	require.Error(t, err)
}
