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

func TestNode_ProfileScopedToOrganization(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	const visibleName = "Org1 Visible Member"
	profileID := factory.CreateUser(org1Owner, factory.Attrs{"fullName": visibleName})
	otherProfileID := factory.CreateUser(org2Owner, factory.Attrs{"fullName": "Org2 Hidden Member"})

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Profile {
					id
					fullName
				}
			}
		}
	`

	var sameOrg struct {
		Node struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"node"`
	}

	err := org1Owner.Execute(query, map[string]any{"id": profileID}, &sameOrg)
	require.NoError(t, err)
	assert.Equal(t, profileID, sameOrg.Node.ID)
	assert.Equal(t, visibleName, sameOrg.Node.FullName)

	var otherOrg struct {
		Node *struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"node"`
	}

	err = org1Owner.Execute(query, map[string]any{"id": otherProfileID}, &otherOrg)
	testutil.AssertNodeNotAccessible(t, err, otherOrg.Node == nil, "profile")
}
