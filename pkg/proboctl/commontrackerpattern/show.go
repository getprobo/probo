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
	"errors"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdShow(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <gid>",
		Short: "Show a single common tracker pattern by GID",
		Args:  cobra.ExactArgs(1),
	}

	output := clicmdutil.AddOutputFlag(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := clicmdutil.ValidateOutputFlag(output); err != nil {
			return err
		}

		id, err := gid.ParseGID(args[0])
		if err != nil {
			return fmt.Errorf("invalid GID %q: %w", args[0], err)
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		var (
			pattern        coredata.CommonTrackerPattern
			thirdPartyName string
		)

		if err := pgClient.WithConn(
			cmd.Context(),
			func(ctx context.Context, conn pg.Querier) error {
				if err := pattern.LoadByID(ctx, conn, id); err != nil {
					if errors.Is(err, coredata.ErrResourceNotFound) {
						return fmt.Errorf("no common tracker pattern found for %q", args[0])
					}

					return fmt.Errorf("cannot load common tracker pattern: %w", err)
				}

				if pattern.CommonThirdPartyID != nil {
					var party coredata.CommonThirdParty
					if err := party.LoadByID(ctx, conn, *pattern.CommonThirdPartyID); err == nil {
						thirdPartyName = party.Name
					}
				}

				return nil
			},
		); err != nil {
			return err
		}

		if *output == clicmdutil.OutputJSON {
			return clicmdutil.PrintJSON(f.IOStreams.Out, pattern)
		}

		return renderPatternDetail(f, pattern, thirdPartyName)
	}

	return cmd
}

func renderPatternDetail(f *cmdutil.Factory, p coredata.CommonTrackerPattern, thirdPartyName string) error {
	out := f.IOStreams.Out
	label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(20)

	row := func(name, value string) {
		_, _ = fmt.Fprintf(out, "%s%s\n", label.Render(name), value)
	}

	row("ID:", p.ID.String())
	row("Tracker type:", string(p.TrackerType))
	row("Match type:", string(p.MatchType))
	row("Pattern:", p.Pattern)
	row("Confidence:", fmt.Sprintf("%.2f", p.Confidence))
	row("State:", enrichmentState(&p))

	if p.MaxAgeSeconds != nil {
		row("Max age (s):", fmt.Sprintf("%d", *p.MaxAgeSeconds))
	}

	if p.CommonThirdPartyID != nil {
		row("Third party:", fmt.Sprintf("%s (%s)", thirdPartyName, p.CommonThirdPartyID.String()))
	} else {
		row("Third party:", "(unlinked)")
	}

	description := p.Description
	if description == "" {
		description = "(none)"
	}

	row("Description:", description)

	if p.EnrichmentRequestedAt != nil {
		row("Enrichment queued:", p.EnrichmentRequestedAt.Format("2006-01-02 15:04:05"))
	}

	if p.EnrichedAt != nil {
		row("Enriched at:", p.EnrichedAt.Format("2006-01-02 15:04:05"))
	}

	row("Created:", p.CreatedAt.Format("2006-01-02 15:04:05"))
	row("Updated:", p.UpdatedAt.Format("2006-01-02 15:04:05"))

	return nil
}
