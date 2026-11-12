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

package cookiebanner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestFreezeCookieBannerSample_LeavesNewerOpenSample(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	first, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	second, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	require.NotEqual(t, first.ID, second.ID)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return coredata.FreezeCookieBannerSample(ctx, tx, fx.scope, first.ID)
	}))

	reloadedFirst := loadSample(t, ctx, client, fx, first.ID)
	require.NotNil(t, reloadedFirst.FrozenAt)

	open := loadLatestSample(t, ctx, client, fx)
	assert.Equal(t, second.ID, open.ID)
	assert.Nil(t, open.FrozenAt)
}

func TestListCookieBannerSamples_OrdersNewestFirst(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedWorkerFixture(t, ctx, client)
	svc := NewService(client, false, 0)

	first, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)
	second, err := svc.StartCookieBannerSample(ctx, fx.scope, fx.banner.ID)
	require.NoError(t, err)

	cursor := page.NewCursor(
		10,
		nil,
		page.Head,
		page.OrderBy[coredata.CookieBannerSampleOrderField]{
			Field:     coredata.CookieBannerSampleOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		},
	)

	samples, err := svc.ListCookieBannerSamples(ctx, fx.scope, fx.banner.ID, cursor)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(samples), 2)
	assert.Equal(t, second.ID, samples[0].ID)
	assert.Equal(t, first.ID, samples[1].ID)
}

func loadLatestSample(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
) coredata.CookieBannerSample {
	t.Helper()

	var sample coredata.CookieBannerSample

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return sample.LoadLatestByCookieBannerID(ctx, conn, fx.scope, fx.banner.ID)
	}))

	return sample
}

func loadSample(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx workerFixture,
	sampleID gid.GID,
) coredata.CookieBannerSample {
	t.Helper()

	var sample coredata.CookieBannerSample

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return sample.LoadByID(ctx, conn, fx.scope, sampleID)
	}))

	return sample
}
