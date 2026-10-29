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

// Package scanner crawls one cookie banner origin in an isolated Chrome
// context and records the cookies, storage keys and resources it observes.
// It does not touch the database. The caller ingests each finished page.
package scanner

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// StaleRunningAfter is how long a running scan may stay silent before
	// recovery pauses it, when it has a checkpoint, or fails it otherwise.
	StaleRunningAfter = 3 * time.Minute

	consentCookieName = "probo_consent"
	consentAction     = "ACCEPT_ALL"

	maxNameBytes = 256
	maxURLBytes  = 2048
	readTimeout  = 5 * time.Second
)

var (
	// ErrNoPageLoaded is returned when a scan ends, other than by pause,
	// without a single page finishing. It wraps the first page error when
	// there is one, so callers can also match ErrUnreachable or ErrBlocked.
	ErrNoPageLoaded = errors.New("scanner: no page loaded")

	// ErrUnreachable means a page could not be loaded because of the
	// network, TLS, or a timeout.
	ErrUnreachable = errors.New("scanner: page unreachable")

	// ErrBlocked means the site refused the scan, or the scanner refused
	// the navigation, for example with HTTP 401, 403, 429 or 503.
	ErrBlocked = errors.New("scanner: page blocked")
)

type (
	// Limits bounds one scan. Zero values are replaced by DefaultLimits.
	// MaxPages and ScanTimeout cover every run of a scan. PagesPerRun and
	// RunTimeout bound one call to Run, which then pauses so the browser
	// can serve other scans.
	Limits struct {
		MaxPages          int
		MaxDepth          int
		PagesPerRun       int
		RunTimeout        time.Duration
		NavigationTimeout time.Duration
		HomepageDwell     time.Duration
		IdleWindow        time.Duration
		LoadIdleTimeout   time.Duration
		ScrollIdleTimeout time.Duration
		MaxSitemapIndexes int
		MaxSitemapBytes   int
		MaxSitemapURLs    int
		MaxSitemapFetches int
		MaxProbes         int
		ProbeConcurrency  int
		ProbeTimeout      time.Duration
		ProbeBudget       time.Duration
		MaxProbeBytes     int
		MaxObservations   int
		MaxQueue          int
		ScanTimeout       time.Duration
	}

	// Consent is the published banner version to seed. Nil means the scan
	// runs with no consent cookie, which is the right behavior when no
	// version is published: nothing is blocked, so every tracker fires.
	Consent struct {
		Version       int
		CategorySlugs []string
	}

	// QueueItem is one not-yet-visited URL in the crawl frontier.
	// EmbedHosts are the third-party hosts its static HTML embeds.
	QueueItem struct {
		URL         string   `json:"url"`
		Depth       int      `json:"depth"`
		FromSitemap bool     `json:"from_sitemap,omitempty"`
		EmbedHosts  []string `json:"embed_hosts,omitempty"`
	}

	// Checkpoint is the crawl frontier persisted between pages. It holds
	// no cookie values. CoveredHosts are the third-party hosts visited
	// pages embedded or requested.
	Checkpoint struct {
		VisitedTemplates []string    `json:"visited_templates"`
		Queue            []QueueItem `json:"queue"`
		CoveredHosts     []string    `json:"covered_hosts,omitempty"`
		PagesScanned     int         `json:"pages_scanned"`
		ElapsedMS        int64       `json:"elapsed_ms"`
		Truncated        bool        `json:"truncated"`
	}

	// Cookie is one observed cookie. Domain is nil for a host-only cookie
	// whose host is not known.
	Cookie struct {
		Name          string
		Domain        *string
		HostOnly      *bool
		MaxAgeSeconds *int
		Source        coredata.CookieSource
		HTTPOnly      *bool
		Secure        *bool
		SameSite      *coredata.CookieSameSite
		Partitioned   *bool
		SetByURL      *string
		InitiatorURL  *string
		ValueSize     int
	}

	// StorageItem is one localStorage or sessionStorage key.
	StorageItem struct {
		Type      coredata.TrackerType
		Key       string
		ValueSize int
	}

	// Resource is one requested tracker resource.
	Resource struct {
		Type         coredata.TrackerResourceType
		Origin       string
		Path         string
		InitiatorURL *string
	}

	// Page is the new observations from one finished page.
	Page struct {
		URL       string
		Cookies   []Cookie
		Storage   []StorageItem
		Resources []Resource
	}

	// OnPage is called after each page that finished loading. The context
	// is detached from scan cancellation so a shutdown can still persist
	// that page. The checkpoint includes the page. The page in flight when
	// the scan context is cancelled is not reported.
	OnPage func(ctx context.Context, page Page, checkpoint Checkpoint) error

	// Input is one scan. Checkpoint nil or empty starts at the origin root.
	Input struct {
		Origin     string
		BannerID   string
		APIBaseURL string
		Consent    *Consent
		Limits     Limits
		Checkpoint *Checkpoint
		OnPage     OnPage
		Logger     *log.Logger
		// BrowserURL is the DevTools websocket of the Chrome instance.
		// Empty skips instrumentation of cross-site iframes and workers.
		BrowserURL string

		httpClient         *http.Client
		skipPublicURLCheck bool
	}

	// Result is the frontier to resume and how far this invocation got.
	// Observations are delivered only through OnPage. Paused is set when
	// the caller's context is cancelled or the run's page or time budget is
	// spent with pages left. Truncated is set when the scan's time budget,
	// page cap, or observation cap is hit. A paused result is not also
	// truncated. The checkpoint still records the cap, so a resume stops.
	Result struct {
		PagesScanned int
		Truncated    bool
		Paused       bool
		Checkpoint   Checkpoint
	}
)

