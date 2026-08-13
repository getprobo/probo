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

// bundledFirstPartyLibraries holds client-side libraries that write local
// keys but egress nothing, so the key they lay is the site's own state: the
// project has a name, but no data recipient and therefore no third party.
//
// Membership turns on one question: does the project operate an ingestion
// endpoint its client code reports to? Any project running a hosted service
// its SDK transmits to is a genuine third party and must never appear here —
// Sentry, PostHog, Mixpanel, Segment, Amplitude and Vercel are all curated
// catalog vendors.
//
// Deliberately narrow: it pins the offenders the prompt has been observed to
// miss, not every library that feels first-party. When in doubt leave a project
// out, since a missing entry costs one discarded attribution while a wrong one
// suppresses a real vendor across every tenant.
//
// Keys are pre-normalised because NormalizeAlnum folds away hyphens, dots and
// casing ("react-i18next" arrives as "reacti18next").
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

// nameIsBundledFirstPartyLibrary reports whether a candidate vendor name is a
// bundled library that egresses nothing. The agent may return a bare project
// name ("i18next"), a package form ("react-i18next"), or a domain form
// ("i18next.com"), so the candidate is checked as-is and reduced to its
// primary domain label.
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
