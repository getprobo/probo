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
query ThirdPartyAgreementList($id: ID!, $first: Int, $after: CursorKey, $orderBy: ThirdPartyAgreementOrder) {
  node(id: $id) {
    __typename
    ... on ThirdParty {
      agreements(first: $first, after: $after, orderBy: $orderBy) {
        edges {
          node {
            id
            title
            type
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

type agreement struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	CreatedAt string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagThirdParty string
		flagLimit      int
		flagAfter      string
		flagOrderBy    string
		flagOrderDir   string
		flagOutput     *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List agreements for a third party",
		Aliases: []string{"ls"},
		Example: `  # List agreements for a third party
  prb third-party agreement list --third-party <third-party-id>`,
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
				"id": flagThirdParty,
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

			agreements, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[agreement]("agreements", "ThirdParty", fmt.Sprintf("third party %s not found", flagThirdParty)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				agreements,
				totalCount,
				cmdutil.ListView[agreement]{
					Empty:      "No agreements found.",
					Noun:       "agreements",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "TITLE", "TYPE", "CREATED"},
					Row: func(item agreement) ([]string, error) {
						return []string{
							item.ID,
							item.Title,
							item.Type,
							cmdutil.FormatTime(item.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagThirdParty, "third-party", "", "Third party ID")
	_ = cmd.MarkFlagRequired("third-party")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of agreements to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
