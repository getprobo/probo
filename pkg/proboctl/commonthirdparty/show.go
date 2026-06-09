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

package commonthirdparty

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdShow(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <gid|slug>",
		Short: "Show a single common third party with its domains and linked pattern count",
		Args:  cobra.ExactArgs(1),
	}

	output := clicmdutil.AddOutputFlag(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := clicmdutil.ValidateOutputFlag(output); err != nil {
			return err
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		var (
			party        coredata.CommonThirdParty
			domains      coredata.CommonThirdPartyDomains
			patternCount int
		)

		if err := pgClient.WithConn(
			cmd.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				party, err = resolveCommonThirdParty(ctx, conn, args[0])
				if err != nil {
					return err
				}

				if err := domains.LoadByCommonThirdPartyID(ctx, conn, party.ID); err != nil {
					return fmt.Errorf("cannot load domains: %w", err)
				}

				var patterns coredata.CommonTrackerPatterns
				if err := patterns.LoadByCommonThirdPartyID(ctx, conn, party.ID); err != nil {
					return fmt.Errorf("cannot load linked patterns: %w", err)
				}

				patternCount = len(patterns)

				return nil
			},
		); err != nil {
			return err
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, map[string]any{
				"thirdParty":         party,
				"domains":            domains,
				"linkedPatternCount": patternCount,
			})
		}

		out := f.IOStreams.Out
		label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(20)
		row := func(name, value string) {
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render(name), value)
		}

		row("ID:", party.ID.String())
		row("Name:", party.Name)
		row("Slug:", party.Slug)
		row("Category:", string(party.Category))

		if party.WebsiteURL != nil {
			row("Website:", *party.WebsiteURL)
		}

		domainNames := make([]string, 0, len(domains))
		for _, d := range domains {
			domainNames = append(domainNames, d.Domain)
		}

		if len(domainNames) > 0 {
			row("Domains:", strings.Join(domainNames, ", "))
		} else {
			row("Domains:", "(none)")
		}

		row("Linked patterns:", fmt.Sprintf("%d", patternCount))
		row("Created:", party.CreatedAt.Format("2006-01-02 15:04:05"))
		row("Updated:", party.UpdatedAt.Format("2006-01-02 15:04:05"))

		return nil
	}

	return cmd
}
