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

	"go.probo.inc/probo/pkg/coredata"
)

const (
	// StaleRunningAfter is how long a running scan may stay silent before
	// recovery pauses it, when it has a checkpoint, or fails it otherwise.
	StaleRunningAfter = 3 * time.Minute

	consentCookieName = "probo_consent"
	consentAction     = "ACCEPT_ALL"
	cookieBinding     = "__proboScanCookie"
)

var ErrNoPageLoaded = errors.New("scanner: no page loaded")

type (
	// Limits bounds one scan. Zero values are replaced by DefaultLimits.
	Limits struct {
		MaxPages          int
		MaxDepth          int
		NavigationTimeout time.Duration
		HomepageDwell     time.Duration
		QuietWindow       time.Duration
		QuietCap          time.Duration
		ScrollQuietCap    time.Duration
		MaxSitemapIndexes int
		MaxSitemapBytes   int
		MaxSitemapURLs    int
		MaxObservations   int
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
	QueueItem struct {
		URL         string `json:"url"`
		Depth       int    `json:"depth"`
		FromSitemap bool   `json:"from_sitemap,omitempty"`
	}

	// Checkpoint is the crawl frontier persisted between pages. It holds
	// no cookie values.
	Checkpoint struct {
		VisitedTemplates []string    `json:"visited_templates"`
		Queue            []QueueItem `json:"queue"`
		PagesScanned     int         `json:"pages_scanned"`
		ElapsedMS        int64       `json:"elapsed_ms"`
		Truncated        bool        `json:"truncated"`
	}

	// Cookie is one observed cookie. Domain is nil for a host-only cookie.
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

		httpClient         *http.Client
		skipPublicURLCheck bool
	}

	// Result is what this invocation observed, plus the frontier to resume.
	// Paused is set when the caller's context is cancelled. Truncated is
	// set when the time budget or the observation cap is hit.
	Result struct {
		Cookies      []Cookie
		Storage      []StorageItem
		Resources    []Resource
		PagesScanned int
		Truncated    bool
		Paused       bool
		Checkpoint   Checkpoint
	}
)

// DefaultLimits returns the production scan bounds.
func DefaultLimits() Limits {
	return Limits{
		MaxPages:          25,
		MaxDepth:          2,
		NavigationTimeout: 30 * time.Second,
		HomepageDwell:     10 * time.Second,
		QuietWindow:       2 * time.Second,
		QuietCap:          15 * time.Second,
		ScrollQuietCap:    10 * time.Second,
		MaxSitemapIndexes: 3,
		MaxSitemapBytes:   5 << 20,
		MaxSitemapURLs:    500,
		MaxObservations:   5000,
		ScanTimeout:       15 * time.Minute,
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

	if l.NavigationTimeout <= 0 {
		l.NavigationTimeout = d.NavigationTimeout
	}

	if l.HomepageDwell <= 0 {
		l.HomepageDwell = d.HomepageDwell
	}

	if l.QuietWindow <= 0 {
		l.QuietWindow = d.QuietWindow
	}

	if l.QuietCap <= 0 {
		l.QuietCap = d.QuietCap
	}

	if l.ScrollQuietCap <= 0 {
		l.ScrollQuietCap = d.ScrollQuietCap
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

	if l.MaxObservations <= 0 {
		l.MaxObservations = d.MaxObservations
	}

	if l.ScanTimeout <= 0 {
		l.ScanTimeout = d.ScanTimeout
	}

	return l
}
