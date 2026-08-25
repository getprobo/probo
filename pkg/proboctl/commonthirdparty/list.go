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
	"strings"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName     string
		flagCategory string
		flagKeyword  string
		flagState    string
		flagStatus   string
		flagReview   string
		flagPatterns bool
		flagSort     string
		flagOrder    string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List common third parties with filters and sorting",
		Args:  cobra.NoArgs,
	}

	output := clicmdutil.AddOutputFlag(cmd)

	cmd.Flags().StringVar(&flagName, "name", "", "Filter by name substring")
	cmd.Flags().StringVar(&flagCategory, "category", "", "Filter by category")
	cmd.Flags().StringVar(&flagKeyword, "keyword", "", "Filter by name/slug substring")
	cmd.Flags().StringVar(&flagState, "state", "", "Filter by enrichment state (queued, enriched, unenriched)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Filter by last enrichment status (done, partial, failed)")
	cmd.Flags().StringVar(&flagReview, "review", "", "Filter by review state (unreviewed, validated, rejected)")
	cmd.Flags().BoolVar(&flagPatterns, "with-patterns", false, "Show each entry's tracker pattern keys, so a review can be judged from the listing instead of one lookup per row")
	cmd.Flags().StringVar(&flagSort, "sort", "name", "Sort field: name, created, updated")
	cmd.Flags().StringVar(&flagOrder, "order", "", "Sort order: asc, desc (default depends on field)")

	pageFlags := cmdutil.AddPageFlags(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := clicmdutil.ValidateOutputFlag(output); err != nil {
			return err
		}

		orderBy, err := parseOrderBy(flagSort, flagOrder)
		if err != nil {
			return err
		}

		cursor, err := cmdutil.NewCursorFromFlags(pageFlags, orderBy)
		if err != nil {
			return err
		}

		filter := coredata.NewCommonThirdPartyFilter(optionalString(flagName))

		if flagCategory != "" {
			cat := coredata.ThirdPartyCategory(flagCategory)
			if !cat.IsValid() {
				return fmt.Errorf("invalid --category value %q", flagCategory)
			}

			filter.WithCategory(&cat)
		}

		if flagKeyword != "" {
			filter.WithKeyword(&flagKeyword)
		}

		if flagReview != "" {
			review, err := parseReview(flagReview)
			if err != nil {
				return err
			}

			filter.WithReview(&review)
		}

		if flagState != "" {
			st, err := parseEnrichmentState(flagState)
			if err != nil {
				return err
			}

			filter.WithState(&st)
		}

		if flagStatus != "" {
			if _, ok := validEnrichmentStatuses[flagStatus]; !ok {
				return fmt.Errorf("invalid --status value %q: valid values are done, partial, failed", flagStatus)
			}

			filter.WithEnrichmentStatus(&flagStatus)
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		var (
			parties  coredata.CommonThirdParties
			pageInfo cmdutil.PageInfo
		)

		if err := pgClient.WithConn(
			cmd.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				p, err := cmdutil.FetchPage(
					ctx,
					cursor,
					func(ctx context.Context, cursor *page.Cursor[coredata.CommonThirdPartyOrderField]) ([]*coredata.CommonThirdParty, error) {
						var ts coredata.CommonThirdParties
						if err := ts.Load(ctx, conn, cursor, filter); err != nil {
							return nil, err
						}

						return ts, nil
					},
				)
				if err != nil {
					return err
				}

				parties = p.Data
				pageInfo = cmdutil.NewPageInfo(p)

				return nil
			},
		); err != nil {
			return err
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, cmdutil.PageOutput{Items: parties, PageInfo: pageInfo})
		}

		if len(parties) == 0 {
			_, _ = fmt.Fprintln(f.IOStreams.Out, "No common third parties found.")
			return nil
		}

		// The pattern keys are what a review is judged on, so with
		// --with-patterns they replace the enrichment columns rather than
		// widening an already wide table. One grouped query serves the whole
		// page: per-row lookups are what make reading the evidence slower
		// than guessing from the name.
		if flagPatterns {
			var byParty map[gid.GID][]coredata.PatternSummary

			if err := pgClient.WithConn(
				cmd.Context(),
				func(ctx context.Context, conn pg.Querier) error {
					var patterns coredata.CommonTrackerPatterns

					var err error

					byParty, err = patterns.LoadSummariesGroupedByCommonThirdPartyID(ctx, conn)

					return err
				},
			); err != nil {
				return fmt.Errorf("cannot load pattern summaries: %w", err)
			}

			table := clicmdutil.NewTable("NAME", "SLUG", "CATEGORY", "REVIEW", "PATTERNS")

			for _, p := range parties {
				table.Row(
					p.Name,
					p.Slug,
					string(p.Category),
					reviewSummary(p),
					summarisePatterns(byParty[p.ID]),
				)
			}

			_, _ = fmt.Fprintln(f.IOStreams.Out, table.Render())
			cmdutil.PrintPageInfo(f.IOStreams.Out, pageInfo)
			_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "Showing %d common third parties.\n", len(parties))

			return nil
		}

		table := clicmdutil.NewTable("ID", "NAME", "SLUG", "CATEGORY", "REVIEW", "STATE", "STATUS", "LAST ATTEMPT", "UPDATED")

		for _, p := range parties {
			lastAttempt := ""
			if p.LastEnrichmentAttemptAt != nil {
				lastAttempt = p.LastEnrichmentAttemptAt.Format("2006-01-02 15:04:05")
			}

			table.Row(
				p.ID.String(),
				p.Name,
				p.Slug,
				string(p.Category),
				reviewSummary(p),
				enrichmentState(p),
				enrichmentStatus(p),
				lastAttempt,
				p.UpdatedAt.Format("2006-01-02 15:04:05"),
			)
		}

		_, _ = fmt.Fprintln(f.IOStreams.Out, table.Render())
		cmdutil.PrintPageInfo(f.IOStreams.Out, pageInfo)
		_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "Showing %d common third parties.\n", len(parties))

		return nil
	}

	return cmd
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}

// summarisePatterns renders the few keys that decide a verdict. Highest
// confidence first (the loader orders them), capped because a handful is
// enough to classify a row and a shared-library namespace runs to dozens.
func summarisePatterns(summaries []coredata.PatternSummary) string {
	const shown = 4

	if len(summaries) == 0 {
		return "(none)"
	}

	parts := make([]string, 0, shown+1)

	for i, s := range summaries {
		if i == shown {
			parts = append(parts, fmt.Sprintf("+%d more", len(summaries)-shown))
			break
		}

		parts = append(parts, fmt.Sprintf("%s [%s]", s.Pattern, storageAbbrev(s.TrackerType)))
	}

	return strings.Join(parts, ", ")
}

// storageAbbrev keeps the storage kind visible without the column dominating
// the row: where a key lives distinguishes a cookie sent to a vendor from
// local state that never leaves the browser.
func storageAbbrev(trackerType coredata.TrackerType) string {
	switch trackerType {
	case coredata.TrackerTypeCookie:
		return "ck"
	case coredata.TrackerTypeLocalStorage:
		return "ls"
	case coredata.TrackerTypeSessionStorage:
		return "ss"
	case coredata.TrackerTypeIndexedDB:
		return "idb"
	case coredata.TrackerTypeCacheStorage:
		return "cs"
	}

	return strings.ToLower(string(trackerType))
}
