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

package repair

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdThirdPartyRegisterNotes(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg             string
		flagDryRun          bool
		flagYes             bool
		flagContinueOnError bool
	)

	cmd := &cobra.Command{
		Use:   "third-party-register-notes",
		Short: "Rewrite raw Markdown notes in published third-party registers",
		Long: "Scan document_versions belonging to generated third-party registers " +
			"and convert risk-assessment Notes that were stored as raw Markdown " +
			"text into structured ProseMirror blocks. Clears cached PDFs " +
			"(file_id) so the PDF worker regenerates exports. Re-running is " +
			"safe: already-converted versions are skipped. For a full " +
			"regeneration from live third-party data, use " +
			"`prb thirdParty publish --org ORG_ID --minor` instead.",
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Limit to a single organization ID")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Report versions that would be rewritten without writing")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")
	cmd.Flags().BoolVar(
		&flagContinueOnError,
		"continue-on-error",
		false,
		"Keep going when a version fails; exit non-zero if any failed",
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		out := f.IOStreams.Out

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		var orgID *gid.GID

		if flagOrg != "" {
			parsed, err := gid.ParseGID(flagOrg)
			if err != nil {
				return fmt.Errorf("invalid --org: %w", err)
			}

			orgID = &parsed
		}

		ids, err := listThirdPartyRegisterVersionIDs(ctx, pgClient, orgID)
		if err != nil {
			return err
		}

		if len(ids) == 0 {
			_, _ = fmt.Fprintln(out, "No third-party register document versions found.")
			return nil
		}

		_, _ = fmt.Fprintf(out, "Scanning %d third-party register document version(s).\n", len(ids))

		if flagDryRun {
			var wouldRewrite int

			for _, id := range ids {
				changed, err := previewThirdPartyRegisterNotesRewrite(ctx, pgClient, id)
				if err != nil {
					return err
				}

				if changed {
					wouldRewrite++
					_, _ = fmt.Fprintf(out, "would rewrite %s\n", id)
				}
			}

			_, _ = fmt.Fprintf(out, "Would rewrite %d of %d version(s).\n", wouldRewrite, len(ids))

			return nil
		}

		if !flagYes {
			return fmt.Errorf(
				"about to rewrite up to %d third-party register version(s); pass --yes to proceed or --dry-run to preview",
				len(ids),
			)
		}

		var (
			rewritten int
			failures  int
		)

		for _, id := range ids {
			changed, err := rewriteThirdPartyRegisterNotesVersion(ctx, pgClient, id)
			if err != nil {
				failures++
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "%v\n", err)

				if !flagContinueOnError {
					return fmt.Errorf("stopped after %d failure(s)", failures)
				}

				continue
			}

			if changed {
				rewritten++
				_, _ = fmt.Fprintf(out, "rewrote %s\n", id)
			}
		}

		_, _ = fmt.Fprintf(
			out,
			"Done. Rewrote %d of %d version(s); %d failure(s).\n",
			rewritten,
			len(ids),
			failures,
		)

		if failures > 0 {
			return fmt.Errorf("%d version(s) failed", failures)
		}

		return nil
	}

	return cmd
}

func listThirdPartyRegisterVersionIDs(
	ctx context.Context,
	pgClient *pg.Client,
	orgID *gid.GID,
) ([]gid.GID, error) {
	var ids []gid.GID

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			q := `
SELECT
	dv.id
FROM
	document_versions dv
INNER JOIN
	generated_documents gd ON gd.third_parties_document_id = dv.document_id
WHERE
	dv.content IS NOT NULL
	AND btrim(dv.content) <> ''
`

			args := pgx.NamedArgs{}

			if orgID != nil {
				q += `
	AND gd.organization_id = @organization_id
`
				args["organization_id"] = *orgID
			}

			q += `
ORDER BY
	dv.id
`

			rows, err := conn.Query(ctx, q, args)
			if err != nil {
				return fmt.Errorf("cannot list third-party register versions: %w", err)
			}
			defer rows.Close()

			for rows.Next() {
				var id gid.GID
				if err := rows.Scan(&id); err != nil {
					return fmt.Errorf("cannot scan document version id: %w", err)
				}

				ids = append(ids, id)
			}

			return rows.Err()
		},
	)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func previewThirdPartyRegisterNotesRewrite(
	ctx context.Context,
	pgClient *pg.Client,
	versionID gid.GID,
) (bool, error) {
	var changed bool

	err := pgClient.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			dv := &coredata.DocumentVersion{}
			if err := dv.LoadByID(ctx, conn, coredata.NewNoScope(), versionID); err != nil {
				return fmt.Errorf("cannot load document version %s: %w", versionID, err)
			}

			_, rewritten, err := probo.RewriteThirdPartyRegisterNotesContent(dv.Content)
			if err != nil {
				return fmt.Errorf("cannot rewrite document version %s: %w", versionID, err)
			}

			changed = rewritten

			return nil
		},
	)

	return changed, err
}

func rewriteThirdPartyRegisterNotesVersion(
	ctx context.Context,
	pgClient *pg.Client,
	versionID gid.GID,
) (bool, error) {
	var changed bool

	err := pgClient.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			dv := &coredata.DocumentVersion{}
			if err := dv.LoadByID(ctx, tx, coredata.NewNoScope(), versionID); err != nil {
				return fmt.Errorf("cannot load document version %s: %w", versionID, err)
			}

			rewritten, didRewrite, err := probo.RewriteThirdPartyRegisterNotesContent(dv.Content)
			if err != nil {
				return fmt.Errorf("cannot rewrite document version %s: %w", versionID, err)
			}

			if !didRewrite {
				return nil
			}

			dv.Content = rewritten
			dv.FileID = nil
			dv.PdfAttemptCount = 0
			dv.UpdatedAt = time.Now()

			if err := dv.Update(ctx, tx, coredata.NewNoScope()); err != nil {
				return fmt.Errorf("cannot update document version %s: %w", versionID, err)
			}

			changed = true

			return nil
		},
	)

	return changed, err
}
