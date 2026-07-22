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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const requestAccessesMutation = `
	mutation RequestAccesses($input: RequestAccessesInput!) {
		requestAccesses(input: $input) {
			documents {
				id
				access { status }
			}
			audits {
				reportFile { id }
			}
			files {
				id
			}
		}
	}
`

// requestAccessesResult mirrors the shape selected by requestAccessesMutation.
type requestAccessesResult struct {
	RequestAccesses struct {
		Documents []struct {
			ID     string `json:"id"`
			Access *struct {
				Status string `json:"status"`
			} `json:"access"`
		} `json:"documents"`
		Audits []struct {
			ReportFile struct {
				ID string `json:"id"`
			} `json:"reportFile"`
		} `json:"audits"`
		Files []struct {
			ID string `json:"id"`
		} `json:"files"`
	} `json:"requestAccesses"`
}

// TestCompliancePortal_RequestAccesses_Batch verifies that an authenticated
// visitor can request access to a specific selection of private documents in a
// single mutation, and that each affected row comes back flagged as REQUESTED.
func TestCompliancePortal_RequestAccesses_Batch(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	documentID := setupPrivatePortalDocument(t, owner)
	compliancePortalID := lookupCompliancePortalID(t, owner)
	trustHost := lookupTrustHost(t, owner, compliancePortalID)

	visitor := testutil.SelfProvisionCompliancePortalVisitor(t, trustHost)

	var result requestAccessesResult
	err := visitor.ExecuteTrust(trustHost, requestAccessesMutation, map[string]any{
		"input": map[string]any{
			"documentIds":             []string{documentID},
			"reportIds":               []string{},
			"compliancePortalFileIds": []string{},
		},
	}, &result)
	require.NoError(t, err, "an authenticated visitor must be able to request access to a selection")

	require.Len(t, result.RequestAccesses.Documents, 1, "the payload must echo the requested document")
	assert.Equal(t, documentID, result.RequestAccesses.Documents[0].ID)
	require.NotNil(t, result.RequestAccesses.Documents[0].Access, "the requested document must carry an access record")
	assert.Equal(t, "REQUESTED", result.RequestAccesses.Documents[0].Access.Status)
	assert.Empty(t, result.RequestAccesses.Audits, "no reports were requested")
	assert.Empty(t, result.RequestAccesses.Files, "no files were requested")
}

// TestCompliancePortal_RequestAccesses_TenantIsolation verifies that a visitor
// on one organization's compliance portal cannot request access to another
// organization's document by supplying a foreign document GID: the request is
// rejected before any access row is written.
func TestCompliancePortal_RequestAccesses_TenantIsolation(t *testing.T) {
	t.Parallel()

	victimOwner := testutil.NewClient(t, testutil.RoleOwner)
	attackerOwner := testutil.NewClient(t, testutil.RoleOwner)

	victimDocumentID := setupPrivatePortalDocument(t, victimOwner)

	attackerCompliancePortalID := lookupCompliancePortalID(t, attackerOwner)
	attackerTrustHost := lookupTrustHost(t, attackerOwner, attackerCompliancePortalID)

	attacker := testutil.SelfProvisionCompliancePortalVisitor(t, attackerTrustHost)

	err := attacker.ExecuteTrust(attackerTrustHost, requestAccessesMutation, map[string]any{
		"input": map[string]any{
			"documentIds":             []string{victimDocumentID},
			"reportIds":               []string{},
			"compliancePortalFileIds": []string{},
		},
	}, nil)
	require.Error(t, err, "a foreign compliance portal must not request access to another org's document")
	assert.Contains(
		t,
		err.Error(),
		"not found",
		"cross-tenant document GID must be rejected as not found",
	)
}

// setupPrivatePortalDocument creates a document and marks it privately visible on
// the owner's compliance portal, returning the document ID.
func setupPrivatePortalDocument(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	documentID := factory.NewDocument(owner).WithTitle(factory.SafeName("Document")).Create()

	const updateMutation = `
		mutation UpdateDocument($input: UpdateDocumentInput!) {
			updateDocument(input: $input) {
				document { id }
			}
		}
	`

	err := owner.Execute(updateMutation, map[string]any{
		"input": map[string]any{
			"id":                         documentID,
			"compliancePortalVisibility": "PRIVATE",
		},
	}, nil)
	require.NoError(t, err)

	return documentID
}
