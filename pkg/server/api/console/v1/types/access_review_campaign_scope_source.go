package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type AccessReviewCampaignScopeSource struct {
	ID                   gid.GID                                        `json:"id"`
	Source               *AccessSource                                  `json:"source"`
	Name                 string                                         `json:"name"`
	FetchStatus          coredata.AccessReviewCampaignSourceFetchStatus `json:"fetchStatus"`
	FetchedAccountsCount int                                            `json:"fetchedAccountsCount"`
	AttemptCount         int                                            `json:"attemptCount"`
	LastError            *string                                        `json:"lastError,omitempty"`
	FetchStartedAt       *time.Time                                     `json:"fetchStartedAt,omitempty"`
	FetchCompletedAt     *time.Time                                     `json:"fetchCompletedAt,omitempty"`
}
