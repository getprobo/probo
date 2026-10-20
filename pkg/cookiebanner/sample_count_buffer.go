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
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const sampleCountFlushInterval = 10 * time.Second

type (
	sampleFamilyAtomics struct {
		chrome  atomic.Int64
		edge    atomic.Int64
		firefox atomic.Int64
		safari  atomic.Int64
		other   atomic.Int64
	}

	samplePageLoadAtomics struct {
		all      sampleFamilyAtomics
		expected sampleFamilyAtomics
	}

	sampleCounts struct {
		chrome  int64
		edge    int64
		firefox int64
		safari  int64
		other   int64
	}

	sampleHitKey struct {
		bannerID  gid.GID
		patternID gid.GID
	}
)

func (c sampleCounts) empty() bool {
	return c.chrome == 0 && c.edge == 0 && c.firefox == 0 && c.safari == 0 && c.other == 0
}

func (c *sampleFamilyAtomics) addFamily(family coredata.BrowserFamily) {
	switch family {
	case coredata.BrowserFamilyChrome:
		c.chrome.Add(1)
	case coredata.BrowserFamilyEdge:
		c.edge.Add(1)
	case coredata.BrowserFamilyFirefox:
		c.firefox.Add(1)
	case coredata.BrowserFamilySafari:
		c.safari.Add(1)
	default:
		c.other.Add(1)
	}
}

func (c *sampleFamilyAtomics) swap(n int64) sampleCounts {
	return sampleCounts{
		chrome:  c.chrome.Swap(n),
		edge:    c.edge.Swap(n),
		firefox: c.firefox.Swap(n),
		safari:  c.safari.Swap(n),
		other:   c.other.Swap(n),
	}
}

func (s *Service) addPageLoad(bannerID gid.GID, family coredata.BrowserFamily, expected bool) {
	actual, _ := s.pageLoadCounts.LoadOrStore(bannerID, &samplePageLoadAtomics{})
	loads := actual.(*samplePageLoadAtomics)
	loads.all.addFamily(family)

	if expected {
		loads.expected.addFamily(family)
	}
}

func (s *Service) addPatternHit(
	bannerID gid.GID,
	patternID gid.GID,
	family coredata.BrowserFamily,
) {
	key := sampleHitKey{bannerID: bannerID, patternID: patternID}
	actual, _ := s.patternHitCounts.LoadOrStore(key, &sampleFamilyAtomics{})
	actual.(*sampleFamilyAtomics).addFamily(family)
}

// isExpectedSampleLoad reports whether a tracker is allowed to fire on this
// view. OPT_IN counts only an explicit accept. OPT_OUT and NOTICE count every
// view except reject-all and Global Privacy Control, including a missing action.
func isExpectedSampleLoad(presentation Presentation, action *coredata.CookieConsentAction) bool {
	switch presentation {
	case PresentationOptOut, PresentationNotice:
		if action == nil {
			return true
		}

		switch *action {
		case coredata.CookieConsentActionRejectAll, coredata.CookieConsentActionGPC:
			return false
		default:
			return true
		}
	default:
		return action != nil && *action == coredata.CookieConsentActionAcceptAll
	}
}

func (s *Service) recordSampleIncrements(
	bannerID gid.GID,
	family coredata.BrowserFamily,
	pageView bool,
	expected bool,
	patternIDs map[gid.GID]struct{},
) {
	if !family.IsValid() {
		family = coredata.BrowserFamilyOther
	}

	if pageView {
		s.addPageLoad(bannerID, family, expected)
	}

	for patternID := range patternIDs {
		s.addPatternHit(bannerID, patternID, family)
	}
}

