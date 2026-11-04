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
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const exportDocumentPDFMutation = `
	mutation($input: ExportDocumentPDFInput!) {
		exportDocumentPDF(input: $input) {
			data
		}
	}
`

const viewerNDADispositionQuery = `
	query {
		currentCompliancePortal {
			nonDisclosureAgreement {
				viewerDisposition
			}
		}
	}
`

// TestCompliancePortal_NDAExceptionSkipsSignatureGate covers both ways a
// visitor is excused from the portal NDA: a per-access bypass, and a PDF
// recorded from another system. Either one lets a restricted export pass the
// signature gate. Clearing both restores NDA_SIGNATURE_REQUIRED. A recorded
// file wins over bypass for viewerDisposition.
func TestCompliancePortal_NDAExceptionSkipsSignatureGate(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := lookupCompliancePortalID(t, owner)
	uploadCompliancePortalNDA(t, owner, compliancePortalID)
	trustHost := lookupTrustHost(t, owner, compliancePortalID)
	// The certificate poll lease is 2 minutes, so the dedicated listener is
	// not servable within the 30s wait inside self-provision.
	testutil.WaitForCompliancePortalHTTPSWithin(t, trustHost, 3*time.Minute)

	visitor := testutil.SelfProvisionCompliancePortalVisitor(t, trustHost)
	updateVisitorFullName(t, visitor, trustHost, "Excused Visitor")

	documentID := factory.NewDocument(owner).
		WithTitle(factory.SafeName("Restricted NDA gate")).
		Create()
	publishRestrictedPortalDocument(t, owner, compliancePortalID, documentID)

	err := visitor.ExecuteTrust(trustHost, exportDocumentPDFMutation, map[string]any{
		"input": map[string]any{"documentId": documentID},
	}, nil)
	require.Error(t, err)
	assertTrustGraphQLCode(t, err, "NDA_SIGNATURE_REQUIRED")
	assert.Equal(t, "SIGN", viewerNDADisposition(t, visitor, trustHost))

	accessID := findCompliancePortalAccessID(t, owner, compliancePortalID, visitor.GetEmail())

	setAccessNDABypass(t, owner, accessID, true)
	assert.Equal(t, "BYPASSED", viewerNDADisposition(t, visitor, trustHost))

	err = visitor.ExecuteTrust(trustHost, exportDocumentPDFMutation, map[string]any{
		"input": map[string]any{"documentId": documentID},
	}, nil)
	require.Error(t, err, "bypass skips the NDA, then the missing grant is forbidden")
	assertTrustGraphQLCode(t, err, "FORBIDDEN")

	setAccessNDABypass(t, owner, accessID, false)
	assert.Equal(t, "SIGN", viewerNDADisposition(t, visitor, trustHost))

	err = visitor.ExecuteTrust(trustHost, exportDocumentPDFMutation, map[string]any{
		"input": map[string]any{"documentId": documentID},
	}, nil)
	require.Error(t, err)
	assertTrustGraphQLCode(t, err, "NDA_SIGNATURE_REQUIRED")

	uploadRecordedNDA(t, owner, accessID)
	assert.Equal(t, "RECORDED", viewerNDADisposition(t, visitor, trustHost))

	setAccessNDABypass(t, owner, accessID, true)
	assert.Equal(t, "RECORDED", viewerNDADisposition(t, visitor, trustHost), "a recorded file wins over bypass")

	err = visitor.ExecuteTrust(trustHost, exportDocumentPDFMutation, map[string]any{
		"input": map[string]any{"documentId": documentID},
	}, nil)
	require.Error(t, err)
	assertTrustGraphQLCode(t, err, "FORBIDDEN")

	setAccessNDABypass(t, owner, accessID, false)
	deleteRecordedNDA(t, owner, accessID)
	assert.Equal(t, "SIGN", viewerNDADisposition(t, visitor, trustHost))

	err = visitor.ExecuteTrust(trustHost, exportDocumentPDFMutation, map[string]any{
		"input": map[string]any{"documentId": documentID},
	}, nil)
	require.Error(t, err)
	assertTrustGraphQLCode(t, err, "NDA_SIGNATURE_REQUIRED")
}

