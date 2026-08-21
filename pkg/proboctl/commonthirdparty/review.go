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
	"fmt"
	"os/user"
	"strings"

	"github.com/spf13/cobra"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdReview(f *cmdutil.Factory) *cobra.Command {
	var (
		flagVerdict string
		flagBy      string
		flagDryRun  bool
		flagYes     bool
	)

	cmd := &cobra.Command{
		Use:   "review <slug-or-gid> <validated|rejected|unreviewed>",
		Short: "Record a human verdict on what a catalog entry names",
		Long: "Record whether a catalog entry names an entity an organization could " +
			"engage.\n\n" +
			"The test is entity nature, not data flow: a payroll processor or an " +
			"auditor is a valid register entry with no browser presence at all, so " +
			"'does it egress' is the wrong question here. Ask instead whether there " +
			"is a company on the other side to sign an agreement with.\n\n" +
			"A rejected entry is kept rather than deleted. Deleting it only means the " +
			"next scan recreates it and someone adjudicates it again, whereas a " +
			"retained rejection tells the mapping pipeline that patterns matching " +
			"this name earn a terminal attribution instead of a vendor link, and " +
			"stops the entry being offered by the catalog import and the " +
			"identification agent.\n\n" +
			"Rejecting therefore requires --verdict, because which terminal " +
			"attribution applies is a property of what the entry names: a bundled " +
			"library egresses nothing and is first-party, while software the visitor " +
			"installed is not-attributable and calling it the operator's own would " +
			"state something false in a register.",
		Args: cobra.ExactArgs(2),
	}

	cmd.Flags().StringVar(&flagVerdict, "verdict", "", "Terminal attribution for a rejection: first-party or not-attributable")
	cmd.Flags().StringVar(&flagBy, "by", "", "Who reviewed it (defaults to the current OS user)")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the change without writing")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		review, err := parseReview(args[1])
		if err != nil {
			return err
		}

		verdict, err := parseRejectedVerdict(review, flagVerdict)
		if err != nil {
			return err
		}

		if !flagDryRun && !flagYes {
			return fmt.Errorf(
				"about to record %s on %q, which the mapping pipeline and the "+
					"catalog import both act on globally; pass --yes to proceed",
				review,
				args[0],
			)
		}

		reviewedBy := strings.TrimSpace(flagBy)
		if reviewedBy == "" {
			if u, err := user.Current(); err == nil {
				reviewedBy = u.Username
			}
		}

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		out := f.IOStreams.Out

		var party coredata.CommonThirdParty

		if err := pgClient.WithTx(
			ctx,
			func(ctx context.Context, tx pg.Tx) error {
				party, err = resolveCommonThirdParty(ctx, tx, args[0])
				if err != nil {
					return err
				}

				if flagDryRun {
					return nil
				}

				return (coredata.CommonThirdParty{}).UpdateReview(
					ctx,
					tx,
					party.ID,
					review,
					verdict,
					reviewedBy,
				)
			},
		); err != nil {
			return fmt.Errorf("cannot review common third party: %w", err)
		}

		verdictSuffix := ""
		if verdict != nil {
			verdictSuffix = fmt.Sprintf(" (%s)", *verdict)
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(
				out,
				"Would mark %q (%s) %s%s.\n",
				party.Name, party.Slug, review, verdictSuffix,
			)

			return nil
		}

		_, _ = fmt.Fprintf(
			out,
			"Marked %q (%s) %s%s.\n",
			party.Name, party.Slug, review, verdictSuffix,
		)

		return nil
	}

	return cmd
}

// parseReview accepts the review states in the lower-case spelling the rest
// of the CLI uses for enums.
func parseReview(raw string) (coredata.CommonThirdPartyReview, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "unreviewed":
		return coredata.CommonThirdPartyReviewUnreviewed, nil
	case "validated":
		return coredata.CommonThirdPartyReviewValidated, nil
	case "rejected":
		return coredata.CommonThirdPartyReviewRejected, nil
	}

	return "", fmt.Errorf(
		"invalid review state %q: want validated, rejected or unreviewed",
		raw,
	)
}

// parseRejectedVerdict pairs the verdict with the review state. A rejection
// without one would leave the mapping pipeline with nothing to apply, and a
// verdict on any other state would claim a terminal attribution for a row
// that still resolves to a vendor.
func parseRejectedVerdict(
	review coredata.CommonThirdPartyReview,
	raw string,
) (*coredata.CommonTrackerPatternAttribution, error) {
	trimmed := strings.ToLower(strings.TrimSpace(raw))

	if review != coredata.CommonThirdPartyReviewRejected {
		if trimmed != "" {
			return nil, fmt.Errorf("--verdict only applies when rejecting an entry")
		}

		return nil, nil
	}

	switch trimmed {
	case "first-party", "first_party":
		verdict := coredata.CommonTrackerPatternAttributionFirstParty
		return &verdict, nil
	case "not-attributable", "not_attributable":
		verdict := coredata.CommonTrackerPatternAttributionNotAttributable
		return &verdict, nil
	case "":
		return nil, fmt.Errorf(
			"--verdict is required when rejecting: first-party for a component that " +
				"egresses nothing, not-attributable for software the visitor installed",
		)
	}

	return nil, fmt.Errorf(
		"invalid --verdict %q: want first-party or not-attributable",
		raw,
	)
}
