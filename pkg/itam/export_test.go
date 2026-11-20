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

package itam

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connectorclient"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func NewAssetSyncHandlerForTest(
	pgClient *pg.Client,
	clients *connectorclient.Client,
	list AssetLister,
) *assetSyncHandler {
	h := newAssetSyncHandler(pgClient, clients, log.NewLogger(log.WithOutput(io.Discard)))
	for i := range h.kinds {
		h.kinds[i].list = list
	}

	return h
}

func SetAssetSyncSourceForTest(h *assetSyncHandler, id gid.GID) {
	h.onlySource = &id
}

func ClaimSourceForTest(
	h *assetSyncHandler,
	ctx context.Context,
	sourceID gid.GID,
) (coredata.AssetSync, error) {
	h.onlySource = &sourceID

	return h.Claim(ctx)
}

func SyncSourceForTest(
	h *assetSyncHandler,
	ctx context.Context,
	sourceID gid.GID,
) (coredata.AssetSync, error) {
	sync, err := ClaimSourceForTest(h, ctx, sourceID)
	if err != nil {
		return coredata.AssetSync{}, err
	}

	if err := h.Process(ctx, sync); err != nil {
		return sync, err
	}

	return sync, nil
}

func ClaimAssetSyncForTest(
	h *assetSyncHandler,
	ctx context.Context,
) (coredata.AssetSync, error) {
	return h.Claim(ctx)
}

func ProcessAssetSyncForTest(
	h *assetSyncHandler,
	ctx context.Context,
	sync coredata.AssetSync,
) error {
	return h.Process(ctx, sync)
}

func RecoverStaleAssetSyncsForTest(h *assetSyncHandler, ctx context.Context) error {
	return h.RecoverStale(ctx)
}

func AssetSyncFailureMessageForTest() string {
	return assetSyncFailureMessage
}

func AssetSyncStaleMessageForTest() string {
	return assetSyncStaleMessage
}

func BackdateAssetSyncForTest(
	ctx context.Context,
	pgClient *pg.Client,
	id gid.GID,
	at time.Time,
) error {
	return pgClient.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, `UPDATE assets_syncs SET started_at = $1 WHERE id = $2`, at, id)
		if err != nil {
			return fmt.Errorf("cannot backdate asset sync: %w", err)
		}

		return nil
	})
}

func CollectPageLimitForTest(limit int) (int, error) {
	calls := 0
	_, err := collectPages(
		context.Background(),
		limit,
		func(context.Context, *string) ([]int, bool, *string, error) {
			calls++
			next := fmt.Sprintf("page-%d", calls)

			return []int{calls}, true, &next, nil
		},
	)

	return calls, err
}

func CollectSinglePageForTest() ([]int, error) {
	return collectPages(
		context.Background(),
		5,
		func(context.Context, *string) ([]int, bool, *string, error) {
			return []int{1, 2}, false, nil, nil
		},
	)
}

func MapCloudFrontForTest() (AssetObservation, error) {
	arn := "arn:aws:cloudfront::123456789012:distribution/EDFDVBD632BHDS5"
	domain := "d111111abcdef8.cloudfront.net"
	enabled := true
	status := "Deployed"
	certificate := true

	return mapCloudFrontDistribution(types.DistributionSummary{
		ARN:        &arn,
		DomainName: &domain,
		Enabled:    &enabled,
		Status:     &status,
		PriceClass: types.PriceClassPriceClass100,
		ViewerCertificate: &types.ViewerCertificate{
			CloudFrontDefaultCertificate: &certificate,
		},
	})
}

func MapCloudFrontMissingARNForTest() error {
	_, err := mapCloudFrontDistribution(types.DistributionSummary{})

	return err
}
