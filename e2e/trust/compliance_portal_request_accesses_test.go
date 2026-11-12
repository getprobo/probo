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
// visitor can request access to a specific selection of restricted documents in a
// single mutation, and that each affected row comes back flagged as REQUESTED.
func TestCompliancePortal_RequestAccesses_Batch(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	documentID, compliancePortalID := setupRestrictedPortalDocument(t, owner)
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

	victimDocumentID, _ := setupRestrictedPortalDocument(t, victimOwner)

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

// TestCompliancePortal_ViewerHasRequestedAccess verifies the portal-level flag
// used to defer the unsigned-NDA banner until the visitor requests private
// access: false before any visitor request, true after requestDocumentAccess.
func TestCompliancePortal_ViewerHasRequestedAccess(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	documentID, compliancePortalID := setupRestrictedPortalDocument(t, owner)
	trustHost := lookupTrustHost(t, owner, compliancePortalID)

	visitor := testutil.SelfProvisionCompliancePortalVisitor(t, trustHost)

	assert.False(
		t,
		queryViewerHasRequestedAccess(t, visitor, trustHost),
		"a newly signed-in visitor must not appear to have requested access yet",
	)

	const mutation = `
		mutation RequestDocumentAccess($input: RequestDocumentAccessInput!) {
			requestDocumentAccess(input: $input) {
				document { id }
				compliancePortal {
					id
					viewerHasRequestedAccess
				}
			}
		}
	`

	var result struct {
		RequestDocumentAccess struct {
			Document struct {
				ID string `json:"id"`
			} `json:"document"`
			CompliancePortal struct {
				ViewerHasRequestedAccess bool `json:"viewerHasRequestedAccess"`
			} `json:"compliancePortal"`
		} `json:"requestDocumentAccess"`
	}

	err := visitor.ExecuteTrust(trustHost, mutation, map[string]any{
		"input": map[string]any{
			"documentId": documentID,
		},
	}, &result)
	require.NoError(t, err, "an authenticated visitor must be able to request document access")

	assert.True(
		t,
		result.RequestDocumentAccess.CompliancePortal.ViewerHasRequestedAccess,
		"the mutation payload must flip viewerHasRequestedAccess so clients can update without a refetch",
	)
	assert.True(
		t,
		queryViewerHasRequestedAccess(t, visitor, trustHost),
		"viewerHasRequestedAccess must remain true on a subsequent portal query",
	)
}

func queryViewerHasRequestedAccess(t *testing.T, visitor *testutil.Client, trustHost string) bool {
	t.Helper()

	const query = `
		query {
			currentCompliancePortal {
				viewerHasRequestedAccess
			}
		}
	`

	var result struct {
		CurrentCompliancePortal struct {
			ViewerHasRequestedAccess bool `json:"viewerHasRequestedAccess"`
		} `json:"currentCompliancePortal"`
	}

	err := visitor.ExecuteTrust(trustHost, query, nil, &result)
	require.NoError(t, err)

	return result.CurrentCompliancePortal.ViewerHasRequestedAccess
}

// TestCompliancePortal_RequestAccesses_EmptyRejects verifies that requestAccesses
// with no document, report, or file ids is rejected — there is no "request all"
// shortcut; callers must always name the targets explicitly.
func TestCompliancePortal_RequestAccesses_EmptyRejects(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	compliancePortalID := lookupCompliancePortalID(t, owner)
	trustHost := lookupTrustHost(t, owner, compliancePortalID)

	visitor := testutil.SelfProvisionCompliancePortalVisitor(t, trustHost)

	err := visitor.ExecuteTrust(trustHost, requestAccessesMutation, map[string]any{
		"input": map[string]any{
			"documentIds":             []string{},
			"reportIds":               []string{},
			"compliancePortalFileIds": []string{},
		},
	}, nil)
	require.Error(t, err, "requestAccesses with empty id lists must be rejected")
	assert.Contains(
		t,
		err.Error(),
		"at least one document, report, or file id is required",
		"empty request must surface a client validation error",
	)
}

// TestCompliancePortal_RequestAccesses_ExistingRow verifies that requesting an
// item that already has an access row succeeds. The second call reactivates
// the existing row instead of inserting a new one.
func TestCompliancePortal_RequestAccesses_ExistingRow(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"document", "report", "file"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			target := setupRestrictedAccessTarget(t, owner, kind)
			visitor := testutil.SelfProvisionCompliancePortalVisitor(t, target.trustHost)

			first := requestRestrictedAccess(t, visitor, target)
			assert.Equal(t, "REQUESTED", first.Status)

			second := requestRestrictedAccess(t, visitor, target)
			assert.Equal(t, "REQUESTED", second.Status)
			assert.Equal(t, first.ID, second.ID, "the second request must reuse the existing access row")
		})
	}
}

