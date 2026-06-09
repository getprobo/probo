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

package coredata

import (
	"encoding"
	"fmt"
)

type AccessReviewCampaignStatus string

const (
	AccessReviewCampaignStatusDraft          AccessReviewCampaignStatus = "DRAFT"
	AccessReviewCampaignStatusInProgress     AccessReviewCampaignStatus = "IN_PROGRESS"
	AccessReviewCampaignStatusPendingActions AccessReviewCampaignStatus = "PENDING_ACTIONS"
	AccessReviewCampaignStatusCompleted      AccessReviewCampaignStatus = "COMPLETED"
	AccessReviewCampaignStatusCancelled      AccessReviewCampaignStatus = "CANCELLED"
)

var (
	_ fmt.Stringer             = AccessReviewCampaignStatus("")
	_ encoding.TextMarshaler   = AccessReviewCampaignStatus("")
	_ encoding.TextUnmarshaler = (*AccessReviewCampaignStatus)(nil)
)

func AccessReviewCampaignStatuses() []AccessReviewCampaignStatus {
	return []AccessReviewCampaignStatus{
		AccessReviewCampaignStatusDraft,
		AccessReviewCampaignStatusInProgress,
		AccessReviewCampaignStatusPendingActions,
		AccessReviewCampaignStatusCompleted,
		AccessReviewCampaignStatusCancelled,
	}
}

func (v AccessReviewCampaignStatus) IsValid() bool {
	switch v {
	case
		AccessReviewCampaignStatusDraft,
		AccessReviewCampaignStatusInProgress,
		AccessReviewCampaignStatusPendingActions,
		AccessReviewCampaignStatusCompleted,
		AccessReviewCampaignStatusCancelled:
		return true
	}

	return false
}

func (v AccessReviewCampaignStatus) String() string {
	return string(v)
}

func (v AccessReviewCampaignStatus) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *AccessReviewCampaignStatus) UnmarshalText(text []byte) error {
	val := AccessReviewCampaignStatus(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid AccessReviewCampaignStatus value: %q", string(text))
	}

	*v = val

	return nil
}
