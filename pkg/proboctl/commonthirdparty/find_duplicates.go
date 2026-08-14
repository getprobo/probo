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

package commonthirdparty

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
	seed "go.probo.inc/probo/pkg/proboctl/seed/common-third-parties"
	"go.probo.inc/probo/pkg/thirdparty"
)

// duplicateEntryJSON is one catalog row in the JSON report.
type duplicateEntryJSON struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	Category        string  `json:"category"`
	Seeded          bool    `json:"seeded"`
	TrackerPatterns int     `json:"tracker_patterns"`
	OrgThirdParties int     `json:"org_third_parties"`
	EnrichmentState string  `json:"enrichment_state"`
	Score           float64 `json:"score,omitempty"`
	Reason          string  `json:"reason,omitempty"`
}

// duplicateClusterJSON is one cluster in the JSON report, shaped so a
// consumer can build a merge invocation from it directly.
type duplicateClusterJSON struct {
	Winner duplicateEntryJSON   `json:"winner"`
	Losers []duplicateEntryJSON `json:"losers"`
}

func newCmdFindDuplicates(f *cmdutil.Factory) *cobra.Command {
	var (
		flagMinScore   float64
		flagCategory   string
		flagSeededOnly bool
		flagLimit      int
	)

	cmd := &cobra.Command{
		Use:   "find-duplicates",
		Short: "Report catalog entries that look like duplicates of each other",
		Long: "Scan the global catalog and group entries that appear to be the same " +
			"vendor, with the entry that should absorb the others.\n\n" +
			"Read-only: it reports candidates, it never merges. Pipe --output json " +
			"into `common-third-party merge` to act on a cluster.\n\n" +
			"The default threshold reports only name-identity matches — two entries " +
			"whose names agree once a legal form, a country qualifier, or a " +
			"parenthesised suffix is removed. Lower it to also see domain- and " +
			"prefix-based candidates, which find real duplicates but equally match a " +
			"vendor's legitimate product family (\"Google\" and \"Google Analytics\" " +
			"share a domain and a prefix yet are distinct entries), so those need " +
			"review before merging.",
		Args: cobra.NoArgs,
	}

	output := clicmdutil.AddOutputFlag(cmd)

	cmd.Flags().Float64Var(&flagMinScore, "min-score", thirdparty.DefaultDuplicateMinScore, "Minimum pair score to report (0.55 to 1.0)")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Restrict the scan to one category")
	cmd.Flags().BoolVar(&flagSeededOnly, "seeded-only", false, "Only clusters containing a curated seed entry")
	cmd.Flags().IntVar(&flagLimit, "limit", 50, "Maximum clusters to print (0 for all)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if err := clicmdutil.ValidateOutputFlag(output); err != nil {
			return err
		}

		// Bounded to the range the scoring actually uses. Below the floor
		// nothing scores lower, so a smaller value only misleads; and NaN
		// would pass every comparison in the threshold check, turning
		// zero-score pairs into clusters.
		if math.IsNaN(flagMinScore) || math.IsInf(flagMinScore, 0) {
			return fmt.Errorf("invalid --min-score value %v: must be a finite number", flagMinScore)
		}

		if flagMinScore < thirdparty.DuplicateScoreBrandPrefix || flagMinScore > thirdparty.DuplicateScoreExactName {
			return fmt.Errorf(
				"invalid --min-score value %v: must be between %.2f and %.2f",
				flagMinScore,
				thirdparty.DuplicateScoreBrandPrefix,
				thirdparty.DuplicateScoreExactName,
			)
		}

		filter := coredata.NewCommonThirdPartyFilter(nil)

		if flagCategory != "" {
			category := coredata.ThirdPartyCategory(flagCategory)
			if !category.IsValid() {
				return fmt.Errorf("invalid --category value %q", flagCategory)
			}

			filter.WithCategory(&category)
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		seededSlugs, err := seed.SeededSlugs()
		if err != nil {
			return fmt.Errorf("cannot load curated seed slugs: %w", err)
		}

		entries, err := loadCatalogEntries(ctx, pgClient, filter, seededSlugs)
		if err != nil {
			return err
		}

		clusters := thirdparty.FindDuplicates(entries, flagMinScore)

		if flagSeededOnly {
			clusters = filterSeededClusters(clusters)
		}

		out := f.IOStreams.Out
		errOut := f.IOStreams.ErrOut

		covered := 0
		for _, c := range clusters {
			covered += 1 + len(c.Losers)
		}

		_, _ = fmt.Fprintf(
			errOut,
			"Scanned %d catalog row(s), found %d cluster(s) covering %d row(s) at min-score %.2f.\n",
			len(entries),
			len(clusters),
			covered,
			flagMinScore,
		)

		shown := clusters
		if flagLimit > 0 && len(shown) > flagLimit {
			shown = shown[:flagLimit]
			_, _ = fmt.Fprintf(errOut, "Printing the first %d; pass --limit 0 for all.\n", flagLimit)
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(out, buildDuplicateJSON(shown))
		}

		if len(shown) == 0 {
			_, _ = fmt.Fprintln(out, "No duplicate candidates found.")
			return nil
		}

		printDuplicateClusters(cmd, out, shown)

		return nil
	}

	return cmd
}

