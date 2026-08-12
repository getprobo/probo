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
	"go.probo.inc/probo/pkg/stringsx"
	"go.probo.inc/probo/pkg/uri"
)

// bundledFirstPartyLibraries holds alphanumeric-normalised names of
// client-side libraries that write local keys but egress no data to any
// endpoint their maintainers operate. They ship inside a site's own
// JavaScript bundle, so the key they lay is the site's own state: the
// project has a name, but there is no data recipient and therefore no
// third party to catalog.
//
// Membership is decided by one question only: does the project operate an
// ingestion endpoint its client code reports to? A library that receives
// nothing belongs here. Any project running a hosted service its SDK
// transmits to is a genuine third party and must never appear here —
// Sentry, PostHog, Mixpanel, Segment, Amplitude, and Vercel are all
// curated catalog vendors. That exclusion mirrors how
// cookieDatabaseAggregators deliberately omits consent-management vendors
// that do set their own product cookies.
//
// The list is deliberately narrow: it pins the recurring offenders the
// identification prompt has been observed to miss, not every library that
// feels first-party. The prompt's egress test handles the long tail. When
// in doubt about a project, leave it out — a missing entry costs one
// discarded attribution the prompt usually catches anyway, while a wrong
// entry silently suppresses a real vendor across every tenant.
//
// Keys are stored already-normalised because stringsx.NormalizeAlnum
// folds away hyphens, dots, and casing (so "react-i18next" arrives as
// "reacti18next").
var bundledFirstPartyLibraries = map[string]struct{}{
	// Internationalisation: stores the selected locale locally.
	"i18next":      {},
	"i18nextlng":   {},
	"reacti18next": {},
	"vuei18n":      {},
	"formatjs":     {},

	// Logging and debugging: stores a verbosity level or namespace filter.
	"loglevel": {},
	"debug":    {},

	// State management and persistence: rehydrates a local store.
	"redux":        {},
	"reduxpersist": {},
	"mobx":         {},
	"zustand":      {},
	"pinia":        {},
	"vuex":         {},

	// UI, utility, and date libraries.
	"jquery":     {},
	"lodash":     {},
	"underscore": {},
	"bootstrap":  {},
	"moment":     {},
	"momentjs":   {},
	"dayjs":      {},
	"datefns":    {},

	// Build and module tooling runtimes.
	"webpack": {},
	"vite":    {},
	"rollup":  {},
	"corejs":  {},
	"babel":   {},
}

// nameIsBundledFirstPartyLibrary reports whether a candidate vendor name is
// a bundled client-side library that egresses nothing and therefore must
// never become a catalog third party. The agent may return a bare project
// name ("i18next"), a package form ("react-i18next"), or a domain form
// ("i18next.com"), so the candidate is checked both as-is and reduced to
// its primary domain label. Both lookups are alphanumeric-normalised so
// spacing, punctuation, and casing do not matter.
func nameIsBundledFirstPartyLibrary(name string) bool {
	if _, ok := bundledFirstPartyLibraries[stringsx.NormalizeAlnum(name)]; ok {
		return true
	}

	label := stringsx.NormalizeAlnum(uri.DomainLabel(name))
	if label == "" {
		return false
	}

	_, ok := bundledFirstPartyLibraries[label]

	return ok
}
