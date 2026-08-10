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
	AiSystemFilter struct {
		status             *AiSystemStatus
		riskClassification *AiSystemRiskClassification
	}
)

func NewAiSystemFilter(
	status *AiSystemStatus,
	riskClassification *AiSystemRiskClassification,
) *AiSystemFilter {
	return &AiSystemFilter{
		status:             status,
		riskClassification: riskClassification,
	}
}

func (f *AiSystemFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"has_status_filter":              false,
		"filter_status":                  nil,
		"has_risk_classification_filter": false,
		"filter_risk_classification":     nil,
	}

	if f.status != nil {
		args["has_status_filter"] = true
		args["filter_status"] = string(*f.status)
	}

	if f.riskClassification != nil {
		args["has_risk_classification_filter"] = true
		args["filter_risk_classification"] = string(*f.riskClassification)
	}

	return args
}

func (f *AiSystemFilter) SQLFragment() string {
	return `
(
    CASE
        WHEN @has_status_filter::boolean = false THEN TRUE
        WHEN @has_status_filter::boolean = true THEN
            status = @filter_status::ai_system_statuses
        ELSE TRUE
    END
    AND
    CASE
        WHEN @has_risk_classification_filter::boolean = false THEN TRUE
        WHEN @has_risk_classification_filter::boolean = true THEN
            risk_classification = @filter_risk_classification::ai_system_risk_classifications
        ELSE TRUE
    END
)`
}
