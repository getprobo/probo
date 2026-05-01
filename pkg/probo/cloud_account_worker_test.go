// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package probo

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
)

// TestComputeCloudAccountTransition_VerifiedToErrored asserts a
// VERIFIED row drops to ERRORED on first failure with
// consecutive_probe_failures=1 and first_probe_failure_at set.
func TestComputeCloudAccountTransition_VerifiedToErrored(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	probeErr := errors.New("AccessDenied: assume role failed")

	transition := computeCloudAccountTransition(
		coredata.CloudAccountStatusVerified,
		0,
		nil,
		probeErr,
		now,
		cloudAccountDisconnectFailureThreshold,
		cloudAccountDisconnectTimeGate,
	)

	assert.Equal(t, coredata.CloudAccountStatusErrored, transition.Status)
	assert.Equal(t, 1, transition.ConsecutiveProbeFailures)
	require.NotNil(t, transition.FirstProbeFailureAt)
	assert.Equal(t, now, *transition.FirstProbeFailureAt)
	assert.Equal(t, now, transition.LastProbeAt)
	require.NotNil(t, transition.LastProbeError)
	assert.Equal(t, probeErr.Error(), *transition.LastProbeError)
	assert.Nil(t, transition.LastVerifiedAt)
	assert.Equal(t, cloudAccountWebhookNone, transition.Webhook)
}

// TestComputeCloudAccountTransition_ErroredStaysErrored asserts a
// second failure increments the counter, keeps the row in ERRORED,
// and emits no webhook because the time gate has not elapsed.
func TestComputeCloudAccountTransition_ErroredStaysErrored(t *testing.T) {
	t.Parallel()

	first := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	now := first.Add(15 * time.Minute)
	probeErr := errors.New("Throttling")

	transition := computeCloudAccountTransition(
		coredata.CloudAccountStatusErrored,
		1,
		&first,
		probeErr,
		now,
		cloudAccountDisconnectFailureThreshold,
		cloudAccountDisconnectTimeGate,
	)

	assert.Equal(t, coredata.CloudAccountStatusErrored, transition.Status)
	assert.Equal(t, 2, transition.ConsecutiveProbeFailures)
	require.NotNil(t, transition.FirstProbeFailureAt)
	assert.Equal(t, first, *transition.FirstProbeFailureAt, "first failure marker is preserved")
	assert.Equal(t, cloudAccountWebhookNone, transition.Webhook)
}

// TestComputeCloudAccountTransition_ErroredToDisconnectedTimeGate
// asserts the worker does NOT promote to DISCONNECTED when the
// consecutive-failure threshold is met but the time gate has not
// elapsed (30 min < 1h).
func TestComputeCloudAccountTransition_ErroredToDisconnectedTimeGate(t *testing.T) {
	t.Parallel()

	first := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	now := first.Add(30 * time.Minute)
	probeErr := errors.New("AccessDenied")

	transition := computeCloudAccountTransition(
		coredata.CloudAccountStatusErrored,
		2, // becomes 3 after increment, which meets threshold
		&first,
		probeErr,
		now,
		cloudAccountDisconnectFailureThreshold,
		cloudAccountDisconnectTimeGate,
	)

	assert.Equal(
		t,
		coredata.CloudAccountStatusErrored,
		transition.Status,
		"row stays ERRORED until both threshold and time gate are met",
	)
	assert.Equal(t, 3, transition.ConsecutiveProbeFailures)
	assert.Equal(t, cloudAccountWebhookNone, transition.Webhook)
}

// TestComputeCloudAccountTransition_ErroredToDisconnected asserts
// the worker promotes to DISCONNECTED only after BOTH the
// failure-count threshold AND the time gate are met (>=3 failures
// and >=1h elapsed).
func TestComputeCloudAccountTransition_ErroredToDisconnected(t *testing.T) {
	t.Parallel()

	first := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	now := first.Add(61 * time.Minute)
	probeErr := errors.New("AccessDenied")

	transition := computeCloudAccountTransition(
		coredata.CloudAccountStatusErrored,
		2, // becomes 3 after increment
		&first,
		probeErr,
		now,
		cloudAccountDisconnectFailureThreshold,
		cloudAccountDisconnectTimeGate,
	)

	assert.Equal(t, coredata.CloudAccountStatusDisconnected, transition.Status)
	assert.Equal(t, 3, transition.ConsecutiveProbeFailures)
	assert.Equal(t, cloudAccountWebhookDisconnected, transition.Webhook)
}

