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
	"time"

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
		Category:               coredata.AccessReviewSourceCategorySaaS,
	}
}

func TestNewAccessReviewCampaignSource_DefaultFetchState(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	sourceID := gid.New(tenantID, coredata.AccessReviewSourceEntityType)
	campaignSource := newTestCampaignSource(tenantID, &sourceID, "Google Workspace")

	got := NewAccessReviewCampaignSource(campaignSource, nil)
	if got.FetchStatus != coredata.AccessReviewCampaignSourceFetchStatusQueued {
		t.Fatalf("fetch status = %q, want QUEUED", got.FetchStatus)
	}

	if got.FetchedAccountsCount != 0 {
		t.Fatalf("fetched accounts count = %d, want 0", got.FetchedAccountsCount)
	}

	if got.AttemptCount != 0 {
		t.Fatalf("attempt count = %d, want 0", got.AttemptCount)
	}

	if got.SourceID == nil || *got.SourceID != sourceID {
		t.Fatalf("source id = %v, want %v", got.SourceID, sourceID)
	}

	if got.Name != "Google Workspace" {
		t.Fatalf("name = %q, want snapshot name", got.Name)
	}
}

func TestNewAccessReviewCampaignSource_UsesLatestAttempt(t *testing.T) {
	t.Parallel()

	now := time.Now()
	errMsg := "We couldn't fetch accounts from this source."
	tenantID := gid.NewTenantID()
	sourceID := gid.New(tenantID, coredata.AccessReviewSourceEntityType)
	campaignSource := newTestCampaignSource(tenantID, &sourceID, "Linear")
	attempt := &coredata.AccessReviewCampaignSourceFetchAttempt{
		Status:               coredata.AccessReviewCampaignSourceFetchStatusFailed,
		FetchedAccountsCount: 42,
		AttemptNumber:        3,
		Error:                &errMsg,
		StartedAt:            &now,
		CompletedAt:          &now,
	}

	got := NewAccessReviewCampaignSource(campaignSource, attempt)
	if got.FetchStatus != coredata.AccessReviewCampaignSourceFetchStatusFailed {
		t.Fatalf("fetch status = %q, want FAILED", got.FetchStatus)
	}

	if got.FetchedAccountsCount != 42 {
		t.Fatalf("fetched accounts count = %d, want 42", got.FetchedAccountsCount)
	}

	if got.AttemptCount != 3 {
		t.Fatalf("attempt count = %d, want 3", got.AttemptCount)
	}

	if got.LastError == nil || *got.LastError != errMsg {
		t.Fatalf("last error = %v, want %q", got.LastError, errMsg)
	}
}

// TestNewAccessReviewCampaignSource_DeletedSource verifies a snapshot whose
// live source has been deleted still renders (with a nil source link) so the
// historical review data remains visible.
func TestNewAccessReviewCampaignSource_DeletedSource(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	campaignSource := newTestCampaignSource(tenantID, nil, "Deleted Source")

	got := NewAccessReviewCampaignSource(campaignSource, nil)
	if got.SourceID != nil {
		t.Fatalf("source id = %v, want nil for deleted source", got.SourceID)
	}

	if got.Name != "Deleted Source" {
		t.Fatalf("name = %q, want snapshot name preserved", got.Name)
	}
}
