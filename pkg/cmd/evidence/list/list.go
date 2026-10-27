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
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: EvidenceOrder) {
  node(id: $id) {
    __typename
    ... on InternalControl {
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
		flagInternalControl string
		flagLimit           int
		flagAfter           string
		flagOrderBy         string
		flagOrderDir        string
		flagOutput          *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List evidences for an internal control",
		Aliases: []string{"ls"},
		Example: `  # List evidences for an internal control
  prb evidence list --internal-control <internal-control-id>

  # Output as JSON
  prb evidence ls --internal-control <internal-control-id> --output json`,
		Args: cobra.NoArgs,
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

			variables := map[string]any{
				"id": flagInternalControl,
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

			evidences, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[evidence]("evidences", "InternalControl", fmt.Sprintf("internal control %s not found", flagInternalControl)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				evidences,
				totalCount,
				cmdutil.ListView[evidence]{
					Empty:      "No evidences found.",
					Noun:       "evidences",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TYPE", "STATE", "DESCRIPTION", "CREATED"},
					Row: func(e evidence) ([]string, error) {
						desc := "-"

						if e.Description != nil && *e.Description != "" {
							d := *e.Description
							if len(d) > 60 {
								d = d[:57] + "..."
							}

							desc = d
						}

						return []string{
							e.ID,
							e.Type,
							e.State,
							desc,
							cmdutil.FormatTime(e.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagInternalControl, "internal-control", "", "Internal control ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of evidences to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("internal-control")

	return cmd
}
