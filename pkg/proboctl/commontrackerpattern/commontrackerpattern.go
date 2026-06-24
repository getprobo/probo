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
	"encoding/json"
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
	cmd.AddCommand(newCmdUpsert(f))
	cmd.AddCommand(newCmdReenrich(f))
	cmd.AddCommand(newCmdStats(f))
	cmd.AddCommand(newCmdLink(f))
	cmd.AddCommand(newCmdUnlink(f))
	cmd.AddCommand(newCmdMarkFirstParty(f))
	cmd.AddCommand(newCmdSetDescription(f))

	return cmd
}

// enrichmentState classifies a pattern's position in the enrichment
// lifecycle for display. A row that has been through the workflow (it
// carries an enrichment payload) reads "enriched" only when every field
// the last run recorded an outcome for resolved a value; otherwise it
// reads "partial (X/Y)".
func enrichmentState(p *coredata.CommonTrackerPattern) string {
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
// value, as opposed to not_found.
var resolvedFieldStatuses = map[string]struct{}{
	"found":           {},
	"exists_external": {},
}

// enrichmentCompleteness counts how many of the fields the last enrichment
// run recorded an outcome for resolved a value (X) versus the total it
// recorded (Y), parsed from the enrichment payload's per-field provenance.
func enrichmentCompleteness(p *coredata.CommonTrackerPattern) (resolved, total int) {
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
