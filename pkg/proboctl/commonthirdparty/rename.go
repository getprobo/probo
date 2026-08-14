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

package commonthirdparty

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
	"go.probo.inc/probo/pkg/slug"
)

func newCmdRename(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName     string
		flagReenrich bool
	)

	cmd := &cobra.Command{
		Use:   "rename <slug-or-gid> --name <new name>",
		Short: "Change a catalog entry's display name",
		Long: "Rename a catalog entry, leaving its slug alone.\n\n" +
			"Use this when two entries are genuinely different vendors and one is " +
			"misnamed — the case `merge` must not be used for.\n\n" +
			"The slug is the entry's identity: dedup resolves against it and the seed " +
			"upserts on it. Renaming deliberately does not touch it, so a correction " +
			"never silently moves an entry's identity. Use `set-slug` for that.",
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&flagName, "name", "", "The new display name")
	cmd.Flags().BoolVar(&flagReenrich, "reenrich", false, "Re-arm enrichment so the profile is re-resolved from the new name")

	_ = cmd.MarkFlagRequired("name")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		name := strings.TrimSpace(flagName)
		if name == "" {
			return fmt.Errorf("--name must not be empty")
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		var party coredata.CommonThirdParty

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				party, err = resolveCommonThirdParty(ctx, tx, args[0])
				if err != nil {
					return err
				}

				return coredata.CommonThirdParty{}.UpdateName(ctx, tx, party.ID, name)
			},
		); err != nil {
			return fmt.Errorf("cannot rename common third party: %w", err)
		}

		out := f.IOStreams.Out

		_, _ = fmt.Fprintf(out, "Renamed %q to %q (slug %s unchanged).\n", party.Name, name, party.Slug)

		// The enrichment worker resolved this row's profile from its old name,
		// so a correction leaves URLs, address and logo describing the wrong
		// company until the row is re-enriched.
		if flagReenrich {
			var requeued int64

			if err := pgClient.WithTx(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					var parties coredata.CommonThirdParties

					var err error
					requeued, err = parties.RequestEnrichmentByIDs(ctx, tx, []gid.GID{party.ID})

					return err
				},
			); err != nil {
				return fmt.Errorf("cannot enqueue enrichment: %w", err)
			}

			_, _ = fmt.Fprintf(out, "Queued %d common third party(ies) for the enrichment worker.\n", requeued)
		} else if len(party.Enrichment) > 0 {
			_, _ = fmt.Fprintf(
				f.IOStreams.ErrOut,
				"note: this row was enriched under its old name, so its profile may describe the wrong company; pass --reenrich to re-resolve it.\n",
			)
		}

		if expected := slug.Make(name); expected != party.Slug {
			_, _ = fmt.Fprintf(
				f.IOStreams.ErrOut,
				"note: the slug no longer matches the name (%s would derive %s); run set-slug if the identity should move too.\n",
				party.Slug,
				expected,
			)
		}

		return nil
	}

	return cmd
}

func newCmdSetSlug(f *cmdutil.Factory) *cobra.Command {
	var (
		flagSlug string
		flagYes  bool
	)

	cmd := &cobra.Command{
		Use:   "set-slug <slug-or-gid> --slug <new slug>",
		Short: "Change a catalog entry's slug",
		Long: "Change the slug of a catalog entry.\n\n" +
			"The slug is the entry's identity, not a label: vendor resolution dedups " +
			"against it, and the seed upserts on it. Changing it therefore changes " +
			"which future attributions land on this entry, and whether a seed run " +
			"recreates an entry under the old slug. It also breaks any external " +
			"reference to the old value, so it requires --yes.",
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().StringVar(&flagSlug, "slug", "", "The new slug")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")

	_ = cmd.MarkFlagRequired("slug")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		newSlug := slug.Make(flagSlug)
		if newSlug == "" {
			return fmt.Errorf("invalid --slug value %q: normalizes to empty", flagSlug)
		}

		if newSlug != flagSlug {
			_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "note: normalized slug to %q.\n", newSlug)
		}

		if !flagYes {
			return fmt.Errorf(
				"about to change the identity of %q to slug %q, which affects future vendor resolution and seeding; pass --yes to proceed",
				args[0],
				newSlug,
			)
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		var party coredata.CommonThirdParty

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				party, err = resolveCommonThirdParty(ctx, tx, args[0])
				if err != nil {
					return err
				}

				return coredata.CommonThirdParty{}.UpdateSlug(ctx, tx, party.ID, newSlug)
			},
		); err != nil {
			if errors.Is(err, coredata.ErrResourceAlreadyExists) {
				return fmt.Errorf("another common third party already uses the slug %q", newSlug)
			}

			return fmt.Errorf("cannot set common third party slug: %w", err)
		}

		_, _ = fmt.Fprintf(
			f.IOStreams.Out,
			"Changed the slug of %q from %s to %s.\n",
			party.Name,
			party.Slug,
			newSlug,
		)

		return nil
	}

	return cmd
}
