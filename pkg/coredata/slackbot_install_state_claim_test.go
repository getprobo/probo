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

func TestSlackbotInstallStateClaim_ProcessingLifecycle(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	pgClient := test.PGClient(t)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organization := coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Slack install state claim test",
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

	now := time.Now().UTC()
	claim := coredata.NewSlackbotInstallStateClaim(
		organization.ID,
		"state-"+organization.ID.String(),
	)
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
					5*time.Minute,
				)

				return err
			},
		),
	)
	assert.True(t, claimed)

	competing := coredata.NewSlackbotInstallStateClaim(
		organization.ID,
		"state-"+organization.ID.String(),
	)

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
					now.Add(time.Minute),
					5*time.Minute,
				)

				return err
			},
		),
	)
	assert.False(t, claimed)

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
					now.Add(6*time.Minute),
					5*time.Minute,
				)
				if err != nil {
					return err
				}

				if !claimed {
					return fmt.Errorf("cannot reclaim stale Slack install state")
				}

				return competing.Complete(
					ctx,
					tx,
					scope,
					"00000000-0000-7000-8000-000000000003",
					now.Add(6*time.Minute),
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

				claimed, err = claim.Claim(
					ctx,
					tx,
					scope,
					"00000000-0000-7000-8000-000000000004",
					now.Add(20*time.Minute),
					5*time.Minute,
				)

				return err
			},
		),
	)
	assert.False(t, claimed)

	released := coredata.NewSlackbotInstallStateClaim(
		organization.ID,
		"released-state-"+organization.ID.String(),
	)

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
					5*time.Minute,
				)
				if err != nil {
					return err
				}

				if !claimed {
					return fmt.Errorf("cannot claim releasable Slack install state")
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
					now.Add(time.Minute),
					5*time.Minute,
				)

				return err
			},
		),
	)
	assert.True(t, claimed)
}