// DefaultLimits returns the production scan bounds.
func DefaultLimits() Limits {
	return Limits{
		MaxPages:          100,
		MaxDepth:          2,
		PagesPerRun:       10,
		RunTimeout:        2 * time.Minute,
		NavigationTimeout: 30 * time.Second,
		HomepageDwell:     10 * time.Second,
		IdleWindow:        2 * time.Second,
		LoadIdleTimeout:   15 * time.Second,
		ScrollIdleTimeout: 10 * time.Second,
		MaxSitemapIndexes: 3,
		MaxSitemapBytes:   5 << 20,
		MaxSitemapURLs:    500,
		MaxSitemapFetches: 10,
		MaxProbes:         150,
		ProbeConcurrency:  4,
		ProbeTimeout:      10 * time.Second,
		ProbeBudget:       time.Minute,
		MaxProbeBytes:     2 << 20,
		MaxObservations:   5000,
		MaxQueue:          500,
		ScanTimeout:       30 * time.Minute,
	}
}

func (l Limits) withDefaults() Limits {
	d := DefaultLimits()
	if l.MaxPages <= 0 {
		l.MaxPages = d.MaxPages
	}

	if l.MaxDepth <= 0 {
		l.MaxDepth = d.MaxDepth
	}

	if l.PagesPerRun <= 0 {
		l.PagesPerRun = d.PagesPerRun
	}

	if l.RunTimeout <= 0 {
		l.RunTimeout = d.RunTimeout
	}

	if l.NavigationTimeout <= 0 {
		l.NavigationTimeout = d.NavigationTimeout
	}

	if l.HomepageDwell <= 0 {
		l.HomepageDwell = d.HomepageDwell
	}

	if l.IdleWindow <= 0 {
		l.IdleWindow = d.IdleWindow
	}

	if l.LoadIdleTimeout <= 0 {
		l.LoadIdleTimeout = d.LoadIdleTimeout
	}

	if l.ScrollIdleTimeout <= 0 {
		l.ScrollIdleTimeout = d.ScrollIdleTimeout
	}

	if l.MaxSitemapIndexes <= 0 {
		l.MaxSitemapIndexes = d.MaxSitemapIndexes
	}

	if l.MaxSitemapBytes <= 0 {
		l.MaxSitemapBytes = d.MaxSitemapBytes
	}

	if l.MaxSitemapURLs <= 0 {
		l.MaxSitemapURLs = d.MaxSitemapURLs
	}

	if l.MaxSitemapFetches <= 0 {
		l.MaxSitemapFetches = d.MaxSitemapFetches
	}

	if l.MaxProbes <= 0 {
		l.MaxProbes = d.MaxProbes
	}

	if l.ProbeConcurrency <= 0 {
		l.ProbeConcurrency = d.ProbeConcurrency
	}

	if l.ProbeTimeout <= 0 {
		l.ProbeTimeout = d.ProbeTimeout
	}

	if l.ProbeBudget <= 0 {
		l.ProbeBudget = d.ProbeBudget
	}

	if l.MaxProbeBytes <= 0 {
		l.MaxProbeBytes = d.MaxProbeBytes
	}

	if l.MaxObservations <= 0 {
		l.MaxObservations = d.MaxObservations
	}

	if l.MaxQueue <= 0 {
		l.MaxQueue = d.MaxQueue
	}

	if l.ScanTimeout <= 0 {
		l.ScanTimeout = d.ScanTimeout
	}

	return l
}
