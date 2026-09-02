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

package listgvlcatalog

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($first: Int, $after: CursorKey, $filter: CommonGVLVendorFilter) {
  commonGVLVendors(first: $first, after: $after, filter: $filter) {
    totalCount
    edges {
      node {
        id
        iabVendorId
        name
        policyUrl
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`

type vendor struct {
	ID          string  `json:"id"`
	IabVendorID int     `json:"iabVendorId"`
	Name        string  `json:"name"`
	PolicyURL   *string `json:"policyUrl"`
}

func NewCmdListGVLCatalog(f *cmdutil.Factory) *cobra.Command {
	var (
		flagQuery          string
		flagCookieBannerID string
		flagMembership     string
		flagLimit          int
		flagOutput         *string
	)

	cmd := &cobra.Command{
		Use:   "list-gvl-catalog",
		Short: "List IAB GVL vendors from the global catalog",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
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

			filter := map[string]any{}
			if flagQuery != "" {
				filter["query"] = flagQuery
			}
			if flagMembership != "" {
				if flagCookieBannerID == "" {
					return fmt.Errorf("--cookie-banner-id is required with --membership")
				}
				filter["cookieBannerId"] = flagCookieBannerID
				filter["membership"] = flagMembership
			}

			variables := map[string]any{}
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			vendors, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[vendor], error) {
					var resp struct {
						CommonGVLVendors api.Connection[vendor] `json:"commonGVLVendors"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					return &resp.CommonGVLVendors, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, vendors)
			}

			if len(vendors) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No GVL vendors found. Run gvl-import to populate the catalog.")
				return nil
			}

			rows := make([][]string, 0, len(vendors))
			for _, v := range vendors {
				policy := ""
				if v.PolicyURL != nil {
					policy = *v.PolicyURL
				}

				rows = append(rows, []string{fmt.Sprintf("%d", v.IabVendorID), v.Name, policy})
			}

			t := cmdutil.NewTable("IAB ID", "NAME", "POLICY URL").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(vendors) {
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "\nShowing %d of %d GVL vendors\n", len(vendors), totalCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagQuery, "query", "", "Search by name or IAB vendor ID")
	cmd.Flags().StringVar(&flagCookieBannerID, "cookie-banner-id", "", "Cookie banner ID (required with --membership)")
	cmd.Flags().StringVar(&flagMembership, "membership", "", "Filter by banner membership: ON_BANNER or NOT_ON_BANNER")
	cmd.Flags().IntVar(&flagLimit, "limit", 50, "Maximum number of vendors to return")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
