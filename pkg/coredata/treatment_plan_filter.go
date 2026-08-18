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
	"github.com/jackc/pgx/v5"
)

type (
	TreatmentPlanFilter struct {
		scoreType  *TreatmentPlanScoreType
		likelihood *int
		impact     *int
	}
)

func NewTreatmentPlanFilter(
	scoreType *TreatmentPlanScoreType,
	likelihood *int,
	impact *int,
) *TreatmentPlanFilter {
	return &TreatmentPlanFilter{
		scoreType:  scoreType,
		likelihood: likelihood,
		impact:     impact,
	}
}

func (f *TreatmentPlanFilter) ScoreType() *TreatmentPlanScoreType {
	if f == nil {
		return nil
	}

	return f.scoreType
}

func (f *TreatmentPlanFilter) Likelihood() *int {
	if f == nil {
		return nil
	}

	return f.likelihood
}

func (f *TreatmentPlanFilter) Impact() *int {
	if f == nil {
		return nil
	}

	return f.impact
}

func (f *TreatmentPlanFilter) IsEmpty() bool {
	if f == nil {
		return true
	}

	return f.scoreType == nil && f.likelihood == nil && f.impact == nil
}

func (f *TreatmentPlanFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_score_type":          nil,
		"filter_likelihood":          nil,
		"filter_impact":              nil,
		"filter_score_type_inherent": TreatmentPlanScoreTypeInherent,
		"filter_score_type_net":      TreatmentPlanScoreTypeNet,
		"filter_score_type_residual": TreatmentPlanScoreTypeResidual,
		"filter_net_implemented":     MeasureStateImplemented,
	}
	if f.scoreType != nil {
		args["filter_score_type"] = string(*f.scoreType)
	}

	if f.likelihood != nil {
		args["filter_likelihood"] = *f.likelihood
	}

	if f.impact != nil {
		args["filter_impact"] = *f.impact
	}

	return args
}

func (f *TreatmentPlanFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_score_type::text IS NULL
			OR @filter_likelihood::int IS NULL
			OR @filter_impact::int IS NULL
		THEN TRUE
		WHEN @filter_score_type::text = @filter_score_type_inherent::text THEN
			inherent_likelihood = @filter_likelihood
			AND inherent_impact = @filter_impact
		WHEN @filter_score_type::text = @filter_score_type_residual::text THEN
			residual_likelihood = @filter_likelihood
			AND residual_impact = @filter_impact
		WHEN @filter_score_type::text = @filter_score_type_net::text THEN
			CASE
				WHEN ` + treatmentPlanAllMeasuresImplementedSQL() + ` THEN
					residual_likelihood = @filter_likelihood
					AND residual_impact = @filter_impact
				ELSE
					inherent_likelihood = @filter_likelihood
					AND inherent_impact = @filter_impact
			END
		ELSE TRUE
	END
)`
}

func treatmentPlanAllMeasuresImplementedSQL() string {
	return `(
EXISTS (
	SELECT 1
	FROM treatment_plans_measures tpm
	WHERE tpm.treatment_plan_id = id
)
AND NOT EXISTS (
	SELECT 1
	FROM treatment_plans_measures tpm
	INNER JOIN measures m ON m.id = tpm.measure_id
	WHERE tpm.treatment_plan_id = id
		AND m.state IS DISTINCT FROM @filter_net_implemented
)
)`
}
