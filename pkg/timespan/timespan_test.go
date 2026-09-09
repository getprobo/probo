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

package timespan_test

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/timespan"
)

func TestParseAndString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		want   timespan.TimeSpan
		output string
	}{
		{
			name:   "zero seconds",
			input:  "PT0S",
			want:   timespan.Zero,
			output: "PT0S",
		},
		{
			name:   "one hour",
			input:  "PT1H",
			want:   timespan.TimeSpan{Microseconds: 3600 * 1_000_000},
			output: "PT1H",
		},
		{
			name:   "hour and minute",
			input:  "PT1H30M",
			want:   timespan.TimeSpan{Microseconds: (3600 + 30*60) * 1_000_000},
			output: "PT1H30M",
		},
		{
			name:   "one day",
			input:  "P1D",
			want:   timespan.TimeSpan{Days: 1},
			output: "P1D",
		},
		{
			name:   "one month",
			input:  "P1M",
			want:   timespan.TimeSpan{Months: 1},
			output: "P1M",
		},
		{
			name:   "one year",
			input:  "P1Y",
			want:   timespan.TimeSpan{Months: 12},
			output: "P1Y",
		},
		{
			name:   "week becomes days",
			input:  "P1W",
			want:   timespan.TimeSpan{Days: 7},
			output: "P7D",
		},
		{
			name:  "mixed calendar and clock",
			input: "P1Y2M3DT4H5M6S",
			want: timespan.TimeSpan{
				Months:       14,
				Days:         3,
				Microseconds: (4*3600 + 5*60 + 6) * 1_000_000,
			},
			output: "P1Y2M3DT4H5M6S",
		},
		{
			name:   "fractional seconds",
			input:  "PT0.5S",
			want:   timespan.TimeSpan{Microseconds: 500_000},
			output: "PT0.5S",
		},
		{
			name:   "fractional month becomes days",
			input:  "P1.5M",
			want:   timespan.TimeSpan{Months: 1, Days: 15},
			output: "P1M15D",
		},
		{
			name:   "negative month",
			input:  "-P1M",
			want:   timespan.TimeSpan{Months: -1},
			output: "-P1M",
		},
		{
			name:   "mixed month and negative days",
			input:  "P1M-5D",
			want:   timespan.TimeSpan{Months: 1, Days: -5},
			output: "P1M-5D",
		},
		{
			name:   "day minus one hour normalizes to twenty three hours",
			input:  "P1DT-1H",
			want:   timespan.TimeSpan{Microseconds: 23 * 3600 * 1_000_000},
			output: "PT23H",
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				got, err := timespan.Parse(tc.input)
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.output, got.String())
			},
		)
	}
}

func TestParseRejectsInvalid(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		"", "1H", "P1H", "PT1D", "P1X", "PT2562047789H",
		"P", "PT", "-P", "-PT", "P1YT", "P1DT", "-P1YT",
	} {
		t.Run(
			input,
			func(t *testing.T) {
				t.Parallel()

				_, err := timespan.Parse(input)
				assert.Error(t, err)
			},
		)
	}
}

func TestStringMinIntegerComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		span timespan.TimeSpan
		want string
	}{
		{
			name: "minimum months",
			span: timespan.TimeSpan{Months: math.MinInt32},
			want: "-P178956970Y8M",
		},
		{
			name: "minimum days",
			span: timespan.TimeSpan{Days: math.MinInt32},
			want: "-P2147483648D",
		},
		{
			name: "minimum microseconds",
			span: timespan.TimeSpan{Microseconds: math.MinInt64},
			want: "-P106751991DT4H54.775808S",
		},
		{
			name: "positive months with minimum microseconds",
			span: timespan.TimeSpan{Months: 1, Microseconds: math.MinInt64},
			want: "P1M-106751991DT-4H-54.775808S",
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, tc.span.String())

				parsed, err := timespan.Parse(tc.want)
				require.NoError(t, err)
				assert.Equal(t, 0, tc.span.Compare(parsed))
			},
		)
	}
}

func TestRoundTripDoesNotFlattenMonths(t *testing.T) {
	t.Parallel()

	parsed, err := timespan.Parse("P1M")
	require.NoError(t, err)
	assert.Equal(t, timespan.TimeSpan{Months: 1}, parsed)

	pgInterval, err := parsed.IntervalValue()
	require.NoError(t, err)
	assert.Equal(t, pgtype.Interval{Months: 1, Valid: true}, pgInterval)

	var scanned timespan.TimeSpan
	require.NoError(t, scanned.ScanInterval(pgInterval))
	assert.Equal(t, "P1M", scanned.String())
}

