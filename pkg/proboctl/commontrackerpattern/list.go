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
		flagTrackerType          string
		flagMatchType            string
		flagCommonThirdParty     string
		flagLinkedBanner         string
		flagLinkedOrg            string
		flagKeyword              string
		flagState                string
		flagAttribution          string
		flagWithCommonThirdParty bool
		flagWithoutDescription   bool
		flagSort                 string
		flagOrder                string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List common tracker patterns with filters and sorting",
		Args:  cobra.NoArgs,
	}

	output := clicmdutil.AddOutputFlag(cmd)

	cmd.Flags().StringVar(&flagTrackerType, "tracker-type", "", "Filter by tracker type (COOKIE, LOCAL_STORAGE, SESSION_STORAGE, INDEXED_DB)")
	cmd.Flags().StringVar(&flagMatchType, "match-type", "", "Filter by match type (EXACT, GLOB, PREFIX)")
	cmd.Flags().StringVar(&flagCommonThirdParty, "common-third-party", "", "Filter by linked common third party (slug or GID)")
	cmd.Flags().StringVar(&flagLinkedBanner, "linked-banner", "", "Filter to catalog rows linked to a cookie banner's patterns (GID)")
	cmd.Flags().StringVar(&flagLinkedOrg, "linked-org", "", "Filter to catalog rows linked to an organization's patterns (GID)")
	cmd.Flags().StringVar(&flagKeyword, "keyword", "", "Filter by pattern/description substring")
	cmd.Flags().StringVar(&flagState, "state", "", "Filter by enrichment state (queued, enriched, unenriched)")
	cmd.Flags().StringVar(&flagAttribution, "attribution", "", "Filter by attribution verdict (UNDETERMINED, THIRD_PARTY, FIRST_PARTY)")
	cmd.Flags().BoolVar(&flagWithCommonThirdParty, "with-common-third-party", false, "Filter by whether the pattern is linked to a common third party (true/false); ignored when not set")
	cmd.Flags().BoolVar(&flagWithoutDescription, "without-description", false, "Only patterns with a blank description")
	cmd.Flags().StringVar(&flagSort, "sort", "confidence", "Sort field: pattern, confidence, created, updated, attempted")
	cmd.Flags().StringVar(&flagOrder, "order", "", "Sort order: asc, desc (default depends on field)")

	pageFlags := cmdutil.AddPageFlags(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := clicmdutil.ValidateOutputFlag(output); err != nil {
			return err
		}

		if flagLinkedBanner != "" && flagLinkedOrg != "" {
			return fmt.Errorf("--linked-banner and --linked-org are mutually exclusive")
		}

		orderBy, err := parseOrderBy(flagSort, flagOrder)
		if err != nil {
			return err
		}

		cursor, err := cmdutil.NewCursorFromFlags(pageFlags, orderBy)
		if err != nil {
			return err
		}

		var withCommonThirdParty *bool
		if cmd.Flags().Changed("with-common-third-party") {
			withCommonThirdParty = &flagWithCommonThirdParty
		}

		var described *bool
		if flagWithoutDescription {
			described = new(false)
		}

		filter, err := buildListFilter(flagTrackerType, flagMatchType, flagKeyword, flagState, flagAttribution, withCommonThirdParty, described)
		if err != nil {
			return err
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		ctx := cmd.Context()

		var (
			patterns coredata.CommonTrackerPatterns
			pageInfo cmdutil.PageInfo
		)

		if err := pgClient.WithConn(
			ctx,
			func(ctx context.Context, conn pg.Querier) error {
				if flagCommonThirdParty != "" {
					id, err := resolveCommonThirdPartyID(ctx, conn, flagCommonThirdParty)
					if err != nil {
						return err
					}

					filter.WithCommonThirdPartyID(&id)
				}

				switch {
				case flagLinkedBanner != "":
					bannerID, err := gid.ParseGID(flagLinkedBanner)
					if err != nil {
						return fmt.Errorf("invalid --linked-banner GID %q: %w", flagLinkedBanner, err)
					}

					var tps coredata.TrackerPatterns

					linkedIDs, err := tps.LoadAllLinkedCommonTrackerPatternIDsByCookieBannerID(ctx, conn, coredata.NewScopeFromObjectID(bannerID), bannerID)
					if err != nil {
						return err
					}

					if len(linkedIDs) == 0 {
						return nil
					}

					filter.WithIDs(linkedIDs)
				case flagLinkedOrg != "":
					orgID, err := gid.ParseGID(flagLinkedOrg)
					if err != nil {
						return fmt.Errorf("invalid --linked-org GID %q: %w", flagLinkedOrg, err)
					}

					var tps coredata.TrackerPatterns

					linkedIDs, err := tps.LoadAllLinkedCommonTrackerPatternIDsByOrganizationID(ctx, conn, coredata.NewScopeFromObjectID(orgID), orgID)
					if err != nil {
						return err
					}

					if len(linkedIDs) == 0 {
						return nil
					}

					filter.WithIDs(linkedIDs)
				}

				p, err := cmdutil.FetchPage(
					ctx,
					cursor,
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

				patterns = p.Data
				pageInfo = cmdutil.NewPageInfo(p)

				return nil
			},
		); err != nil {
			return err
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, cmdutil.PageOutput{Items: patterns, PageInfo: pageInfo})
		}

		return renderPatternTable(cmd, f, patterns, pageInfo)
	}

	return cmd
}

func renderPatternTable(cmd *cobra.Command, f *cmdutil.Factory, patterns coredata.CommonTrackerPatterns, pageInfo cmdutil.PageInfo) error {
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

	table := clicmdutil.NewTable("ID", "TYPE", "MATCH", "PATTERN", "CONF", "VERDICT", "STATE", "THIRD PARTY", "LAST ATTEMPT", "CREATED", "UPDATED")

	for _, p := range patterns {
		thirdParty := ""
		if p.CommonThirdPartyID != nil {
			thirdParty = names[*p.CommonThirdPartyID]
		}

		lastAttempt := ""
		if p.LastEnrichmentAttemptAt != nil {
			lastAttempt = p.LastEnrichmentAttemptAt.Format("2006-01-02 15:04:05")
		}

		table.Row(
			p.ID.String(),
			string(p.TrackerType),
			string(p.MatchType),
			p.Pattern,
			fmt.Sprintf("%.2f", p.Confidence),
			string(p.Attribution),
			enrichmentState(p),
			thirdParty,
			lastAttempt,
			p.CreatedAt.Format("2006-01-02 15:04:05"),
			p.UpdatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	_, _ = fmt.Fprintln(out, table.Render())
	cmdutil.PrintPageInfo(out, pageInfo)

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
	case "attempted":
		field, defaultDesc = coredata.CommonTrackerPatternOrderFieldLastEnrichmentAttemptAt, true
	default:
		return zeroOrderBy, fmt.Errorf("invalid --sort value %q: valid values are pattern, confidence, created, updated, attempted", sort)
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
	trackerType, matchType, keyword, state, attribution string,
	withCommonThirdParty, described *bool,
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

	if attribution != "" {
		attr := coredata.CommonTrackerPatternAttribution(attribution)
		if !attr.IsValid() {
			return nil, fmt.Errorf("invalid --attribution value %q: valid values are UNDETERMINED, THIRD_PARTY, FIRST_PARTY", attribution)
		}

		filter.WithAttribution(&attr)
	}

	if withCommonThirdParty != nil {
		filter.WithLinked(withCommonThirdParty)
	}

	if described != nil {
		filter.WithDescribed(described)
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
