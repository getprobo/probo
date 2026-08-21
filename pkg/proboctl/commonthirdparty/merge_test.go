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

package commonthirdparty

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slug"
)

func seedPreviewParty(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	prefix string,
) coredata.CommonThirdParty {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	id := gid.New(gid.NilTenant, coredata.CommonThirdPartyEntityType)
	name := prefix + " " + id.String()
	party := coredata.CommonThirdParty{
		ID:             id,
		Name:           name,
		Slug:           slug.Make(name),
		Category:       coredata.ThirdPartyCategoryAnalytics,
		Certifications: []string{},
		Review:         coredata.CommonThirdPartyReviewUnreviewed,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return party.Insert(ctx, tx)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			return party.Delete(ctx, tx, party.ID)
		})
	})

	return party
}

func seedPreviewDomain(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	partyID gid.GID,
	domain string,
) {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	ownedDomain := coredata.CommonThirdPartyDomain{
		ID:                 gid.New(gid.NilTenant, coredata.CommonThirdPartyDomainEntityType),
		CommonThirdPartyID: partyID,
		Domain:             domain,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return ownedDomain.Insert(ctx, tx)
	}))
}

func TestPreviewMerges_AccountsForEarlierLosers(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()

	winner := seedPreviewParty(t, ctx, client, "preview winner")
	firstLoser := seedPreviewParty(t, ctx, client, "preview first loser")
	secondLoser := seedPreviewParty(t, ctx, client, "preview second loser")
	sharedDomain := "shared-" + winner.Slug + ".example"

	seedPreviewDomain(t, ctx, client, firstLoser.ID, sharedDomain)
	seedPreviewDomain(t, ctx, client, secondLoser.ID, sharedDomain)

	var out bytes.Buffer

	require.NoError(
		t,
		previewMerges(
			ctx,
			client,
			&out,
			winner,
			[]coredata.CommonThirdParty{firstLoser, secondLoser},
		),
	)

	output := out.String()
	firstResult := "domains:            1 moved, 0 dropped as duplicate"
	secondResult := "domains:            0 moved, 1 dropped as duplicate"

	require.Contains(t, output, firstResult)
	require.Contains(t, output, secondResult)
	assert.Equal(t, 1, strings.Count(output, firstResult))
	assert.Equal(t, 1, strings.Count(output, secondResult))
	assert.Less(t, strings.Index(output, firstResult), strings.Index(output, secondResult))

	for partyID, expectedDomains := range map[gid.GID]int{
		winner.ID:      0,
		firstLoser.ID:  1,
		secondLoser.ID: 1,
	} {
		var domains coredata.CommonThirdPartyDomains

		require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return domains.LoadByCommonThirdPartyID(ctx, conn, partyID)
		}))

		assert.Len(t, domains, expectedDomains, "the preview transaction must roll back")
	}
}
