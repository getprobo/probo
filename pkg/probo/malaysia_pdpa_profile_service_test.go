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

package probo_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/validator"
)

func validMalaysiaPDPAProfileRequest() probo.UpdateMalaysiaPDPAProfileRequest {
	tenantID := gid.NewTenantID()

	return probo.UpdateMalaysiaPDPAProfileRequest{
		OrganizationID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TotalDataSubjects:     100,
		SensitiveDataSubjects: 10,
		AssessedByProfileID:   gid.New(tenantID, coredata.MembershipProfileEntityType),
	}
}

func TestUpdateMalaysiaPDPAProfileRequest_Validate(t *testing.T) {
	t.Parallel()

	t.Run(
		"accepts a valid assessment",
		func(t *testing.T) {
			t.Parallel()

			req := validMalaysiaPDPAProfileRequest()

			assert.NoError(t, req.Validate())
		},
	)

	t.Run(
		"rejects a sensitive count above the total",
		func(t *testing.T) {
			t.Parallel()

			req := validMalaysiaPDPAProfileRequest()
			req.SensitiveDataSubjects = req.TotalDataSubjects + 1

			err := req.Validate()
			require.Error(t, err)

			validationErrors, ok := err.(validator.ValidationErrors)
			require.True(t, ok)
			assert.NotEmpty(t, validationErrors.ByField("sensitive_data_subjects"))
		},
	)

	t.Run(
		"requires an appointment date with a DPO profile",
		func(t *testing.T) {
			t.Parallel()

			req := validMalaysiaPDPAProfileRequest()
			dpoProfileID := gid.New(req.OrganizationID.TenantID(), coredata.MembershipProfileEntityType)
			req.DPOProfileID = &dpoProfileID

			err := req.Validate()
			require.Error(t, err)

			validationErrors, ok := err.(validator.ValidationErrors)
			require.True(t, ok)
			assert.NotEmpty(t, validationErrors.ByField("dpo_profile_id"))
		},
	)

	t.Run(
		"rejects notification before appointment",
		func(t *testing.T) {
			t.Parallel()

			req := validMalaysiaPDPAProfileRequest()
			dpoProfileID := gid.New(req.OrganizationID.TenantID(), coredata.MembershipProfileEntityType)
			appointedAt := time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)
			notifiedAt := appointedAt.Add(-time.Hour)
			req.DPOProfileID = &dpoProfileID
			req.DPOAppointedAt = &appointedAt
			req.CommissionerNotifiedAt = &notifiedAt

			err := req.Validate()
			require.Error(t, err)

			validationErrors, ok := err.(validator.ValidationErrors)
			require.True(t, ok)
			assert.NotEmpty(t, validationErrors.ByField("commissioner_notified_at"))
		},
	)
}
