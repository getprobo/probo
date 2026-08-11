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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdPartyService_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ThirdPartyID := factory.NewThirdParty(org1Owner).WithName("Org1 ThirdParty for Service").Create()

	var createResult struct {
		CreateThirdPartyService struct {
			ThirdPartyServiceEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyServiceEdge"`
		} `json:"createThirdPartyService"`
	}

	err := org1Owner.Execute(`
		mutation($input: CreateThirdPartyServiceInput!) {
			createThirdPartyService(input: $input) {
				thirdPartyServiceEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"thirdPartyId": org1ThirdPartyID,
			"name":         "Org1 Service",
		},
	}, &createResult)
	require.NoError(t, err)

	serviceID := createResult.CreateThirdPartyService.ThirdPartyServiceEdge.Node.ID

	t.Run("cannot update thirdPartyService from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: UpdateThirdPartyServiceInput!) {
				updateThirdPartyService(input: $input) {
					thirdPartyService { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":   serviceID,
				"name": "Hijacked Service",
			},
		})
		require.Error(t, err, "Should not be able to update thirdPartyService from another org")
	})

	t.Run("cannot delete thirdPartyService from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: DeleteThirdPartyServiceInput!) {
				deleteThirdPartyService(input: $input) {
					deletedThirdPartyServiceId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"thirdPartyServiceId": serviceID,
			},
		})
		require.Error(t, err, "Should not be able to delete thirdPartyService from another org")
	})

	t.Run("cannot create thirdPartyService on a thirdParty from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: CreateThirdPartyServiceInput!) {
				createThirdPartyService(input: $input) {
					thirdPartyServiceEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"thirdPartyId": org1ThirdPartyID,
				"name":         "Attacker Service",
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})
}