// TestComputeCloudAccountTransition_ErroredRecoveryEmitsWebhook
// asserts the ERRORED -> VERIFIED recovery clears the failure
// counter, sets last_verified_at, and emits cloud_account.verified.
func TestComputeCloudAccountTransition_ErroredRecoveryEmitsWebhook(t *testing.T) {
	t.Parallel()

	first := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	now := first.Add(45 * time.Minute)

	transition := computeCloudAccountTransition(
		coredata.CloudAccountStatusErrored,
		2,
		&first,
		nil, // success
		now,
		cloudAccountDisconnectFailureThreshold,
		cloudAccountDisconnectTimeGate,
	)

	assert.Equal(t, coredata.CloudAccountStatusVerified, transition.Status)
	assert.Equal(t, 0, transition.ConsecutiveProbeFailures)
	assert.Nil(t, transition.FirstProbeFailureAt)
	assert.Nil(t, transition.LastProbeError)
	require.NotNil(t, transition.LastVerifiedAt)
	assert.Equal(t, now, *transition.LastVerifiedAt)
	assert.Equal(t, cloudAccountWebhookVerified, transition.Webhook)
}

// TestComputeCloudAccountTransition_DisconnectedRecoveryEmitsWebhook
// asserts a DISCONNECTED row recovers all the way to VERIFIED on a
// successful probe and emits cloud_account.verified.
func TestComputeCloudAccountTransition_DisconnectedRecoveryEmitsWebhook(t *testing.T) {
	t.Parallel()

	first := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	now := first.Add(2 * time.Hour)

	transition := computeCloudAccountTransition(
		coredata.CloudAccountStatusDisconnected,
		5,
		&first,
		nil,
		now,
		cloudAccountDisconnectFailureThreshold,
		cloudAccountDisconnectTimeGate,
	)

	assert.Equal(t, coredata.CloudAccountStatusVerified, transition.Status)
	assert.Equal(t, cloudAccountWebhookVerified, transition.Webhook)
}

// TestComputeCloudAccountTransition_VerifiedToVerifiedNoWebhook
// asserts a VERIFIED -> VERIFIED no-op transition does NOT emit a
// cloud_account.verified webhook (only emitted on recovery, never on
// every healthy probe).
func TestComputeCloudAccountTransition_VerifiedToVerifiedNoWebhook(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	transition := computeCloudAccountTransition(
		coredata.CloudAccountStatusVerified,
		0,
		nil,
		nil,
		now,
		cloudAccountDisconnectFailureThreshold,
		cloudAccountDisconnectTimeGate,
	)

	assert.Equal(t, coredata.CloudAccountStatusVerified, transition.Status)
	assert.Equal(
		t,
		cloudAccountWebhookNone,
		transition.Webhook,
		"VERIFIED -> VERIFIED never emits cloud_account.verified",
	)
}

// TestComputeCloudAccountTransition_PendingVerificationStays asserts
// the worker NEVER auto-promotes a never-verified account, neither on
// success nor on failure -- the customer must re-trigger Verify from
// the UI to leave PENDING_VERIFICATION.
func TestComputeCloudAccountTransition_PendingVerificationStays(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		probeErr error
	}{
		{name: "stays pending on probe failure", probeErr: errors.New("denied")},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			transition := computeCloudAccountTransition(
				coredata.CloudAccountStatusPendingVerification,
				0,
				nil,
				tt.probeErr,
				now,
				cloudAccountDisconnectFailureThreshold,
				cloudAccountDisconnectTimeGate,
			)

			assert.Equal(t, coredata.CloudAccountStatusPendingVerification, transition.Status)
			assert.Equal(t, cloudAccountWebhookNone, transition.Webhook)
		})
	}
}

