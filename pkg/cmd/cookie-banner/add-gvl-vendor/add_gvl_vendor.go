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

package addgvlvendor

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const addMutation = `
mutation($input: AddCookieBannerGVLVendorInput!) {
  addCookieBannerGVLVendor(input: $input) {
    commonGVLVendor {
      iabVendorId
      name
    }
    cookieBanner {
      id
    }
  }
}
`

type addResponse struct {
	AddCookieBannerGVLVendor struct {
		CommonGVLVendor struct {
			IabVendorID int    `json:"iabVendorId"`
			Name        string `json:"name"`
		} `json:"commonGVLVendor"`
		CookieBanner struct {
			ID string `json:"id"`
		} `json:"cookieBanner"`
	} `json:"addCookieBannerGVLVendor"`
}

func NewCmdAddGVLVendor(f *cmdutil.Factory) *cobra.Command {
	var flagIABVendorID int

	cmd := &cobra.Command{
		Use:   "add-gvl-vendor <cookie-banner-id>",
		Short: "Add an IAB GVL vendor to a cookie banner",
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

			data, err := client.Do(addMutation, map[string]any{
				"input": map[string]any{
					"cookieBannerId": args[0],
					"iabVendorId":    flagIABVendorID,
				},
			})
			if err != nil {
				return err
			}

			var resp addResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			v := resp.AddCookieBannerGVLVendor.CommonGVLVendor
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Added GVL vendor %d (%s) to cookie banner %s\n", v.IabVendorID, v.Name, args[0])

			return nil
		},
	}

	cmd.Flags().IntVar(&flagIABVendorID, "iab-vendor-id", 0, "IAB GVL vendor ID (required)")

	return cmd
}
