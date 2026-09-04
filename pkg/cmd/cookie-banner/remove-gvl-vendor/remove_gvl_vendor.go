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

package removegvlvendor

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const removeMutation = `
mutation($input: RemoveCookieBannerGVLVendorInput!) {
  removeCookieBannerGVLVendor(input: $input) {
    cookieBanner {
      id
    }
  }
}
`

func NewCmdRemoveGVLVendor(f *cmdutil.Factory) *cobra.Command {
	var flagIABVendorID int

	cmd := &cobra.Command{
		Use:   "remove-gvl-vendor <cookie-banner-id>",
		Short: "Remove an IAB GVL vendor from a cookie banner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagIABVendorID < 1 {
				return fmt.Errorf("iab-vendor-id is required")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			if _, err := client.Do(removeMutation, map[string]any{
				"input": map[string]any{
					"cookieBannerId": args[0],
					"iabVendorId":    flagIABVendorID,
				},
			}); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Removed GVL vendor %d from cookie banner %s\n", flagIABVendorID, args[0])

			return nil
		},
	}

	cmd.Flags().IntVar(&flagIABVendorID, "iab-vendor-id", 0, "IAB GVL vendor ID (required)")

	return cmd
}
