// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package thirdparty

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slug"
)

func discardLogger() *log.Logger {
	return log.NewLogger(log.WithOutput(io.Discard))
}

// seedCatalogThirdParty inserts a catalog third party with the given
// name and slug and registers its cleanup.
func seedCatalogThirdParty(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	name string,
	slugValue string,
) coredata.CommonThirdParty {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	party := coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           name,
		Slug:           slugValue,
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return party.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, party.ID)
			return err
		})
	})

	return party
}

// TestResolveOrCreateCommonThirdParty pins the catalog dedup that the
// mapping and enrichment workers reuse to link a vendor: an exact name
// match and a slug match both return the existing row, and a name absent
// from the catalog creates a fresh row (name, slug, category) rather than
// duplicating one.
func TestResolveOrCreateCommonThirdParty(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	logger := discardLogger()

	token := slug.Make(gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType).String())

	t.Run("returns existing row on name match", func(t *testing.T) {
		name := "Hotjar " + token
		party := seedCatalogThirdParty(t, ctx, client, name, slug.Make(name))

		var got *gid.GID

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			id, err := ResolveOrCreateCommonThirdParty(
				ctx,
				tx,
				logger,
				name,
				coredata.ThirdPartyCategoryAnalytics,
			)
			got = id

			return err
		}))

		require.NotNil(t, got)
		assert.Equal(t, party.ID, *got)
	})

	t.Run("returns existing row on slug match", func(t *testing.T) {
		name := "Matomo " + token
		party := seedCatalogThirdParty(t, ctx, client, name, slug.Make(name))

		// A differently-spelled name that normalizes to the same slug
		// must resolve to the existing row, not create a duplicate.
		variant := "Matomo  " + token + "!!!"
		require.NotEqual(t, name, variant)
		require.Equal(t, slug.Make(name), slug.Make(variant))

		var got *gid.GID

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			id, err := ResolveOrCreateCommonThirdParty(
				ctx,
				tx,
				logger,
				variant,
				coredata.ThirdPartyCategoryAnalytics,
			)
			got = id

			return err
		}))

		require.NotNil(t, got)
		assert.Equal(t, party.ID, *got)
	})

	t.Run("creates a new row when absent", func(t *testing.T) {
		name := "Freshvendor " + token
		expectedSlug := slug.Make(name)

		var got *gid.GID

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			id, err := ResolveOrCreateCommonThirdParty(
				ctx,
				tx,
				logger,
				name,
				coredata.ThirdPartyCategoryMarketing,
			)
			got = id

			return err
		}))

		require.NotNil(t, got)

		t.Cleanup(func() {
			_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
				_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, *got)
				return err
			})
		})

		var created coredata.CommonThirdParty

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return created.LoadByID(ctx, conn, *got)
		}))

		assert.Equal(t, name, created.Name)
		assert.Equal(t, expectedSlug, created.Slug)
		assert.Equal(t, coredata.ThirdPartyCategoryMarketing, created.Category)
	})
}
