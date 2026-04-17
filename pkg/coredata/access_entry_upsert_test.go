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
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// testPgDSNEnvVar is the environment variable that points the integration
// tests at a migrated test database. When unset, the tests are skipped so
// `make test` stays a pure unit-test run.
const testPgDSNEnvVar = "PROBO_TEST_PG_URL"

// newTestPgClient returns a pg.Client connected to the test database, or
// skips the test if no DSN is configured.
func newTestPgClient(t *testing.T) *pg.Client {
	t.Helper()

	dsn := os.Getenv(testPgDSNEnvVar)
	if dsn == "" {
		t.Skipf("skipping: %s not set (requires a migrated test database)", testPgDSNEnvVar)
	}

	u, err := url.Parse(dsn)
	require.NoError(t, err, "invalid %s value", testPgDSNEnvVar)

	// Each test builds its own pg.Client, so we provide a fresh Prometheus
	// registry every time to avoid "duplicate collector" panics when tests
	// run in parallel.
	opts := []pg.Option{pg.WithRegisterer(prometheus.NewRegistry())}
	if u.Host != "" {
		opts = append(opts, pg.WithAddr(u.Host))
	}
	if u.User != nil {
		opts = append(opts, pg.WithUser(u.User.Username()))
		if password, ok := u.User.Password(); ok {
			opts = append(opts, pg.WithPassword(password))
		}
	}
	if len(u.Path) > 1 {
		opts = append(opts, pg.WithDatabase(u.Path[1:]))
	}

	client, err := pg.NewClient(opts...)
	require.NoError(t, err)

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// accessEntryFixture bootstraps the parent rows (organization, campaign,
// source) that the access_entries FKs require.
type accessEntryFixture struct {
	scope          *coredata.Scope
	organizationID gid.GID
	campaignID     gid.GID
	sourceID       gid.GID
	accountKey     string
}

func seedAccessEntryFixture(t *testing.T, ctx context.Context, client *pg.Client) accessEntryFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	campaignID := gid.New(tenantID, coredata.AccessReviewCampaignEntityType)
	sourceID := gid.New(tenantID, coredata.AccessSourceEntityType)
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

		source := &coredata.AccessSource{
			ID:             sourceID,
			OrganizationID: organizationID,
			Name:           "Upsert Freeze Test Source",
			Category:       coredata.AccessSourceCategorySaaS,
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

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			// Delete access_entries first (no ON DELETE CASCADE for the org side),
			// then parents.
			if _, err := tx.Exec(ctx, `DELETE FROM access_entries WHERE access_review_campaign_id = $1`, campaignID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DELETE FROM access_review_campaigns WHERE id = $1`, campaignID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DELETE FROM access_sources WHERE id = $1`, sourceID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID); err != nil {
				return err
			}
			return nil
		})
	})

	return accessEntryFixture{
		scope:          scope,
		organizationID: organizationID,
		campaignID:     campaignID,
		sourceID:       sourceID,
		accountKey:     accountKey,
	}
}

func TestAccessEntry_Upsert_FreezesDecidedFields(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedAccessEntryFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	originalFlagReasons := []string{"original-flag-reason"}
	originalFlags := []coredata.AccessEntryFlag{coredata.AccessEntryFlagNew}
	originalEmail := "old@example.com"
	originalFullName := "Old Name"
	originalRole := "viewer"

	t0 := time.Now().UTC().Truncate(time.Microsecond)

	// Step 1: Initial Upsert with PENDING decision.
	entryID := gid.New(tenantID, coredata.AccessEntryEntityType)
	initial := &coredata.AccessEntry{
		ID:                     entryID,
		OrganizationID:         fx.organizationID,
		AccessReviewCampaignID: fx.campaignID,
		AccessSourceID:         fx.sourceID,
		Email:                  originalEmail,
		FullName:               originalFullName,
		Role:                   originalRole,
		JobTitle:               "",
		IsAdmin:                false,
		MFAStatus:              coredata.MFAStatusUnknown,
		AuthMethod:             coredata.AccessEntryAuthMethodUnknown,
		AccountType:            coredata.AccessEntryAccountTypeUser,
		ExternalID:             "ext-1",
		AccountKey:             fx.accountKey,
		IncrementalTag:         coredata.AccessEntryIncrementalTagNew,
		Flags:                  originalFlags,
		FlagReasons:            originalFlagReasons,
		Decision:               coredata.AccessEntryDecisionPending,
		DecisionNote:           nil,
		DecidedBy:              nil,
		DecidedAt:              nil,
		CreatedAt:              t0,
		UpdatedAt:              t0,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return initial.Upsert(ctx, tx, fx.scope)
	}))

	// Step 2: Record a decision via Update — APPROVED with decided_by / decided_at.
	decisionTime := t0.Add(1 * time.Hour)
	decidedBy := gid.New(tenantID, coredata.OrganizationEntityType) // opaque ID suffices: decided_by has no FK.
	decisionNote := "looks good"

	decided := &coredata.AccessEntry{
		ID:           entryID,
		Flags:        originalFlags,
		FlagReasons:  originalFlagReasons,
		Decision:     coredata.AccessEntryDecisionApproved,
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
	secondRole := "admin"
	refresh := &coredata.AccessEntry{
		ID:                     gid.New(tenantID, coredata.AccessEntryEntityType), // ignored by ON CONFLICT
		OrganizationID:         fx.organizationID,
		AccessReviewCampaignID: fx.campaignID,
		AccessSourceID:         fx.sourceID,
		Email:                  secondEmail,
		FullName:               secondFullName,
		Role:                   secondRole,
		JobTitle:               "",
		IsAdmin:                true,
		MFAStatus:              coredata.MFAStatusEnabled,
		AuthMethod:             coredata.AccessEntryAuthMethodSSO,
		AccountType:            coredata.AccessEntryAccountTypeUser,
		ExternalID:             "ext-1",
		AccountKey:             fx.accountKey,
		IncrementalTag:         coredata.AccessEntryIncrementalTagUnchanged,
		Flags:                  []coredata.AccessEntryFlag{coredata.AccessEntryFlagInactive},
		FlagReasons:            []string{"refreshed-flag-reason"},
		Decision:               coredata.AccessEntryDecisionPending,
		DecisionNote:           nil,
		DecidedBy:              nil,
		DecidedAt:              nil,
		CreatedAt:              t2,
		UpdatedAt:              t2,
	}
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return refresh.Upsert(ctx, tx, fx.scope)
	}))

	// Step 4: Load and assert the freeze semantics.
	loaded := &coredata.AccessEntry{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, fx.scope, entryID)
	}))

	// Decision fields are FROZEN at APPROVED / decided_by / decided_at /
	// decision_note from the Update call.
	assert.Equal(t, coredata.AccessEntryDecisionApproved, loaded.Decision, "decision must be frozen once locked")
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
	assert.Equal(t, secondRole, loaded.Role)
	assert.True(t, loaded.IsAdmin)
	assert.Equal(t, coredata.MFAStatusEnabled, loaded.MFAStatus)
	assert.Equal(t, coredata.AccessEntryAuthMethodSSO, loaded.AuthMethod)
	assert.WithinDuration(t, t2, loaded.UpdatedAt, time.Second)
}

