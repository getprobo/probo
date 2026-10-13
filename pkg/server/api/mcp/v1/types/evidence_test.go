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

package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

func TestNewEvidence_FileDownloadURL(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	now := time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)
	fileID := gid.New(tenantID, coredata.FileEntityType)
	downloadURL := "https://probo.example/api/files/v1/" + fileID.String()
	file := &types.File{
		ID:          fileID,
		MimeType:    "application/pdf",
		FileName:    "access-review.pdf",
		Size:        128,
		DownloadURL: downloadURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	got := types.NewEvidence(
		&coredata.Evidence{
			ID:                gid.New(tenantID, coredata.EvidenceEntityType),
			OrganizationID:    gid.New(tenantID, coredata.OrganizationEntityType),
			InternalControlID: gid.New(tenantID, coredata.InternalControlEntityType),
			State:             coredata.EvidenceStateFulfilled,
			ReferenceID:       "EVD-1",
			Type:              coredata.EvidenceTypeFile,
			EvidenceFileId:    &fileID,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		file,
	)

	require.NotNil(t, got.File)
	assert.Equal(t, fileID, got.File.ID)
	assert.Equal(t, "access-review.pdf", got.File.FileName)
	assert.Equal(t, downloadURL, got.URL)
	assert.Equal(t, types.EvidenceType(coredata.EvidenceTypeFile), got.Type)
}

func TestNewEvidence_LinkKeepsStoredURL(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	now := time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)
	link := "https://example.com/policy"

	got := types.NewEvidence(
		&coredata.Evidence{
			ID:                gid.New(tenantID, coredata.EvidenceEntityType),
			OrganizationID:    gid.New(tenantID, coredata.OrganizationEntityType),
			InternalControlID: gid.New(tenantID, coredata.InternalControlEntityType),
			State:             coredata.EvidenceStateFulfilled,
			ReferenceID:       "EVD-2",
			Type:              coredata.EvidenceTypeLink,
			URL:               link,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		nil,
	)

	assert.Nil(t, got.File)
	assert.Equal(t, link, got.URL)
}

func TestNewListInternalControlEvidencesOutput_AttachesFile(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	now := time.Date(2026, 10, 13, 9, 0, 0, 0, time.UTC)
	fileID := gid.New(tenantID, coredata.FileEntityType)
	evidenceID := gid.New(tenantID, coredata.EvidenceEntityType)
	downloadURL := "https://probo.example/api/files/v1/" + fileID.String()

	output := types.NewListInternalControlEvidencesOutput(
		&page.Page[*coredata.Evidence, coredata.EvidenceOrderField]{
			Data: []*coredata.Evidence{
				{
					ID:                evidenceID,
					OrganizationID:    gid.New(tenantID, coredata.OrganizationEntityType),
					InternalControlID: gid.New(tenantID, coredata.InternalControlEntityType),
					State:             coredata.EvidenceStateFulfilled,
					ReferenceID:       "EVD-3",
					Type:              coredata.EvidenceTypeFile,
					EvidenceFileId:    &fileID,
					CreatedAt:         now,
					UpdatedAt:         now,
				},
			},
			Cursor: &page.Cursor[coredata.EvidenceOrderField]{
				OrderBy: page.OrderBy[coredata.EvidenceOrderField]{
					Field:     coredata.EvidenceOrderFieldCreatedAt,
					Direction: page.OrderDirectionDesc,
				},
			},
		},
		map[gid.GID]*types.File{
			fileID: {
				ID:          fileID,
				DownloadURL: downloadURL,
				FileName:    "screenshot.png",
			},
		},
	)

	require.Len(t, output.Evidences, 1)
	assert.Equal(t, evidenceID, output.Evidences[0].ID)
	assert.Equal(t, downloadURL, output.Evidences[0].URL)
	require.NotNil(t, output.Evidences[0].File)
	assert.Equal(t, "screenshot.png", output.Evidences[0].File.FileName)
}
