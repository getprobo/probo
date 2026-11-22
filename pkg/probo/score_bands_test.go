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

func TestRiskScoreLevelLabel_UsesConfiguredMapping(t *testing.T) {
	t.Parallel()

	levels := coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 6},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 13},
	}

	score := 1
	assert.Equal(t, "Low", riskScoreLevelLabel(score, levels))
	assert.Equal(t, "1 — Low", formatScoreCell(&score, riskScoreLevelLabel(score, levels)))
	assert.Equal(t, "Low", riskScoreLevelLabel(5, levels))
	assert.Equal(t, "Medium", riskScoreLevelLabel(12, levels))
	assert.Equal(t, "High", riskScoreLevelLabel(25, levels))
}

func TestRiskScoreLevelLabel_DefaultMatrix(t *testing.T) {
	t.Parallel()

	levels := riskmanagement.DefaultScoreLevels(5, 5)

	assert.Equal(t, "Low", riskScoreLevelLabel(1, levels))
	assert.Equal(t, "Medium", riskScoreLevelLabel(9, levels))
	assert.Equal(t, "High", riskScoreLevelLabel(10, levels))
	assert.Equal(t, "Critical", riskScoreLevelLabel(25, levels))
}

func TestRiskScoreLevelLabel_UsesHighestMatchingMinimum(t *testing.T) {
	t.Parallel()

	// High sits between two lower bands, so only the highest eligible minimum selects it.
	levels := coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 13},
		{Level: coredata.RiskScoreLevelMedium, MinScore: 6},
	}

	assert.Equal(t, "High", riskScoreLevelLabel(13, levels))
	assert.Equal(t, "Medium", riskScoreLevelLabel(7, levels))
	assert.Equal(t, "", riskScoreLevelLabel(1, nil))
}
