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
query($id: ID!, $first: Int, $after: CursorKey, $filter: CookieConsentRecordFilter) {
  node(id: $id) {
    __typename
    ... on CookieBanner {
      consentRecords(first: $first, after: $after, filter: $filter) {
        totalCount
        edges {
          node {
            id
            visitorId
            action
            sdkVersion
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
	ID         string `json:"id"`
	VisitorID  string `json:"visitorId"`
	Action     string `json:"action"`
	SDKVersion string `json:"sdkVersion"`
	CreatedAt  string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagBannerID  string
		flagAction    string
		flagVisitorID string
		flagVersion   int
		flagLimit     int
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List cookie consent records for a banner",
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

			variables := map[string]any{"id": flagBannerID}

			filter := map[string]any{}
			if cmd.Flags().Changed("action") {
				filter["action"] = flagAction
			}
			if cmd.Flags().Changed("visitor-id") {
				filter["visitorId"] = flagVisitorID
			}
			if cmd.Flags().Changed("version") {
				filter["version"] = flagVersion
			}
			if len(filter) > 0 {
				variables["filter"] = filter
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
						return nil, fmt.Errorf("cookie banner %s not found", flagBannerID)
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
				rows = append(rows, []string{r.ID, r.VisitorID, r.Action, r.SDKVersion, r.CreatedAt})
			}

			t := cmdutil.NewTable("ID", "VISITOR ID", "ACTION", "SDK VERSION", "CREATED AT").Rows(rows...)
			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(records) {
				_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "\nShowing %d of %d consent records\n", len(records), totalCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagBannerID, "banner-id", "", "Cookie banner ID (required)")
	_ = cmd.MarkFlagRequired("banner-id")
	cmd.Flags().StringVar(&flagAction, "action", "", "Filter by action")
	cmd.Flags().StringVar(&flagVisitorID, "visitor-id", "", "Filter by visitor ID")
	cmd.Flags().IntVar(&flagVersion, "version", 0, "Filter by version")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
