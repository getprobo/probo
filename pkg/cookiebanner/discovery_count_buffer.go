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

	discoveryHitKey struct {
		bannerID  gid.GID
		patternID gid.GID
	}

	bannerDiscoveryFlush struct {
		pageLoads coredata.DiscoveryFamilyDeltas
		hits      map[gid.GID]coredata.DiscoveryFamilyDeltas
	}
)

func (c *discoveryFamilyAtomics) addDeltas(deltas coredata.DiscoveryFamilyDeltas) {
	if deltas.Chrome != 0 {
		c.chrome.Add(deltas.Chrome)
	}

	if deltas.Edge != 0 {
		c.edge.Add(deltas.Edge)
	}

	if deltas.Firefox != 0 {
		c.firefox.Add(deltas.Firefox)
	}

	if deltas.Safari != 0 {
		c.safari.Add(deltas.Safari)
	}

	if deltas.Other != 0 {
		c.other.Add(deltas.Other)
	}
}

func (c *discoveryFamilyAtomics) swap(n int64) coredata.DiscoveryFamilyDeltas {
	return coredata.DiscoveryFamilyDeltas{
		Chrome:  c.chrome.Swap(n),
		Edge:    c.edge.Swap(n),
		Firefox: c.firefox.Swap(n),
		Safari:  c.safari.Swap(n),
		Other:   c.other.Swap(n),
	}
}

func (s *Service) addDiscoveryFamily(
	key any,
	family coredata.DiscoveryBrowserFamily,
	pageLoad bool,
) {
	s.addDiscoveryDeltas(key, coredata.DiscoveryFamilyDeltasFor(family, 1), pageLoad)
}

func (s *Service) addDiscoveryDeltas(
	key any,
	deltas coredata.DiscoveryFamilyDeltas,
	pageLoad bool,
) {
	if deltas.Empty() {
		return
	}

	m := &s.patternHitCounts
	if pageLoad {
		m = &s.pageLoadCounts
	}

	actual, _ := m.LoadOrStore(key, &discoveryFamilyAtomics{})
	c := actual.(*discoveryFamilyAtomics)
	c.addDeltas(deltas)

	if current, ok := m.Load(key); ok && current == c {
		return
	}

	// Flush deleted the entry after we loaded it. Put the increment
	// on whatever key is live now so it is not stranded on an orphan.
	actual, _ = m.LoadOrStore(key, &discoveryFamilyAtomics{})
	actual.(*discoveryFamilyAtomics).addDeltas(deltas)
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
		s.addDiscoveryFamily(bannerID, family, true)
	}

	for patternID := range patternIDs {
		s.addDiscoveryFamily(
			discoveryHitKey{bannerID: bannerID, patternID: patternID},
			family,
			false,
		)
	}
}

func (s *Service) FlushDiscoveryCounts(ctx context.Context) error {
	byBanner := make(map[gid.GID]*bannerDiscoveryFlush)

	s.pageLoadCounts.Range(
		func(key, value any) bool {
			bannerID := key.(gid.GID)

			deltas := value.(*discoveryFamilyAtomics).swap(0)
			if deltas.Empty() {
				s.pageLoadCounts.Delete(key)
				return true
			}

			flush := bannerDiscoveryFlushFor(byBanner, bannerID)
			flush.pageLoads = deltas

			return true
		},
	)

	s.patternHitCounts.Range(
		func(key, value any) bool {
			hitKey := key.(discoveryHitKey)

			deltas := value.(*discoveryFamilyAtomics).swap(0)
			if deltas.Empty() {
				s.patternHitCounts.Delete(key)
				return true
			}

			flush := bannerDiscoveryFlushFor(byBanner, hitKey.bannerID)
			if flush.hits == nil {
				flush.hits = make(map[gid.GID]coredata.DiscoveryFamilyDeltas)
			}

			flush.hits[hitKey.patternID] = deltas

			return true
		},
	)

	var firstErr error

	for bannerID, flush := range byBanner {
		if err := s.flushBannerDiscoveryCounts(ctx, bannerID, flush); err != nil {
			s.restoreBannerDiscoveryCounts(bannerID, flush)

			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (s *Service) FlushDiscoveryCountsForBanner(ctx context.Context, bannerID gid.GID) error {
	flush := &bannerDiscoveryFlush{}

	if actual, ok := s.pageLoadCounts.Load(bannerID); ok {
		deltas := actual.(*discoveryFamilyAtomics).swap(0)
		if deltas.Empty() {
			s.pageLoadCounts.Delete(bannerID)
		} else {
			flush.pageLoads = deltas
		}
	}

	s.patternHitCounts.Range(
		func(key, value any) bool {
			hitKey := key.(discoveryHitKey)
			if hitKey.bannerID != bannerID {
				return true
			}

			deltas := value.(*discoveryFamilyAtomics).swap(0)
			if deltas.Empty() {
				s.patternHitCounts.Delete(key)
				return true
			}

			if flush.hits == nil {
				flush.hits = make(map[gid.GID]coredata.DiscoveryFamilyDeltas)
			}

			flush.hits[hitKey.patternID] = deltas

			return true
		},
	)

	if flush.pageLoads.Empty() && len(flush.hits) == 0 {
		return nil
	}

	if err := s.flushBannerDiscoveryCounts(ctx, bannerID, flush); err != nil {
		s.restoreBannerDiscoveryCounts(bannerID, flush)
		return err
	}

	return nil
}

func (s *Service) restoreBannerDiscoveryCounts(bannerID gid.GID, flush *bannerDiscoveryFlush) {
	s.addDiscoveryDeltas(bannerID, flush.pageLoads, true)

	for patternID, deltas := range flush.hits {
		s.addDiscoveryDeltas(
			discoveryHitKey{bannerID: bannerID, patternID: patternID},
			deltas,
			false,
		)
	}
}

func (s *Service) flushBannerDiscoveryCounts(
	ctx context.Context,
	bannerID gid.GID,
	flush *bannerDiscoveryFlush,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(bannerID)

			if err := coredata.RecordDiscoveryPageLoad(ctx, tx, scope, bannerID, flush.pageLoads); err != nil {
				return fmt.Errorf("cannot record discovery page load: %w", err)
			}

			for patternID, deltas := range flush.hits {
				if err := coredata.RecordDiscoveryPatternHit(ctx, tx, scope, bannerID, patternID, deltas); err != nil {
					return fmt.Errorf("cannot record discovery pattern hit: %w", err)
				}
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

func bannerDiscoveryFlushFor(
	byBanner map[gid.GID]*bannerDiscoveryFlush,
	bannerID gid.GID,
) *bannerDiscoveryFlush {
	flush, ok := byBanner[bannerID]
	if !ok {
		flush = &bannerDiscoveryFlush{}
		byBanner[bannerID] = flush
	}

	return flush
}
