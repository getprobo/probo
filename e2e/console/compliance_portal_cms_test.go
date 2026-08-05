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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCompliancePortal_CMS_OwnerLifecycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()

	portalID := factory.CreateCompliancePortal(
		owner,
		factory.Attrs{"entityName": factory.SafeName("CMS portal")},
	)

	t.Run("references", func(t *testing.T) {
		refName := factory.SafeName("CMS reference")
		node := createCompliancePortalCMSReference(
			t,
			owner,
			portalID,
			refName,
			"https://example.com/cms-reference",
		)
		require.NotEmpty(t, node.ID)
		assert.Equal(t, refName, node.Name)
		assert.NotEmpty(t, node.Logo.ID)
		assert.NotEmpty(t, node.Logo.DownloadURL)
		assert.True(t, node.CanUpdate)
		assert.True(t, node.CanDelete)

		refTotal, refID := queryCompliancePortalCMSReferences(t, owner, portalID)
		assert.Equal(t, 1, refTotal)
		assert.Equal(t, node.ID, refID)

		updatedName := factory.SafeName("CMS reference updated")
		assert.Equal(t, updatedName, updateCompliancePortalCMSReferenceName(t, owner, node.ID, updatedName))
		assert.Equal(t, node.ID, deleteCompliancePortalCMSReference(t, owner, node.ID))
		assert.Equal(t, 0, queryCompliancePortalCMSReferenceCount(t, owner, portalID))
	})

	t.Run("commitment groups and commitments", func(t *testing.T) {
		groupTitle := factory.SafeName("CMS commitment group")
		group := createCompliancePortalCMSCommitmentGroup(
			t,
			owner,
			portalID,
			groupTitle,
			"Group description",
		)
		require.NotEmpty(t, group.ID)
		assert.Equal(t, groupTitle, group.Title)
		assert.True(t, group.CanUpdate)
		assert.True(t, group.CanDelete)

		commitmentTitle := factory.SafeName("CMS commitment")
		commitment := createCompliancePortalCMSCommitment(t, owner, group.ID, commitmentTitle)
		require.NotEmpty(t, commitment.ID)
		assert.Equal(t, commitmentTitle, commitment.Title)
		assert.Equal(t, "SHIELD_CHECK", commitment.Icon)
		assert.True(t, commitment.CanUpdate)
		assert.True(t, commitment.CanDelete)

		groupTotal, listedGroupID, commitmentTotal, listedCommitmentID := queryCompliancePortalCMSCommitmentGroups(
			t,
			owner,
			portalID,
		)
		assert.Equal(t, 1, groupTotal)
		assert.Equal(t, group.ID, listedGroupID)
		assert.Equal(t, 1, commitmentTotal)
		assert.Equal(t, commitment.ID, listedCommitmentID)

		err := owner.Execute(
			`
				mutation($input: UpdateCompliancePortalCommitmentInput!) {
					updateCompliancePortalCommitment(input: $input) {
						compliancePortalCommitment { id }
					}
				}
			`,
			map[string]any{
				"input": map[string]any{
					"id":    commitment.ID,
					"title": factory.SafeName("CMS commitment updated"),
				},
			},
			nil,
		)
		require.NoError(t, err)

		err = owner.Execute(
			`
				mutation($input: UpdateCompliancePortalCommitmentGroupInput!) {
					updateCompliancePortalCommitmentGroup(input: $input) {
						compliancePortalCommitmentGroup { id }
					}
				}
			`,
			map[string]any{
				"input": map[string]any{
					"id":    group.ID,
					"title": factory.SafeName("CMS group updated"),
				},
			},
			nil,
		)
		require.NoError(t, err)

		err = owner.Execute(
			`
				mutation($input: DeleteCompliancePortalCommitmentInput!) {
					deleteCompliancePortalCommitment(input: $input) {
						deletedCompliancePortalCommitmentId
					}
				}
			`,
			map[string]any{"input": map[string]any{"id": commitment.ID}},
			nil,
		)
		require.NoError(t, err)

		var deleteGroupResult struct {
			DeleteCompliancePortalCommitmentGroup struct {
				DeletedCompliancePortalCommitmentGroupID string `json:"deletedCompliancePortalCommitmentGroupId"`
			} `json:"deleteCompliancePortalCommitmentGroup"`
		}

		err = owner.Execute(
			`
				mutation($input: DeleteCompliancePortalCommitmentGroupInput!) {
					deleteCompliancePortalCommitmentGroup(input: $input) {
						deletedCompliancePortalCommitmentGroupId
					}
				}
			`,
			map[string]any{"input": map[string]any{"id": group.ID}},
			&deleteGroupResult,
		)
		require.NoError(t, err)
		assert.Equal(t, group.ID, deleteGroupResult.DeleteCompliancePortalCommitmentGroup.DeletedCompliancePortalCommitmentGroupID)
	})

	t.Run("custom links", func(t *testing.T) {
		linkName := factory.SafeName("CMS custom link")
		link := createCompliancePortalCMSCustomLink(
			t,
			owner,
			portalID,
			linkName,
			"https://example.com/cms-link",
		)
		require.NotEmpty(t, link.ID)
		assert.Equal(t, linkName, link.Name)
		assert.True(t, link.CanUpdate)
		assert.True(t, link.CanDelete)

		linkCount, linkID := queryCompliancePortalCMSCustomLinks(t, owner, portalID)
		assert.Equal(t, 1, linkCount)
		assert.Equal(t, link.ID, linkID)

		err := owner.Execute(
			`
				mutation($input: UpdateComplianceCustomLinkInput!) {
					updateComplianceCustomLink(input: $input) {
						complianceCustomLink { id }
					}
				}
			`,
			map[string]any{
				"input": map[string]any{
					"id":   link.ID,
					"name": factory.SafeName("CMS link updated"),
					"url":  "https://example.com/cms-link-updated",
				},
			},
			nil,
		)
		require.NoError(t, err)

		var deleteResult struct {
			DeleteComplianceCustomLink struct {
				DeletedComplianceCustomLinkID string `json:"deletedComplianceCustomLinkId"`
			} `json:"deleteComplianceCustomLink"`
		}

		err = owner.Execute(
			`
				mutation($input: DeleteComplianceCustomLinkInput!) {
					deleteComplianceCustomLink(input: $input) {
						deletedComplianceCustomLinkId
					}
				}
			`,
			map[string]any{"input": map[string]any{"id": link.ID}},
			&deleteResult,
		)
		require.NoError(t, err)
		assert.Equal(t, link.ID, deleteResult.DeleteComplianceCustomLink.DeletedComplianceCustomLinkID)
	})

	t.Run("portal files", func(t *testing.T) {
		fileName := factory.SafeName("cms-file") + ".pdf"
		file := createCompliancePortalCMSPortalFile(t, owner, portalID, fileName)
		require.NotEmpty(t, file.ID)
		assert.Equal(t, fileName, file.Name)
		assert.Equal(t, "Security", file.Category)
		assert.Equal(t, "RESTRICTED", file.CompliancePortalVisibility)
		assert.Equal(t, fileName, file.File.FileName)
		assert.Equal(t, "application/pdf", file.File.MimeType)
		assert.NotEmpty(t, file.File.DownloadURL)
		assert.True(
			t,
			strings.Contains(file.File.DownloadURL, "/api/files/v1/"),
			"downloadUrl must route through the files API, got %q",
			file.File.DownloadURL,
		)
		assert.True(t, file.CanUpdate)
		assert.True(t, file.CanDelete)

		fileTotal, listedFileID := queryCompliancePortalCMSPortalFiles(t, owner, portalID)
		assert.Equal(t, 1, fileTotal)
		assert.Equal(t, file.ID, listedFileID)

		updatedDisplayName := factory.SafeName("cms-file-renamed") + ".pdf"

		var updateResult struct {
			UpdateCompliancePortalFile struct {
				CompliancePortalFile struct {
					Name                       string `json:"name"`
					Category                   string `json:"category"`
					CompliancePortalVisibility string `json:"compliancePortalVisibility"`
				} `json:"compliancePortalFile"`
			} `json:"updateCompliancePortalFile"`
		}

		err := owner.Execute(
			`
				mutation($input: UpdateCompliancePortalFileInput!) {
					updateCompliancePortalFile(input: $input) {
						compliancePortalFile {
							name category compliancePortalVisibility
						}
					}
				}
			`,
			map[string]any{
				"input": map[string]any{
					"id":                         file.ID,
					"name":                       updatedDisplayName,
					"category":                   "Compliance",
					"compliancePortalVisibility": "PUBLIC",
				},
			},
			&updateResult,
		)
		require.NoError(t, err)
		assert.Equal(t, updatedDisplayName, updateResult.UpdateCompliancePortalFile.CompliancePortalFile.Name)
		assert.Equal(t, "Compliance", updateResult.UpdateCompliancePortalFile.CompliancePortalFile.Category)
		assert.Equal(t, "PUBLIC", updateResult.UpdateCompliancePortalFile.CompliancePortalFile.CompliancePortalVisibility)

		var getResult struct {
			GetCompliancePortalFile struct {
				CompliancePortalFile struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"compliancePortalFile"`
			} `json:"getCompliancePortalFile"`
		}

		err = owner.Execute(
			`
				mutation($input: GetCompliancePortalFileInput!) {
					getCompliancePortalFile(input: $input) {
						compliancePortalFile { id name }
					}
				}
			`,
			map[string]any{"input": map[string]any{"id": file.ID}},
			&getResult,
		)
		require.NoError(t, err)
		assert.Equal(t, file.ID, getResult.GetCompliancePortalFile.CompliancePortalFile.ID)
		assert.Equal(t, updatedDisplayName, getResult.GetCompliancePortalFile.CompliancePortalFile.Name)

		var deleteResult struct {
			DeleteCompliancePortalFile struct {
				DeletedCompliancePortalFileID string `json:"deletedCompliancePortalFileId"`
			} `json:"deleteCompliancePortalFile"`
		}

		err = owner.Execute(
			`
				mutation($input: DeleteCompliancePortalFileInput!) {
					deleteCompliancePortalFile(input: $input) {
						deletedCompliancePortalFileId
					}
				}
			`,
			map[string]any{"input": map[string]any{"id": file.ID}},
			&deleteResult,
		)
		require.NoError(t, err)
		assert.Equal(t, file.ID, deleteResult.DeleteCompliancePortalFile.DeletedCompliancePortalFileID)

		fileTotal, _ = queryCompliancePortalCMSPortalFiles(t, owner, portalID)
		assert.Equal(t, 0, fileTotal)
	})

	t.Run("portal overview", func(t *testing.T) {
		assertCompliancePortalCMSOverview(
			t,
			owner,
			portalID,
			orgID,
			factory.SafeName("CMS profile entity"),
		)
	})
}
