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

	t.Run("an operating frequency change leaves evidence and test dues", func(t *testing.T) {
		t.Parallel()

		quarterly := mustMeasureSpan(t, "P3M")
		event := "when someone leaves"
		freq := &coredata.InternalControlOperatingFrequency{
			Mode:     coredata.InternalControlOperatingModePeriodic,
			Interval: &quarterly,
			Event:    &event,
		}
		due := now.AddDate(0, 1, 0)
		measure := &coredata.Measure{
			EvidenceCadence: &quarterly,
			NextEvidenceDue: &due,
			TestingCadence:  &quarterly,
			NextTestDue:     &due,
		}

		applyMeasureUpdate(
			&UpdateMeasureRequest{OperatingFrequency: &freq},
			measure,
			now,
		)

		require.NotNil(t, measure.OperatingMode)
		assert.Equal(t, coredata.InternalControlOperatingModePeriodic, *measure.OperatingMode)
		require.NotNil(t, measure.OperatingInterval)
		assert.Equal(t, quarterly, *measure.OperatingInterval)
		assert.Nil(t, measure.OperatingEvent)
		require.NotNil(t, measure.NextEvidenceDue)
		assert.True(t, measure.NextEvidenceDue.Equal(due))
		require.NotNil(t, measure.NextTestDue)
		assert.True(t, measure.NextTestDue.Equal(due))
	})

	t.Run("clearing the operating frequency clears its columns", func(t *testing.T) {
		t.Parallel()

		quarterly := mustMeasureSpan(t, "P1D")
		mode := coredata.InternalControlOperatingModePeriodic

		var cleared *coredata.InternalControlOperatingFrequency

		measure := &coredata.Measure{
			OperatingMode:     &mode,
			OperatingInterval: &quarterly,
		}

		applyMeasureUpdate(
			&UpdateMeasureRequest{OperatingFrequency: &cleared},
			measure,
			now,
		)

		assert.Nil(t, measure.OperatingMode)
		assert.Nil(t, measure.OperatingInterval)
		assert.Nil(t, measure.OperatingEvent)
	})
}

func TestValidOperatingFrequency(t *testing.T) {
	t.Parallel()

	t.Run("an empty value is allowed", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, validOperatingFrequency()(nil))
	})

	t.Run("continuous rejects an interval or an event", func(t *testing.T) {
		t.Parallel()

		quarterly := mustMeasureSpan(t, "P3M")
		event := "when someone leaves"

		assert.Nil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode: coredata.InternalControlOperatingModeContinuous,
		}))
		assert.NotNil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode:     coredata.InternalControlOperatingModeContinuous,
			Interval: &quarterly,
		}))
		assert.NotNil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode:  coredata.InternalControlOperatingModeContinuous,
			Event: &event,
		}))
	})

	t.Run("event keeps optional text and rejects an interval", func(t *testing.T) {
		t.Parallel()

		quarterly := mustMeasureSpan(t, "P3M")
		event := "when someone leaves"

		assert.Nil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode: coredata.InternalControlOperatingModeEvent,
		}))
		assert.Nil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode:  coredata.InternalControlOperatingModeEvent,
			Event: &event,
		}))
		assert.NotNil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode:     coredata.InternalControlOperatingModeEvent,
			Interval: &quarterly,
		}))
	})

	t.Run("periodic requires a positive duration and rejects an event", func(t *testing.T) {
		t.Parallel()

		quarterly := mustMeasureSpan(t, "P3M")
		negative := mustMeasureSpan(t, "-P1M")
		event := "when someone leaves"
		zero := timespan.TimeSpan{}

		assert.Nil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode:     coredata.InternalControlOperatingModePeriodic,
			Interval: &quarterly,
		}))
		assert.NotNil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode: coredata.InternalControlOperatingModePeriodic,
		}))
		assert.NotNil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode:     coredata.InternalControlOperatingModePeriodic,
			Interval: &zero,
		}))
		assert.NotNil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode:     coredata.InternalControlOperatingModePeriodic,
			Interval: &negative,
		}))
		assert.NotNil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode:     coredata.InternalControlOperatingModePeriodic,
			Interval: &quarterly,
			Event:    &event,
		}))
	})

	t.Run("an unknown mode is rejected", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, validOperatingFrequency()(coredata.InternalControlOperatingFrequency{
			Mode: "WEEKLY",
		}))
	})
}

func TestNormalizeOperatingFrequency(t *testing.T) {
	t.Parallel()

	t.Run("blank event text becomes empty", func(t *testing.T) {
		t.Parallel()

		blank := "  "
		freq := &coredata.InternalControlOperatingFrequency{
			Mode:  coredata.InternalControlOperatingModeEvent,
			Event: &blank,
		}

		normalizeOperatingFrequency(freq)

		assert.Nil(t, freq.Event)
	})
}

func TestNonPositiveCadenceBecomesUnscheduled(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	negative := mustMeasureSpan(t, "-P1M")
	zero := timespan.TimeSpan{}
	create := &CreateMeasureRequest{
		OrganizationID:  gid.New(tenantID, coredata.OrganizationEntityType),
		Name:            "Access reviews",
		Category:        "Access",
		EvidenceCadence: &negative,
		TestingCadence:  &zero,
	}

	require.NoError(t, create.Validate())
	assert.Nil(t, create.EvidenceCadence)
	assert.Nil(t, create.TestingCadence)

	quarterly := mustMeasureSpan(t, "P3M")
	kept := &quarterly
	update := &UpdateMeasureRequest{
		ID:              gid.New(tenantID, coredata.MeasureEntityType),
		EvidenceCadence: &kept,
	}

	require.NoError(t, update.Validate())
	require.NotNil(t, update.EvidenceCadence)
	require.NotNil(t, *update.EvidenceCadence)
	assert.Equal(t, quarterly, **update.EvidenceCadence)

	cleared := &negative
	update.EvidenceCadence = &cleared

	require.NoError(t, update.Validate())
	require.NotNil(t, update.EvidenceCadence)
	assert.Nil(t, *update.EvidenceCadence)
}

func mustMeasureSpan(t *testing.T, raw string) timespan.TimeSpan {
	t.Helper()

	span, err := timespan.Parse(raw)
	require.NoError(t, err)

	return span
}
