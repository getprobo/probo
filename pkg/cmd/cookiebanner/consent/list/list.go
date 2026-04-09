// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
    ... on CookieBanner {
      consentRecords(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            visitorId
            action
            bannerVersion
            createdAt
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

type consentRecord struct {
	ID            string `json:"id"`
	VisitorID     string `json:"visitorId"`
	Action        string `json:"action"`
	BannerVersion int    `json:"bannerVersion"`
	CreatedAt     string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBanner string
		flagLimit  int
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List consent records for a banner",
		Aliases: []string{"ls"},
		Example: `  # List consent records for a cookie banner
  prb cookie-banner consent list --banner <banner-id>`,
		Args: cobra.NoArgs,
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
			)

			if flagBanner == "" {
				return fmt.Errorf("banner is required; pass --banner")
			}

			variables := map[string]any{
				"id": flagBanner,
			}

			records, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[consentRecord], error) {
					var resp struct {
						Node *struct {
							Typename       string                        `json:"__typename"`
							ConsentRecords api.Connection[consentRecord] `json:"consentRecords"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("cookie banner %s not found", flagBanner)
					}
					if resp.Node.Typename != "CookieBanner" {
						return nil, fmt.Errorf("expected CookieBanner node, got %s", resp.Node.Typename)
					}
					return &resp.Node.ConsentRecords, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, records)
			}

			if len(records) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No consent records found.")
				return nil
			}

			rows := make([][]string, 0, len(records))
			for _, r := range records {
				rows = append(rows, []string{
					r.ID,
					r.VisitorID,
					r.Action,
					fmt.Sprintf("%d", r.BannerVersion),
					cmdutil.FormatTime(r.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "VISITOR", "ACTION", "VERSION", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(records) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d consent records\n",
					len(records),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagBanner, "banner", "", "Cookie banner ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of records to list")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("banner")

	return cmd
}
