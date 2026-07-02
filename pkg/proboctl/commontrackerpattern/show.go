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

package commontrackerpattern

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	clicmdutil "go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

// patternEnrichmentMetadataView mirrors the subset of the common tracker
// pattern enrichment payload (written by the enrichment worker) that show
// renders. It is decoded locally to avoid a dependency on the cookiebanner
// package.
type patternEnrichmentMetadataView struct {
	Model       string                                `json:"model"`
	AttemptedAt time.Time                             `json:"attempted_at"`
	Status      string                                `json:"status"`
	Error       string                                `json:"error"`
	Fields      map[string]patternEnrichmentFieldView `json:"fields"`
	Attribution *patternEnrichmentAttributionView     `json:"attribution"`
}

type patternEnrichmentFieldView struct {
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type patternEnrichmentAttributionView struct {
	ThirdPartyName string  `json:"third_party_name"`
	Category       string  `json:"category"`
	Confidence     float64 `json:"confidence"`
	Linked         bool    `json:"linked"`
}

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
					if err := party.LoadByID(ctx, conn, *pattern.CommonThirdPartyID); err != nil {
						if !errors.Is(err, coredata.ErrResourceNotFound) {
							return fmt.Errorf("cannot load common third party: %w", err)
						}
					} else {
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
	row("Verdict:", string(p.Attribution))
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

	row("Enrichment attempts:", fmt.Sprintf("%d", p.EnrichmentAttempts))

	if p.EnrichmentRequestedAt != nil {
		row("Enrichment queued:", p.EnrichmentRequestedAt.Format("2006-01-02 15:04:05"))
	}

	if p.LastEnrichmentAttemptAt != nil {
		row("Last attempt:", p.LastEnrichmentAttemptAt.Format("2006-01-02 15:04:05"))
	}

	row("Created:", p.CreatedAt.Format("2006-01-02 15:04:05"))
	row("Updated:", p.UpdatedAt.Format("2006-01-02 15:04:05"))

	printPatternEnrichmentDetails(out, label, p)

	return nil
}

// printPatternEnrichmentDetails renders the run-level status (done,
// partial, no_result), the agent attribution, and the per-field
// provenance recorded in the enrichment payload, when present.
func printPatternEnrichmentDetails(out io.Writer, label lipgloss.Style, p coredata.CommonTrackerPattern) {
	if len(p.Enrichment) == 0 {
		return
	}

	var meta patternEnrichmentMetadataView
	if err := json.Unmarshal(p.Enrichment, &meta); err != nil {
		return
	}

	row := func(name, value string) {
		_, _ = fmt.Fprintf(out, "%s%s\n", label.Render(name), value)
	}

	if meta.Status != "" {
		row("Last run status:", meta.Status)
	}

	if !meta.AttemptedAt.IsZero() {
		row("Last run recorded:", meta.AttemptedAt.Format("2006-01-02 15:04:05"))
	}

	if meta.Model != "" {
		row("Enrichment model:", meta.Model)
	}

	if meta.Error != "" {
		row("Last error:", meta.Error)
	}

	if meta.Attribution != nil {
		name := meta.Attribution.ThirdPartyName
		if name == "" {
			name = "(none)"
		}

		linked := "no"
		if meta.Attribution.Linked {
			linked = "yes"
		}

		row("Agent attribution:", fmt.Sprintf("%s [%s] conf %.2f linked=%s", name, meta.Attribution.Category, meta.Attribution.Confidence, linked))
	}

	if len(meta.Fields) > 0 {
		names := make([]string, 0, len(meta.Fields))
		for name := range meta.Fields {
			names = append(names, name)
		}

		sort.Strings(names)

		_, _ = fmt.Fprintln(out)

		table := clicmdutil.NewTable("FIELD", "STATUS", "UPDATED")

		for _, name := range names {
			fm := meta.Fields[name]

			updated := ""
			if !fm.UpdatedAt.IsZero() {
				updated = fm.UpdatedAt.Format("2006-01-02 15:04:05")
			}

			table.Row(name, fm.Status, updated)
		}

		_, _ = fmt.Fprintln(out, table.Render())
	}
}