// loadCatalogEntries assembles the whole catalog with the reference counts
// and domains duplicate detection ranks on.
//
// The reference counts are fetched as two histograms rather than per row:
// the scan covers every catalog entry, so per-row counting would be one
// round trip each. Organization third parties are counted across all
// tenants because the catalog is global.
func loadCatalogEntries(
	ctx context.Context,
	pgClient *pg.Client,
	filter *coredata.CommonThirdPartyFilter,
	seededSlugs map[string]struct{},
) ([]*thirdparty.CatalogEntry, error) {
	var entries []*thirdparty.CatalogEntry

	if err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			parties, err := page.LoadAll(
				ctx,
				page.OrderBy[coredata.CommonThirdPartyOrderField]{
					Field:     coredata.CommonThirdPartyOrderFieldName,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.CommonThirdPartyOrderField]) ([]*coredata.CommonThirdParty, error) {
					var batch coredata.CommonThirdParties
					if err := batch.Load(ctx, conn, cursor, filter); err != nil {
						return nil, fmt.Errorf("cannot load common third parties: %w", err)
					}

					return batch, nil
				},
			)
			if err != nil {
				return err
			}

			var patterns coredata.CommonTrackerPatterns

			patternCounts, err := patterns.CountByCommonThirdPartyID(ctx, conn)
			if err != nil {
				return err
			}

			var orgParties coredata.ThirdParties

			orgCounts, err := orgParties.CountByCommonThirdPartyID(ctx, conn, coredata.NewNoScope())
			if err != nil {
				return err
			}

			var domains coredata.CommonThirdPartyDomains

			domainsByParty, err := domains.LoadAllGroupedByCommonThirdPartyID(ctx, conn)
			if err != nil {
				return err
			}

			entries = make([]*thirdparty.CatalogEntry, 0, len(parties))

			for _, party := range parties {
				_, seeded := seededSlugs[party.Slug]

				entries = append(entries, &thirdparty.CatalogEntry{
					Party:           party,
					Domains:         domainsByParty[party.ID],
					TrackerPatterns: patternCounts[party.ID],
					OrgThirdParties: orgCounts[party.ID],
					Seeded:          seeded,
				})
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return entries, nil
}

// filterSeededClusters keeps only clusters touching a curated entry, which
// is where merging is both safest and most valuable: the curated row is the
// one the next seed run recreates.
func filterSeededClusters(clusters []thirdparty.DuplicateCluster) []thirdparty.DuplicateCluster {
	kept := make([]thirdparty.DuplicateCluster, 0, len(clusters))

	for _, c := range clusters {
		seeded := c.Winner.Seeded

		for _, l := range c.Losers {
			seeded = seeded || l.Seeded
		}

		if seeded {
			kept = append(kept, c)
		}
	}

	return kept
}

