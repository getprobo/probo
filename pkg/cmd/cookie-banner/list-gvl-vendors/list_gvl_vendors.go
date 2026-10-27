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

package listgvlvendors

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    __typename
    ... on CookieBanner {
      gvlVendors(first: $first, after: $after) {
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
  }
}
`

type vendor struct {
	ID          string  `json:"id"`
	IabVendorID int     `json:"iabVendorId"`
	Name        string  `json:"name"`
	PolicyURL   *string `json:"policyUrl"`
}

func NewCmdListGVLVendors(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit  int
		flagAfter  string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "list-gvl-vendors <cookie-banner-id>",
		Short: "List IAB GVL vendors selected on a cookie banner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{"id": args[0]}

			vendors, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[vendor]("gvlVendors", "CookieBanner", fmt.Sprintf("cookie banner %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				vendors,
				totalCount,
				cmdutil.ListView[vendor]{
					Empty:      "No GVL vendors selected on this cookie banner.",
					Noun:       "GVL vendors",
					NextCursor: nextCursor,
					Headers:    []string{"IAB ID", "NAME", "POLICY URL"},
					Row: func(v vendor) ([]string, error) {
						policy := ""
						if v.PolicyURL != nil {
							policy = *v.PolicyURL
						}

						return []string{fmt.Sprintf("%d", v.IabVendorID), v.Name, policy}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVar(&flagLimit, "limit", 50, "Maximum number of vendors to return")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
