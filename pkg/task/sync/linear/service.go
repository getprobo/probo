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

package linear

import (
	"context"
	"net/http"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/task/sync/job"
)

type (
	Service struct {
		pg                     *pg.Client
		encryptionKey          cipher.EncryptionKey
		connectorRegistry      *connectorprotocol.Registry
		baseURL                string
		linearAPIBaseURL       string
		logger                 *log.Logger
		recordUpdateActivities job.RecordUpdateActivitiesFunc
		recordImportedTask     job.RecordImportedTaskFunc
		attachments            job.Reader
		stores                 job.Writer
		linearAssets           linearAssetCache
		inboundFiles           inboundFileCache

		oauthHTTP       func(context.Context, *connectorprotocol.OAuth2Connection, coredata.ConnectorProvider) (*http.Client, error)
		actorProfile    func(context.Context, pg.Querier, coredata.Scoper, *gid.GID, gid.GID) (*gid.GID, error)
		enqueueOutbound func(context.Context, pg.Tx, coredata.Scoper, *coredata.TaskExternalLink, job.SyncAction) error
	}

	Config struct {
		PG                     *pg.Client
		EncryptionKey          cipher.EncryptionKey
		ConnectorRegistry      *connectorprotocol.Registry
		BaseURL                string
		APIBaseURL             string
		Logger                 *log.Logger
		RecordUpdateActivities job.RecordUpdateActivitiesFunc
		RecordImportedTask     job.RecordImportedTaskFunc
		Attachments            job.Reader
		Stores                 job.Writer
		OAuthHTTP              func(context.Context, *connectorprotocol.OAuth2Connection, coredata.ConnectorProvider) (*http.Client, error)
		ActorProfile           func(context.Context, pg.Querier, coredata.Scoper, *gid.GID, gid.GID) (*gid.GID, error)
		EnqueueOutbound        func(context.Context, pg.Tx, coredata.Scoper, *coredata.TaskExternalLink, job.SyncAction) error
	}
)

func New(cfg Config) *Service {
	return &Service{
		pg:                     cfg.PG,
		encryptionKey:          cfg.EncryptionKey,
		connectorRegistry:      cfg.ConnectorRegistry,
		baseURL:                cfg.BaseURL,
		linearAPIBaseURL:       cfg.APIBaseURL,
		logger:                 cfg.Logger,
		recordUpdateActivities: cfg.RecordUpdateActivities,
		recordImportedTask:     cfg.RecordImportedTask,
		attachments:            cfg.Attachments,
		stores:                 cfg.Stores,
		linearAssets:           linearAssetCache{urls: map[string]string{}},
		oauthHTTP:              cfg.OAuthHTTP,
		actorProfile:           cfg.ActorProfile,
		enqueueOutbound:        cfg.EnqueueOutbound,
	}
}
