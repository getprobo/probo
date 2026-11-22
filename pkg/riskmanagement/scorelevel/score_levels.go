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

package scorelevel

import "go.probo.inc/probo/pkg/coredata"

// DefaultScoreLevels is the historical ratio mapping: low below 5/25 of the
// maximum score, medium below 10/25, high below 15/25, and critical after that.
func DefaultScoreLevels(rows, cols int) coredata.ScoreLevels {
	maxScore := rows * cols
	if maxScore <= 0 {
		maxScore = 25
	}

	return coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelMedium, MinScore: ceilRatio(maxScore, 5)},
		{Level: coredata.RiskScoreLevelHigh, MinScore: ceilRatio(maxScore, 10)},
		{Level: coredata.RiskScoreLevelCritical, MinScore: ceilRatio(maxScore, 15)},
	}
}

func ceilRatio(maxScore, numerator int) int {
	return (maxScore*numerator + 24) / 25
}
