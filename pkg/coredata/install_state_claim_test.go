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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// installStateClaimStaleAfter is the PRODUCTION window, not a test-local one:
// binding to it is what makes a future tuning of the constant show up here
// rather than pass silently against a window nothing runs under. The offsets
// below are derived from it for the same reason.
const installStateClaimStaleAfter = coredata.InstallStateStaleAfter

// installStateClaimLedgers is every ledger the shared primitive writes to. The
// two differ only by table, so the contract below is asserted against both
// rather than against whichever one a change happened to touch.
func installStateClaimLedgers() map[string]func(gid.GID, string) *coredata.InstallStateClaim {
	return map[string]func(gid.GID, string) *coredata.InstallStateClaim{
		"connector": coredata.NewConnectorInstallStateClaim,
		"slackbot":  coredata.NewSlackbotInstallStateClaim,
	}
}

func newInstallStateClaimOrganization(
	t *testing.T,
	pgClient *pg.Client,
	name string,
) (coredata.Organization, coredata.Scoper) {
	t.Helper()

	ctx := t.Context()
	tenantID := gid.NewTenantID()
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(
		t,
		pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Insert(ctx, tx)
			},
		),
	)
	t.Cleanup(func() {
		_ = pgClient.WithTx(
			context.Background(),
			func(ctx context.Context, tx pg.Tx) error {
				return organization.Delete(ctx, tx, organization.ID)
			},
		)
	})

	return organization, coredata.NewScope(tenantID)
}

func TestInstallStateClaim_ProcessingLifecycle(t *testing.T) {
	t.Parallel()

	for ledger, newClaim := range installStateClaimLedgers() {
		t.Run(ledger, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			pgClient := test.PGClient(t)
			organization, scope := newInstallStateClaimOrganization(
				t,
				pgClient,
				ledger+" install state claim test",
			)

			now := time.Now().UTC()
			state := "state-" + organization.ID.String()
			claim := newClaim(organization.ID, state)
			claimed := false

			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						var err error

						claimed, err = claim.Claim(
							ctx,
							tx,
							scope,
							"00000000-0000-7000-8000-000000000001",
							now,
							installStateClaimStaleAfter,
						)

						return err
					},
				),
			)
			assert.True(t, claimed, "first claim must succeed")

			// A second request landing inside the stale window - a refresh, or
			// a link-preview prefetch - gets nothing.
			competing := newClaim(organization.ID, state)

			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						var err error

						claimed, err = competing.Claim(
							ctx,
							tx,
							scope,
							"00000000-0000-7000-8000-000000000002",
							now.Add(installStateClaimStaleAfter/2),
							installStateClaimStaleAfter,
						)

						return err
					},
				),
			)
			assert.False(t, claimed, "a live claim must not be stealable")

			// Past the stale window the claim is reclaimable, so a crashed
			// process does not cost the customer their window.
			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						claimed, err := competing.Claim(
							ctx,
							tx,
							scope,
							"00000000-0000-7000-8000-000000000003",
							now.Add(installStateClaimStaleAfter+time.Second),
							installStateClaimStaleAfter,
						)
						if err != nil {
							return err
						}

						if !claimed {
							return fmt.Errorf("cannot reclaim stale install state")
						}

						return competing.Complete(
							ctx,
							tx,
							scope,
							"00000000-0000-7000-8000-000000000003",
							now.Add(installStateClaimStaleAfter+time.Second),
						)
					},
				),
			)

			// Completed is forever: no later claim reopens a burnt state.
			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						var err error

						claimed, err = claim.Claim(
							ctx,
							tx,
							scope,
							"00000000-0000-7000-8000-000000000004",
							now.Add(10*installStateClaimStaleAfter),
							installStateClaimStaleAfter,
						)

						return err
					},
				),
			)
			assert.False(t, claimed, "a completed state must never be reclaimable")

			// A released state is spendable again inside its TTL.
			released := newClaim(organization.ID, "released-state-"+organization.ID.String())

			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						claimed, err := released.Claim(
							ctx,
							tx,
							scope,
							"00000000-0000-7000-8000-000000000005",
							now,
							installStateClaimStaleAfter,
						)
						if err != nil {
							return err
						}

						if !claimed {
							return fmt.Errorf("cannot claim releasable install state")
						}

						return released.Release(
							ctx,
							tx,
							scope,
							"00000000-0000-7000-8000-000000000005",
						)
					},
				),
			)
			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						var err error

						claimed, err = released.Claim(
							ctx,
							tx,
							scope,
							"00000000-0000-7000-8000-000000000006",
							now.Add(installStateClaimStaleAfter/2),
							installStateClaimStaleAfter,
						)

						return err
					},
				),
			)
			assert.True(t, claimed, "a released state must be reclaimable")
		})
	}
}

