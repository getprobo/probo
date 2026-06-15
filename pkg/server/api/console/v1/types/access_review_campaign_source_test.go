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
