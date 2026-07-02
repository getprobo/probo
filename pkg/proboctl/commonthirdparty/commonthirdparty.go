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
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

// NewCmdCommonThirdParty is the entry point for inspecting and
// re-enriching the global common third party catalog.
func NewCmdCommonThirdParty(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "common-third-party <command>",
		Aliases: []string{"ctp3"},
		Short:   "Inspect and re-enrich the global common third party catalog",
	}

	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdShow(f))
	cmd.AddCommand(newCmdDomains(f))
	cmd.AddCommand(newCmdUpsert(f))
	cmd.AddCommand(newCmdReenrich(f))
	cmd.AddCommand(newCmdStats(f))

	return cmd
}

// enrichmentState classifies a common third party's position in the
// enrichment lifecycle for display. A row that has been through the
// workflow (it carries an enrichment payload) reads "enriched" only when
// every field the last run recorded an outcome for resolved a value;
// otherwise it reads "partial (X/Y)".
func enrichmentState(p *coredata.CommonThirdParty) string {
	switch {
	case p.EnrichmentRequestedAt != nil:
		return "queued"
	case len(p.Enrichment) > 0:
		resolved, total := enrichmentCompleteness(p)
		if total == 0 || resolved == total {
			return "enriched"
		}

		return fmt.Sprintf("partial (%d/%d)", resolved, total)
	default:
		return "unenriched"
	}
}

// resolvedFieldStatuses are the per-field enrichment statuses that carry a
// value, as opposed to not_found / low_confidence.
var resolvedFieldStatuses = map[string]struct{}{
	"found":                 {},
	"exists_external":       {},
	"fallback_display_name": {},
}

// enrichmentCompleteness counts how many of the fields the last enrichment
// run recorded an outcome for resolved a value (X) versus the total it
// recorded (Y), parsed from the enrichment payload's per-field provenance.
func enrichmentCompleteness(p *coredata.CommonThirdParty) (resolved, total int) {
	if len(p.Enrichment) == 0 {
		return 0, 0
	}

	var meta struct {
		Fields map[string]struct {
			Status string `json:"status"`
		} `json:"fields"`
	}

	if err := json.Unmarshal(p.Enrichment, &meta); err != nil {
		return 0, 0
	}

	for _, f := range meta.Fields {
		total++

		if _, ok := resolvedFieldStatuses[f.Status]; ok {
			resolved++
		}
	}

	return resolved, total
}

// enrichmentStatus returns the run-level status recorded in the
// enrichment payload (done, partial, failed), or an empty string when
// the row has never been enriched or the payload is malformed.
func enrichmentStatus(p *coredata.CommonThirdParty) string {
	if len(p.Enrichment) == 0 {
		return ""
	}

	var meta struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal(p.Enrichment, &meta); err != nil {
		return ""
	}

	return meta.Status
}

// parseEnrichmentState maps the --state flag to a coredata enrichment
// state.
func parseEnrichmentState(value string) (coredata.CommonThirdPartyEnrichmentState, error) {
	switch value {
	case "queued":
		return coredata.CommonThirdPartyEnrichmentStateQueued, nil
	case "enriched":
		return coredata.CommonThirdPartyEnrichmentStateEnriched, nil
	case "unenriched":
		return coredata.CommonThirdPartyEnrichmentStateUnenriched, nil
	default:
		return "", fmt.Errorf("invalid --state value %q: valid values are queued, enriched, unenriched", value)
	}
}

// validEnrichmentStatuses are the run-level statuses the enrichment
// worker records in the payload.
var validEnrichmentStatuses = map[string]struct{}{
	"done":    {},
	"partial": {},
	"failed":  {},
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

// resolveCommonThirdParty loads a common third party by GID or slug.
func resolveCommonThirdParty(ctx context.Context, conn pg.Querier, value string) (coredata.CommonThirdParty, error) {
	var party coredata.CommonThirdParty

	if id, err := gid.ParseGID(value); err == nil {
		if err := party.LoadByID(ctx, conn, id); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return party, fmt.Errorf("no common third party found for %q", value)
			}

			return party, fmt.Errorf("cannot load common third party: %w", err)
		}

		return party, nil
	}

	if err := party.LoadBySlug(ctx, conn, value); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return party, fmt.Errorf("no common third party found for %q (pass a slug or GID)", value)
		}

		return party, fmt.Errorf("cannot load common third party: %w", err)
	}

	return party, nil
}

// parseOrderBy maps the --sort/--order flags to a page.OrderBy. Name
// defaults to ascending; the time fields default to descending.
func parseOrderBy(sort, order string) (page.OrderBy[coredata.CommonThirdPartyOrderField], error) {
	var (
		field       coredata.CommonThirdPartyOrderField
		defaultDesc bool
		zero        page.OrderBy[coredata.CommonThirdPartyOrderField]
	)

	switch sort {
	case "name":
		field = coredata.CommonThirdPartyOrderFieldName
	case "created":
		field, defaultDesc = coredata.CommonThirdPartyOrderFieldCreatedAt, true
	case "updated":
		field, defaultDesc = coredata.CommonThirdPartyOrderFieldUpdatedAt, true
	default:
		return zero, fmt.Errorf("invalid --sort value %q: valid values are name, created, updated", sort)
	}

	direction := page.OrderDirectionAsc
	if defaultDesc {
		direction = page.OrderDirectionDesc
	}

	switch order {
	case "":
	case "asc":
		direction = page.OrderDirectionAsc
	case "desc":
		direction = page.OrderDirectionDesc
	default:
		return zero, fmt.Errorf("invalid --order value %q: valid values are asc, desc", order)
	}

	return page.OrderBy[coredata.CommonThirdPartyOrderField]{Field: field, Direction: direction}, nil
}
