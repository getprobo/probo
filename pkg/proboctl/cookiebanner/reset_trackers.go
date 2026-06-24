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

package cookiebanner

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/proboctl/cmdutil"
)

func newCmdResetTrackers(f *cmdutil.Factory) *cobra.Command {
	var (
		flagMappingOnly bool
		flagKeyword     string
		flagDryRun      bool
		flagYes         bool
	)

	cmd := &cobra.Command{
		Use:   "reset-trackers <banner-gid>",
		Short: "Rebuild a banner's tracker patterns from detections and re-arm the analysis + mapping workers",
		Long: "Destructive, tenant-scoped operator action. For a banner's uncategorised, " +
			"non-excluded patterns it clears catalog/vendor links, rebuilds the raw exact " +
			"patterns from detected_trackers (decomposing derived globs), and re-arms the " +
			"pattern-analysis and mapping workers. User-categorised and excluded patterns are " +
			"preserved. With --mapping-only it skips the rebuild and only re-arms mapping. " +
			"With --keyword the rebuild and mapping reset are scoped to patterns whose pattern " +
			"or display name contains the substring.",
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().BoolVar(&flagMappingOnly, "mapping-only", false, "Only re-arm mapping (skip the detection rebuild and analysis)")
	cmd.Flags().StringVar(&flagKeyword, "keyword", "", "Only reset patterns whose pattern or display name contains this substring")
	cmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Print the target banner without writing")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Skip confirmation")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		bannerID, err := gid.ParseGID(args[0])
		if err != nil {
			return fmt.Errorf("invalid banner GID %q: %w", args[0], err)
		}

		ctx := cmd.Context()

		pgClient, err := f.PgClient()
		if err != nil {
			return err
		}

		scope := coredata.NewScopeFromObjectID(bannerID)

		out := f.IOStreams.Out

		mode := "full reset"
		if flagMappingOnly {
			mode = "mapping-only reset"
		}

		var keyword *string
		if flagKeyword != "" {
			keyword = &flagKeyword
			mode = fmt.Sprintf("%s scoped to keyword %q", mode, flagKeyword)
		}

		if flagDryRun {
			_, _ = fmt.Fprintf(out, "Would run %s on banner %s.\n", mode, bannerID.String())
			return nil
		}

		if !flagYes {
			return fmt.Errorf("about to run %s on banner %s; pass --yes to proceed or --dry-run to preview", mode, bannerID.String())
		}

		_, _ = fmt.Fprintf(out, "Running %s on banner %s.\n", mode, bannerID.String())

		result, err := cookiebanner.ResetBannerTrackers(ctx, pgClient, scope, bannerID, flagMappingOnly, keyword)
		if err != nil {
			return fmt.Errorf("cannot reset banner %s: %w", bannerID, err)
		}

		_, _ = fmt.Fprintf(
			out,
			"%s: reset %d pattern(s), decomposed %d glob(s) into %d exact(s), relinked %d detection(s), analysis_requested=%t\n",
			bannerID.String(),
			result.PatternsReset,
			result.GlobsDecomposed,
			result.ExactsCreated,
			result.DetectionsRelinked,
			result.AnalysisRequested,
		)

		return nil
	}

	return cmd
}
