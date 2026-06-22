// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