// TestCompliancePortal_RequestAccesses_RetryAfterRejection verifies that a
// visitor can request an item again after the owner rejects it, and that the
// same access row returns to REQUESTED.
func TestCompliancePortal_RequestAccesses_RetryAfterRejection(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"document", "report", "file"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			owner := testutil.NewClient(t, testutil.RoleOwner)
			target := setupRestrictedAccessTarget(t, owner, kind)
			visitor := testutil.SelfProvisionCompliancePortalVisitor(t, target.trustHost)

			first := requestRestrictedAccess(t, visitor, target)
			assert.Equal(t, "REQUESTED", first.Status)

			accessID := lookupVisitorAccessID(t, owner, target.compliancePortalID, visitor.GetEmail())
			rejectRestrictedAccess(t, owner, accessID, target)
			assert.Equal(t, "REJECTED", visitorAccessStatus(t, visitor, target))

			retried := requestRestrictedAccess(t, visitor, target)
			assert.Equal(t, "REQUESTED", retried.Status)
			assert.Equal(t, first.ID, retried.ID, "the retry must reactivate the rejected access row")
		})
	}
}

// setupRestrictedPortalDocument creates a published document with restricted
// visibility on the owner's compliance portal.
func setupRestrictedPortalDocument(t *testing.T, owner *testutil.Client) (documentID, compliancePortalID string) {
	t.Helper()

	compliancePortalID = lookupCompliancePortalID(t, owner)
	documentID = factory.NewDocument(owner).WithTitle(factory.SafeName("Document")).Create()

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
			"changelog":  "Publish for portal catalog",
		},
	}, nil)
	require.NoError(t, err)

	const updateMutation = `
		mutation UpdateCompliancePortalDocumentVisibility($input: UpdateCompliancePortalDocumentVisibilityInput!) {
			updateCompliancePortalDocumentVisibility(input: $input) {
				catalogDocument {
					visibility
					document { id }
				}
			}
		}
	`

	err = owner.Execute(updateMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId":         compliancePortalID,
			"documentId":                 documentID,
			"compliancePortalVisibility": "RESTRICTED",
		},
	}, nil)
	require.NoError(t, err)

	return documentID, compliancePortalID
}

const requestAccessStatusMutation = `
	mutation RequestAccesses($input: RequestAccessesInput!) {
		requestAccesses(input: $input) {
			documents {
				id
				access { id status }
			}
			audits {
				reportFile {
					id
					access { id status }
				}
			}
			files {
				id
				access { id status }
			}
		}
	}
`

type (
	requestedAccessStatus struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	requestAccessStatusResult struct {
		RequestAccesses struct {
			Documents []struct {
				ID     string                 `json:"id"`
				Access *requestedAccessStatus `json:"access"`
			} `json:"documents"`
			Audits []struct {
				ReportFile struct {
					ID     string                 `json:"id"`
					Access *requestedAccessStatus `json:"access"`
				} `json:"reportFile"`
			} `json:"audits"`
			Files []struct {
				ID     string                 `json:"id"`
				Access *requestedAccessStatus `json:"access"`
			} `json:"files"`
		} `json:"requestAccesses"`
	}

	restrictedAccessTarget struct {
		kind               string
		targetID           string
		compliancePortalID string
		trustHost          string
	}
)

func setupRestrictedAccessTarget(
	t *testing.T,
	owner *testutil.Client,
	kind string,
) restrictedAccessTarget {
	t.Helper()

	var (
		targetID           string
		compliancePortalID string
	)

	switch kind {
	case "document":
		targetID, compliancePortalID = setupRestrictedPortalDocument(t, owner)
	case "report":
		targetID, compliancePortalID = setupRestrictedPortalReport(t, owner)
	case "file":
		compliancePortalID = lookupCompliancePortalID(t, owner)
		targetID = setupRestrictedPortalFile(t, owner, compliancePortalID)
	default:
		t.Fatalf("unknown access target %q", kind)
	}

	return restrictedAccessTarget{
		kind:               kind,
		targetID:           targetID,
		compliancePortalID: compliancePortalID,
		trustHost:          lookupTrustHost(t, owner, compliancePortalID),
	}
}

