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

package list

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    __typename
    ... on Organization {
      cookieBanners(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            name
            origin
            state
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

type banner struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Origin string `json:"origin"`
	State  string `json:"state"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagLimit  int
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List cookie banners in an organization",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			variables := map[string]any{"id": flagOrg}

			banners, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[banner], error) {
					var resp struct {
						Node *struct {
							Typename      string                 `json:"__typename"`
							CookieBanners api.Connection[banner] `json:"cookieBanners"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}

					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}

					return &resp.Node.CookieBanners, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, banners)
			}

			if len(banners) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No cookie banners found.")
				return nil
			}

			rows := make([][]string, 0, len(banners))
			for _, b := range banners {
				rows = append(rows, []string{b.ID, b.Name, b.Origin, b.State})
			}

			t := cmdutil.NewTable("ID", "NAME", "ORIGIN", "STATE").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(banners) {
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "\nShowing %d of %d cookie banners\n", len(banners), totalCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
