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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

// manualEnrichmentPayload builds the enrichment provenance written when an
// operator sets a description by hand. It records only the description
// outcome (resolved), so the row reads "enriched" and the enrichment
// worker leaves it alone, while marking the source as manual for audit.
func manualEnrichmentPayload() json.RawMessage {
	now := time.Now()

	payload := struct {
		Status      string    `json:"status"`
		Source      string    `json:"source"`
		AttemptedAt time.Time `json:"attempted_at"`
		Fields      map[string]struct {
			Status    string    `json:"status"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"fields"`
	}{
		Status:      "manual",
		Source:      "manual",
		AttemptedAt: now,
		Fields: map[string]struct {
			Status    string    `json:"status"`
			UpdatedAt time.Time `json:"updated_at"`
		}{
			"description": {Status: "found", UpdatedAt: now},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil
	}

	return raw
}

func newCmdSetDescription(f *cmdutil.Factory) *cobra.Command {
	var (
		flagDescription string
		flagYes         bool
	)

	cmd := &cobra.Command{
		Use:   "set-description <gid>",
		Short: "Set a common tracker pattern's description and backfill org patterns",
		Long: "Write a description on a common tracker pattern and mark it enriched so " +
			"the enrichment worker leaves it alone, then backfill the description onto " +
			"every linked org tracker pattern that does not already have one.",
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&flagDescription, "description", "", "Description to set (required)")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")

	_ = cmd.MarkFlagRequired("description")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		if flagDescription == "" {
			return fmt.Errorf("--description must not be empty")
		}

		id, err := gid.ParseGID(args[0])
		if err != nil {
			return fmt.Errorf("invalid GID %q: %w", args[0], err)
		}

		if !flagYes {
			return fmt.Errorf("about to set the description on %s; pass --yes to proceed", id.String())
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		out := f.IOStreams.Out

		var backfilled int64

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				var pattern coredata.CommonTrackerPattern
				if err := pattern.LoadByID(ctx, tx, id); err != nil {
					if errors.Is(err, coredata.ErrResourceNotFound) {
						return fmt.Errorf("no common tracker pattern found for %q", args[0])
					}

					return fmt.Errorf("cannot load common tracker pattern: %w", err)
				}

				if err := pattern.UpdateEnrichment(ctx, tx, flagDescription, nil, manualEnrichmentPayload()); err != nil {
					return fmt.Errorf("cannot set common tracker pattern description: %w", err)
				}

				var tps coredata.TrackerPatterns

				backfilled, err = tps.BackfillDescriptionByCommonTrackerPatternID(ctx, tx, id, flagDescription)
				if err != nil {
					return err
				}

				return nil
			},
		); err != nil {
			return err
		}

		_, _ = fmt.Fprintf(out, "Set description on %s, backfilled %d org tracker pattern(s).\n", id.String(), backfilled)

		return nil
	}

	return cmd
}
