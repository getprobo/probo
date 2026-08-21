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

package thirdparty

import (
	"context"
	"io"
	"strconv"
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

// seedCatalogThirdPartyAt inserts a catalog third party with an explicit
// creation timestamp, so a test can pin which of two same-named rows is
// the oldest.
func seedCatalogThirdPartyAt(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	name string,
	slugValue string,
	createdAt time.Time,
) coredata.CommonThirdParty {
	t.Helper()

	party := seedCatalogThirdParty(t, ctx, client, name, slugValue)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`UPDATE common_third_parties SET created_at = $2 WHERE id = $1`,
			party.ID,
			createdAt,
		)

		return err
	}))

	party.CreatedAt = createdAt

	return party
}

// TestResolveOrCreateCommonThirdParty_TrimsName pins that the resolver
// trims the incoming name itself. The mapping worker forwards the agent's
// output verbatim, so an untrimmed name would otherwise be stored with its
// padding and miss the name lookup on the next resolution.
func TestResolveOrCreateCommonThirdParty_TrimsName(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	token := slug.Make(gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType).String())
	name := "Padded " + token

	var got *gid.GID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		id, err := ResolveOrCreateCommonThirdParty(
			ctx,
			tx,
			discardLogger(),
			"  "+name+"  ",
			coredata.ThirdPartyCategoryAnalytics,
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
	assert.Equal(t, slug.Make(name), created.Slug)
}

// TestResolveOrCreateCommonThirdParty_BlankNames pins that a name with no
// slug-able characters resolves to no vendor at all rather than inserting
// a row with an empty slug.
func TestResolveOrCreateCommonThirdParty_BlankNames(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	for _, name := range []string{"", "   ", "!!!"} {
		t.Run("name "+strconv.Quote(name), func(t *testing.T) {
			t.Parallel()

			var got *gid.GID

			require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
				id, err := ResolveOrCreateCommonThirdParty(
					ctx,
					tx,
					discardLogger(),
					name,
					coredata.ThirdPartyCategoryAnalytics,
				)
				got = id

				return err
			}))

			assert.Nil(t, got)
		})
	}
}

// TestResolveOrCreateCommonThirdParty_ReturnsOldestRowWhenNameDuplicated
// pins that a duplicated name resolves deterministically to the oldest
// row. lower(name) is not unique, so without an explicit ordering the
// lookup could return either row and links would scatter across a
// duplicate set between calls.
func TestResolveOrCreateCommonThirdParty_ReturnsOldestRowWhenNameDuplicated(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	token := slug.Make(gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType).String())
	name := "Doubled " + token
	now := time.Now().UTC().Truncate(time.Microsecond)

	// Two rows sharing a name but not a slug: only slug is unique, so
	// this is a state the catalog can genuinely reach.
	//
	// The newer row is inserted FIRST so it comes earlier in heap order.
	// An unordered LIMIT 1 then returns it, which is what makes this test
	// fail if the explicit ordering is ever dropped.
	newest := seedCatalogThirdPartyAt(t, ctx, client, name, slug.Make(name)+"-2", now.Add(-1*time.Hour))
	oldest := seedCatalogThirdPartyAt(t, ctx, client, name, slug.Make(name), now.Add(-48*time.Hour))
	require.NotEqual(t, oldest.ID, newest.ID)

	// Resolve twice: the point is stability across calls, not just that
	// one call happens to land on the older row.
	for range 2 {
		var got *gid.GID

		require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			id, err := ResolveOrCreateCommonThirdParty(
				ctx,
				tx,
				discardLogger(),
				name,
				coredata.ThirdPartyCategoryAnalytics,
			)
			got = id

			return err
		}))

		require.NotNil(t, got)
		assert.Equal(t, oldest.ID, *got)
	}
}

// countCatalogRowsBySlugPrefix counts catalog rows whose slug starts with
// the given prefix, so a test can assert that resolving a name variant
// reused a row instead of adding one.
func countCatalogRowsBySlugPrefix(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	prefix string,
) int {
	t.Helper()

	var count int

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return conn.QueryRow(
			ctx,
			`SELECT count(*) FROM common_third_parties WHERE slug LIKE $1 || '%'`,
			prefix,
		).Scan(&count)
	}))

	return count
}

