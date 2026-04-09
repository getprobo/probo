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

package consent

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

type (
	Service struct {
		pg *pg.Client
	}

	RecordConsentRequest struct {
		VisitorID   string
		ConsentData json.RawMessage
		Action      coredata.ConsentAction
	}

	PublicBannerConfig struct {
		Banner     *coredata.CookieBanner
		Categories coredata.CookieCategories
	}
)

const (
	NameMaxLength        = 100
	TitleMaxLength       = 1000
	ContentMaxLength     = 5000
	ConsentExpiryDaysMin = 1
	ConsentExpiryDaysMax = 3650
)

func NewService(pgClient *pg.Client) *Service {
	return &Service{pg: pgClient}
}

func (s *Service) GetPublishedBannerConfig(
	ctx context.Context,
	bannerID gid.GID,
) (*PublicBannerConfig, error) {
	var (
		banner     = &coredata.CookieBanner{}
		categories coredata.CookieCategories
	)

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := banner.LoadPublishedByID(ctx, conn, bannerID); err != nil {
				return fmt.Errorf("cannot load published cookie banner: %w", err)
			}

			if err := categories.LoadAllPublicByCookieBannerID(ctx, conn, bannerID); err != nil {
				return fmt.Errorf("cannot load cookie categories: %w", err)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return &PublicBannerConfig{
		Banner:     banner,
		Categories: categories,
	}, nil
}

func (s *Service) RecordConsent(
	ctx context.Context,
	bannerID gid.GID,
	req RecordConsentRequest,
	ipAddress string,
	userAgent string,
) error {
	return s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			banner := &coredata.CookieBanner{}
			if err := banner.LoadPublishedByID(ctx, conn, bannerID); err != nil {
				return fmt.Errorf("cannot load published cookie banner: %w", err)
			}

			var (
				tenantID     = banner.ID.TenantID()
				scope        = coredata.NewScope(tenantID)
				anonymizedIP = anonymizeIP(ipAddress)
			)

			record := &coredata.ConsentRecord{
				ID:             gid.New(tenantID, coredata.ConsentRecordEntityType),
				CookieBannerID: bannerID,
				VisitorID:      req.VisitorID,
				IPAddress:      &anonymizedIP,
				UserAgent:      &userAgent,
				ConsentData:    req.ConsentData,
				Action:         req.Action,
				BannerVersion:  banner.Version,
				CreatedAt:      time.Now(),
			}

			if err := record.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert consent record: %w", err)
			}

			return nil
		},
	)
}

func anonymizeIP(ip string) string {
	// r.RemoteAddr includes a port (e.g. "192.168.1.42:54321"),
	// strip it before parsing.
	host, _, err := net.SplitHostPort(ip)
	if err != nil {
		// No port — use as-is (e.g. Unix socket or plain IP).
		host = ip
	}

	parsed := net.ParseIP(host)
	if parsed == nil {
		return ""
	}

	if v4 := parsed.To4(); v4 != nil {
		v4[3] = 0
		return v4.String()
	}

	// IPv6: zero last 80 bits (10 bytes)
	for i := 6; i < 16; i++ {
		parsed[i] = 0
	}
	return parsed.String()
}
