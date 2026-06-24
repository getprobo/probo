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

func TestResourceAlias_SetAndRead(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID := factory.NewDocument(owner).WithTitle("Alias Test Document").Create()

	const setAliasMutation = `
		mutation($input: SetResourceAliasInput!) {
			setResourceAlias(input: $input) {
				resourceAlias {
					resourceId
					alias
				}
			}
		}
	`

	var setResult struct {
		SetResourceAlias struct {
			ResourceAlias struct {
				ResourceID string `json:"resourceId"`
				Alias      string `json:"alias"`
			} `json:"resourceAlias"`
		} `json:"setResourceAlias"`
	}

	err := owner.Execute(setAliasMutation, map[string]any{
		"input": map[string]any{
			"resourceId": documentID,
			"alias":      "privacy-policy",
		},
	}, &setResult)
	require.NoError(t, err)
	assert.Equal(t, documentID, setResult.SetResourceAlias.ResourceAlias.ResourceID)
	assert.Equal(t, "privacy-policy", setResult.SetResourceAlias.ResourceAlias.Alias)

	const getDocumentQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					id
					alias
				}
			}
		}
	`

	var getResult struct {
		Node struct {
			ID    string  `json:"id"`
			Alias *string `json:"alias"`
		} `json:"node"`
	}

	err = owner.Execute(getDocumentQuery, map[string]any{"id": documentID}, &getResult)
	require.NoError(t, err)
	require.NotNil(t, getResult.Node.Alias)
	assert.Equal(t, "privacy-policy", *getResult.Node.Alias)
}

func TestResourceAlias_Conflict(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	docA := factory.NewDocument(owner).WithTitle("Alias Conflict A").Create()
	docB := factory.NewDocument(owner).WithTitle("Alias Conflict B").Create()

	const setAliasMutation = `
		mutation($input: SetResourceAliasInput!) {
			setResourceAlias(input: $input) {
				resourceAlias { alias }
			}
		}
	`

	err := owner.Execute(setAliasMutation, map[string]any{
		"input": map[string]any{
			"resourceId": docA,
			"alias":      "shared-alias",
		},
	}, nil)
	require.NoError(t, err)

	_, err = owner.Do(setAliasMutation, map[string]any{
		"input": map[string]any{
			"resourceId": docB,
			"alias":      "shared-alias",
		},
	})
	require.Error(t, err)
}
