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

package scorelevel_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/riskmanagement/scorelevel"
)

func TestDefaultScoreLevels(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 5},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 10},
		{Level: coredata.RiskScoreLevelCritical, MinScore: 15},
	}, scorelevel.DefaultScoreLevels(5, 5))

	assert.Equal(t, coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 2},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 4},
		{Level: coredata.RiskScoreLevelCritical, MinScore: 6},
	}, scorelevel.DefaultScoreLevels(3, 3))

	assert.Equal(t, coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 4},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 7},
		{Level: coredata.RiskScoreLevelCritical, MinScore: 10},
	}, scorelevel.DefaultScoreLevels(4, 4))

	t.Run("non-positive dimensions use a maximum of 25", func(t *testing.T) {
		t.Parallel()

		expected := coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelLow, MinScore: 1},
			{Level: coredata.RiskScoreLevelMedium, MinScore: 5},
			{Level: coredata.RiskScoreLevelHigh, MinScore: 10},
			{Level: coredata.RiskScoreLevelCritical, MinScore: 15},
		}

		assert.Equal(t, expected, scorelevel.DefaultScoreLevels(0, 0))
		assert.Equal(t, expected, scorelevel.DefaultScoreLevels(0, 5))
		assert.Equal(t, expected, scorelevel.DefaultScoreLevels(-1, 4))
	})
}