func TestCompare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		left  timespan.TimeSpan
		right timespan.TimeSpan
		want  int
	}{
		{
			name:  "equal clock spans",
			left:  timespan.FromDuration(time.Hour),
			right: timespan.FromDuration(time.Hour),
			want:  0,
		},
		{
			name:  "month equals thirty days",
			left:  timespan.TimeSpan{Months: 1},
			right: timespan.TimeSpan{Days: 30},
			want:  0,
		},
		{
			name:  "month less than one thousand hours",
			left:  timespan.TimeSpan{Months: 1},
			right: timespan.FromDuration(1000 * time.Hour),
			want:  -1,
		},
		{
			name:  "two months greater than one thousand hours",
			left:  timespan.TimeSpan{Months: 2},
			right: timespan.FromDuration(1000 * time.Hour),
			want:  1,
		},
		{
			name:  "day equals twenty four hours",
			left:  timespan.TimeSpan{Days: 1},
			right: timespan.FromDuration(24 * time.Hour),
			want:  0,
		},
	}

	for _, tc := range tests {
		t.Run(
			tc.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tc.want, tc.left.Compare(tc.right))
				assert.Equal(t, -tc.want, tc.right.Compare(tc.left))
			},
		)
	}
}

func TestCompareDuration(t *testing.T) {
	t.Parallel()

	t.Run(
		"equal microsecond bound",
		func(t *testing.T) {
			t.Parallel()

			span := timespan.FromDuration(time.Microsecond)
			assert.Equal(t, 0, span.CompareDuration(time.Microsecond))
		},
	)

	t.Run(
		"zero below nanosecond bound",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, -1, timespan.Zero.CompareDuration(time.Nanosecond))
		},
	)

	t.Run(
		"microsecond below leftover-nanosecond bound",
		func(t *testing.T) {
			t.Parallel()

			span := timespan.FromDuration(time.Microsecond)
			assert.Equal(t, -1, span.CompareDuration(time.Microsecond+time.Nanosecond))
		},
	)

	t.Run(
		"microsecond above leftover-nanosecond bound",
		func(t *testing.T) {
			t.Parallel()

			span := timespan.FromDuration(time.Microsecond)
			assert.Equal(t, 1, span.CompareDuration(time.Microsecond-time.Nanosecond))
		},
	)
}

func TestFromDurationAndClockDuration(t *testing.T) {
	t.Parallel()

	clock := timespan.FromDuration(90 * time.Minute)
	got, ok := clock.ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 90*time.Minute, got)

	_, ok = timespan.TimeSpan{Months: 1}.ClockDuration()
	assert.False(t, ok)

	got, ok = timespan.TimeSpan{Days: 1}.ClockDuration()
	require.True(t, ok)
	assert.Equal(t, 24*time.Hour, got)

	_, ok = timespan.TimeSpan{Days: 213504}.ClockDuration()
	assert.False(t, ok)

	maxDays := int32(int64(math.MaxInt64) / int64(24*time.Hour))
	got, ok = timespan.TimeSpan{Days: maxDays}.ClockDuration()
	require.True(t, ok)
	assert.Equal(t, time.Duration(maxDays)*24*time.Hour, got)

	_, ok = timespan.TimeSpan{Days: maxDays + 1}.ClockDuration()
	assert.False(t, ok)

	dayNs := int64(maxDays) * int64(24*time.Hour)
	overflowUs := (math.MaxInt64-dayNs)/int64(time.Microsecond) + 1
	_, ok = timespan.TimeSpan{Days: maxDays, Microseconds: overflowUs}.ClockDuration()
	assert.False(t, ok)
}

func TestJSONRoundTripMixedSigns(t *testing.T) {
	t.Parallel()

	original := timespan.TimeSpan{Months: 1, Days: -5}
	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Equal(t, `"P1M-5D"`, string(data))

	var parsed timespan.TimeSpan
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, original, parsed)

	var scanned timespan.TimeSpan
	require.NoError(t, scanned.ScanInterval(pgtype.Interval{Months: 1, Days: -5, Valid: true}))
	assert.Equal(t, original, scanned)

	data, err = json.Marshal(scanned)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, original, parsed)
}

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := timespan.TimeSpan{Months: 1, Days: 2, Microseconds: 3_000_000}
	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Equal(t, `"P1M2DT3S"`, string(data))

	var parsed timespan.TimeSpan
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, original, parsed)
}
