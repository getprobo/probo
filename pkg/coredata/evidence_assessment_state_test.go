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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// These integration tests exercise the assessment status state machine — the
// concurrency-sensitive guards the evidence-assessment worker relies on. They
// reuse newTestPgClient (access_entry_upsert_test.go) and skip when
// PROBO_TEST_PG_URL is unset, so `make test` stays a pure unit run.
//
// They lock in particular the claim-ownership guard on the terminal
// transitions: a worker whose claim was recycled by stale recovery and
// re-claimed by another worker must NOT clobber the new owner's PROCESSING
// row when its late SetAssessmentCompleted/SetAssessmentFailed fires.
//
// LoadNextPendingAssessmentForUpdateSkipLocked is intentionally not asserted
// here: it scans globally (no tenant scope), so it cannot be isolated against
// other rows in a shared parallel test database; the worker integration path
// covers its selection.

type assessmentEvidenceFixture struct {
	scope          *coredata.Scope
	organizationID gid.GID
	measureID      gid.GID
}

// seedAssessmentFixture bootstraps the organization and measure that an
// evidence row's FKs require.
func seedAssessmentFixture(t *testing.T, ctx context.Context, client *pg.Client) assessmentEvidenceFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	measureID := gid.New(tenantID, coredata.MeasureEntityType)
	now := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Evidence Assessment Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		measure := &coredata.Measure{
			ID:             measureID,
			OrganizationID: organizationID,
			Category:       "Test",
			Name:           "Evidence Assessment Test Measure",
			State:          coredata.MeasureStateNotImplemented,
			ReferenceID:    "measure-assessment-" + measureID.String(),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		return measure.Insert(ctx, tx, scope)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(ctx, `DELETE FROM evidences WHERE measure_id = $1`, measureID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DELETE FROM measures WHERE id = $1`, measureID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID); err != nil {
				return err
			}

			return nil
		})
	})

	return assessmentEvidenceFixture{
		scope:          scope,
		organizationID: organizationID,
		measureID:      measureID,
	}
}

// insertPendingEvidence inserts a PENDING evidence with a file attached.
func (fx assessmentEvidenceFixture) insertPendingEvidence(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	referenceID string,
	createdAt time.Time,
) gid.GID {
	t.Helper()

	tenantID := fx.scope.GetTenantID()
	evidenceID := gid.New(tenantID, coredata.EvidenceEntityType)
	fileID := gid.New(tenantID, coredata.FileEntityType)

	evidence := coredata.Evidence{
		ID:               evidenceID,
		OrganizationID:   fx.organizationID,
		MeasureID:        fx.measureID,
		State:            coredata.EvidenceStateFulfilled,
		ReferenceID:      referenceID,
		Type:             coredata.EvidenceTypeFile,
		EvidenceFileID:   &fileID,
		AssessmentStatus: coredata.EvidenceAssessmentStatusPending,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return evidence.Insert(ctx, tx, fx.scope)
	}))

	return evidenceID
}

func loadEvidence(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	id gid.GID,
) coredata.Evidence {
	t.Helper()

	var e coredata.Evidence
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return e.LoadByID(ctx, conn, scope, id)
	}))

	return e
}

// claimEvidence transitions a row to PROCESSING with the given started_at,
// mirroring what the worker's Claim does, and returns the claimed snapshot
// (carrying that started_at) so the caller can drive a terminal transition.
func claimEvidence(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	id gid.GID,
	startedAt time.Time,
) coredata.Evidence {
	t.Helper()

	e := loadEvidence(t, ctx, client, scope, id)
	e.AssessmentStatus = coredata.EvidenceAssessmentStatusProcessing
	e.AssessmentProcessingStartedAt = &startedAt
	e.UpdatedAt = startedAt

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return e.Update(ctx, tx, scope)
	}))

	return e
}

func TestEvidence_SetAssessmentCompleted_TransitionsAndPreservesUserColumns(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedAssessmentFixture(t, ctx, client)

	t0 := time.Now().UTC().Truncate(time.Microsecond)
	id := fx.insertPendingEvidence(t, ctx, client, "ev-complete-"+t0.Format(time.RFC3339Nano), t0)

	claimedAt := t0.Add(time.Minute)
	claimed := claimEvidence(t, ctx, client, fx.scope, id, claimedAt)

	// Simulate a concurrent user edit to a user-owned column while the
	// assessment ran; the terminal write must not clobber it.
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE evidences SET url = $1 WHERE id = $2`, "https://edited.example", id)
		return err
	}))

	summary := "Admin console shows enforced MFA."
	claimed.Description = &summary
	require.NoError(t, claimed.SetAssessment(map[string]any{"summary": summary, "confidence": "HIGH"}))

	var updated bool
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error
		updated, err = claimed.SetAssessmentCompleted(ctx, tx, fx.scope)
		return err
	}))
	assert.True(t, updated, "completing a PROCESSING row owned by this claim must update it")

	loaded := loadEvidence(t, ctx, client, fx.scope, id)
	assert.Equal(t, coredata.EvidenceAssessmentStatusCompleted, loaded.AssessmentStatus)
	require.NotNil(t, loaded.Description)
	assert.Equal(t, summary, *loaded.Description)
	assert.NotEmpty(t, loaded.Assessment)
	assert.Nil(t, loaded.AssessmentProcessingStartedAt)
	assert.Equal(t, "https://edited.example", loaded.URL, "a user edit during the run must be preserved")
}

