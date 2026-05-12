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
query($first: Int, $after: CursorKey, $orderBy: WebhookSubscriptionOrder) {
  viewer {
    organization {
      webhookSubscriptions(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            endpointUrl
            selectedEvents
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

type webhookSubscription struct {
	ID             string   `json:"id"`
	EndpointURL    string   `json:"endpointUrl"`
	SelectedEvents []string `json:"selectedEvents"`
	CreatedAt      string   `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List webhook subscriptions",
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

			variables := map[string]any{}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			webhooks, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[webhookSubscription], error) {
					var resp struct {
						Viewer struct {
							Organization struct {
								WebhookSubscriptions api.Connection[webhookSubscription] `json:"webhookSubscriptions"`
							} `json:"organization"`
						} `json:"viewer"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					return &resp.Viewer.Organization.WebhookSubscriptions, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				if webhooks == nil {
					webhooks = []webhookSubscription{}
				}
				return cmdutil.PrintJSON(f.IOStreams.Out, webhooks)
			}

			if len(webhooks) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No webhook subscriptions found.")
				return nil
			}

			rows := make([][]string, 0, len(webhooks))
			for _, w := range webhooks {
				rows = append(rows, []string{
					w.ID,
					w.EndpointURL,
					fmt.Sprintf("%d events", len(w.SelectedEvents)),
					cmdutil.FormatTime(w.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "ENDPOINT", "EVENTS", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(webhooks) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d webhook subscriptions\n",
					len(webhooks),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of webhook subscriptions to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
