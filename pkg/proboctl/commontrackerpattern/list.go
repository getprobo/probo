// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package commontrackerpattern

import (
	"context"
	"fmt"

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
		flagTrackerType string
		flagMatchType   string
		flagThirdParty  string
		flagKeyword     string
		flagState       string
		flagLinked      bool
		flagUnlinked    bool
		flagSort        string
		flagOrder       string
		flagLimit       int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List common tracker patterns with filters and sorting",
		Args:  cobra.NoArgs,
	}

	output := clicmdutil.AddOutputFlag(cmd)

	cmd.Flags().StringVar(&flagTrackerType, "tracker-type", "", "Filter by tracker type (COOKIE, LOCAL_STORAGE, SESSION_STORAGE, INDEXED_DB)")
	cmd.Flags().StringVar(&flagMatchType, "match-type", "", "Filter by match type (EXACT, GLOB, PREFIX)")
	cmd.Flags().StringVar(&flagThirdParty, "third-party", "", "Filter by linked common third party (slug or GID)")
	cmd.Flags().StringVar(&flagKeyword, "keyword", "", "Filter by pattern/description substring")
	cmd.Flags().StringVar(&flagState, "state", "", "Filter by enrichment state (queued, enriched, unenriched)")
	cmd.Flags().BoolVar(&flagLinked, "linked", false, "Only patterns linked to a common third party")
	cmd.Flags().BoolVar(&flagUnlinked, "unlinked", false, "Only patterns not linked to a common third party")
	cmd.Flags().StringVar(&flagSort, "sort", "confidence", "Sort field: pattern, confidence, created, updated, enriched")
	cmd.Flags().StringVar(&flagOrder, "order", "", "Sort order: asc, desc (default depends on field)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 50, "Maximum rows to return (0 for all)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := clicmdutil.ValidateOutputFlag(output); err != nil {
			return err
		}

		if flagLinked && flagUnlinked {
			return fmt.Errorf("--linked and --unlinked are mutually exclusive")
		}

		orderBy, err := parseOrderBy(flagSort, flagOrder)
		if err != nil {
			return err
		}

		filter, err := buildListFilter(flagTrackerType, flagMatchType, flagKeyword, flagState, flagLinked, flagUnlinked)
		if err != nil {
			return err
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()

		var patterns coredata.CommonTrackerPatterns

		if err := pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				if flagThirdParty != "" {
					id, err := resolveCommonThirdPartyID(ctx, conn, flagThirdParty)
					if err != nil {
						return err
					}

					filter.WithCommonThirdPartyID(&id)
				}

				rows, err := cmdutil.Paginate(
					ctx,
					orderBy,
					flagLimit,
					func(ctx context.Context, cursor *page.Cursor[coredata.CommonTrackerPatternOrderField]) ([]*coredata.CommonTrackerPattern, error) {
						var ps coredata.CommonTrackerPatterns
						if err := ps.Load(ctx, conn, cursor, filter); err != nil {
							return nil, err
						}

						return ps, nil
					},
				)
				if err != nil {
					return err
				}

				patterns = rows

				return nil
			},
		); err != nil {
			return err
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, patterns)
		}

		return renderPatternTable(cmd, f, patterns)
	}

	return cmd
}

func renderPatternTable(cmd *cobra.Command, f *cmdutil.Factory, patterns coredata.CommonTrackerPatterns) error {
	out := f.IOStreams.Out

	if len(patterns) == 0 {
		_, _ = fmt.Fprintln(out, "No common tracker patterns found.")
		return nil
	}

	var linkedIDs []gid.GID

	for _, p := range patterns {
		if p.CommonThirdPartyID != nil {
			linkedIDs = append(linkedIDs, *p.CommonThirdPartyID)
		}
	}

	pgClient, err := f.PgClient()
	if err != nil {
		return err
	}

	var names map[gid.GID]string

	if err := pgClient.WithConn(
		cmd.Context(),
		func(ctx context.Context, conn pg.Querier) error {
			names, err = thirdPartyNamesByID(ctx, conn, linkedIDs)
			return err
		},
	); err != nil {
		return err
	}

	table := clicmdutil.NewTable("ID", "TYPE", "MATCH", "PATTERN", "CONF", "STATE", "THIRD PARTY")

	for _, p := range patterns {
		thirdParty := ""
		if p.CommonThirdPartyID != nil {
			thirdParty = names[*p.CommonThirdPartyID]
		}

		table.Row(
			p.ID.String(),
			string(p.TrackerType),
			string(p.MatchType),
			p.Pattern,
			fmt.Sprintf("%.2f", p.Confidence),
			enrichmentState(p),
			thirdParty,
		)
	}

	_, _ = fmt.Fprintln(out, table.Render())
	_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "Showing %d common tracker patterns.\n", len(patterns))

	return nil
}