// TestInstallStateClaim_RejectsForeignTenant pins the isolation the ON CONFLICT
// predicate carries: the ledger is keyed on the state digest alone, so without
// the tenant and organization guards a second tenant replaying a leaked state
// would take the claim out from under its owner.
func TestInstallStateClaim_RejectsForeignTenant(t *testing.T) {
	t.Parallel()

	for ledger, newClaim := range installStateClaimLedgers() {
		t.Run(ledger, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			pgClient := test.PGClient(t)
			owner, ownerScope := newInstallStateClaimOrganization(
				t,
				pgClient,
				ledger+" install state claim owner",
			)
			intruder, intruderScope := newInstallStateClaimOrganization(
				t,
				pgClient,
				ledger+" install state claim intruder",
			)

			now := time.Now().UTC()
			state := "shared-state-" + owner.ID.String()
			claimed := false

			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						var err error

						claimed, err = newClaim(owner.ID, state).Claim(
							ctx,
							tx,
							ownerScope,
							"00000000-0000-7000-8000-00000000000a",
							now,
							installStateClaimStaleAfter,
						)

						return err
					},
				),
			)
			require.True(t, claimed)

			// The attack shape: the intruder names the OWNER's organization —
			// so the organization guard matches — while scoping the write to
			// their own tenant. Only the tenant guard separates the two, which
			// is why the organization id here is the owner's and not the
			// intruder's.
			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						var err error

						claimed, err = newClaim(owner.ID, state).Claim(
							ctx,
							tx,
							intruderScope,
							"00000000-0000-7000-8000-00000000000b",
							// Well past the stale window: staleness must not be
							// what saves us here.
							now.Add(10*installStateClaimStaleAfter),
							installStateClaimStaleAfter,
						)

						return err
					},
				),
			)
			assert.False(t, claimed, "a foreign tenant must never take the claim")

			// And the mirror image: the intruder's OWN organization under the
			// owner's tenant, which the organization guard alone must refuse.
			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						var err error

						claimed, err = newClaim(intruder.ID, state).Claim(
							ctx,
							tx,
							ownerScope,
							"00000000-0000-7000-8000-00000000000d",
							now.Add(10*installStateClaimStaleAfter),
							installStateClaimStaleAfter,
						)

						return err
					},
				),
			)
			assert.False(t, claimed, "a foreign organization must never take the claim")

			// Nor may it burn or release what it does not hold.
			ownerClaim := newClaim(owner.ID, state)

			require.ErrorIs(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						return ownerClaim.Complete(
							ctx,
							tx,
							intruderScope,
							"00000000-0000-7000-8000-00000000000a",
							now,
						)
					},
				),
				coredata.ErrResourceNotFound,
			)

			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						return ownerClaim.Release(
							ctx,
							tx,
							intruderScope,
							"00000000-0000-7000-8000-00000000000a",
						)
					},
				),
			)

			// The owner's claim survived the intruder's release attempt.
			require.NoError(
				t,
				pgClient.WithTx(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						var err error

						claimed, err = newClaim(owner.ID, state).Claim(
							ctx,
							tx,
							ownerScope,
							"00000000-0000-7000-8000-00000000000c",
							now.Add(installStateClaimStaleAfter/2),
							installStateClaimStaleAfter,
						)

						return err
					},
				),
			)
			assert.False(t, claimed, "the owner's live claim must have survived")
		})
	}
}
