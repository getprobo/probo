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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: CompliancePortalCommitmentOrder) {
  node(id: $id) {
    __typename
    ... on CompliancePortalCommitmentGroup {
      commitments(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            icon
            eyebrow
            title
            description
            rank
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

type commitment struct {
	ID          string `json:"id"`
	Icon        string `json:"icon"`
	Eyebrow     string `json:"eyebrow"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Rank        int    `json:"rank"`
	CreatedAt   string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagGroup    string
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List compliance portal commitments",
		Aliases: []string{"ls"},
		Example: `  # List commitments in a group
  prb trust-center commitment list --group <group-id>

  # List commitments sorted by rank
  prb trust-center cmt ls --group <group-id> --order-by RANK`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if flagGroup == "" {
				return fmt.Errorf("group is required; pass --group")
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagGroup,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum("order-by", flagOrderBy, []string{"RANK", "CREATED_AT", "UPDATED_AT"}); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			commitments, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				api.NodeConnection[commitment]("commitments", "CompliancePortalCommitmentGroup", fmt.Sprintf("commitment group %s not found", flagGroup)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				commitments,
				totalCount,
				cmdutil.ListView[commitment]{
					Empty:   "No commitments found.",
					Noun:    "commitments",
					Headers: []string{"ID", "ICON", "EYEBROW", "TITLE", "RANK"},
					Row: func(c commitment) ([]string, error) {
						return []string{
							c.ID,
							c.Icon,
							c.Eyebrow,
							c.Title,
							fmt.Sprintf("%d", c.Rank),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagGroup, "group", "", "Commitment group ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of commitments to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (RANK, CREATED_AT, UPDATED_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
