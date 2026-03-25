// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package accessreview

import (
	"context"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	worker           *SourceFetchWorker
	sourceNameWorker *SourceNameWorker
}

type CampaignReader interface {
	Get(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error)
}

type TenantRuntime interface {
	AccessReviewCampaigns() CampaignReader
	SnapshotSource(ctx context.Context, campaign *coredata.AccessReviewCampaign, sourceID gid.GID) (int, error)
}

type TenantRuntimeProvider func(tenantID gid.TenantID) TenantRuntime

func NewService(
	pgClient *pg.Client,
	logger *log.Logger,
	interval time.Duration,
	tenantRuntimeProvider TenantRuntimeProvider,
	encryptionKey cipher.EncryptionKey,
) *Service {
	return &Service{
		worker: NewSourceFetchWorker(
			pgClient,
			tenantRuntimeProvider,
			logger,
			WithSourceFetchWorkerInterval(interval),
		),
		sourceNameWorker: NewSourceNameWorker(
			pgClient,
			encryptionKey,
			logger.Named("source-name"),
		),
	}
}

func (s *Service) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return s.worker.Run(ctx) })
	g.Go(func() error { return s.sourceNameWorker.Run(ctx) })
	return g.Wait()
}
