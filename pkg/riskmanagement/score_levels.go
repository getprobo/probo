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

package riskmanagement

import (
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/riskmanagement/scorelevel"
	"go.probo.inc/probo/pkg/validator"
)

// DefaultScoreLevels is the historical ratio mapping: low below 5/25 of the
// maximum score, medium below 10/25, high below 15/25, and critical after that.
func DefaultScoreLevels(rows, cols int) coredata.ScoreLevels {
	return scorelevel.DefaultScoreLevels(rows, cols)
}

func (r *CreateRiskAnalysisRequest) applyDefaultScoreLevels() {
	if r.ScoreLevels != nil || matrixMaxScore(r.MatrixSize) == 0 {
		return
	}

	levels := DefaultScoreLevels(r.MatrixSize.Rows, r.MatrixSize.Cols)
	r.ScoreLevels = &levels
}

func (r *CreateRiskAnalysisRequest) scoreLevelsForValidation() coredata.ScoreLevels {
	if r.ScoreLevels == nil {
		return nil
	}

	return *r.ScoreLevels
}

func validateScoreLevels(v *validator.Validator, levels coredata.ScoreLevels, maxScore int) {
	v.Check(levels, "score_levels", validator.Required())

	if levels == nil {
		return
	}

	if len(levels) < 2 {
		v.Check(
			levels,
			"score_levels",
			func(any) *validator.ValidationError {
				return &validator.ValidationError{
					Code:    validator.ErrorCodeTooShort,
					Message: "must include at least two levels",
				}
			},
		)

		return
	}

	rank := map[coredata.RiskScoreLevel]int{}
	for i, level := range coredata.RiskScoreLevels() {
		rank[level] = i
	}

	previousRank := -1
	previousMin := 0

	for i, mapping := range levels {
		field := fmt.Sprintf("score_levels.%d", i)
		levelRank, known := rank[mapping.Level]

		if !known || levelRank <= previousRank {
			v.Check(
				mapping.Level,
				field+".level",
				func(any) *validator.ValidationError {
					return &validator.ValidationError{
						Code:    validator.ErrorCodeInvalidEnum,
						Message: "must be LOW, MEDIUM, HIGH, or CRITICAL, in that order, without duplicates",
					}
				},
			)

			return
		}

		previousRank = levelRank

		if mapping.MinScore < 1 || (maxScore > 0 && mapping.MinScore > maxScore) {
			message := "must be at least 1"
			if maxScore > 0 {
				message = fmt.Sprintf("must be between 1 and %d", maxScore)
			}

			v.Check(
				mapping.MinScore,
				field+".min_score",
				func(any) *validator.ValidationError {
					return &validator.ValidationError{
						Code:    validator.ErrorCodeOutOfRange,
						Message: message,
					}
				},
			)

			continue
		}

		if i == 0 && mapping.MinScore != 1 {
			v.Check(
				mapping.MinScore,
				field+".min_score",
				func(any) *validator.ValidationError {
					return &validator.ValidationError{
						Code:    validator.ErrorCodeOutOfRange,
						Message: "must be 1 so every score maps to a level",
					}
				},
			)
		}

		if i > 0 && mapping.MinScore <= previousMin {
			v.Check(
				mapping.MinScore,
				field+".min_score",
				func(any) *validator.ValidationError {
					return &validator.ValidationError{
						Code:    validator.ErrorCodeOutOfRange,
						Message: "must be greater than the previous level",
					}
				},
			)
		}

		previousMin = mapping.MinScore
	}
}
