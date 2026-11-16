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

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/riskmanagement"
)

func TestFormatScoreBands_OmitsCritical(t *testing.T) {
	t.Parallel()

	bands := formatScoreBands(riskmanagement.DefaultScoreLevels(5, 5), 25)
	assert.Equal(t, "Low (1-4), Medium (5-9), High (10-14), Critical (15-25)", bands)

	bands = formatScoreBands(coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 6},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 13},
	}, 25)

	assert.Equal(t, "Low (1-5), Medium (6-12), High (13-25)", bands)
	assert.NotContains(t, bands, "Critical")
}

func TestRiskScoreRange_UsesConfiguredMapping(t *testing.T) {
	t.Parallel()

	levels := coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 6},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 13},
	}

	assert.Equal(t, "1-5", riskScoreRange(1, levels, 25))
	assert.Equal(t, "1-5", riskScoreRange(5, levels, 25))
	assert.Equal(t, "6-12", riskScoreRange(12, levels, 25))
	assert.Equal(t, "13-25", riskScoreRange(25, levels, 25))
}

func TestRiskScoreRange_DefaultMatrix(t *testing.T) {
	t.Parallel()

	levels := riskmanagement.DefaultScoreLevels(5, 5)

	assert.Equal(t, "5-9", riskScoreRange(9, levels, 25))
	assert.Equal(t, "10-14", riskScoreRange(10, levels, 25))
	assert.Equal(t, "15-25", riskScoreRange(25, levels, 25))
}
