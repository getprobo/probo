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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const compliancePortalCMSPDF = "%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF"

const (
	compliancePortalCMSReferencesListQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					references(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`

	compliancePortalCMSReferencesCountQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					references(first: 10) { totalCount }
				}
			}
		}
	`

	compliancePortalCMSCommitmentGroupsQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					commitmentGroups(first: 10) {
						totalCount
						edges {
							node {
								id
								commitments(first: 10) {
									totalCount
									edges { node { id } }
								}
							}
						}
					}
				}
			}
		}
	`

	compliancePortalCMSCustomLinksQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					customLinks(first: 10) {
						edges { node { id } }
					}
				}
			}
		}
	`

	compliancePortalCMSPortalFilesQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					compliancePortalFiles(first: 10) {
						totalCount
						edges { node { id } }
					}
				}
			}
		}
	`

	compliancePortalCMSOverviewQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					entityName
					description
					websiteUrl
					publicUrl
					slug
					organization { id }
					defaultDomain { id domain }
					accesses(first: 5) {
						edges {
							node {
								pendingRequestCount
								activeCount
								profile { id }
							}
						}
					}
					canUpdate: permission(action: "compliance-portal:portal:update")
					canList: permission(action: "compliance-portal:portal:list")
				}
			}
		}
	`
)

type (
	cmsReferenceWireNode struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		WebsiteURL string `json:"websiteUrl"`
		Logo       struct {
			ID          string `json:"id"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"logo"`
		CanUpdate bool `json:"canUpdate"`
		CanDelete bool `json:"canDelete"`
	}

	cmsCommitmentGroupWireNode struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		CanUpdate   bool   `json:"canUpdate"`
		CanDelete   bool   `json:"canDelete"`
	}

	cmsCommitmentWireNode struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Icon      string `json:"icon"`
		CanUpdate bool   `json:"canUpdate"`
		CanDelete bool   `json:"canDelete"`
	}

	cmsCustomLinkWireNode struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		URL       string `json:"url"`
		CanUpdate bool   `json:"canUpdate"`
		CanDelete bool   `json:"canDelete"`
	}

	cmsPortalFileWireNode struct {
		ID                         string `json:"id"`
		Name                       string `json:"name"`
		Category                   string `json:"category"`
		CompliancePortalVisibility string `json:"compliancePortalVisibility"`
		File                       struct {
			FileName    string `json:"fileName"`
			MimeType    string `json:"mimeType"`
			DownloadURL string `json:"downloadUrl"`
		} `json:"file"`
		CanUpdate bool `json:"canUpdate"`
		CanDelete bool `json:"canDelete"`
	}

	cmsPortalOverviewWire struct {
		EntityName   string `json:"entityName"`
		Description  string `json:"description"`
		WebsiteURL   string `json:"websiteUrl"`
		PublicURL    string `json:"publicUrl"`
		Slug         string `json:"slug"`
		Organization struct {
			ID string `json:"id"`
		} `json:"organization"`
		DefaultDomain *struct {
			ID     string `json:"id"`
			Domain string `json:"domain"`
		} `json:"defaultDomain"`
		Accesses struct {
			Edges []struct {
				Node struct {
					PendingRequestCount int `json:"pendingRequestCount"`
					ActiveCount         int `json:"activeCount"`
					Profile             *struct {
						ID string `json:"id"`
					} `json:"profile"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"accesses"`
		CanUpdate bool `json:"canUpdate"`
		CanList   bool `json:"canList"`
	}
)

func compliancePortalCMSPNGBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}
}

func compliancePortalCMSPNGUpload(filename string) testutil.UploadFile {
	return testutil.UploadFile{
		Filename:    filename,
		ContentType: "image/png",
		Content:     compliancePortalCMSPNGBytes(),
	}
}

