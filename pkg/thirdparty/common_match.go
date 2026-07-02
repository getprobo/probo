// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package thirdparty

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/slug"
	"go.probo.inc/probo/pkg/uri"
)

// Heuristic thresholds for matching a CommonThirdParty to an existing
// org ThirdParty. Exported so callers can short-circuit explicitly
// (skip the agent, leave the pattern unlinked, etc.) instead of
// duplicating magic numbers.
const (
	// HighConfidenceScore is the floor at which a heuristic match
	// is treated as obvious (exact name, suffix-stripped name, slug
	// equality). Callers typically link without consulting the
	// agent at this score.
	HighConfidenceScore = 0.85

	// MinAgentScore is the floor below which a candidate is
	// statistical noise and should not be shown to the
	// disambiguation agent. Below this, callers typically prefer
	// to create a fresh row over asking the model to pick among
	// weak candidates.
	MinAgentScore = 0.6

	// MaxAgentCandidates caps the candidate list shown to the
	// disambiguation agent. The list is heuristic-ranked, so the
	// top few are the only ones worth the agent's tokens.
	MaxAgentCandidates = 5
)

// ScoredCandidate is a scored heuristic-match candidate. It is the
// unified currency between the heuristic ranker (RankCandidates) and
// the disambiguation agent (Disambiguate): the agent renders the
// `ThirdParty` fields plus the score directly into its prompt, with
// no intermediate DTO.
type ScoredCandidate struct {
	ThirdParty *coredata.ThirdParty
	Score      float64
}

// corporateSuffixes are the legal-form noise words stripped when
// comparing third-party names heuristically. The list is intentionally
// short and conservative: matching "Foo Inc" to "Foo" is safe, but
// stripping "Group" or "Services" would over-match unrelated entries.
//
// Order matters: stripCorporateSuffixes returns on the first match,
// so longer / comma-prefixed forms must come before their shorter
// siblings (", inc." before " inc.", which itself comes before " inc").
var corporateSuffixes = []string{
	" incorporated",
	" corporation",
	", inc.",
	", inc",
	" l.l.c.",
	" s.a.s.",
	" inc.",
	" inc",
	" llc",
	" ltd.",
	" ltd",
	" limited",
	" gmbh",
	" s.a.",
	" sas",
	" sa",
	" ag",
	" plc",
	" corp.",
	" corp",
	" co.",
	" co",
	" b.v.",
	" bv",
}

// RankCandidates ranks org ThirdParty rows by how likely each is to
// represent the given CommonThirdParty. Returned slice is sorted by
// descending score; only candidates with score > 0 are kept. Pure
// function: no I/O, deterministic on its inputs.
//
// Scoring (highest match wins; website-host overlap can lift a name
// miss to 0.8):
//
//   - exact lowercase name                                      = 1.0
//   - lowercase name with corporate suffix stripped, equal      = 0.9
//   - slug equality (slug.Make on the org's name)               = 0.85
//   - website host (eTLD+1) overlap with the catalog domain set = 0.8
func RankCandidates(
	commonParty coredata.CommonThirdParty,
	commonDomains coredata.CommonThirdPartyDomains,
	candidates coredata.ThirdParties,
) []ScoredCandidate {
	commonName := strings.ToLower(strings.TrimSpace(commonParty.Name))
	commonStripped := stripCorporateSuffixes(commonName)
	commonSlug := commonParty.Slug

	commonHost := ""
	if commonParty.WebsiteURL != nil {
		commonHost = uri.ExtractDomain(*commonParty.WebsiteURL)
	}

	commonDomainSet := make(map[string]struct{}, len(commonDomains))
	for _, d := range commonDomains {
		commonDomainSet[strings.ToLower(d.Domain)] = struct{}{}
	}

	if commonHost != "" {
		commonDomainSet[commonHost] = struct{}{}
	}

	scored := make([]ScoredCandidate, 0, len(candidates))

	for _, tp := range candidates {
		score := 0.0

		orgName := strings.ToLower(strings.TrimSpace(tp.Name))
		orgStripped := stripCorporateSuffixes(orgName)

		switch {
		case orgName != "" && orgName == commonName:
			score = 1.0
		case orgStripped != "" && orgStripped == commonStripped:
			score = 0.9
		case commonSlug != "" && slug.Make(tp.Name) == commonSlug:
			score = 0.85
		}

		if tp.WebsiteURL != nil {
			orgHost := uri.ExtractDomain(*tp.WebsiteURL)
			if orgHost != "" {
				if _, hit := commonDomainSet[orgHost]; hit {
					if score < 0.8 {
						score = 0.8
					}
				}
			}
		}

		if score == 0 {
			continue
		}

		scored = append(scored, ScoredCandidate{
			ThirdParty: tp,
			Score:      score,
		})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	return scored
}

// stripCorporateSuffixes removes a single trailing legal-form suffix
// from a lowercased name. Only one suffix is stripped to avoid
// mangling names that happen to end in two stop-words (e.g. "Foo Inc
// LLC" → "Foo Inc", not "Foo").
func stripCorporateSuffixes(lowerName string) string {
	for _, s := range corporateSuffixes {
		if before, ok := strings.CutSuffix(lowerName, s); ok {
			return strings.TrimSpace(before)
		}
	}

	return lowerName
}

// LinkToCommon writes common_third_party_id onto an org ThirdParty so
// future matches against the same CommonThirdParty can short-circuit
// to the exact-link path in O(1). No-op when the field is already set
// (to any value) — we never overwrite an existing catalog link because
// a heuristic or agent false-positive must not corrupt a previous,
// possibly more accurate, association.
func LinkToCommon(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	orgThirdParty *coredata.ThirdParty,
	commonID gid.GID,
) error {
	if orgThirdParty.CommonThirdPartyID != nil {
		return nil
	}

	orgThirdParty.CommonThirdPartyID = &commonID

	if err := orgThirdParty.Update(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update third party with common id: %w", err)
	}

	return nil
}
