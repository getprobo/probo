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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

func TestScoreLevels_LevelFor(t *testing.T) {
	t.Parallel()

	levels := coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 6},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 13},
	}

	assert.Equal(t, coredata.RiskScoreLevelLow, levels.LevelFor(1))
	assert.Equal(t, coredata.RiskScoreLevelLow, levels.LevelFor(5))
	assert.Equal(t, coredata.RiskScoreLevelMedium, levels.LevelFor(6))
	assert.Equal(t, coredata.RiskScoreLevelMedium, levels.LevelFor(12))
	assert.Equal(t, coredata.RiskScoreLevelHigh, levels.LevelFor(13))
	assert.Equal(t, coredata.RiskScoreLevelHigh, levels.LevelFor(25))
}

func TestParseScoreLevels_RoundTrip(t *testing.T) {
	t.Parallel()

	levels := coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 5},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 10},
		{Level: coredata.RiskScoreLevelCritical, MinScore: 15},
	}
	parsed, err := coredata.ParseScoreLevels(levels.Spec())
	require.NoError(t, err)
	assert.Equal(t, levels, parsed)

	parsed, err = coredata.ParseScoreLevels(" low = 1 , high = 10 ")
	require.NoError(t, err)
	assert.Equal(t, coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 10},
	}, parsed)
}

func TestScoreLevels_ScanRejectsEmpty(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		value any
	}{
		{name: "nil", value: nil},
		{name: "empty string", value: ""},
		{name: "null string", value: "null"},
		{name: "empty object string", value: "{}"},
		{name: "empty array string", value: "[]"},
		{name: "empty bytes", value: []byte{}},
		{name: "null bytes", value: []byte("null")},
		{name: "empty object bytes", value: []byte("{}")},
		{name: "empty array bytes", value: []byte("[]")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var levels coredata.ScoreLevels

			err := levels.Scan(tc.value)
			if tc.value == nil {
				require.ErrorContains(t, err, "value is nil")
				return
			}

			require.ErrorContains(t, err, "value is empty")
		})
	}
}

func TestScoreLevels_ScanAcceptsJSON(t *testing.T) {
	t.Parallel()

	raw := `[{"level":"LOW","min_score":1},{"level":"HIGH","min_score":10}]`
	expected := coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 10},
	}

	for _, tc := range []struct {
		name  string
		value any
	}{
		{name: "string", value: raw},
		{name: "bytes", value: []byte(raw)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var levels coredata.ScoreLevels

			require.NoError(t, levels.Scan(tc.value))
			assert.Equal(t, expected, levels)
		})
	}
}
