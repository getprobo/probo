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

package commonthirdparty

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName     string
		flagCategory string
		flagKeyword  string
		flagSort     string
		flagOrder    string
		flagLimit    int
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
	cmd.Flags().StringVar(&flagSort, "sort", "name", "Sort field: name, created, updated")
	cmd.Flags().StringVar(&flagOrder, "order", "", "Sort order: asc, desc (default depends on field)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 50, "Maximum rows to return (0 for all)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := clicmdutil.ValidateOutputFlag(output); err != nil {
			return err
		}

		orderBy, err := parseOrderBy(flagSort, flagOrder)
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

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		var parties coredata.CommonThirdParties

		if err := pgClient.WithConn(
			cmd.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				rows, err := cmdutil.Paginate(
					ctx,
					orderBy,
					flagLimit,
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

				parties = rows

				return nil
			},
		); err != nil {
			return err
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, parties)
		}

		if len(parties) == 0 {
			_, _ = fmt.Fprintln(f.IOStreams.Out, "No common third parties found.")
			return nil
		}

		table := clicmdutil.NewTable("ID", "NAME", "SLUG", "CATEGORY")
		for _, p := range parties {
			table.Row(p.ID.String(), p.Name, p.Slug, string(p.Category))
		}

		_, _ = fmt.Fprintln(f.IOStreams.Out, table.Render())
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