func publishRestrictedPortalDocument(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	documentID string,
) {
	t.Helper()

	err := owner.Execute(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				documentVersion { status }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":      true,
			"documentId": documentID,
			"changelog":  "Publish for the NDA gate",
		},
	}, nil)
	require.NoError(t, err)

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
			"compliancePortalVisibility": "RESTRICTED",
		},
	}, nil)
	require.NoError(t, err)
}

func viewerNDADisposition(t *testing.T, visitor *testutil.Client, trustHost string) string {
	t.Helper()

	var result struct {
		CurrentCompliancePortal struct {
			NonDisclosureAgreement struct {
				ViewerDisposition string `json:"viewerDisposition"`
			} `json:"nonDisclosureAgreement"`
		} `json:"currentCompliancePortal"`
	}

	err := visitor.ExecuteTrust(trustHost, viewerNDADispositionQuery, nil, &result)
	require.NoError(t, err)

	return result.CurrentCompliancePortal.NonDisclosureAgreement.ViewerDisposition
}

func findCompliancePortalAccessID(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	email string,
) string {
	t.Helper()

	const query = `
		query($id: ID!, $query: String!) {
			node(id: $id) {
				... on CompliancePortal {
					accesses(first: 5, filter: {query: $query}) {
						edges {
							node {
								id
								identity { email }
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Accesses struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						Identity struct {
							Email string `json:"email"`
						} `json:"identity"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"accesses"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id":    compliancePortalID,
		"query": email,
	}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.Accesses.Edges)

	for _, edge := range result.Node.Accesses.Edges {
		if edge.Node.Identity.Email == email {
			return edge.Node.ID
		}
	}

	t.Fatalf("compliance portal access for %s not found", email)

	return ""
}

func setAccessNDABypass(t *testing.T, owner *testutil.Client, accessID string, bypassed bool) {
	t.Helper()

	err := owner.Execute(`
		mutation($input: SetCompliancePortalAccessNDABypassInput!) {
			setCompliancePortalAccessNDABypass(input: $input) {
				compliancePortalAccess { ndaBypassed }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":       accessID,
			"bypassed": bypassed,
		},
	}, nil)
	require.NoError(t, err)
}

func uploadRecordedNDA(t *testing.T, owner *testutil.Client, accessID string) {
	t.Helper()

	pdfContent, err := base64.StdEncoding.DecodeString(minimalPDFBase64)
	require.NoError(t, err)

	err = owner.ExecuteWithFile(`
		mutation($input: UploadCompliancePortalAccessRecordedNDAInput!) {
			uploadCompliancePortalAccessRecordedNDA(input: $input) {
				compliancePortalAccess { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"compliancePortalAccessId": accessID,
			"fileName":                 "recorded-nda.pdf",
			"file":                     nil,
			"signedAt":                 time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339),
		},
	}, "input.file", testutil.UploadFile{
		Filename:    "recorded-nda.pdf",
		ContentType: "application/pdf",
		Content:     pdfContent,
	}, nil)
	require.NoError(t, err)
}

func deleteRecordedNDA(t *testing.T, owner *testutil.Client, accessID string) {
	t.Helper()

	err := owner.Execute(`
		mutation($input: DeleteCompliancePortalAccessRecordedNDAInput!) {
			deleteCompliancePortalAccessRecordedNDA(input: $input) {
				compliancePortalAccess { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"compliancePortalAccessId": accessID,
		},
	}, nil)
	require.NoError(t, err)
}

func assertTrustGraphQLCode(t *testing.T, err error, code string) {
	t.Helper()

	gqlErrs, ok := err.(testutil.GraphQLErrors)
	require.True(t, ok, "expected a GraphQL error, got %T: %v", err, err)
	require.NotEmpty(t, gqlErrs)
	assert.Equal(t, code, gqlErrs[0].Code())
}
