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

func TestThirdParty_ImportFromCommon(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	commonName := factory.SafeName("ImportTP")
	commonThirdPartyID := seedCommonThirdParty(t, commonName)

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
