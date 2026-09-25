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

package riskmanagement_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/riskmanagement"
)

func TestNetScores(t *testing.T) {
	t.Parallel()

	t.Run("no measures stays at inherent", func(t *testing.T) {
		t.Parallel()

		for _, action := range []coredata.RiskTreatment{
			coredata.RiskTreatmentMitigated,
			coredata.RiskTreatmentAvoided,
			coredata.RiskTreatmentTransferred,
			coredata.RiskTreatmentAccepted,
		} {
			likelihood, impact, score := riskmanagement.NetScores(
				plan(action, 4, 4, 1, 1),
				riskmanagement.TreatmentProgress{},
			)
			assert.Equal(t, 4, likelihood, action)
			assert.Equal(t, 4, impact, action)
			assert.Equal(t, 16, score, action)
		}
	})

	t.Run("all measures implemented is residual", func(t *testing.T) {
		t.Parallel()

		for _, action := range []coredata.RiskTreatment{
			coredata.RiskTreatmentMitigated,
			coredata.RiskTreatmentAvoided,
			coredata.RiskTreatmentTransferred,
		} {
			likelihood, impact, score := riskmanagement.NetScores(
				plan(action, 4, 4, 1, 2),
				riskmanagement.TreatmentProgress{Done: 2, Total: 2},
			)
			assert.Equal(t, 1, likelihood, action)
			assert.Equal(t, 2, impact, action)
			assert.Equal(t, 2, score, action)
		}
	})

	t.Run("partial progress stays at inherent", func(t *testing.T) {
		t.Parallel()

		likelihood, impact, score := riskmanagement.NetScores(
			plan(coredata.RiskTreatmentMitigated, 4, 4, 2, 2),
			riskmanagement.TreatmentProgress{Done: 1, Total: 2},
		)
		assert.Equal(t, 4, likelihood)
		assert.Equal(t, 4, impact)
		assert.Equal(t, 16, score)
	})
}

func plan(
	action coredata.RiskTreatment,
	inherentL, inherentI, residualL, residualI int,
) *coredata.TreatmentPlan {
	return &coredata.TreatmentPlan{
		Treatment:          action,
		InherentLikelihood: inherentL,
		InherentImpact:     inherentI,
		ResidualLikelihood: residualL,
		ResidualImpact:     residualI,
	}
}
