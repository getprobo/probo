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
	"go.probo.inc/probo/pkg/timespan"
)

func TestApplyMeasureUpdate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	t.Run("repeating the current status keeps a legacy state", func(t *testing.T) {
		t.Parallel()

		status := coredata.InternalControlImplementationStatusNotImplemented
		state := coredata.MeasureStateNotStarted
		measure := &coredata.Measure{
			ImplementationStatus: status,
			State:                state,
		}

		applyMeasureUpdate(
			&UpdateMeasureRequest{
				ImplementationStatus: &status,
				State:                &state,
			},
			measure,
			now,
		)

		assert.Equal(t, coredata.MeasureStateNotStarted, measure.State)
		assert.Equal(t, status, measure.ImplementationStatus)
	})

	t.Run("repeating an operating status keeps it operating", func(t *testing.T) {
		t.Parallel()

		status := coredata.InternalControlImplementationStatusOperating
		state := coredata.MeasureStateImplemented
		measure := &coredata.Measure{
			ImplementationStatus: status,
			State:                state,
		}

		applyMeasureUpdate(
			&UpdateMeasureRequest{
				ImplementationStatus: &status,
				State:                &state,
			},
			measure,
			now,
		)

		assert.Equal(t, coredata.InternalControlImplementationStatusOperating, measure.ImplementationStatus)
		assert.Equal(t, coredata.MeasureStateImplemented, measure.State)
	})

	t.Run("a status change wins over a state sent with it", func(t *testing.T) {
		t.Parallel()

		status := coredata.InternalControlImplementationStatusOperating
		state := coredata.MeasureStateNotApplicable
		measure := &coredata.Measure{
			ImplementationStatus: coredata.InternalControlImplementationStatusNotImplemented,
			State:                coredata.MeasureStateNotStarted,
		}

		applyMeasureUpdate(
			&UpdateMeasureRequest{
				ImplementationStatus: &status,
				State:                &state,
			},
			measure,
			now,
		)

		assert.Equal(t, coredata.InternalControlImplementationStatusOperating, measure.ImplementationStatus)
		assert.Equal(t, coredata.MeasureStateImplemented, measure.State)
	})

	t.Run("a state change applies when the status is unchanged", func(t *testing.T) {
		t.Parallel()

		status := coredata.InternalControlImplementationStatusNotImplemented
		state := coredata.MeasureStateNotApplicable
		measure := &coredata.Measure{
			ImplementationStatus: status,
			State:                coredata.MeasureStateNotStarted,
		}

		applyMeasureUpdate(
			&UpdateMeasureRequest{
				ImplementationStatus: &status,
				State:                &state,
			},
			measure,
			now,
		)

		assert.Equal(t, coredata.MeasureStateNotApplicable, measure.State)
		assert.Equal(t, coredata.InternalControlImplementationStatusNotImplemented, measure.ImplementationStatus)
	})

	t.Run("changing the evidence cadence sets the next due date", func(t *testing.T) {
		t.Parallel()

		quarterly := mustMeasureSpan(t, "P3M")
		cadence := &quarterly
		measure := &coredata.Measure{}

		applyMeasureUpdate(
			&UpdateMeasureRequest{EvidenceCadence: &cadence},
			measure,
			now,
		)

		require.NotNil(t, measure.EvidenceCadence)
		assert.Equal(t, quarterly, *measure.EvidenceCadence)
		require.NotNil(t, measure.NextEvidenceDue)
		assert.True(t, measure.NextEvidenceDue.Equal(now.AddDate(0, 3, 0)))
	})

	t.Run("repeating a cadence keeps the existing due date", func(t *testing.T) {
		t.Parallel()

		quarterly := mustMeasureSpan(t, "P3M")
		cadence := &quarterly
		due := now.AddDate(0, 1, 0)
		measure := &coredata.Measure{
			EvidenceCadence: &quarterly,
			NextEvidenceDue: &due,
		}

		applyMeasureUpdate(
			&UpdateMeasureRequest{EvidenceCadence: &cadence},
			measure,
			now,
		)

		require.NotNil(t, measure.NextEvidenceDue)
		assert.True(t, measure.NextEvidenceDue.Equal(due))
	})

	t.Run("clearing a cadence clears the due date", func(t *testing.T) {
		t.Parallel()

		quarterly := mustMeasureSpan(t, "P3M")
		var cleared *timespan.TimeSpan
		due := now
		measure := &coredata.Measure{
			EvidenceCadence: &quarterly,
			NextEvidenceDue: &due,
		}

		applyMeasureUpdate(
			&UpdateMeasureRequest{EvidenceCadence: &cleared},
			measure,
			now,
		)

		assert.Nil(t, measure.EvidenceCadence)
		assert.Nil(t, measure.NextEvidenceDue)
	})
}

func mustMeasureSpan(t *testing.T, raw string) timespan.TimeSpan {
	t.Helper()

	span, err := timespan.Parse(raw)
	require.NoError(t, err)

	return span
}
