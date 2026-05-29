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

package cookiebanner

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func newEnrichmentHandler(client *pg.Client) *commonPatternEnrichmentHandler {
	return &commonPatternEnrichmentHandler{
		pg:     client,
		logger: log.NewLogger(log.WithOutput(io.Discard)),
	}
}

// seedEnrichmentThirdParty inserts a collision-free catalog third party
// for the resolver to match against.
func seedEnrichmentThirdParty(t *testing.T, ctx context.Context, client *pg.Client) coredata.CommonThirdParty {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	suffix := id.String()

	party := coredata.CommonThirdParty{
		ID:             id,
		Name:           "Hotjar " + suffix,
		Slug:           "hotjar-" + suffix,
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
			_, err := tx.Exec(ctx, `DELETE FROM common_third_parties WHERE id = $1`, id)
			return err
		})
	})

	return party
}

// TestResolveThirdPartyID pins the enrichment worker's third-party
// resolution: it links the agent's returned company to an existing
// catalog row by name, but only when the pattern has no third party yet,
// and it never invents one for a name absent from the catalog.
func TestResolveThirdPartyID(t *testing.T) {
	t.Parallel()

	client := newTestPgClient(t)
	ctx := context.Background()
	h := newEnrichmentHandler(client)

	party := seedEnrichmentThirdParty(t, ctx, client)
	existingID := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)

	t.Run("links existing catalog third party when unset", func(t *testing.T) {
		cp := coredata.CommonTrackerPattern{}

		got, err := h.resolveThirdPartyID(ctx, cp, party.Name)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, party.ID, *got)
	})

	t.Run("matches catalog name case-insensitively", func(t *testing.T) {
		cp := coredata.CommonTrackerPattern{}

		got, err := h.resolveThirdPartyID(ctx, cp, "hOtJaR "+party.ID.String())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, party.ID, *got)
	})

	t.Run("does not resolve when pattern already linked", func(t *testing.T) {
		cp := coredata.CommonTrackerPattern{CommonThirdPartyID: &existingID}

		got, err := h.resolveThirdPartyID(ctx, cp, party.Name)
		require.NoError(t, err)
		assert.Nil(t, got, "must not override an existing third-party link")
	})

	t.Run("returns nil for a name absent from the catalog", func(t *testing.T) {
		cp := coredata.CommonTrackerPattern{}

		got, err := h.resolveThirdPartyID(ctx, cp, "Nonexistent Vendor "+party.ID.String())
		require.NoError(t, err)
		assert.Nil(t, got, "must not invent a third party")
	})

	t.Run("returns nil for an empty name", func(t *testing.T) {
		cp := coredata.CommonTrackerPattern{}

		got, err := h.resolveThirdPartyID(ctx, cp, "")
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}
