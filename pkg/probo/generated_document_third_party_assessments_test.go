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

package probo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestLatestActiveThirdPartyRiskAssessment(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 30, 12, 0, 0, 0, time.UTC)
	tenantID := gid.NewTenantID()

	newAssessment := func(createdAt, expiresAt time.Time) *coredata.ThirdPartyRiskAssessment {
		return &coredata.ThirdPartyRiskAssessment{
			ID:        gid.New(tenantID, coredata.ThirdPartyRiskAssessmentEntityType),
			CreatedAt: createdAt,
			ExpiresAt: expiresAt,
		}
	}

	expired := newAssessment(now.AddDate(0, -1, 0), now.Add(-time.Minute))
	olderActive := newAssessment(now.AddDate(0, 0, -7), now.AddDate(1, 0, 0))
	latestActive := newAssessment(now.AddDate(0, 0, -1), now.AddDate(1, 0, 0))
	expiringNow := newAssessment(now, now)

	t.Run(
		"returns the most recently created unexpired assessment",
		func(t *testing.T) {
			t.Parallel()

			got := latestActiveThirdPartyRiskAssessment(
				coredata.ThirdPartyRiskAssessments{expired, olderActive, latestActive},
				now,
			)
			require.NotNil(t, got)
			assert.Equal(t, latestActive.ID, got.ID)
		},
	)

	t.Run(
		"omits expired assessments even when they are newest",
		func(t *testing.T) {
			t.Parallel()

			got := latestActiveThirdPartyRiskAssessment(
				coredata.ThirdPartyRiskAssessments{latestActive, expired},
				now,
			)
			require.NotNil(t, got)
			assert.Equal(t, latestActive.ID, got.ID)
		},
	)

	t.Run(
		"returns nil when every assessment is expired",
		func(t *testing.T) {
			t.Parallel()

			got := latestActiveThirdPartyRiskAssessment(
				coredata.ThirdPartyRiskAssessments{expired, expiringNow},
				now,
			)
			assert.Nil(t, got)
		},
	)

	t.Run(
		"returns nil for an empty list",
		func(t *testing.T) {
			t.Parallel()

			got := latestActiveThirdPartyRiskAssessment(nil, now)
			assert.Nil(t, got)
		},
	)
}