func compliancePortalCMSPDFUpload(filename string) testutil.UploadFile {
	return testutil.UploadFile{
		Filename:    filename,
		ContentType: "application/pdf",
		Content:     []byte(compliancePortalCMSPDF),
	}
}

func queryCompliancePortalCMSReferenceCount(
	t *testing.T,
	client *testutil.Client,
	portalID string,
) int {
	t.Helper()

	var result struct {
		Node struct {
			References struct {
				TotalCount int `json:"totalCount"`
			} `json:"references"`
		} `json:"node"`
	}

	err := client.Execute(
		compliancePortalCMSReferencesCountQuery,
		map[string]any{"id": portalID},
		&result,
	)
	require.NoError(t, err)

	return result.Node.References.TotalCount
}

func queryCompliancePortalCMSReferences(
	t *testing.T,
	client *testutil.Client,
	portalID string,
) (totalCount int, firstID string) {
	t.Helper()

	var result struct {
		Node struct {
			References struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"references"`
		} `json:"node"`
	}

	err := client.Execute(
		compliancePortalCMSReferencesListQuery,
		map[string]any{"id": portalID},
		&result,
	)
	require.NoError(t, err)

	if len(result.Node.References.Edges) > 0 {
		firstID = result.Node.References.Edges[0].Node.ID
	}

	return result.Node.References.TotalCount, firstID
}

func createCompliancePortalCMSReference(
	t *testing.T,
	client *testutil.Client,
	portalID string,
	name string,
	websiteURL string,
) cmsReferenceWireNode {
	t.Helper()

	var result struct {
		CreateCompliancePortalReference struct {
			CompliancePortalReferenceEdge struct {
				Node cmsReferenceWireNode `json:"node"`
			} `json:"compliancePortalReferenceEdge"`
		} `json:"createCompliancePortalReference"`
	}

	err := client.ExecuteWithFile(
		`
			mutation($input: CreateCompliancePortalReferenceInput!) {
				createCompliancePortalReference(input: $input) {
					compliancePortalReferenceEdge {
						node {
							id name websiteUrl
							logo { id downloadUrl }
							canUpdate: permission(action: "compliance-portal:portal-reference:update")
							canDelete: permission(action: "compliance-portal:portal-reference:delete")
						}
					}
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"name":               name,
				"description":        "Initial reference",
				"websiteUrl":         websiteURL,
				"logoFile":           nil,
			},
		},
		"input.logoFile",
		compliancePortalCMSPNGUpload("reference-logo.png"),
		&result,
	)
	require.NoError(t, err)

	return result.CreateCompliancePortalReference.CompliancePortalReferenceEdge.Node
}

func updateCompliancePortalCMSReferenceName(
	t *testing.T,
	client *testutil.Client,
	referenceID string,
	name string,
) string {
	t.Helper()

	var result struct {
		UpdateCompliancePortalReference struct {
			CompliancePortalReference struct {
				Name string `json:"name"`
			} `json:"compliancePortalReference"`
		} `json:"updateCompliancePortalReference"`
	}

	err := client.Execute(
		`
			mutation($input: UpdateCompliancePortalReferenceInput!) {
				updateCompliancePortalReference(input: $input) {
					compliancePortalReference { name }
				}
			}
		`,
		map[string]any{"input": map[string]any{"id": referenceID, "name": name}},
		&result,
	)
	require.NoError(t, err)

	return result.UpdateCompliancePortalReference.CompliancePortalReference.Name
}

func deleteCompliancePortalCMSReference(
	t *testing.T,
	client *testutil.Client,
	referenceID string,
) string {
	t.Helper()

	var result struct {
		DeleteCompliancePortalReference struct {
			DeletedCompliancePortalReferenceID string `json:"deletedCompliancePortalReferenceId"`
		} `json:"deleteCompliancePortalReference"`
	}

	err := client.Execute(
		`
			mutation($input: DeleteCompliancePortalReferenceInput!) {
				deleteCompliancePortalReference(input: $input) {
					deletedCompliancePortalReferenceId
				}
			}
		`,
		map[string]any{"input": map[string]any{"id": referenceID}},
		&result,
	)
	require.NoError(t, err)

	return result.DeleteCompliancePortalReference.DeletedCompliancePortalReferenceID
}

func tryCreateCompliancePortalCMSReference(
	client *testutil.Client,
	portalID string,
	name string,
	websiteURL string,
) error {
	var result struct {
		CreateCompliancePortalReference struct {
			CompliancePortalReferenceEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"compliancePortalReferenceEdge"`
		} `json:"createCompliancePortalReference"`
	}

	return client.ExecuteWithFile(
		`
			mutation($input: CreateCompliancePortalReferenceInput!) {
				createCompliancePortalReference(input: $input) {
					compliancePortalReferenceEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"name":               name,
				"websiteUrl":         websiteURL,
				"logoFile":           nil,
			},
		},
		"input.logoFile",
		compliancePortalCMSPNGUpload("reference-logo.png"),
		&result,
	)
}

func createCompliancePortalCMSCommitmentGroup(
	t *testing.T,
	client *testutil.Client,
	portalID string,
	title string,
	description string,
) cmsCommitmentGroupWireNode {
	t.Helper()

	var result struct {
		CreateCompliancePortalCommitmentGroup struct {
			CompliancePortalCommitmentGroupEdge struct {
				Node cmsCommitmentGroupWireNode `json:"node"`
			} `json:"compliancePortalCommitmentGroupEdge"`
		} `json:"createCompliancePortalCommitmentGroup"`
	}

	err := client.Execute(
		`
			mutation($input: CreateCompliancePortalCommitmentGroupInput!) {
				createCompliancePortalCommitmentGroup(input: $input) {
					compliancePortalCommitmentGroupEdge {
						node {
							id title description
							canUpdate: permission(action: "compliance-portal:commitment-group:update")
							canDelete: permission(action: "compliance-portal:commitment-group:delete")
						}
					}
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"title":              title,
				"description":        description,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.CreateCompliancePortalCommitmentGroup.CompliancePortalCommitmentGroupEdge.Node
}

func createCompliancePortalCMSCommitment(
	t *testing.T,
	client *testutil.Client,
	groupID string,
	title string,
) cmsCommitmentWireNode {
	t.Helper()

	var result struct {
		CreateCompliancePortalCommitment struct {
			CompliancePortalCommitmentEdge struct {
				Node cmsCommitmentWireNode `json:"node"`
			} `json:"compliancePortalCommitmentEdge"`
		} `json:"createCompliancePortalCommitment"`
	}

	err := client.Execute(
		`
			mutation($input: CreateCompliancePortalCommitmentInput!) {
				createCompliancePortalCommitment(input: $input) {
					compliancePortalCommitmentEdge {
						node {
							id title icon
							canUpdate: permission(action: "compliance-portal:commitment:update")
							canDelete: permission(action: "compliance-portal:commitment:delete")
						}
					}
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"groupId":     groupID,
				"icon":        "SHIELD_CHECK",
				"eyebrow":     "Security",
				"title":       title,
				"description": "We encrypt data at rest.",
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.CreateCompliancePortalCommitment.CompliancePortalCommitmentEdge.Node
}

func queryCompliancePortalCMSCommitmentGroups(
	t *testing.T,
	client *testutil.Client,
	portalID string,
) (groupCount int, groupID string, commitmentCount int, commitmentID string) {
	t.Helper()

	var result struct {
		Node struct {
			CommitmentGroups struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID          string `json:"id"`
						Commitments struct {
							TotalCount int `json:"totalCount"`
							Edges      []struct {
								Node struct {
									ID string `json:"id"`
								} `json:"node"`
							} `json:"edges"`
						} `json:"commitments"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"commitmentGroups"`
		} `json:"node"`
	}

	err := client.Execute(
		compliancePortalCMSCommitmentGroupsQuery,
		map[string]any{"id": portalID},
		&result,
	)
	require.NoError(t, err)

	groupCount = result.Node.CommitmentGroups.TotalCount
	if len(result.Node.CommitmentGroups.Edges) > 0 {
		groupEdge := result.Node.CommitmentGroups.Edges[0].Node
		groupID = groupEdge.ID

		commitmentCount = groupEdge.Commitments.TotalCount
		if len(groupEdge.Commitments.Edges) > 0 {
			commitmentID = groupEdge.Commitments.Edges[0].Node.ID
		}
	}

	return groupCount, groupID, commitmentCount, commitmentID
}

func tryCreateCompliancePortalCMSCommitmentGroup(
	client *testutil.Client,
	portalID string,
	title string,
) error {
	return client.Execute(
		`
			mutation($input: CreateCompliancePortalCommitmentGroupInput!) {
				createCompliancePortalCommitmentGroup(input: $input) {
					compliancePortalCommitmentGroupEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"title":              title,
				"description":        "forbidden",
			},
		},
		nil,
	)
}

func createCompliancePortalCMSCustomLink(
	t *testing.T,
	client *testutil.Client,
	portalID string,
	name string,
	url string,
) cmsCustomLinkWireNode {
	t.Helper()

	var result struct {
		CreateComplianceCustomLink struct {
			ComplianceCustomLinkEdge struct {
				Node cmsCustomLinkWireNode `json:"node"`
			} `json:"complianceCustomLinkEdge"`
		} `json:"createComplianceCustomLink"`
	}

	err := client.Execute(
		`
			mutation($input: CreateComplianceCustomLinkInput!) {
				createComplianceCustomLink(input: $input) {
					complianceCustomLinkEdge {
						node {
							id name url
							canUpdate: permission(action: "compliance-portal:compliance-custom-link:update")
							canDelete: permission(action: "compliance-portal:compliance-custom-link:delete")
						}
					}
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"name":               name,
				"url":                url,
			},
		},
		&result,
	)
	require.NoError(t, err)

	return result.CreateComplianceCustomLink.ComplianceCustomLinkEdge.Node
}

func queryCompliancePortalCMSCustomLinks(
	t *testing.T,
	client *testutil.Client,
	portalID string,
) (count int, firstID string) {
	t.Helper()

	var result struct {
		Node struct {
			CustomLinks struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"customLinks"`
		} `json:"node"`
	}

	err := client.Execute(
		compliancePortalCMSCustomLinksQuery,
		map[string]any{"id": portalID},
		&result,
	)
	require.NoError(t, err)

	if len(result.Node.CustomLinks.Edges) > 0 {
		firstID = result.Node.CustomLinks.Edges[0].Node.ID
	}

	return len(result.Node.CustomLinks.Edges), firstID
}

func tryCreateCompliancePortalCMSCustomLink(
	client *testutil.Client,
	portalID string,
	name string,
	url string,
) error {
	return client.Execute(
		`
			mutation($input: CreateComplianceCustomLinkInput!) {
				createComplianceCustomLink(input: $input) {
					complianceCustomLinkEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"name":               name,
				"url":                url,
			},
		},
		nil,
	)
}

func createCompliancePortalCMSPortalFile(
	t *testing.T,
	client *testutil.Client,
	portalID string,
	fileName string,
) cmsPortalFileWireNode {
	t.Helper()

	var result struct {
		CreateCompliancePortalFile struct {
			CompliancePortalFileEdge struct {
				Node cmsPortalFileWireNode `json:"node"`
			} `json:"compliancePortalFileEdge"`
		} `json:"createCompliancePortalFile"`
	}

	err := client.ExecuteWithFile(
		`
			mutation($input: CreateCompliancePortalFileInput!) {
				createCompliancePortalFile(input: $input) {
					compliancePortalFileEdge {
						node {
							id name category compliancePortalVisibility
							file { fileName mimeType downloadUrl }
							canUpdate: permission(action: "compliance-portal:portal-file:update")
							canDelete: permission(action: "compliance-portal:portal-file:delete")
						}
					}
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId":         portalID,
				"name":                       fileName,
				"category":                   "Security",
				"file":                       nil,
				"compliancePortalVisibility": "RESTRICTED",
			},
		},
		"input.file",
		compliancePortalCMSPDFUpload(fileName),
		&result,
	)
	require.NoError(t, err)

	return result.CreateCompliancePortalFile.CompliancePortalFileEdge.Node
}

func queryCompliancePortalCMSPortalFiles(
	t *testing.T,
	client *testutil.Client,
	portalID string,
) (totalCount int, firstID string) {
	t.Helper()

	var result struct {
		Node struct {
			CompliancePortalFiles struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"compliancePortalFiles"`
		} `json:"node"`
	}

	err := client.Execute(
		compliancePortalCMSPortalFilesQuery,
		map[string]any{"id": portalID},
		&result,
	)
	require.NoError(t, err)

	if len(result.Node.CompliancePortalFiles.Edges) > 0 {
		firstID = result.Node.CompliancePortalFiles.Edges[0].Node.ID
	}

	return result.Node.CompliancePortalFiles.TotalCount, firstID
}

func tryCreateCompliancePortalCMSPortalFile(
	client *testutil.Client,
	portalID string,
) error {
	var result struct {
		CreateCompliancePortalFile struct {
			CompliancePortalFileEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"compliancePortalFileEdge"`
		} `json:"createCompliancePortalFile"`
	}

	return client.ExecuteWithFile(
		`
			mutation($input: CreateCompliancePortalFileInput!) {
				createCompliancePortalFile(input: $input) {
					compliancePortalFileEdge { node { id } }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId":         portalID,
				"name":                       "rbac.pdf",
				"category":                   "Security",
				"file":                       nil,
				"compliancePortalVisibility": "NONE",
			},
		},
		"input.file",
		compliancePortalCMSPDFUpload("rbac.pdf"),
		&result,
	)
}

func assertCompliancePortalCMSOverview(
	t *testing.T,
	owner *testutil.Client,
	portalID string,
	orgID string,
	profileName string,
) cmsPortalOverviewWire {
	t.Helper()

	err := owner.Execute(
		`
			mutation($input: UpdateCompliancePortalInput!) {
				updateCompliancePortal(input: $input) {
					compliancePortal { id }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId": portalID,
				"entityName":         profileName,
				"description":        "CMS overview description",
				"websiteUrl":         "https://example.com/cms",
			},
		},
		nil,
	)
	require.NoError(t, err)

	var result struct {
		Node cmsPortalOverviewWire `json:"node"`
	}

	ok := testutil.Poll(
		t,
		30*time.Second,
		500*time.Millisecond,
		func() bool {
			err = owner.Execute(
				compliancePortalCMSOverviewQuery,
				map[string]any{"id": portalID},
				&result,
			)

			return err == nil && result.Node.DefaultDomain != nil
		},
	)
	require.True(t, ok, "default managed domain should be provisioned")
	require.NoError(t, err)

	node := result.Node
	assert.Equal(t, profileName, node.EntityName)
	assert.Equal(t, "CMS overview description", node.Description)
	assert.Equal(t, "https://example.com/cms", node.WebsiteURL)
	assert.NotEmpty(t, node.PublicURL)
	assert.NotEmpty(t, node.Slug)
	assert.Equal(t, orgID, node.Organization.ID)
	require.NotNil(t, node.DefaultDomain)
	assert.NotEmpty(t, node.DefaultDomain.ID)
	assert.NotEmpty(t, node.DefaultDomain.Domain)
	assert.True(t, node.CanUpdate)
	assert.True(t, node.CanList)
	assert.Empty(t, node.Accesses.Edges)

	return node
}