// scoreForEntry returns the best pair score and reason connecting an entry
// to the rest of its cluster, so each loser can be shown with the evidence
// that put it there.
func scoreForEntry(cluster thirdparty.DuplicateCluster, entry *thirdparty.CatalogEntry) (float64, string) {
	var (
		best   float64
		reason string
	)

	for _, p := range cluster.Pairs {
		if p.A != entry.Party.ID && p.B != entry.Party.ID {
			continue
		}

		if p.Score > best {
			best, reason = p.Score, p.Reason
		}
	}

	return best, reason
}

func buildDuplicateJSON(clusters []thirdparty.DuplicateCluster) []duplicateClusterJSON {
	out := make([]duplicateClusterJSON, 0, len(clusters))

	for _, c := range clusters {
		cluster := duplicateClusterJSON{
			Winner: duplicateEntryJSON{
				ID:              c.Winner.Party.ID.String(),
				Name:            c.Winner.Party.Name,
				Slug:            c.Winner.Party.Slug,
				Category:        string(c.Winner.Party.Category),
				Seeded:          c.Winner.Seeded,
				TrackerPatterns: c.Winner.TrackerPatterns,
				OrgThirdParties: c.Winner.OrgThirdParties,
				EnrichmentState: enrichmentState(c.Winner.Party),
			},
			Losers: make([]duplicateEntryJSON, 0, len(c.Losers)),
		}

		for _, l := range c.Losers {
			score, reason := scoreForEntry(c, l)

			cluster.Losers = append(cluster.Losers, duplicateEntryJSON{
				ID:              l.Party.ID.String(),
				Name:            l.Party.Name,
				Slug:            l.Party.Slug,
				Category:        string(l.Party.Category),
				Seeded:          l.Seeded,
				TrackerPatterns: l.TrackerPatterns,
				OrgThirdParties: l.OrgThirdParties,
				EnrichmentState: enrichmentState(l.Party),
				Score:           score,
				Reason:          reason,
			})
		}

		out = append(out, cluster)
	}

	return out
}

func printDuplicateClusters(
	cmd *cobra.Command,
	out io.Writer,
	clusters []thirdparty.DuplicateCluster,
) {
	for i, c := range clusters {
		if i > 0 {
			_, _ = fmt.Fprintln(out)
		}

		_, _ = fmt.Fprintf(
			out,
			"WINNER  %s (%s)%s  patterns=%d org=%d  %s\n",
			c.Winner.Party.Name,
			c.Winner.Party.Slug,
			seededMarker(c.Winner.Seeded),
			c.Winner.TrackerPatterns,
			c.Winner.OrgThirdParties,
			enrichmentState(c.Winner.Party),
		)

		for _, l := range c.Losers {
			score, reason := scoreForEntry(c, l)

			_, _ = fmt.Fprintf(
				out,
				"  loser %s (%s)%s  patterns=%d org=%d  score=%.2f  %s\n",
				l.Party.Name,
				l.Party.Slug,
				seededMarker(l.Seeded),
				l.TrackerPatterns,
				l.OrgThirdParties,
				score,
				reason,
			)
		}

		slugs := make([]string, 0, len(c.Losers))
		for _, l := range c.Losers {
			slugs = append(slugs, l.Party.Slug)
		}

		_, _ = fmt.Fprintf(
			out,
			"  merge:  %s common-third-party merge --into %s %s --dry-run\n",
			cmd.Root().Name(),
			c.Winner.Party.Slug,
			strings.Join(slugs, " "),
		)
	}
}

func seededMarker(seeded bool) string {
	if seeded {
		return " [seeded]"
	}

	return ""
}
