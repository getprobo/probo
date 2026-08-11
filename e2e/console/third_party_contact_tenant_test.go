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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdPartyContact_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ThirdPartyID := factory.NewThirdParty(org1Owner).WithName("Org1 ThirdParty for Contact").Create()

	var createResult struct {
		CreateThirdPartyContact struct {
			ThirdPartyContactEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyContactEdge"`
		} `json:"createThirdPartyContact"`
	}

	err := org1Owner.Execute(`
		mutation($input: CreateThirdPartyContactInput!) {
			createThirdPartyContact(input: $input) {
				thirdPartyContactEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"thirdPartyId": org1ThirdPartyID,
			"fullName":     "Org1 Contact",
			"email":        fmt.Sprintf("org1.contact.%d@thirdParty.com", time.Now().UnixNano()),
		},
	}, &createResult)
	require.NoError(t, err)

	contactID := createResult.CreateThirdPartyContact.ThirdPartyContactEdge.Node.ID

	t.Run("cannot update thirdPartyContact from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: UpdateThirdPartyContactInput!) {
				updateThirdPartyContact(input: $input) {
					thirdPartyContact { id }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"id":       contactID,
				"fullName": "Hijacked Contact",
			},
		})
		require.Error(t, err, "Should not be able to update thirdPartyContact from another org")
	})

	t.Run("cannot delete thirdPartyContact from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: DeleteThirdPartyContactInput!) {
				deleteThirdPartyContact(input: $input) {
					deletedThirdPartyContactId
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"thirdPartyContactId": contactID,
			},
		})
		require.Error(t, err, "Should not be able to delete thirdPartyContact from another org")
	})

	t.Run("cannot create thirdPartyContact on a thirdParty from another organization", func(t *testing.T) {
		_, err := org2Owner.Do(`
			mutation($input: CreateThirdPartyContactInput!) {
				createThirdPartyContact(input: $input) {
					thirdPartyContactEdge { node { id } }
				}
			}
		`, map[string]any{
			"input": map[string]any{
				"thirdPartyId": org1ThirdPartyID,
				"fullName":     "Attacker Contact",
				"email":        fmt.Sprintf("attacker.%d@thirdParty.com", time.Now().UnixNano()),
			},
		})
		require.Error(t, err, "must not accept a thirdPartyId belonging to another organization")
	})
}
