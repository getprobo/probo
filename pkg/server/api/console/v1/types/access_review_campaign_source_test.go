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

package types

import (
	"testing"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func newTestCampaignSource(tenantID gid.TenantID, sourceID *gid.GID, name string) *coredata.AccessReviewCampaignSource {
	return &coredata.AccessReviewCampaignSource{
		ID:                     gid.New(tenantID, coredata.AccessReviewCampaignSourceEntityType),
		TenantID:               tenantID,
		AccessReviewCampaignID: gid.New(tenantID, coredata.AccessReviewCampaignEntityType),
		AccessReviewSourceID:   sourceID,
		Name:                   name,
	}
}

func TestNewAccessReviewCampaignSource(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	sourceID := gid.New(tenantID, coredata.AccessReviewSourceEntityType)
	campaignSource := newTestCampaignSource(tenantID, &sourceID, "Google Workspace")

	got := NewAccessReviewCampaignSource(campaignSource)
	if got.ID != campaignSource.ID {
		t.Fatalf("id = %v, want %v", got.ID, campaignSource.ID)
	}

	if got.Campaign == nil || got.Campaign.ID != campaignSource.AccessReviewCampaignID {
		t.Fatalf("campaign id = %v, want %v", got.Campaign, campaignSource.AccessReviewCampaignID)
	}

	if got.SourceID == nil || *got.SourceID != sourceID {
		t.Fatalf("source id = %v, want %v", got.SourceID, sourceID)
	}

	if got.Name != "Google Workspace" {
		t.Fatalf("name = %q, want snapshot name", got.Name)
	}
}

// TestNewAccessReviewCampaignSource_DeletedSource verifies a snapshot whose
// live source has been deleted still renders (with a nil source link) so the
// historical review data remains visible.
func TestNewAccessReviewCampaignSource_DeletedSource(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	campaignSource := newTestCampaignSource(tenantID, nil, "Deleted Source")

	got := NewAccessReviewCampaignSource(campaignSource)
	if got.SourceID != nil {
		t.Fatalf("source id = %v, want nil for deleted source", got.SourceID)
	}

	if got.Name != "Deleted Source" {
		t.Fatalf("name = %q, want snapshot name preserved", got.Name)
	}
}
