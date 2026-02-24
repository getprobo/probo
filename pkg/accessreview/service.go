package accessreview

import (
	"context"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type Service struct {
	worker *SourceFetchWorker
}

type CampaignReader interface {
	Get(ctx context.Context, campaignID gid.GID) (*coredata.AccessReviewCampaign, error)
}

type TenantRuntime interface {
	AccessReviewCampaigns() CampaignReader
	SnapshotSource(ctx context.Context, campaign *coredata.AccessReviewCampaign, sourceID gid.GID) (int, error)
	Diff(ctx context.Context, campaign *coredata.AccessReviewCampaign) error
}

type TenantRuntimeProvider func(tenantID gid.TenantID) TenantRuntime

func NewService(
	pgClient *pg.Client,
	logger *log.Logger,
	interval time.Duration,
	tenantRuntimeProvider TenantRuntimeProvider,
) *Service {
	return &Service{
		worker: newSourceFetchWorker(
			pgClient,
			tenantRuntimeProvider,
			logger,
			WithSourceFetchWorkerInterval(interval),
		),
	}
}

func (s *Service) Run(ctx context.Context) error {
	return s.worker.Run(ctx)
}
