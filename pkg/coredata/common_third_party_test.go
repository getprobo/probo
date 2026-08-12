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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// TestCommonThirdPartyUpsert_SyncsReceiverToWrittenRow pins the invariant
// the common-third-parties seed relies on: after Upsert the receiver
// identifies the row that was actually written, so the seed can attach an
// entry's domains using party.ID without reloading.
//
// The seed used to reload by name on the update path. lower(name) is not
// unique, so that reload could return a different row and misattribute the
// entry's domains to it. Upsert conflicts on slug and returns the
// conflicting row, which is what makes the reload unnecessary.
func TestCommonThirdPartyUpsert_SyncsReceiverToWrittenRow(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	existing := seedCommonThirdParty(t, ctx, client)

	// A second row sharing the name but not the slug: the state that made
	// the old reload-by-name ambiguous.
	decoy := coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           existing.Name,
		Slug:           existing.Slug + "-decoy",
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		CreatedAt:      existing.CreatedAt,
		UpdatedAt:      existing.UpdatedAt,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return decoy.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, decoy.ID)
			return err
		})
	})

	// Re-seeding mints a fresh GID and conflicts on the existing slug,
	// exactly as the seed command does on a second run.
	now := time.Now().UTC().Truncate(time.Microsecond)
	reseeded := coredata.CommonThirdParty{
		ID:             gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType),
		Name:           existing.Name,
		Slug:           existing.Slug,
		Category:       coredata.ThirdPartyCategoryMarketing,
		Certifications: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	var inserted bool

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var err error
		inserted, err = reseeded.Upsert(ctx, tx)

		return err
	}))

	assert.False(t, inserted, "conflicting slug must report an update, not an insert")
	assert.Equal(t, existing.ID, reseeded.ID, "receiver must carry the id of the row written")
	assert.NotEqual(t, decoy.ID, reseeded.ID, "receiver must not pick up the same-named decoy row")
	assert.Equal(t, existing.Slug, reseeded.Slug)
}