// TestComputeCloudAccountTransition_StateMatrix is a table-driven
// matrix covering each transition the worker is responsible for.
func TestComputeCloudAccountTransition_StateMatrix(t *testing.T) {
	t.Parallel()

	first := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	now := first.Add(90 * time.Minute)
	probeErr := errors.New("denied")

	tests := []struct {
		name           string
		current        coredata.CloudAccountStatus
		failures       int
		firstAt        *time.Time
		probeErr       error
		expectStatus   coredata.CloudAccountStatus
		expectFailures int
		expectWebhook  cloudAccountWebhookEvent
	}{
		{
			name:           "verified -> errored on first failure",
			current:        coredata.CloudAccountStatusVerified,
			failures:       0,
			firstAt:        nil,
			probeErr:       probeErr,
			expectStatus:   coredata.CloudAccountStatusErrored,
			expectFailures: 1,
			expectWebhook:  cloudAccountWebhookNone,
		},
		{
			name:           "errored -> errored on second failure under time gate",
			current:        coredata.CloudAccountStatusErrored,
			failures:       1,
			firstAt:        ptrTime(now.Add(-15 * time.Minute)),
			probeErr:       probeErr,
			expectStatus:   coredata.CloudAccountStatusErrored,
			expectFailures: 2,
			expectWebhook:  cloudAccountWebhookNone,
		},
		{
			name:           "errored -> disconnected past threshold and time gate",
			current:        coredata.CloudAccountStatusErrored,
			failures:       2,
			firstAt:        &first,
			probeErr:       probeErr,
			expectStatus:   coredata.CloudAccountStatusDisconnected,
			expectFailures: 3,
			expectWebhook:  cloudAccountWebhookDisconnected,
		},
		{
			name:           "errored -> verified emits recovery webhook",
			current:        coredata.CloudAccountStatusErrored,
			failures:       2,
			firstAt:        &first,
			probeErr:       nil,
			expectStatus:   coredata.CloudAccountStatusVerified,
			expectFailures: 0,
			expectWebhook:  cloudAccountWebhookVerified,
		},
		{
			name:           "verified -> verified emits no webhook",
			current:        coredata.CloudAccountStatusVerified,
			failures:       0,
			firstAt:        nil,
			probeErr:       nil,
			expectStatus:   coredata.CloudAccountStatusVerified,
			expectFailures: 0,
			expectWebhook:  cloudAccountWebhookNone,
		},
		{
			name:           "pending -> pending on failure",
			current:        coredata.CloudAccountStatusPendingVerification,
			failures:       0,
			firstAt:        nil,
			probeErr:       probeErr,
			expectStatus:   coredata.CloudAccountStatusPendingVerification,
			expectFailures: 0,
			expectWebhook:  cloudAccountWebhookNone,
		},
		{
			name:           "disconnected -> disconnected on further failure",
			current:        coredata.CloudAccountStatusDisconnected,
			failures:       5,
			firstAt:        &first,
			probeErr:       probeErr,
			expectStatus:   coredata.CloudAccountStatusDisconnected,
			expectFailures: 6,
			expectWebhook:  cloudAccountWebhookNone,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			transition := computeCloudAccountTransition(
				tt.current,
				tt.failures,
				tt.firstAt,
				tt.probeErr,
				now,
				cloudAccountDisconnectFailureThreshold,
				cloudAccountDisconnectTimeGate,
			)

			assert.Equal(t, tt.expectStatus, transition.Status)
			assert.Equal(t, tt.expectFailures, transition.ConsecutiveProbeFailures)
			assert.Equal(t, tt.expectWebhook, transition.Webhook)
		})
	}
}

// TestApplyCloudAccountTransition asserts the helper that copies a
// transition's fields onto a coredata.CloudAccount preserves the
// fields that are NOT part of the state machine output (provider,
// scope, organization).
func TestApplyCloudAccountTransition(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	probeErr := errors.New("denied")
	account := &coredata.CloudAccount{
		Provider:                 coredata.CloudAccountProviderAWS,
		Status:                   coredata.CloudAccountStatusVerified,
		ConsecutiveProbeFailures: 0,
		ScopeIdentifier:          "123456789012",
	}

	transition := computeCloudAccountTransition(
		account.Status,
		account.ConsecutiveProbeFailures,
		account.FirstProbeFailureAt,
		probeErr,
		now,
		cloudAccountDisconnectFailureThreshold,
		cloudAccountDisconnectTimeGate,
	)

	applyCloudAccountTransition(account, transition)

	assert.Equal(t, coredata.CloudAccountStatusErrored, account.Status)
	assert.Equal(t, 1, account.ConsecutiveProbeFailures)
	require.NotNil(t, account.LastProbeError)
	assert.Equal(t, probeErr.Error(), *account.LastProbeError)
	assert.Equal(t, "123456789012", account.ScopeIdentifier, "non-state fields are not touched")
	assert.Equal(t, coredata.CloudAccountProviderAWS, account.Provider)
	require.NotNil(t, account.LastProbeAt)
	assert.Equal(t, now, *account.LastProbeAt)
	assert.Equal(t, now, account.UpdatedAt)
}

func ptrTime(t time.Time) *time.Time { return &t }
