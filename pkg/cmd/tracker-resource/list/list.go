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
    ... on CookieCategory {
      trackerResources(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            type
            origin
            path
            displayName
            excluded
            lastDetectedAt
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

type trackerResource struct {
	ID             string  `json:"id"`
	Type           string  `json:"type"`
	Origin         string  `json:"origin"`
	Path           string  `json:"path"`
	DisplayName    string  `json:"displayName"`
	Excluded       bool    `json:"excluded"`
	LastDetectedAt *string `json:"lastDetectedAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagCategoryID string
		flagLimit      int
		flagOutput     *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tracker resources in a category",
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

			variables := map[string]any{"id": flagCategoryID}

			resources, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[trackerResource], error) {
					var resp struct {
						Node *struct {
							Typename         string                          `json:"__typename"`
							TrackerResources api.Connection[trackerResource] `json:"trackerResources"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("cookie category %s not found", flagCategoryID)
					}

					if resp.Node.Typename != "CookieCategory" {
						return nil, fmt.Errorf("expected CookieCategory node, got %s", resp.Node.Typename)
					}

					return &resp.Node.TrackerResources, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resources)
			}

			if len(resources) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No tracker resources found.")
				return nil
			}

			rows := make([][]string, 0, len(resources))
			for _, r := range resources {
				excluded := ""
				if r.Excluded {
					excluded = "yes"
				}

				lastDetected := ""
				if r.LastDetectedAt != nil {
					lastDetected = cmdutil.FormatTime(*r.LastDetectedAt)
				}

				rows = append(rows, []string{r.ID, r.Type, r.Origin, r.Path, r.DisplayName, excluded, lastDetected})
			}

			t := cmdutil.NewTable("ID", "TYPE", "ORIGIN", "PATH", "DISPLAY NAME", "EXCLUDED", "LAST DETECTED").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(resources) {
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "\nShowing %d of %d tracker resources\n", len(resources), totalCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagCategoryID, "category-id", "", "Cookie category ID (required)")
	_ = cmd.MarkFlagRequired("category-id")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