func requestRestrictedAccess(
	t *testing.T,
	visitor *testutil.Client,
	target restrictedAccessTarget,
) requestedAccessStatus {
	t.Helper()

	input := map[string]any{
		"documentIds":             []string{},
		"reportIds":               []string{},
		"compliancePortalFileIds": []string{},
	}

	switch target.kind {
	case "document":
		input["documentIds"] = []string{target.targetID}
	case "report":
		input["reportIds"] = []string{target.targetID}
	case "file":
		input["compliancePortalFileIds"] = []string{target.targetID}
	}

	var result requestAccessStatusResult

	err := visitor.ExecuteTrust(
		target.trustHost,
		requestAccessStatusMutation,
		map[string]any{"input": input},
		&result,
	)
	require.NoError(t, err, "visitor must be able to request restricted %s access", target.kind)

	var access *requestedAccessStatus

	switch target.kind {
	case "document":
		require.Len(t, result.RequestAccesses.Documents, 1)
		access = result.RequestAccesses.Documents[0].Access
	case "report":
		require.Len(t, result.RequestAccesses.Audits, 1)
		access = result.RequestAccesses.Audits[0].ReportFile.Access
	case "file":
		require.Len(t, result.RequestAccesses.Files, 1)
		access = result.RequestAccesses.Files[0].Access
	}

	require.NotNil(t, access, "requested %s must carry an access record", target.kind)

	return *access
}

func rejectRestrictedAccess(
	t *testing.T,
	owner *testutil.Client,
	accessID string,
	target restrictedAccessTarget,
) {
	t.Helper()

	input := map[string]any{
		"id": accessID,
	}
	entries := []map[string]any{
		{"id": target.targetID, "status": "REJECTED"},
	}

	switch target.kind {
	case "document":
		input["documents"] = entries
	case "report":
		input["reports"] = entries
	case "file":
		input["compliancePortalFiles"] = entries
	}

	err := owner.Execute(
		`
		mutation($input: UpdateCompliancePortalAccessInput!) {
			updateCompliancePortalAccess(input: $input) {
				compliancePortalAccess { id }
			}
		}
	`,
		map[string]any{"input": input},
		nil,
	)
	require.NoError(t, err, "owner must be able to reject the visitor %s request", target.kind)
}

func visitorAccessStatus(
	t *testing.T,
	visitor *testutil.Client,
	target restrictedAccessTarget,
) string {
	t.Helper()

	var query string

	switch target.kind {
	case "document":
		query = `
			query($id: ID!) {
				node(id: $id) {
					... on Document { access { status } }
				}
			}
		`
	case "report":
		query = `
			query($id: ID!) {
				node(id: $id) {
					... on AuditReport { access { status } }
				}
			}
		`
	case "file":
		query = `
			query($id: ID!) {
				node(id: $id) {
					... on CompliancePortalFile { access { status } }
				}
			}
		`
	}

	var result struct {
		Node struct {
			Access *struct {
				Status string `json:"status"`
			} `json:"access"`
		} `json:"node"`
	}

	err := visitor.ExecuteTrust(
		target.trustHost,
		query,
		map[string]any{"id": target.targetID},
		&result,
	)
	require.NoError(t, err)
	require.NotNil(t, result.Node.Access, "visitor must see an access record for the %s", target.kind)

	return result.Node.Access.Status
}

// setupRestrictedPortalReport creates an audit with an uploaded report and
// restricted portal visibility. The returned id is the report file id.
func setupRestrictedPortalReport(t *testing.T, owner *testutil.Client) (reportFileID, compliancePortalID string) {
	t.Helper()

	frameworkID := factory.NewFramework(owner).WithName(factory.SafeName("Framework")).Create()
	auditID := factory.NewAudit(owner, frameworkID).WithName(factory.SafeName("Audit")).Create()

	var uploadResult struct {
		UploadAuditReport struct {
			Audit struct {
				ReportFile struct {
					ID string `json:"id"`
				} `json:"reportFile"`
			} `json:"audit"`
		} `json:"uploadAuditReport"`
	}

	err := owner.ExecuteWithFile(
		`
		mutation UploadAuditReport($input: UploadAuditReportInput!) {
			uploadAuditReport(input: $input) {
				audit {
					reportFile { id }
				}
			}
		}
	`,
		map[string]any{
			"input": map[string]any{
				"auditId": auditID,
				"file":    nil,
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "audit-report.pdf",
			ContentType: "application/pdf",
			Content:     []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF"),
		},
		&uploadResult,
	)
	require.NoError(t, err)

	reportFileID = uploadResult.UploadAuditReport.Audit.ReportFile.ID
	require.NotEmpty(t, reportFileID)

	compliancePortalID = lookupCompliancePortalID(t, owner)

	err = owner.Execute(
		`
		mutation UpdateCompliancePortalAuditVisibility($input: UpdateCompliancePortalAuditVisibilityInput!) {
			updateCompliancePortalAuditVisibility(input: $input) {
				catalogAudit {
					visibility
					audit { id }
				}
			}
		}
	`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId":         compliancePortalID,
				"auditId":                    auditID,
				"compliancePortalVisibility": "RESTRICTED",
			},
		},
		nil,
	)
	require.NoError(t, err)

	return reportFileID, compliancePortalID
}
