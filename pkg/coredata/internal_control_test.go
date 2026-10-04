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
	"go.probo.inc/probo/pkg/timespan"
)

func mustTimeSpan(t *testing.T, raw string) *timespan.TimeSpan {
	t.Helper()

	span, err := timespan.Parse(raw)
	require.NoError(t, err)

	return &span
}

func TestInternalControlOperatingFrequency_Columns(t *testing.T) {
	t.Parallel()

	quarterly := mustTimeSpan(t, "P3M")
	event := "when someone leaves"

	t.Run("nil clears every column", func(t *testing.T) {
		t.Parallel()

		var freq *InternalControlOperatingFrequency
		mode, interval, text := freq.Columns()

		assert.Nil(t, mode)
		assert.Nil(t, interval)
		assert.Nil(t, text)
	})

	t.Run("continuous keeps only the mode", func(t *testing.T) {
		t.Parallel()

		freq := &InternalControlOperatingFrequency{
			Mode:     InternalControlOperatingModeContinuous,
			Interval: quarterly,
			Event:    &event,
		}
		mode, interval, text := freq.Columns()

		require.NotNil(t, mode)
		assert.Equal(t, InternalControlOperatingModeContinuous, *mode)
		assert.Nil(t, interval)
		assert.Nil(t, text)
	})

	t.Run("event keeps the text and drops the interval", func(t *testing.T) {
		t.Parallel()

		freq := &InternalControlOperatingFrequency{
			Mode:     InternalControlOperatingModeEvent,
			Interval: quarterly,
			Event:    &event,
		}
		mode, interval, text := freq.Columns()

		require.NotNil(t, mode)
		assert.Equal(t, InternalControlOperatingModeEvent, *mode)
		assert.Nil(t, interval)
		require.NotNil(t, text)
		assert.Equal(t, event, *text)
	})

	t.Run("periodic keeps the interval and drops the event", func(t *testing.T) {
		t.Parallel()

		freq := &InternalControlOperatingFrequency{
			Mode:     InternalControlOperatingModePeriodic,
			Interval: quarterly,
			Event:    &event,
		}
		mode, interval, text := freq.Columns()

		require.NotNil(t, mode)
		assert.Equal(t, InternalControlOperatingModePeriodic, *mode)
		require.NotNil(t, interval)
		assert.Equal(t, *quarterly, *interval)
		assert.Nil(t, text)
	})
}

func TestNextInternalControlDue(t *testing.T) {
	t.Parallel()

	from := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)

	t.Run("positive durations advance from the anchor", func(t *testing.T) {
		t.Parallel()

		cases := []struct {
			raw  string
			want time.Time
		}{
			{"P1D", from.AddDate(0, 0, 1)},
			{"P7D", from.AddDate(0, 0, 7)},
			{"P1M", from.AddDate(0, 1, 0)},
			{"P3M", from.AddDate(0, 3, 0)},
			{"P6M", from.AddDate(0, 6, 0)},
			{"P1Y", from.AddDate(1, 0, 0)},
		}

		for _, tc := range cases {
			got := NextInternalControlDue(from, mustTimeSpan(t, tc.raw))
			require.NotNil(t, got)
			assert.True(t, tc.want.Equal(*got), tc.raw)
		}
	})

	t.Run("empty and non-positive durations have no due date", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, NextInternalControlDue(from, nil))

		zero := timespan.TimeSpan{}
		assert.Nil(t, NextInternalControlDue(from, &zero))
		assert.Nil(t, NextInternalControlDue(from, mustTimeSpan(t, "-P1D")))
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
