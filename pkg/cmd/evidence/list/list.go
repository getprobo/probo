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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: EvidenceOrder) {
  node(id: $id) {
    __typename
    ... on Measure {
      evidences(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            state
            type
            url
            description
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

type evidence struct {
	ID          string  `json:"id"`
	State       string  `json:"state"`
	Type        string  `json:"type"`
	URL         string  `json:"url"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagMeasure  string
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List evidences for a measure",
		Aliases: []string{"ls"},
		Example: `  # List evidences for a measure
  prb evidence list --measure <measure-id>

  # Output as JSON
  prb evidence ls --measure <measure-id> --output json`,
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

			variables := map[string]any{
				"id": flagMeasure,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			evidences, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[evidence], error) {
					var resp struct {
						Node *struct {
							Typename  string                   `json:"__typename"`
							Evidences api.Connection[evidence] `json:"evidences"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("measure %s not found", flagMeasure)
					}
					if resp.Node.Typename != "Measure" {
						return nil, fmt.Errorf("expected Measure node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Evidences, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, evidences)
			}

			if len(evidences) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No evidences found.")
				return nil
			}

			rows := make([][]string, 0, len(evidences))
			for _, e := range evidences {
				desc := "-"
				if e.Description != nil && *e.Description != "" {
					d := *e.Description
					if len(d) > 60 {
						d = d[:57] + "..."
					}
					desc = d
				}
				rows = append(rows, []string{
					e.ID,
					e.Type,
					e.State,
					desc,
					cmdutil.FormatTime(e.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "TYPE", "STATE", "DESCRIPTION", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(evidences) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d evidences\n",
					len(evidences),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagMeasure, "measure", "", "Measure ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of evidences to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("measure")

	return cmd
}
