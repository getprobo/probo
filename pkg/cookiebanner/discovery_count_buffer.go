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
	"fmt"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const discoveryCountFlushInterval = 10 * time.Second

type (
	discoveryFamilyAtomics struct {
		chrome  atomic.Int64
		edge    atomic.Int64
		firefox atomic.Int64
		safari  atomic.Int64
		other   atomic.Int64
	}

	discoveryCounts struct {
		chrome  int64
		edge    int64
		firefox int64
		safari  int64
		other   int64
	}

	discoveryHitKey struct {
		bannerID  gid.GID
		patternID gid.GID
	}
)

func (c discoveryCounts) empty() bool {
	return c.chrome == 0 && c.edge == 0 && c.firefox == 0 && c.safari == 0 && c.other == 0
}

func (c *discoveryFamilyAtomics) addFamily(family coredata.DiscoveryBrowserFamily) {
	switch family {
	case coredata.DiscoveryBrowserFamilyChrome:
		c.chrome.Add(1)
	case coredata.DiscoveryBrowserFamilyEdge:
		c.edge.Add(1)
	case coredata.DiscoveryBrowserFamilyFirefox:
		c.firefox.Add(1)
	case coredata.DiscoveryBrowserFamilySafari:
		c.safari.Add(1)
	default:
		c.other.Add(1)
	}
}

func (c *discoveryFamilyAtomics) swap(n int64) discoveryCounts {
	return discoveryCounts{
		chrome:  c.chrome.Swap(n),
		edge:    c.edge.Swap(n),
		firefox: c.firefox.Swap(n),
		safari:  c.safari.Swap(n),
		other:   c.other.Swap(n),
	}
}

func (s *Service) addPageLoad(bannerID gid.GID, family coredata.DiscoveryBrowserFamily) {
	actual, _ := s.pageLoadCounts.LoadOrStore(bannerID, &discoveryFamilyAtomics{})
	actual.(*discoveryFamilyAtomics).addFamily(family)
}

func (s *Service) addPatternHit(
	bannerID gid.GID,
	patternID gid.GID,
	family coredata.DiscoveryBrowserFamily,
) {
	key := discoveryHitKey{bannerID: bannerID, patternID: patternID}
	actual, _ := s.patternHitCounts.LoadOrStore(key, &discoveryFamilyAtomics{})
	actual.(*discoveryFamilyAtomics).addFamily(family)
}

func (s *Service) recordDiscoveryIncrements(
	bannerID gid.GID,
	family coredata.DiscoveryBrowserFamily,
	pageView bool,
	patternIDs map[gid.GID]struct{},
) {
	if !family.IsValid() {
		family = coredata.DiscoveryBrowserFamilyOther
	}

	if pageView {
		s.addPageLoad(bannerID, family)
	}

	for patternID := range patternIDs {
		s.addPatternHit(bannerID, patternID, family)
	}
}

func (s *Service) FlushDiscoveryCounts(ctx context.Context) error {
	var firstErr error

	s.pageLoadCounts.Range(
		func(key, value any) bool {
			bannerID := key.(gid.GID)

			counts := value.(*discoveryFamilyAtomics).swap(0)
			if counts.empty() {
				return true
			}

			if err := s.writePageLoads(ctx, bannerID, counts); err != nil && firstErr == nil {
				firstErr = err
			}

			return true
		},
	)

	s.patternHitCounts.Range(
		func(key, value any) bool {
			hitKey := key.(discoveryHitKey)

			counts := value.(*discoveryFamilyAtomics).swap(0)
			if counts.empty() {
				return true
			}

			if err := s.writePatternHits(ctx, hitKey.bannerID, hitKey.patternID, counts); err != nil && firstErr == nil {
				firstErr = err
			}

			return true
		},
	)

	return firstErr
}

func (s *Service) writePageLoads(
	ctx context.Context,
	bannerID gid.GID,
	counts discoveryCounts,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(bannerID)

			if err := coredata.RecordDiscoveryPageLoad(
				ctx,
				tx,
				scope,
				bannerID,
				counts.chrome,
				counts.edge,
				counts.firefox,
				counts.safari,
				counts.other,
			); err != nil {
				return fmt.Errorf("cannot record discovery page load: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) writePatternHits(
	ctx context.Context,
	bannerID gid.GID,
	patternID gid.GID,
	counts discoveryCounts,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(bannerID)

			if err := coredata.RecordDiscoveryPatternHit(
				ctx,
				tx,
				scope,
				bannerID,
				patternID,
				counts.chrome,
				counts.edge,
				counts.firefox,
				counts.safari,
				counts.other,
			); err != nil {
				return fmt.Errorf("cannot record discovery pattern hit: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) RunDiscoveryCountFlusher(ctx context.Context, logger *log.Logger) error {
	logger = logger.Named("discovery-count-flusher")

	ticker := time.NewTicker(discoveryCountFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			flushCtx := context.WithoutCancel(ctx)
			if err := s.FlushDiscoveryCounts(flushCtx); err != nil {
				logger.ErrorCtx(flushCtx, "cannot flush discovery counts on shutdown", log.Error(err))
			}

			return nil
		case <-ticker.C:
			if err := s.FlushDiscoveryCounts(ctx); err != nil {
				logger.ErrorCtx(ctx, "cannot flush discovery counts", log.Error(err))
			}
		}
	}
}
