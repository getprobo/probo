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
	"go.probo.inc/probo/pkg/coredata"
)

func TestNextRecurrenceDeadline(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.July, 28, 12, 0, 0, 0, time.UTC)
	oneDayOverdue := now.AddDate(0, 0, -1)

	tests := []struct {
		name     string
		deadline time.Time
		unit     coredata.TaskRecurrenceIntervalUnit
		count    int
		want     time.Time
	}{
		{
			name:     "every 2 days, overdue by a day",
			deadline: oneDayOverdue,
			unit:     coredata.TaskRecurrenceIntervalUnitDay,
			count:    2,
			want:     oneDayOverdue.AddDate(0, 0, 2),
		},
		{
			name:     "every 3 weeks, overdue by a day",
			deadline: oneDayOverdue,
			unit:     coredata.TaskRecurrenceIntervalUnitWeek,
			count:    3,
			want:     oneDayOverdue.AddDate(0, 0, 21),
		},
		{
			name:     "every month, overdue by a day",
			deadline: oneDayOverdue,
			unit:     coredata.TaskRecurrenceIntervalUnitMonth,
			count:    1,
			want:     oneDayOverdue.AddDate(0, 1, 0),
		},
		{
			name:     "quarterly (every 3 months), not yet overdue",
			deadline: now.AddDate(0, 1, 0),
			unit:     coredata.TaskRecurrenceIntervalUnitMonth,
			count:    3,
			want:     now.AddDate(0, 1, 0),
		},
		{
			name:     "every year, overdue by a day",
			deadline: oneDayOverdue,
			unit:     coredata.TaskRecurrenceIntervalUnitYear,
			count:    1,
			want:     oneDayOverdue.AddDate(1, 0, 0),
		},
		{
			name:     "long outage collapses to a single future occurrence",
			deadline: now.AddDate(-1, 0, 0),
			unit:     coredata.TaskRecurrenceIntervalUnitDay,
			count:    2,
			want:     now.AddDate(-1, 0, 0).AddDate(0, 0, 366),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := nextRecurrenceDeadline(tt.deadline, tt.unit, tt.count, now)

			assert.True(t, got.After(now), "next deadline must be strictly after now")
			assert.True(t, got.Equal(tt.want), "got %v, want %v", got, tt.want)
		})
	}
}
