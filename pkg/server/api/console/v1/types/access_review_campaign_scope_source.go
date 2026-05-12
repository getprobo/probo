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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type AccessReviewCampaignScopeSource struct {
	ID                   gid.GID                                        `json:"id"`
	CampaignID           gid.GID                                        `json:"-"`
	Source               *AccessSource                                  `json:"source"`
	Name                 string                                         `json:"name"`
	FetchStatus          coredata.AccessReviewCampaignSourceFetchStatus `json:"fetchStatus"`
	FetchedAccountsCount int                                            `json:"fetchedAccountsCount"`
	AttemptCount         int                                            `json:"attemptCount"`
	LastError            *string                                        `json:"lastError,omitempty"`
	FetchStartedAt       *time.Time                                     `json:"fetchStartedAt,omitempty"`
	FetchCompletedAt     *time.Time                                     `json:"fetchCompletedAt,omitempty"`
}
