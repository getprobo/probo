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
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

// NewCmdCommonThirdParty is the entry point for inspecting the global
// common third party catalog.
func NewCmdCommonThirdParty(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "common-third-party <command>",
		Aliases: []string{"ctp3"},
		Short:   "Inspect the global common third party catalog",
	}

	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdShow(f))
	cmd.AddCommand(newCmdDomains(f))
	cmd.AddCommand(newCmdUpsert(f))

	return cmd
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
