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

package listapprovalquorums

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DocumentVersionApprovalQuorumOrder) {
  node(id: $id) {
    __typename
    ... on DocumentVersion {
      approvalQuorums(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            status
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

type approvalQuorum struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"`
}

func NewCmdListApprovalQuorums(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit    int
		flagAfter    string
		flagOrderBy  string
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "list-approval-quorums <document-version-id>",
		Short: "List approval quorums for a document version",
		Example: `  # List approval quorums for a document version
  prb document list-approval-quorums <document-version-id>`,
		Args: cobra.ExactArgs(1),
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
				"id": args[0],
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			quorums, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[approvalQuorum]("approvalQuorums", "DocumentVersion", fmt.Sprintf("document version %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				quorums,
				totalCount,
				cmdutil.ListView[approvalQuorum]{
					Empty:      "No approval quorums found.",
					Noun:       "approval quorums",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "STATUS", "CREATED AT"},
					Row: func(q approvalQuorum) ([]string, error) {
						return []string{
							q.ID,
							q.Status,
							cmdutil.FormatTime(q.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of approval quorums to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT)")
	cmdutil.AddOrderDirectionFlag(cmd, &flagOrderDir, "DESC")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
