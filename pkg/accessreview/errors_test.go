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
	"go.probo.inc/probo/pkg/gid"
)

func TestCampaignClientErrors(t *testing.T) {
	t.Parallel()

	statusErr := &CampaignInvalidStatusError{
		Operation: "start",
		Status:    coredata.AccessReviewCampaignStatusInProgress,
		Expected:  coredata.AccessReviewCampaignStatusDraft,
	}
	sourceErr := &CampaignSourceOrganizationMismatchError{
		Operation: "update",
		SourceID:  gid.GID("source-id"),
	}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "no scope sources",
			err:  ErrCampaignNoScopeSources,
			want: true,
		},
		{
			name: "invalid status",
			err:  statusErr,
			want: true,
		},
		{
			name: "wrapped invalid status",
			err:  fmt.Errorf("cannot start access review campaign: %w", statusErr),
			want: true,
		},
		{
			name: "source organization mismatch",
			err:  sourceErr,
			want: true,
		},
		{
			name: "internal error",
			err:  errors.New("cannot lock campaign: timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, IsCampaignClientError(tt.err))
		})
	}
}
