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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: ControlOrder, $filter: ControlFilter) {
  node(id: $id) {
    __typename
    ... on Framework {
      controls(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            sectionTitle
            name
            description
            bestPractice
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

type control struct {
	ID           string  `json:"id"`
	SectionTitle string  `json:"sectionTitle"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	BestPractice bool    `json:"bestPractice"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagFramework string
		flagLimit     int
		flagOrderBy   string
		flagOrderDir  string
		flagFilter    string
		flagOutput    *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List controls in a framework",
		Aliases: []string{"ls"},
		Example: `  # List controls in a framework
  prb control list --framework <framework-id>

  # Filter and output as JSON
  prb control ls --framework <framework-id> --filter "access" --json`,
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
				"id": flagFramework,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"CREATED_AT", "SECTION_TITLE"}); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			if flagFilter != "" {
				variables["filter"] = map[string]any{
					"query": flagFilter,
				}
			}

			controls, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[control], error) {
					var resp struct {
						Node *struct {
							Typename string                  `json:"__typename"`
							Controls api.Connection[control] `json:"controls"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("framework %s not found", flagFramework)
					}
					if resp.Node.Typename != "Framework" {
						return nil, fmt.Errorf("expected Framework node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Controls, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, controls)
			}

			if len(controls) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No controls found.")
				return nil
			}

			rows := make([][]string, 0, len(controls))
			for _, c := range controls {
				bp := "No"
				if c.BestPractice {
					bp = "Yes"
				}
				rows = append(rows, []string{
					c.ID,
					c.SectionTitle,
					c.Name,
					bp,
				})
			}

			t := cmdutil.NewTable("ID", "SECTION", "NAME", "BEST PRACTICE").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(controls) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d controls\n",
					len(controls),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagFramework, "framework", "", "Framework ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of controls to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, SECTION_TITLE)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVarP(&flagFilter, "filter", "q", "", "Filter controls by search query")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("framework")

	return cmd
}
