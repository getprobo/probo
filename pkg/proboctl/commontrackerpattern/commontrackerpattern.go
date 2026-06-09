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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

// NewCmdCommonTrackerPattern is the entry point for inspecting and
// re-enriching the global common tracker pattern catalog.
func NewCmdCommonTrackerPattern(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "common-tracker-pattern <command>",
		Aliases: []string{"ctp"},
		Short:   "Inspect and re-enrich the global common tracker pattern catalog",
	}

	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdShow(f))
	cmd.AddCommand(newCmdReenrich(f))
	cmd.AddCommand(newCmdStats(f))
	cmd.AddCommand(newCmdLink(f))
	cmd.AddCommand(newCmdUnlink(f))
	cmd.AddCommand(newCmdSetDescription(f))

	return cmd
}

// enrichmentState classifies a pattern's position in the enrichment
// lifecycle for display.
func enrichmentState(p *coredata.CommonTrackerPattern) string {
	switch {
	case p.EnrichmentRequestedAt != nil:
		return "queued"
	case p.EnrichedAt != nil && p.Description == "":
		return "enriched (no description)"
	case p.EnrichedAt != nil:
		return "enriched"
	default:
		return "unenriched"
	}
}

// resolveCommonThirdPartyID accepts either a common third party GID or a
// slug and returns the corresponding id.
func resolveCommonThirdPartyID(ctx context.Context, conn pg.Querier, value string) (gid.GID, error) {
	if id, err := gid.ParseGID(value); err == nil {
		return id, nil
	}

	var party coredata.CommonThirdParty
	if err := party.LoadBySlug(ctx, conn, value); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return gid.GID{}, fmt.Errorf("no common third party found for %q (pass a slug or GID)", value)
		}

		return gid.GID{}, fmt.Errorf("cannot resolve common third party %q: %w", value, err)
	}

	return party.ID, nil
}

// thirdPartyNamesByID loads display names for the given common third
// party ids, skipping nil/empty inputs. It is used to render the linked
// vendor column without per-row queries.
func thirdPartyNamesByID(ctx context.Context, conn pg.Querier, ids []gid.GID) (map[gid.GID]string, error) {
	names := make(map[gid.GID]string)
	if len(ids) == 0 {
		return names, nil
	}

	var parties coredata.CommonThirdParties
	if err := parties.LoadByIDs(ctx, conn, ids); err != nil {
		return nil, fmt.Errorf("cannot load common third parties: %w", err)
	}

	for _, p := range parties {
		names[p.ID] = p.Name
	}

	return names, nil
}
