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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/gid"
)

func TestThirdParty_ImportFromCommon(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Catalog: a common third party and a common tracker pattern linked
	// to it.
	commonName := factory.SafeName("ImportTP")
	commonThirdPartyID := seedCommonThirdParty(t, commonName)
	commonPatternID := seedCommonTrackerPattern(t)
	linkCommonTrackerPatternToVendor(t, commonPatternID, commonThirdPartyID)

	// An org tracker pattern linked to that catalog row but with no org
	// third party yet (the state the mapping worker now leaves behind).
	bannerID := factory.CreateCookieBanner(owner)
	categoryID := factory.CreateCookieCategory(owner, bannerID)
	patternID := factory.CreateTrackerPattern(owner, categoryID)
	linkTrackerPatternToCommon(t, patternID, commonPatternID)

	const mutation = `
		mutation($input: ImportThirdPartyFromCommonInput!) {
			importThirdPartyFromCommon(input: $input) {
				created
				thirdPartyEdge {
					node {
						id
						name
					}
				}
			}
		}
	`

	input := map[string]any{
		"organizationId":     owner.GetOrganizationID().String(),
		"commonThirdPartyId": commonThirdPartyID.String(),
	}

	var first struct {
		ImportThirdPartyFromCommon struct {
			Created        bool `json:"created"`
			ThirdPartyEdge struct {
				Node struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"importThirdPartyFromCommon"`
	}

	require.NoError(t, owner.Execute(mutation, map[string]any{"input": input}, &first))
	assert.True(t, first.ImportThirdPartyFromCommon.Created, "first import must create the org third party")

	importedID := first.ImportThirdPartyFromCommon.ThirdPartyEdge.Node.ID
	require.NotEmpty(t, importedID)
	assert.Equal(t, commonName, first.ImportThirdPartyFromCommon.ThirdPartyEdge.Node.Name, "the org third party is seeded from the catalog name")

	// The linked tracker pattern is backfilled to the imported org vendor.
	const patternQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on TrackerPattern {
					thirdParty {
						id
					}
				}
			}
		}
	`

	var patternResult struct {
		Node struct {
			ThirdParty *struct {
				ID string `json:"id"`
			} `json:"thirdParty"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(patternQuery, map[string]any{"id": patternID}, &patternResult))
	require.NotNil(t, patternResult.Node.ThirdParty, "the tracker pattern must be linked to the imported org third party")
	assert.Equal(t, importedID, patternResult.Node.ThirdParty.ID)

	// Re-importing the same catalog vendor is idempotent: it returns the
	// same row and reports that nothing was created.
	var second struct {
		ImportThirdPartyFromCommon struct {
			Created        bool `json:"created"`
			ThirdPartyEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyEdge"`
		} `json:"importThirdPartyFromCommon"`
	}

	require.NoError(t, owner.Execute(mutation, map[string]any{"input": input}, &second))
	assert.False(t, second.ImportThirdPartyFromCommon.Created, "re-import must not create a duplicate")
	assert.Equal(t, importedID, second.ImportThirdPartyFromCommon.ThirdPartyEdge.Node.ID, "re-import must return the existing org third party")
}

// linkCommonTrackerPatternToVendor attaches a catalog tracker pattern to
// a common third party, the link the import action follows to resolve
// which org vendor a pattern belongs to.
func linkCommonTrackerPatternToVendor(t *testing.T, commonPatternID gid.GID, commonThirdPartyID gid.GID) {
	t.Helper()

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	_, err := conn.Exec(
		ctx,
		`UPDATE common_tracker_patterns SET common_third_party_id = $1 WHERE id = $2`,
		commonThirdPartyID,
		commonPatternID,
	)
	require.NoError(t, err)
}
