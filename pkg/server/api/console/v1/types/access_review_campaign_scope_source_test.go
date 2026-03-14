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
		AccessReviewID: gid.New(tenantID, coredata.AccessReviewEntityType),
		Name:           "Google Workspace",
	}

	got := NewAccessReviewCampaignScopeSource(source, nil)
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
		AccessReviewID: gid.New(tenantID, coredata.AccessReviewEntityType),
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

	got := NewAccessReviewCampaignScopeSource(source, fetch)
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