func TestEvidence_SetAssessmentCompleted_NoOpOnSupersededClaim(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedAssessmentFixture(t, ctx, client)

	t0 := time.Now().UTC().Truncate(time.Microsecond)
	id := fx.insertPendingEvidence(t, ctx, client, "ev-superseded-complete-"+t0.Format(time.RFC3339Nano), t0)

	// Worker A claims the row.
	firstClaim := claimEvidence(t, ctx, client, fx.scope, id, t0.Add(time.Minute))

	// Stale recovery recycled A's claim and worker B re-claimed it with a
	// fresh started_at — the row is now owned by B.
	secondClaimAt := t0.Add(3 * time.Minute)
	_ = claimEvidence(t, ctx, client, fx.scope, id, secondClaimAt)

	// Worker A finishes late and tries to commit its result.
	summary := "stale result from A"
	firstClaim.Description = &summary
	require.NoError(t, firstClaim.SetAssessment(map[string]any{"summary": summary}))

	var updated bool
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error
		updated, err = firstClaim.SetAssessmentCompleted(ctx, tx, fx.scope)
		return err
	}))
	assert.False(t, updated, "a superseded claim must not overwrite the re-claimed row")

	loaded := loadEvidence(t, ctx, client, fx.scope, id)
	assert.Equal(t, coredata.EvidenceAssessmentStatusProcessing, loaded.AssessmentStatus, "row stays owned by the re-claim")
	assert.Empty(t, loaded.Assessment, "the stale result must not be persisted")
	require.NotNil(t, loaded.AssessmentProcessingStartedAt)
	assert.WithinDuration(t, secondClaimAt, *loaded.AssessmentProcessingStartedAt, time.Second)
}

func TestEvidence_SetAssessmentFailed_TransitionsOwningClaim(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedAssessmentFixture(t, ctx, client)

	t0 := time.Now().UTC().Truncate(time.Microsecond)
	id := fx.insertPendingEvidence(t, ctx, client, "ev-fail-"+t0.Format(time.RFC3339Nano), t0)

	claimed := claimEvidence(t, ctx, client, fx.scope, id, t0.Add(time.Minute))

	var updated bool
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error
		updated, err = claimed.SetAssessmentFailed(ctx, tx, fx.scope)
		return err
	}))
	assert.True(t, updated, "failing a PROCESSING row owned by this claim must update it")

	loaded := loadEvidence(t, ctx, client, fx.scope, id)
	assert.Equal(t, coredata.EvidenceAssessmentStatusFailed, loaded.AssessmentStatus)
	assert.Nil(t, loaded.AssessmentProcessingStartedAt)
}

func TestEvidence_SetAssessmentFailed_NoOpOnSupersededClaim(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedAssessmentFixture(t, ctx, client)

	t0 := time.Now().UTC().Truncate(time.Microsecond)
	id := fx.insertPendingEvidence(t, ctx, client, "ev-superseded-fail-"+t0.Format(time.RFC3339Nano), t0)

	// Worker A claims, then B re-claims with a fresh started_at.
	firstClaim := claimEvidence(t, ctx, client, fx.scope, id, t0.Add(time.Minute))
	secondClaimAt := t0.Add(3 * time.Minute)
	_ = claimEvidence(t, ctx, client, fx.scope, id, secondClaimAt)

	// Worker A's late failure must not flip B's live claim to FAILED.
	var updated bool
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error
		updated, err = firstClaim.SetAssessmentFailed(ctx, tx, fx.scope)
		return err
	}))
	assert.False(t, updated, "a stale worker's failure must not clobber the re-claimed row")

	loaded := loadEvidence(t, ctx, client, fx.scope, id)
	assert.Equal(t, coredata.EvidenceAssessmentStatusProcessing, loaded.AssessmentStatus, "row must remain PROCESSING under the new owner")
	require.NotNil(t, loaded.AssessmentProcessingStartedAt)
	assert.WithinDuration(t, secondClaimAt, *loaded.AssessmentProcessingStartedAt, time.Second)
}

func TestResetStaleAssessmentProcessing_ResetsOnlyStaleClaims(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedAssessmentFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	staleID := fx.insertPendingEvidence(t, ctx, client, "ev-stale-"+now.Format(time.RFC3339Nano), now.Add(-time.Hour))
	freshID := fx.insertPendingEvidence(t, ctx, client, "ev-fresh-"+now.Format(time.RFC3339Nano), now.Add(-time.Hour))

	// One claim started long ago (stale), one started now (fresh).
	_ = claimEvidence(t, ctx, client, fx.scope, staleID, now.Add(-10*time.Minute))
	_ = claimEvidence(t, ctx, client, fx.scope, freshID, now)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return coredata.ResetStaleAssessmentProcessing(ctx, conn, 5*time.Minute)
	}))

	stale := loadEvidence(t, ctx, client, fx.scope, staleID)
	assert.Equal(t, coredata.EvidenceAssessmentStatusPending, stale.AssessmentStatus, "a stale claim is recycled to PENDING")
	assert.Nil(t, stale.AssessmentProcessingStartedAt)

	fresh := loadEvidence(t, ctx, client, fx.scope, freshID)
	assert.Equal(t, coredata.EvidenceAssessmentStatusProcessing, fresh.AssessmentStatus, "a fresh claim is left running")
	require.NotNil(t, fresh.AssessmentProcessingStartedAt)
}
