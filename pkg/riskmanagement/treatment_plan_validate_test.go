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

func TestRequireAcceptedScoresMatchInherent(t *testing.T) {
	t.Parallel()

	t.Run("mitigated allows different residual", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		requireAcceptedScoresMatchInherent(
			v,
			coredata.RiskTreatmentMitigated,
			4,
			4,
			1,
			2,
		)
		assert.NoError(t, v.Error())
	})

	t.Run("accepted matching residual passes", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		requireAcceptedScoresMatchInherent(
			v,
			coredata.RiskTreatmentAccepted,
			4,
			4,
			4,
			4,
		)
		assert.NoError(t, v.Error())
	})

	t.Run("accepted mismatched residual fails", func(t *testing.T) {
		t.Parallel()

		v := validator.New()
		requireAcceptedScoresMatchInherent(
			v,
			coredata.RiskTreatmentAccepted,
			4,
			4,
			1,
			2,
		)
		err := v.Error()
		require.Error(t, err)

		var errs validator.ValidationErrors
		require.ErrorAs(t, err, &errs)
		assert.NotEmpty(t, errs.ByField("residual_likelihood"))
		assert.NotEmpty(t, errs.ByField("residual_impact"))
		assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeInvalidFormat))
	})
}

func TestCreateTreatmentPlanRequest_OwnerRequired(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	req := CreateTreatmentPlanRequest{
		RiskID:             gid.New(tenantID, coredata.RiskEntityType),
		RiskAnalysisID:     gid.New(tenantID, coredata.RiskAnalysisEntityType),
		Treatment:          coredata.RiskTreatmentMitigated,
		InherentLikelihood: 2,
		InherentImpact:     3,
	}

	err := req.Validate()
	require.Error(t, err)

	var errs validator.ValidationErrors
	require.ErrorAs(t, err, &errs)
	assert.NotEmpty(t, errs.ByField("owner_id"))
	assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeInvalidGID))
}

func TestUpdateTreatmentPlanRequest_RejectsClearedOwner(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()

	var cleared *gid.GID

	req := UpdateTreatmentPlanRequest{
		ID:      gid.New(tenantID, coredata.TreatmentPlanEntityType),
		OwnerID: &cleared,
	}

	err := req.Validate()
	require.Error(t, err)

	var errs validator.ValidationErrors
	require.ErrorAs(t, err, &errs)
	assert.NotEmpty(t, errs.ByField("owner_id"))
	assert.NotEmpty(t, errs.ByCode(validator.ErrorCodeRequired))
}
