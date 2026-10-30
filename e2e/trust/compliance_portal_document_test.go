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

package trust_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const compliancePortalDocumentPageQuery = `
	query DocumentViewerPage($alias: String!) {
		aliasedNode(alias: $alias) {
			__typename
			... on Document {
				id
				title
				documentType
				isUserAuthorized
				access {
					id
					status
				}
			}
		}
	}
`

// TestCompliancePortal_DocumentPageLoadsPublishedSnapshot verifies that the
// compliance portal document page loads a published document and shows the
// published title and type after the working copy has moved on.
func TestCompliancePortal_DocumentPageLoadsPublishedSnapshot(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := lookupCompliancePortalID(t, owner)
	trustHost := lookupTrustHost(t, owner, compliancePortalID)
	// The certificate is not ready within the 30s default.
	testutil.WaitForCompliancePortalHTTPSWithin(t, trustHost, time.Minute)

	publishedTitle := factory.SafeName("Published policy")
	draftTitle := factory.SafeName("Working draft")
	documentID := factory.NewDocument(owner).
		WithTitle(publishedTitle).
		WithDocumentType("POLICY").
		Create()

	err := owner.Execute(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				documentVersion { id status }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":      true,
			"documentId": documentID,
			"changelog":  "Publish for the compliance portal",
		},
	}, nil)
	require.NoError(t, err, "publishing the document must succeed")

	err = owner.Execute(`
		mutation($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":           documentID,
			"title":        draftTitle,
			"documentType": "PROCEDURE",
		},
	}, nil)
	require.NoError(t, err, "rewriting the working copy must succeed")

	var workingCopy struct {
		Node struct {
			Title        string `json:"title"`
			DocumentType string `json:"documentType"`
		} `json:"node"`
	}

	err = owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					title
					documentType
				}
			}
		}
	`, map[string]any{"id": documentID}, &workingCopy)
	require.NoError(t, err)
	assert.Equal(t, draftTitle, workingCopy.Node.Title, "the console keeps the unpublished rewrite")
	assert.Equal(t, "PROCEDURE", workingCopy.Node.DocumentType)

	err = owner.Execute(`
		mutation($input: UpdateCompliancePortalDocumentVisibilityInput!) {
			updateCompliancePortalDocumentVisibility(input: $input) {
				catalogDocument { visibility }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"compliancePortalId":         compliancePortalID,
			"documentId":                 documentID,
			"compliancePortalVisibility": "PUBLIC",
		},
	}, nil)
	require.NoError(t, err, "publishing the document on the compliance portal must succeed")

	var page struct {
		AliasedNode struct {
			Typename         string `json:"__typename"`
			ID               string `json:"id"`
			Title            string `json:"title"`
			DocumentType     string `json:"documentType"`
			IsUserAuthorized bool   `json:"isUserAuthorized"`
			Access           *struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"access"`
		} `json:"aliasedNode"`
	}

	err = owner.ExecuteTrust(trustHost, compliancePortalDocumentPageQuery, map[string]any{
		"alias": documentID,
	}, &page)
	require.NoError(t, err, "the compliance portal document page must load the published document")
	assert.Equal(t, "Document", page.AliasedNode.Typename)
	assert.Equal(t, documentID, page.AliasedNode.ID)
	assert.Equal(t, publishedTitle, page.AliasedNode.Title, "visitors must see the published title")
	assert.Equal(t, "POLICY", page.AliasedNode.DocumentType, "visitors must see the published document type")
	assert.True(t, page.AliasedNode.IsUserAuthorized, "a public document is readable without an access grant")
	assert.Nil(t, page.AliasedNode.Access)
}