// TestResolveOrCreateCommonThirdParty_SuffixStrippedSlugReuse pins the
// third dedup rung: a name that differs from a catalog entry only by a
// trailing legal form resolves to that entry instead of minting a variant.
func TestResolveOrCreateCommonThirdParty_SuffixStrippedSlugReuse(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	token := slug.Make(gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType).String())
	brand := "Hotjar " + token
	party := seedCatalogThirdParty(t, ctx, client, brand, slug.Make(brand))

	variants := []string{
		brand + " Ltd",
		brand + " Limited",
		brand + " Inc.",
		brand + ", Inc",
		brand + " GmbH",
		brand + " LLC",
		brand + " B.V.",
	}

	for _, variant := range variants {
		t.Run(variant, func(t *testing.T) {
			var got *gid.GID

			require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
				id, err := ResolveOrCreateCommonThirdParty(
					ctx,
					tx,
					discardLogger(),
					variant,
					coredata.ThirdPartyCategoryAnalytics,
				)
				got = id

				return err
			}))

			require.NotNil(t, got)
			assert.Equal(t, party.ID, *got)
		})
	}

	assert.Equal(
		t,
		1,
		countCatalogRowsBySlugPrefix(t, ctx, client, slug.Make(brand)),
		"legal-form variants must reuse the brand row, not add rows",
	)
}

// TestResolveOrCreateCommonThirdParty_DoesNotStripCountryCode pins the
// boundary the write path deliberately declines to cross.
//
// "OVHcloud US" is a separately incorporated subsidiary with its own data
// residency and DPA, so folding it into "OVHcloud" would erase exactly the
// jurisdictional distinction the register records — and unrecoverably,
// since no split exists. Trailing country codes are a duplicate-candidate
// signal for the operator merge tooling instead.
func TestResolveOrCreateCommonThirdParty_DoesNotStripCountryCode(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	token := slug.Make(gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType).String())
	parent := "OVHcloud " + token
	party := seedCatalogThirdParty(t, ctx, client, parent, slug.Make(parent))

	for _, variant := range []string{parent + " US", parent + " France", parent + " UK"} {
		t.Run(variant, func(t *testing.T) {
			var got *gid.GID

			require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
				id, err := ResolveOrCreateCommonThirdParty(
					ctx,
					tx,
					discardLogger(),
					variant,
					coredata.ThirdPartyCategoryCloudProvider,
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

			assert.NotEqual(t, party.ID, *got, "a country-qualified entity must stay distinct")
		})
	}
}

// TestResolveOrCreateCommonThirdParty_DoesNotFoldProductIntoParent guards
// against importing the disambiguation agent's product-to-parent rule into
// the resolver. That rule is right when matching one organization's vendor
// list and wrong for the catalog, where a parent's products are distinct
// entries with distinct categories and cookie families.
func TestResolveOrCreateCommonThirdParty_DoesNotFoldProductIntoParent(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	token := slug.Make(gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType).String())
	parent := "Google " + token
	party := seedCatalogThirdParty(t, ctx, client, parent, slug.Make(parent))

	var got *gid.GID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		id, err := ResolveOrCreateCommonThirdParty(
			ctx,
			tx,
			discardLogger(),
			parent+" Analytics",
			coredata.ThirdPartyCategoryAnalytics,
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

	assert.NotEqual(t, party.ID, *got)
}

// TestResolveOrCreateCommonThirdParty_SuffixOnlyName pins the edge case the
// third rung makes reachable: a name that is nothing but a legal form.
// stripCorporateSuffixes only removes a space-prefixed suffix, so the name
// survives intact and inserts as its own row rather than reducing to an
// empty slug.
func TestResolveOrCreateCommonThirdParty_SuffixOnlyName(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	var got *gid.GID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		id, err := ResolveOrCreateCommonThirdParty(
			ctx,
			tx,
			discardLogger(),
			"Ltd",
			coredata.ThirdPartyCategoryOther,
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

	assert.Equal(t, "Ltd", created.Name)
	assert.Equal(t, "ltd", created.Slug)
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

// A row created from agent identification has nobody's judgement behind it,
// so it must land in the review backlog. It is created with Insert and a nil
// Review, which the statement resolves to UNREVIEWED — the column is NOT NULL
// with no database default, so this is worth pinning rather than assuming.
func TestResolveOrCreateCommonThirdParty_StartsUnreviewed(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	name := "Resolver Review " + gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType).String()

	var id *gid.GID

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error

		id, err = ResolveOrCreateCommonThirdParty(
			ctx,
			tx,
			log.NewLogger(log.WithOutput(io.Discard)),
			name,
			coredata.ThirdPartyCategoryAnalytics,
		)

		return err
	}))

	require.NotNil(t, id)

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, *id)
			return err
		})
	})

	party := coredata.CommonThirdParty{}
	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return party.LoadByID(ctx, tx, *id)
	}))

	require.NotNil(t, party.Review, "the column is NOT NULL, so a stored row always has a state")
	assert.Equal(t, coredata.CommonThirdPartyReviewUnreviewed, *party.Review)
	assert.Nil(t, party.RejectedVerdict)
}
