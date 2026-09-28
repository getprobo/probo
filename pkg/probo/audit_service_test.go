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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/validator"
)

func TestCreateAuditRequest_Validate_AuditDates(t *testing.T) {
	t.Parallel()

	day := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	later := day.AddDate(0, 0, 14)
	earlier := day.AddDate(0, 0, -1)

	tests := []struct {
		name      string
		start     *time.Time
		end       *time.Time
		wantError bool
	}{
		{
			name:  "same day",
			start: &day,
			end:   &day,
		},
		{
			name:  "end after start",
			start: &day,
			end:   &later,
		},
		{
			name: "end only",
			end:  &day,
		},
		{
			name:      "end before start",
			start:     &day,
			end:       &earlier,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := newCreateAuditRequest()
			req.AuditStartDate = tt.start
			req.AuditEndDate = tt.end

			err := req.Validate()
			if !tt.wantError {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
			require.True(t, ok)
			assert.NotEmpty(t, validationErrors.ByField("audit_end_date"))
			assert.NotEmpty(t, validationErrors.ByCode(validator.ErrorCodeOutOfRange))
		})
	}
}

func TestCreateAuditRequest_Validate_RejectsEqualValidity(t *testing.T) {
	t.Parallel()

	day := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	req := newCreateAuditRequest()
	req.ValidFrom = &day
	req.ValidUntil = &day

	err := req.Validate()
	require.Error(t, err)

	validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
	require.True(t, ok)
	assert.NotEmpty(t, validationErrors.ByField("valid_until"))
}

func TestUpdateAuditRequest_Validate_AuditDates(t *testing.T) {
	t.Parallel()

	day := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	later := day.AddDate(0, 0, 7)
	earlier := day.AddDate(0, 0, -1)

	tests := []struct {
		name      string
		start     *time.Time
		end       *time.Time
		wantError bool
	}{
		{
			name:  "same day",
			start: &day,
			end:   &day,
		},
		{
			name:  "end after start",
			start: &day,
			end:   &later,
		},
		{
			name:      "end before start",
			start:     &day,
			end:       &earlier,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := newUpdateAuditRequest()
			req.AuditStartDate = tt.start
			req.AuditEndDate = tt.end

			err := req.Validate()
			if !tt.wantError {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			validationErrors, ok := errors.AsType[validator.ValidationErrors](err)
			require.True(t, ok)
			assert.NotEmpty(t, validationErrors.ByField("audit_end_date"))
			assert.NotEmpty(t, validationErrors.ByCode(validator.ErrorCodeOutOfRange))
		})
	}
}

func newCreateAuditRequest() CreateAuditRequest {
	tenantID := gid.NewTenantID()

	return CreateAuditRequest{
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		FrameworkID:    gid.New(tenantID, coredata.FrameworkEntityType),
	}
}

func newUpdateAuditRequest() UpdateAuditRequest {
	return UpdateAuditRequest{
		ID: gid.New(gid.NewTenantID(), coredata.AuditEntityType),
	}
}
