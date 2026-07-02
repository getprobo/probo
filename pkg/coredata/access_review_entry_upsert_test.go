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

package coredata_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// accessEntryFixture bootstraps the parent rows (organization, campaign,
// source) that the access_review_entries FKs require.
type accessEntryFixture struct {
	scope            *coredata.Scope
	organizationID   gid.GID
	campaignID       gid.GID
	sourceID         gid.GID
	campaignSourceID gid.GID
	accountKey       string
}

func seedAccessReviewEntryFixture(t *testing.T, ctx context.Context, client *pg.Client) accessEntryFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	campaignID := gid.New(tenantID, coredata.AccessReviewCampaignEntityType)
	sourceID := gid.New(tenantID, coredata.AccessReviewSourceEntityType)
	campaignSourceID := gid.New(tenantID, coredata.AccessReviewCampaignSourceEntityType)
	accountKey := "upsert-freeze-test@example.com"
	now := time.Now().UTC()

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Upsert Freeze Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		source := &coredata.AccessReviewSource{
			ID:             sourceID,
			OrganizationID: organizationID,
			Name:           "Upsert Freeze Test Source",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := source.Insert(ctx, tx, scope); err != nil {
			return err
		}

		campaign := &coredata.AccessReviewCampaign{
			ID:             campaignID,
			OrganizationID: organizationID,
			Name:           "Upsert Freeze Test Campaign",
			Status:         coredata.AccessReviewCampaignStatusDraft,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := campaign.Insert(ctx, tx, scope); err != nil {
			return err
		}

		campaignSource := &coredata.AccessReviewCampaignSource{
			ID:                     campaignSourceID,
			OrganizationID:         organizationID,
			TenantID:               tenantID,
			AccessReviewCampaignID: campaignID,
			AccessReviewSourceID:   &sourceID,
			Name:                   "Upsert Freeze Test Source",
			CreatedAt:              now,
			UpdatedAt:              now,
		}
		if err := campaignSource.Upsert(ctx, tx, scope); err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			// Delete access_review_entries first (no ON DELETE CASCADE for the org side),
			// then parents.
			if _, err := tx.Exec(ctx, `DELETE FROM access_review_entries WHERE access_review_campaign_id = $1`, campaignID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM access_review_campaign_sources WHERE access_review_campaign_id = $1`, campaignID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM access_review_campaigns WHERE id = $1`, campaignID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM access_review_sources WHERE id = $1`, sourceID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID); err != nil {
				return err
			}

			return nil
		})
	})

	return accessEntryFixture{
		scope:            scope,
		organizationID:   organizationID,
		campaignID:       campaignID,
		sourceID:         sourceID,
		campaignSourceID: campaignSourceID,
		accountKey:       accountKey,
	}
}

func TestAccessReviewEntry_Upsert_FreezesDecidedFields(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	originalFlagReasons := []string{"original-flag-reason"}
	originalFlags := []coredata.AccessReviewEntryFlag{coredata.AccessReviewEntryFlagNew}
	originalEmail := "old@example.com"
	originalFullName := "Old Name"
	originalRoles := []string{"viewer"}

	t0 := time.Now().UTC().Truncate(time.Microsecond)

	// Step 1: Initial Upsert with PENDING decision.
	entryID := gid.New(tenantID, coredata.AccessReviewEntryEntityType)
	initial := &coredata.AccessReviewEntry{
		ID:                           entryID,
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignID:       fx.campaignID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Email:                        originalEmail,
		FullName:                     originalFullName,
		Roles:                        originalRoles,
		JobTitle:                     "",
		IsAdmin:                      false,
		MFAStatus:                    coredata.MFAStatusUnknown,
		AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:                   "ext-1",
		AccountKey:                   fx.accountKey,
		IncrementalTag:               coredata.AccessReviewEntryIncrementalTagNew,
		Flags:                        originalFlags,
		FlagReasons:                  originalFlagReasons,
		Decision:                     coredata.AccessReviewEntryDecisionPending,
		DecisionNote:                 nil,
		DecidedBy:                    nil,
		DecidedAt:                    nil,
		CreatedAt:                    t0,
		UpdatedAt:                    t0,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return initial.Upsert(ctx, tx, fx.scope)
	}))

	// Step 2: Record a decision via Update — APPROVED with decided_by / decided_at.
	decisionTime := t0.Add(1 * time.Hour)
	decidedBy := gid.New(tenantID, coredata.OrganizationEntityType) // opaque ID suffices: decided_by has no FK.
	decisionNote := "looks good"

	decided := &coredata.AccessReviewEntry{
		ID:           entryID,
		Flags:        originalFlags,
		FlagReasons:  originalFlagReasons,
		Decision:     coredata.AccessReviewEntryDecisionApproved,
		DecisionNote: &decisionNote,
		DecidedBy:    &decidedBy,
		DecidedAt:    &decisionTime,
		UpdatedAt:    decisionTime,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return decided.Update(ctx, tx, fx.scope)
	}))

	// Step 3: Second Upsert with the same unique key but new flags, new
	// flag reasons, PENDING decision, nil note/decidedBy/decidedAt, and
	// refreshed top-level fields (email, full_name, role).
	t2 := decisionTime.Add(1 * time.Hour)
	secondEmail := "new@example.com"
	secondFullName := "New Name"
	secondRoles := []string{"admin"}
	refresh := &coredata.AccessReviewEntry{
		ID:                           gid.New(tenantID, coredata.AccessReviewEntryEntityType), // ignored by ON CONFLICT
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignID:       fx.campaignID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Email:                        secondEmail,
		FullName:                     secondFullName,
		Roles:                        secondRoles,
		JobTitle:                     "",
		IsAdmin:                      true,
		MFAStatus:                    coredata.MFAStatusEnabled,
		AuthMethod:                   coredata.AccessReviewEntryAuthMethodSSO,
		AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:                   "ext-1",
		AccountKey:                   fx.accountKey,
		IncrementalTag:               coredata.AccessReviewEntryIncrementalTagUnchanged,
		Flags:                        []coredata.AccessReviewEntryFlag{coredata.AccessReviewEntryFlagInactive},
		FlagReasons:                  []string{"refreshed-flag-reason"},
		Decision:                     coredata.AccessReviewEntryDecisionPending,
		DecisionNote:                 nil,
		DecidedBy:                    nil,
		DecidedAt:                    nil,
		CreatedAt:                    t2,
		UpdatedAt:                    t2,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return refresh.Upsert(ctx, tx, fx.scope)
	}))

	// Step 4: Load and assert the freeze semantics.
	loaded := &coredata.AccessReviewEntry{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, fx.scope, entryID)
	}))

	// Decision fields are FROZEN at APPROVED / decided_by / decided_at /
	// decision_note from the Update call.
	assert.Equal(t, coredata.AccessReviewEntryDecisionApproved, loaded.Decision, "decision must be frozen once locked")
	require.NotNil(t, loaded.DecidedBy, "decided_by must be preserved")
	assert.Equal(t, decidedBy, *loaded.DecidedBy)
	require.NotNil(t, loaded.DecidedAt, "decided_at must be preserved")
	assert.WithinDuration(t, decisionTime, *loaded.DecidedAt, time.Second)
	require.NotNil(t, loaded.DecisionNote, "decision_note must be preserved")
	assert.Equal(t, decisionNote, *loaded.DecisionNote)

	// Flags / flag_reasons are FROZEN (the new guard from Task 1): once a
	// reviewer locks a decision, the evidence that drove that decision must
	// not be silently replaced by a subsequent poll.
	assert.Equal(t, originalFlags, loaded.Flags, "flags must be frozen once decision is locked")
	assert.Equal(t, originalFlagReasons, loaded.FlagReasons, "flag_reasons must be frozen once decision is locked")

	// Columns that ARE refreshed on every poll.
	assert.Equal(t, secondEmail, loaded.Email)
	assert.Equal(t, secondFullName, loaded.FullName)
	assert.Equal(t, secondRoles, loaded.Roles)
	assert.True(t, loaded.IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, loaded.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, loaded.AuthMethod)
	assert.WithinDuration(t, t2, loaded.UpdatedAt, time.Second)
}

// TestAccessReviewEntry_Upsert_RefreshesSourceTrackingFields pins the contract of
// the ON CONFLICT DO UPDATE SET clause: across repeated polls of the same
// (campaign, source, account_key), the columns that track live source state
// (email, full_name, role, is_admin, MFA, auth_method, last_login, etc.)
// move forward to the latest values, while the verdict-related columns
// (flags, flag_reasons, decision, decision_note, decided_by, decided_at) are
// never written by a re-poll -- those can only change through Update.
func TestAccessReviewEntry_Upsert_RefreshesSourceTrackingFields(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	t0 := time.Now().UTC().Truncate(time.Microsecond)

	entryID := gid.New(tenantID, coredata.AccessReviewEntryEntityType)
	first := &coredata.AccessReviewEntry{
		ID:                           entryID,
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignID:       fx.campaignID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Email:                        "old@example.com",
		FullName:                     "Old Name",
		Roles:                        []string{"viewer"},
		MFAStatus:                    coredata.MFAStatusUnknown,
		AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:                   "ext-2",
		AccountKey:                   fx.accountKey,
		IncrementalTag:               coredata.AccessReviewEntryIncrementalTagNew,
		Flags:                        []coredata.AccessReviewEntryFlag{},
		FlagReasons:                  []string{},
		Decision:                     coredata.AccessReviewEntryDecisionPending,
		CreatedAt:                    t0,
		UpdatedAt:                    t0,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return first.Upsert(ctx, tx, fx.scope)
	}))

	t1 := t0.Add(1 * time.Hour)
	second := &coredata.AccessReviewEntry{
		ID:                           gid.New(tenantID, coredata.AccessReviewEntryEntityType),
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignID:       fx.campaignID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Email:                        "new@example.com",
		FullName:                     "New Name",
		Roles:                        []string{"admin"},
		MFAStatus:                    coredata.MFAStatusEnabled,
		AuthMethod:                   coredata.AccessReviewEntryAuthMethodSSO,
		AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:                   "ext-2",
		AccountKey:                   fx.accountKey,
		IncrementalTag:               coredata.AccessReviewEntryIncrementalTagUnchanged,
		Flags:                        []coredata.AccessReviewEntryFlag{},
		FlagReasons:                  []string{},
		Decision:                     coredata.AccessReviewEntryDecisionPending,
		CreatedAt:                    t1,
		UpdatedAt:                    t1,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return second.Upsert(ctx, tx, fx.scope)
	}))

	loaded := &coredata.AccessReviewEntry{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, fx.scope, entryID)
	}))

	// Source-tracking columns advanced to the second poll's values.
	assert.Equal(t, "new@example.com", loaded.Email)
	assert.Equal(t, "New Name", loaded.FullName)
	assert.Equal(t, []string{"admin"}, loaded.Roles)
	assert.Equal(t, coredata.MFAStatusEnabled, loaded.MFAStatus)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSO, loaded.AuthMethod)

	// Verdict-related columns stayed at whatever the first Upsert set (empty /
	// PENDING); the second Upsert did not touch them.
	assert.Equal(t, coredata.AccessReviewEntryDecisionPending, loaded.Decision)
	assert.Equal(t, []coredata.AccessReviewEntryFlag{}, loaded.Flags)
	assert.Equal(t, []string{}, loaded.FlagReasons)
	assert.Nil(t, loaded.DecisionNote)
	assert.Nil(t, loaded.DecidedBy)
	assert.Nil(t, loaded.DecidedAt)
}

// TestAccessReviewEntry_Upsert_InsertsActiveAccount covers the shape FetchSource
// builds for an active account: a PENDING decision and explicit empty
// flags / flag_reasons slices. The access_review_entries.flags and flag_reasons
// columns are declared NOT NULL, so the caller (FetchSource) is responsible
// for passing non-nil slices.
func TestAccessReviewEntry_Upsert_InsertsActiveAccount(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	t0 := time.Now().UTC().Truncate(time.Microsecond)

	entryID := gid.New(tenantID, coredata.AccessReviewEntryEntityType)
	entry := &coredata.AccessReviewEntry{
		ID:                           entryID,
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignID:       fx.campaignID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Email:                        "active@example.com",
		FullName:                     "Active User",
		Roles:                        []string{"member"},
		MFAStatus:                    coredata.MFAStatusUnknown,
		AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:                   "ext-active",
		AccountKey:                   fx.accountKey,
		IncrementalTag:               coredata.AccessReviewEntryIncrementalTagNew,
		Flags:                        []coredata.AccessReviewEntryFlag{},
		FlagReasons:                  []string{},
		Decision:                     coredata.AccessReviewEntryDecisionPending,
		CreatedAt:                    t0,
		UpdatedAt:                    t0,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return entry.Upsert(ctx, tx, fx.scope)
	}))

	loaded := &coredata.AccessReviewEntry{}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, fx.scope, entryID)
	}))

	assert.Equal(t, coredata.AccessReviewEntryDecisionPending, loaded.Decision)
	assert.Equal(t, []coredata.AccessReviewEntryFlag{}, loaded.Flags)
	assert.Equal(t, []string{}, loaded.FlagReasons)
	assert.Nil(t, loaded.DecisionNote)
	assert.Nil(t, loaded.DecidedBy)
	assert.Nil(t, loaded.DecidedAt)
}

func TestAccessReviewEntry_Upsert_NilRolesWritesEmptyArray(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedAccessReviewEntryFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	t0 := time.Now().UTC().Truncate(time.Microsecond)

	entryID := gid.New(tenantID, coredata.AccessReviewEntryEntityType)
	entry := &coredata.AccessReviewEntry{
		ID:                           entryID,
		OrganizationID:               fx.organizationID,
		AccessReviewCampaignID:       fx.campaignID,
		AccessReviewCampaignSourceID: fx.campaignSourceID,
		Email:                        "nil-roles@example.com",
		FullName:                     "Nil Roles User",
		Roles:                        nil,
		MFAStatus:                    coredata.MFAStatusUnknown,
		AuthMethod:                   coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType:                  coredata.AccessReviewEntryAccountTypeUser,
		ExternalID:                   "ext-nil-roles",
		AccountKey:                   "nil-roles@example.com",
		IncrementalTag:               coredata.AccessReviewEntryIncrementalTagNew,
		Flags:                        []coredata.AccessReviewEntryFlag{},
		FlagReasons:                  []string{},
		Decision:                     coredata.AccessReviewEntryDecisionPending,
		CreatedAt:                    t0,
		UpdatedAt:                    t0,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return entry.Upsert(ctx, tx, fx.scope)
	}))

	var roles []string

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return conn.QueryRow(
			ctx,
			`SELECT roles FROM access_review_entries WHERE id = @id`,
			pgx.StrictNamedArgs{"id": entryID},
		).Scan(&roles)
	}))

	assert.Equal(t, []string{}, roles)
}