// parseOrderBy maps the --sort/--order flags to a page.OrderBy. When
// --order is empty it defaults to descending for the time/score fields
// and ascending for pattern.
func parseOrderBy(sort, order string) (page.OrderBy[coredata.CommonTrackerPatternOrderField], error) {
	var (
		field       coredata.CommonTrackerPatternOrderField
		defaultDesc bool
		zeroOrderBy page.OrderBy[coredata.CommonTrackerPatternOrderField]
	)

	switch sort {
	case "pattern":
		field = coredata.CommonTrackerPatternOrderFieldPattern
	case "confidence":
		field, defaultDesc = coredata.CommonTrackerPatternOrderFieldConfidence, true
	case "created":
		field, defaultDesc = coredata.CommonTrackerPatternOrderFieldCreatedAt, true
	case "updated":
		field, defaultDesc = coredata.CommonTrackerPatternOrderFieldUpdatedAt, true
	case "enriched":
		field, defaultDesc = coredata.CommonTrackerPatternOrderFieldEnrichedAt, true
	default:
		return zeroOrderBy, fmt.Errorf("invalid --sort value %q: valid values are pattern, confidence, created, updated, enriched", sort)
	}

	direction := page.OrderDirectionAsc
	if defaultDesc {
		direction = page.OrderDirectionDesc
	}

	switch order {
	case "":
		// keep field default
	case "asc":
		direction = page.OrderDirectionAsc
	case "desc":
		direction = page.OrderDirectionDesc
	default:
		return zeroOrderBy, fmt.Errorf("invalid --order value %q: valid values are asc, desc", order)
	}

	return page.OrderBy[coredata.CommonTrackerPatternOrderField]{Field: field, Direction: direction}, nil
}

func buildListFilter(
	trackerType, matchType, keyword, state string,
	linked, unlinked bool,
) (*coredata.CommonTrackerPatternFilter, error) {
	filter := coredata.NewCommonTrackerPatternFilter()

	if trackerType != "" {
		tt := coredata.TrackerType(trackerType)
		if !tt.IsValid() {
			return nil, fmt.Errorf("invalid --tracker-type value %q", trackerType)
		}

		filter.WithTrackerType(&tt)
	}

	if matchType != "" {
		mt := coredata.TrackerPatternMatchType(matchType)
		if !mt.IsValid() {
			return nil, fmt.Errorf("invalid --match-type value %q", matchType)
		}

		filter.WithMatchType(&mt)
	}

	if keyword != "" {
		filter.WithKeyword(&keyword)
	}

	if state != "" {
		st, err := parseEnrichmentState(state)
		if err != nil {
			return nil, err
		}

		filter.WithState(&st)
	}

	switch {
	case linked:
		v := true
		filter.WithLinked(&v)
	case unlinked:
		v := false
		filter.WithLinked(&v)
	}

	return filter, nil
}

func parseEnrichmentState(value string) (coredata.CommonTrackerPatternEnrichmentState, error) {
	switch value {
	case "queued":
		return coredata.CommonTrackerPatternEnrichmentStateQueued, nil
	case "enriched":
		return coredata.CommonTrackerPatternEnrichmentStateEnriched, nil
	case "unenriched":
		return coredata.CommonTrackerPatternEnrichmentStateUnenriched, nil
	default:
		return "", fmt.Errorf("invalid --state value %q: valid values are queued, enriched, unenriched", value)
	}
}
