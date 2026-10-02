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

package coredata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextInternalControlDue(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("scheduled cadences advance from the anchor", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			cadence InternalControlCadence
			want    time.Time
		}{
			{InternalControlCadenceDaily, from.AddDate(0, 0, 1)},
			{InternalControlCadenceWeekly, from.AddDate(0, 0, 7)},
			{InternalControlCadenceMonthly, from.AddDate(0, 1, 0)},
			{InternalControlCadenceQuarterly, from.AddDate(0, 3, 0)},
			{InternalControlCadenceSemiannually, from.AddDate(0, 6, 0)},
			{InternalControlCadenceAnnually, from.AddDate(1, 0, 0)},
		}

		for _, tc := range cases {
			got := NextInternalControlDue(from, tc.cadence)
			require.NotNil(t, got)
			assert.True(t, tc.want.Equal(*got), tc.cadence)
		}
	})

	t.Run("continuous and ad hoc have no due date", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, NextInternalControlDue(from, InternalControlCadenceContinuous))
		assert.Nil(t, NextInternalControlDue(from, InternalControlCadenceAdHoc))
	})
}

func TestImplementationStatusMapping(t *testing.T) {
	t.Parallel()

	t.Run("operating counts as implemented for treatment progress", func(t *testing.T) {
		t.Parallel()

		state, ok := MeasureStateForImplementationStatus(InternalControlImplementationStatusOperating)
		require.True(t, ok)
		assert.Equal(t, MeasureStateImplemented, state)
	})

	t.Run("legacy states without an operational equivalent stay not implemented", func(t *testing.T) {
		t.Parallel()

		assert.Equal(
			t,
			InternalControlImplementationStatusNotImplemented,
			ImplementationStatusForMeasureState(MeasureStateNotStarted),
		)
		assert.Equal(
			t,
			InternalControlImplementationStatusNotImplemented,
			ImplementationStatusForMeasureState(MeasureStateNotApplicable),
		)
		assert.Equal(
			t,
			InternalControlImplementationStatusImplemented,
			ImplementationStatusForMeasureState(MeasureStateImplemented),
		)
	})
}
