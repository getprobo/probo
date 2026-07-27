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

package accessreview

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// Campaign is the business view of a loaded access review campaign record.
type Campaign struct {
	ID     gid.GID
	status coredata.AccessReviewCampaignStatus
}

func NewCampaignFromCoredata(record *coredata.AccessReviewCampaign) Campaign {
	return Campaign{
		ID:     record.ID,
		status: record.Status,
	}
}

func (c Campaign) IsDraft() bool {
	return c.status == coredata.AccessReviewCampaignStatusDraft
}

func (c Campaign) IsDeletable() bool {
	return c.IsDraft() || c.status == coredata.AccessReviewCampaignStatusCancelled
}

func (c Campaign) IsPendingActions() bool {
	return c.status == coredata.AccessReviewCampaignStatusPendingActions
}

func (c Campaign) IsCompleted() bool {
	return c.status == coredata.AccessReviewCampaignStatusCompleted
}

func (c Campaign) IsCancelled() bool {
	return c.status == coredata.AccessReviewCampaignStatusCancelled
}
