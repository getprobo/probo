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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/malaysiapdpa"
	"go.probo.inc/probo/pkg/validator"
)

func validCreateMalaysiaPDPABreachRequest() CreateMalaysiaPDPABreachRequest {
	tenantID := gid.NewTenantID()
	discoveredAt := time.Date(2026, time.August, 4, 8, 0, 0, 0, time.UTC)

	return CreateMalaysiaPDPABreachRequest{
		OrganizationID:       gid.New(tenantID, coredata.OrganizationEntityType),
		Title:                "Compromised customer database",
		DiscoveredAt:         discoveredAt,
		AwarenessAt:          discoveredAt.Add(time.Hour),
		AffectedDataSubjects: 200,
		AffectedDataRecords:  500,
		PersonalDataTypes:    "Names and email addresses",
		NotificationDecision: coredata.MalaysiaPDPABreachNotificationDecisionPending,
		ActorProfileID:       gid.New(tenantID, coredata.MembershipProfileEntityType),
	}
}

func TestCreateMalaysiaPDPABreachRequestValidate(t *testing.T) {
	t.Parallel()

	t.Run("accepts a valid pending assessment", func(t *testing.T) {
		t.Parallel()

		req := validCreateMalaysiaPDPABreachRequest()
		assert.NoError(t, req.Validate())
	})

	t.Run("rejects awareness before discovery", func(t *testing.T) {
		t.Parallel()

		req := validCreateMalaysiaPDPABreachRequest()
		req.AwarenessAt = req.DiscoveredAt.Add(-time.Minute)

		err := req.Validate()
		require.Error(t, err)
		validationErrors, ok := err.(validator.ValidationErrors)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("awareness_at"))
	})

	t.Run("requires rationale for a recorded decision", func(t *testing.T) {
		t.Parallel()

		req := validCreateMalaysiaPDPABreachRequest()
		req.NotificationDecision = coredata.MalaysiaPDPABreachNotificationDecisionNotRequired

		err := req.Validate()
		require.Error(t, err)
		validationErrors, ok := err.(validator.ValidationErrors)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("decision_rationale"))
	})

	t.Run("requires Commissioner notification before data-subject notice", func(t *testing.T) {
		t.Parallel()

		req := validCreateMalaysiaPDPABreachRequest()
		notifiedAt := req.AwarenessAt.Add(time.Hour)
		req.DataSubjectsNotifiedAt = &notifiedAt

		err := req.Validate()
		require.Error(t, err)
		validationErrors, ok := err.(validator.ValidationErrors)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("data_subjects_notified_at"))
	})
}

func TestValidateMalaysiaPDPABreachIncidentRequiresLateEvidence(t *testing.T) {
	t.Parallel()

	awarenessAt := time.Date(2026, time.August, 4, 8, 0, 0, 0, time.UTC)
	commissionerNotifiedAt := awarenessAt.Add(malaysiapdpa.CommissionerNotificationWindow + time.Minute)
	rationale := "Sensitive personal data was involved"

	incident := &coredata.MalaysiaPDPABreachIncident{
		Title:                      "Late notification",
		DiscoveredAt:               awarenessAt.Add(-time.Hour),
		AwarenessAt:                awarenessAt,
		AffectedDataSubjects:       1,
		AffectedDataRecords:        1,
		PersonalDataTypes:          "Medical record",
		SensitivePersonalData:      true,
		SignificantHarm:            true,
		NotificationRecommendation: coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
		NotificationDecision:       coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
		DecisionRationale:          &rationale,
		CommissionerNotifiedAt:     &commissionerNotifiedAt,
	}

	err := validateMalaysiaPDPABreachIncident(incident)
	require.Error(t, err)
	validationErrors, ok := err.(validator.ValidationErrors)
	require.True(t, ok)
	assert.NotEmpty(t, validationErrors.ByField("delayed_notification_reason"))
	assert.NotEmpty(t, validationErrors.ByField("delayed_notification_evidence"))
}

func TestMalaysiaPDPABreachStatusTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		from    coredata.MalaysiaPDPABreachStatus
		to      coredata.MalaysiaPDPABreachStatus
		allowed bool
	}{
		{name: "start assessment", from: coredata.MalaysiaPDPABreachStatusOpen, to: coredata.MalaysiaPDPABreachStatusAssessing, allowed: true},
		{name: "contain assessed incident", from: coredata.MalaysiaPDPABreachStatusAssessing, to: coredata.MalaysiaPDPABreachStatusContained, allowed: true},
		{name: "close contained incident", from: coredata.MalaysiaPDPABreachStatusContained, to: coredata.MalaysiaPDPABreachStatusClosed, allowed: true},
		{name: "reopen closed incident", from: coredata.MalaysiaPDPABreachStatusClosed, to: coredata.MalaysiaPDPABreachStatusAssessing, allowed: true},
		{name: "cannot close open incident", from: coredata.MalaysiaPDPABreachStatusOpen, to: coredata.MalaysiaPDPABreachStatusClosed, allowed: false},
		{name: "cannot repeat status", from: coredata.MalaysiaPDPABreachStatusAssessing, to: coredata.MalaysiaPDPABreachStatusAssessing, allowed: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.allowed, isMalaysiaPDPABreachStatusTransitionAllowed(test.from, test.to))
		})
	}
}

func TestValidateMalaysiaPDPABreachClosure(t *testing.T) {
	t.Parallel()

	t.Run("requires Commissioner notification for significant scale", func(t *testing.T) {
		t.Parallel()

		incident := &coredata.MalaysiaPDPABreachIncident{
			NotificationRecommendation: coredata.MalaysiaPDPABreachNotificationDecisionCommissionerOnly,
			NotificationDecision:       coredata.MalaysiaPDPABreachNotificationDecisionPending,
		}

		err := validateMalaysiaPDPABreachClosure(incident)
		require.Error(t, err)
		validationErrors, ok := err.(validator.ValidationErrors)
		require.True(t, ok)
		assert.NotEmpty(t, validationErrors.ByField("commissioner_notified_at"))
	})

	t.Run("requires both notices for significant harm", func(t *testing.T) {
		t.Parallel()

		notifiedAt := time.Now()
		incident := &coredata.MalaysiaPDPABreachIncident{
			NotificationRecommendation: coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
			NotificationDecision:       coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
			CommissionerNotifiedAt:     &notifiedAt,
		}

		err := validateMalaysiaPDPABreachClosure(incident)
		require.Error(t, err)
		validationErrors, ok := err.(validator.ValidationErrors)
		require.True(t, ok)
		assert.Empty(t, validationErrors.ByField("commissioner_notified_at"))
		assert.NotEmpty(t, validationErrors.ByField("data_subjects_notified_at"))
	})

	t.Run("allows closure when required notices are recorded", func(t *testing.T) {
		t.Parallel()

		notifiedAt := time.Now()
		dataSubjectsNotifiedAt := notifiedAt.Add(time.Hour)
		incident := &coredata.MalaysiaPDPABreachIncident{
			NotificationRecommendation: coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
			NotificationDecision:       coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
			CommissionerNotifiedAt:     &notifiedAt,
			DataSubjectsNotifiedAt:     &dataSubjectsNotifiedAt,
		}

		assert.NoError(t, validateMalaysiaPDPABreachClosure(incident))
	})
}
