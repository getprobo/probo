// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

func TestNewAccessReviewCampaignScopeSource_DefaultFetchState(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	source := &coredata.AccessSource{
		ID:             gid.New(tenantID, coredata.AccessSourceEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Name:           "Google Workspace",
	}

	campaignID := gid.New(tenantID, coredata.AccessReviewCampaignEntityType)
	got := NewAccessReviewCampaignScopeSource(campaignID, source, nil)
	if got.FetchStatus != coredata.AccessReviewCampaignSourceFetchStatusQueued {
		t.Fatalf("fetch status = %q, want QUEUED", got.FetchStatus)
	}
	if got.FetchedAccountsCount != 0 {
		t.Fatalf("fetched accounts count = %d, want 0", got.FetchedAccountsCount)
	}
	if got.AttemptCount != 0 {
		t.Fatalf("attempt count = %d, want 0", got.AttemptCount)
	}
}

func TestNewAccessReviewCampaignScopeSource_UsesFetchState(t *testing.T) {
	t.Parallel()

	now := time.Now()
	errMsg := "connector timeout"
	tenantID := gid.NewTenantID()
	source := &coredata.AccessSource{
		ID:             gid.New(tenantID, coredata.AccessSourceEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Name:           "Linear",
	}
	fetch := &coredata.AccessReviewCampaignSourceFetch{
		Status:               coredata.AccessReviewCampaignSourceFetchStatusFailed,
		FetchedAccountsCount: 42,
		AttemptCount:         3,
		LastError:            &errMsg,
		StartedAt:            &now,
		CompletedAt:          &now,
	}

	campaignID := gid.New(tenantID, coredata.AccessReviewCampaignEntityType)
	got := NewAccessReviewCampaignScopeSource(campaignID, source, fetch)
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