func TestAccessEntry_Upsert_RefreshesWhilePending(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	fx := seedAccessEntryFixture(t, ctx, client)

	tenantID := fx.scope.GetTenantID()
	t0 := time.Now().UTC().Truncate(time.Microsecond)

	entryID := gid.New(tenantID, coredata.AccessEntryEntityType)
	first := &coredata.AccessEntry{
		ID:                     entryID,
		OrganizationID:         fx.organizationID,
		AccessReviewCampaignID: fx.campaignID,
		AccessSourceID:         fx.sourceID,
		Email:                  "old@example.com",
		FullName:               "Old Name",
		Role:                   "viewer",
		MFAStatus:              coredata.MFAStatusUnknown,
		AuthMethod:             coredata.AccessEntryAuthMethodUnknown,
		AccountType:            coredata.AccessEntryAccountTypeUser,
		ExternalID:             "ext-2",
		AccountKey:             fx.accountKey,
		IncrementalTag:         coredata.AccessEntryIncrementalTagNew,
		Flags:                  []coredata.AccessEntryFlag{coredata.AccessEntryFlagNew},
		FlagReasons:            []string{"first-reason"},
		Decision:               coredata.AccessEntryDecisionPending,
		CreatedAt:              t0,
		UpdatedAt:              t0,
	}
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return first.Upsert(ctx, tx, fx.scope)
	}))

	t1 := t0.Add(1 * time.Hour)
	second := &coredata.AccessEntry{
		ID:                     gid.New(tenantID, coredata.AccessEntryEntityType),
		OrganizationID:         fx.organizationID,
		AccessReviewCampaignID: fx.campaignID,
		AccessSourceID:         fx.sourceID,
		Email:                  "new@example.com",
		FullName:               "New Name",
		Role:                   "admin",
		MFAStatus:              coredata.MFAStatusEnabled,
		AuthMethod:             coredata.AccessEntryAuthMethodSSO,
		AccountType:            coredata.AccessEntryAccountTypeUser,
		ExternalID:             "ext-2",
		AccountKey:             fx.accountKey,
		IncrementalTag:         coredata.AccessEntryIncrementalTagUnchanged,
		Flags:                  []coredata.AccessEntryFlag{coredata.AccessEntryFlagInactive},
		FlagReasons:            []string{"second-reason"},
		Decision:               coredata.AccessEntryDecisionPending,
		CreatedAt:              t1,
		UpdatedAt:              t1,
	}
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return second.Upsert(ctx, tx, fx.scope)
	}))

	loaded := &coredata.AccessEntry{}
	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByID(ctx, conn, fx.scope, entryID)
	}))

	assert.Equal(t, coredata.AccessEntryDecisionPending, loaded.Decision)
	assert.Equal(t, []coredata.AccessEntryFlag{coredata.AccessEntryFlagInactive}, loaded.Flags)
	assert.Equal(t, []string{"second-reason"}, loaded.FlagReasons)
	assert.Equal(t, "new@example.com", loaded.Email)
	assert.Equal(t, "New Name", loaded.FullName)
	assert.Equal(t, "admin", loaded.Role)
	assert.Equal(t, coredata.MFAStatusEnabled, loaded.MFAStatus)
	assert.Equal(t, coredata.AccessEntryAuthMethodSSO, loaded.AuthMethod)
}
