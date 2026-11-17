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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func TestCreateRiskAnalysisRequest_Validate_AppliesDefault(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	req := CreateRiskAnalysisRequest{
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Name:           "Quarterly",
		MatrixSize:     &MatrixSize{Rows: 3, Cols: 3},
	}

	require.NoError(t, req.Validate())
	require.NotNil(t, req.ScoreLevels)
	assert.Equal(t, DefaultScoreLevels(3, 3), *req.ScoreLevels)
}

func TestCreateRiskAnalysisRequest_Validate_RejectsMissingScoreLevelsWhenMatrixIsInvalid(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()

	for _, tc := range []struct {
		name string
		size *MatrixSize
	}{
		{name: "missing matrix", size: nil},
		{name: "unsupported size", size: &MatrixSize{Rows: 6, Cols: 6}},
		{name: "unequal sides", size: &MatrixSize{Rows: 3, Cols: 4}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := CreateRiskAnalysisRequest{
				OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
				Name:           "Quarterly",
				MatrixSize:     tc.size,
			}

			var errs validator.ValidationErrors
			require.ErrorAs(t, req.Validate(), &errs)
			scoreErrs := errs.ByField("score_levels")
			require.Len(t, scoreErrs, 1)
			assert.Equal(t, validator.ErrorCodeRequired, scoreErrs[0].Code)
			assert.Nil(t, req.ScoreLevels)
		})
	}
}

func TestCreateRiskAnalysisRequest_Validate_KeepsProvidedLevels(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	levels := coredata.ScoreLevels{
		{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		{Level: coredata.RiskScoreLevelHigh, MinScore: 4},
	}
	req := CreateRiskAnalysisRequest{
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Name:           "Quarterly",
		MatrixSize:     &MatrixSize{Rows: 3, Cols: 3},
		ScoreLevels:    &levels,
	}

	require.NoError(t, req.Validate())
	require.NotNil(t, req.ScoreLevels)
	assert.Equal(t, levels, *req.ScoreLevels)
}

func TestValidateScoreLevels(t *testing.T) {
	t.Parallel()

	t.Run("accepts omitting critical", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelLow, MinScore: 1},
			{Level: coredata.RiskScoreLevelMedium, MinScore: 6},
			{Level: coredata.RiskScoreLevelHigh, MinScore: 13},
		}, 25)
		assert.NoError(t, v.Error())
	})

	t.Run("accepts omitting low", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelHigh, MinScore: 1},
			{Level: coredata.RiskScoreLevelCritical, MinScore: 10},
		}, 25)
		assert.NoError(t, v.Error())
	})

	t.Run("requires at least two levels", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelLow, MinScore: 1},
		}, 25)

		var errs validator.ValidationErrors
		require.ErrorAs(t, v.Error(), &errs)
		assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeTooShort))
	})

	t.Run("rejects levels out of order", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelHigh, MinScore: 1},
			{Level: coredata.RiskScoreLevelLow, MinScore: 2},
		}, 25)

		var errs validator.ValidationErrors
		require.ErrorAs(t, v.Error(), &errs)
		assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeInvalidEnum))
	})

	t.Run("rejects a duplicate level", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelLow, MinScore: 1},
			{Level: coredata.RiskScoreLevelLow, MinScore: 2},
		}, 25)

		var errs validator.ValidationErrors
		require.ErrorAs(t, v.Error(), &errs)
		assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeInvalidEnum))
	})

	t.Run("requires the first minimum to be 1", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelLow, MinScore: 2},
			{Level: coredata.RiskScoreLevelHigh, MinScore: 10},
		}, 25)

		var errs validator.ValidationErrors
		require.ErrorAs(t, v.Error(), &errs)
		assert.NotEmpty(t, errs.ByField("score_levels.0.min_score"))
	})

	t.Run("rejects a minimum equal to the previous level", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelLow, MinScore: 1},
			{Level: coredata.RiskScoreLevelMedium, MinScore: 5},
			{Level: coredata.RiskScoreLevelHigh, MinScore: 5},
		}, 25)

		var errs validator.ValidationErrors
		require.ErrorAs(t, v.Error(), &errs)
		assert.NotEmpty(t, errs.ByField("score_levels.2.min_score"))
		assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeOutOfRange))
	})

	t.Run("rejects a minimum below the previous level", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelLow, MinScore: 1},
			{Level: coredata.RiskScoreLevelMedium, MinScore: 8},
			{Level: coredata.RiskScoreLevelHigh, MinScore: 4},
		}, 25)

		var errs validator.ValidationErrors
		require.ErrorAs(t, v.Error(), &errs)
		assert.NotEmpty(t, errs.ByField("score_levels.2.min_score"))
		assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeOutOfRange))
	})

	t.Run("rejects a minimum above the matrix", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		validateScoreLevels(v, coredata.ScoreLevels{
			{Level: coredata.RiskScoreLevelLow, MinScore: 1},
			{Level: coredata.RiskScoreLevelHigh, MinScore: 10},
		}, 9)

		var errs validator.ValidationErrors
		require.ErrorAs(t, v.Error(), &errs)
		assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeOutOfRange))
	})
}