func (s *Service) FlushSampleCounts(ctx context.Context) error {
	var firstErr error

	s.pageLoadCounts.Range(
		func(key, value any) bool {
			bannerID := key.(gid.GID)
			loads := value.(*samplePageLoadAtomics)
			all := loads.all.swap(0)

			expected := loads.expected.swap(0)
			if all.empty() && expected.empty() {
				return true
			}

			if err := s.writePageLoads(ctx, bannerID, all, expected); err != nil && firstErr == nil {
				firstErr = err
			}

			return true
		},
	)

	s.patternHitCounts.Range(
		func(key, value any) bool {
			hitKey := key.(sampleHitKey)

			counts := value.(*sampleFamilyAtomics).swap(0)
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
	all sampleCounts,
	expected sampleCounts,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(bannerID)

			sampleID, ok, err := resolveWritableSample(ctx, tx, scope, bannerID)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}

			if err := coredata.RecordSamplePageLoads(
				ctx,
				tx,
				scope,
				sampleID,
				all.chrome,
				all.edge,
				all.firefox,
				all.safari,
				all.other,
				expected.chrome,
				expected.edge,
				expected.firefox,
				expected.safari,
				expected.other,
			); err != nil {
				return fmt.Errorf("cannot record sample page load: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) writePatternHits(
	ctx context.Context,
	bannerID gid.GID,
	patternID gid.GID,
	counts sampleCounts,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(bannerID)

			sampleID, ok, err := resolveWritableSample(ctx, tx, scope, bannerID)
			if err != nil {
				return err
			}

			if !ok {
				return nil
			}

			if err := coredata.RecordSamplePatternHit(
				ctx,
				tx,
				scope,
				sampleID,
				bannerID,
				patternID,
				counts.chrome,
				counts.edge,
				counts.firefox,
				counts.safari,
				counts.other,
			); err != nil {
				return fmt.Errorf("cannot record sample pattern hit: %w", err)
			}

			return nil
		},
	)
}

func resolveWritableSample(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	bannerID gid.GID,
) (gid.GID, bool, error) {
	var open coredata.CookieBannerSample

	err := open.LoadOpenByCookieBannerID(ctx, tx, scope, bannerID)
	if err == nil {
		return open.ID, true, nil
	}

	if !errors.Is(err, coredata.ErrResourceNotFound) {
		return gid.GID{}, false, fmt.Errorf("cannot load open cookie banner sample: %w", err)
	}

	var samples coredata.CookieBannerSamples

	count, err := samples.CountByCookieBannerID(ctx, tx, scope, bannerID)
	if err != nil {
		return gid.GID{}, false, err
	}

	if count > 0 {
		return gid.GID{}, false, nil
	}

	var banner coredata.CookieBanner
	if err := banner.LoadByID(ctx, tx, scope, bannerID); err != nil {
		return gid.GID{}, false, fmt.Errorf("cannot load cookie banner: %w", err)
	}

	now := time.Now()

	sample := coredata.CookieBannerSample{
		ID:             gid.New(scope.GetTenantID(), coredata.CookieBannerSampleEntityType),
		OrganizationID: banner.OrganizationID,
		CookieBannerID: bannerID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	inserted, err := sample.InsertOpen(ctx, tx, scope)
	if err != nil {
		return gid.GID{}, false, err
	}

	if !inserted {
		if err := open.LoadOpenByCookieBannerID(ctx, tx, scope, bannerID); err != nil {
			return gid.GID{}, false, fmt.Errorf("cannot reload open cookie banner sample: %w", err)
		}

		return open.ID, true, nil
	}

	return sample.ID, true, nil
}

func (s *Service) RunSampleCountFlusher(ctx context.Context, logger *log.Logger) error {
	logger = logger.Named("sample-count-flusher")

	ticker := time.NewTicker(sampleCountFlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			flushCtx := context.WithoutCancel(ctx)
			if err := s.FlushSampleCounts(flushCtx); err != nil {
				logger.ErrorCtx(flushCtx, "cannot flush sample counts on shutdown", log.Error(err))
			}

			return nil
		case <-ticker.C:
			if err := s.FlushSampleCounts(ctx); err != nil {
				logger.ErrorCtx(ctx, "cannot flush sample counts", log.Error(err))
			}
		}
	}
}
