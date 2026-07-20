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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestCampaignStatusError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status coredata.AccessReviewCampaignStatus
		want   error
	}{
		{status: coredata.AccessReviewCampaignStatusInProgress, want: ErrCampaignInProgress},
		{status: coredata.AccessReviewCampaignStatusPendingActions, want: ErrCampaignPendingActions},
		{status: coredata.AccessReviewCampaignStatusCompleted, want: ErrCampaignCompleted},
		{status: coredata.AccessReviewCampaignStatusCancelled, want: ErrCampaignCancelled},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()

			err := fmt.Errorf("cannot start campaign: %w", CampaignStatusError(tt.status))
			assert.ErrorIs(t, err, tt.want)
		})
	}
}
