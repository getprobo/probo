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
	"go.probo.inc/probo/pkg/cmd/task/comment/owner"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TaskCommentOrder) {
  node(id: $id) {
    __typename
    ... on Task {
      comments(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            content
            createdAt
            owner {
              id
              fullName
            }
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

type comment struct {
	ID        string       `json:"id"`
	Content   string       `json:"content"`
	CreatedAt string       `json:"createdAt"`
	Owner     *owner.Owner `json:"owner"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTask     string
		flagLimit    int
		flagOrderDir string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List comments on a task",
		Aliases: []string{"ls"},
		Example: `  # List comments on a task
  prb task comment list --task <task-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if flagTask == "" {
				return fmt.Errorf("task is required; pass --task")
			}

			if err := cmdutil.ValidateEnum("order-direction", flagOrderDir, []string{"ASC", "DESC"}); err != nil {
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
				"id": flagTask,
				"orderBy": map[string]any{
					"field":     "CREATED_AT",
					"direction": flagOrderDir,
				},
			}

			comments, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				api.NodeConnection[comment]("comments", "Task", fmt.Sprintf("task %s not found", flagTask)),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				comments,
				totalCount,
				cmdutil.ListView[comment]{
					Empty:   "No comments found.",
					Noun:    "comments",
					Headers: []string{"ID", "OWNER", "CONTENT", "CREATED"},
					Row: func(c comment) ([]string, error) {
						content, err := cmdutil.FormatRichText(c.Content)
						if err != nil {
							return nil, fmt.Errorf("cannot format comment content: %w", err)
						}

						return []string{
							c.ID,
							owner.Name(c.Owner),
							content,
							cmdutil.FormatTime(c.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagTask, "task", "", "Task ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of comments to list")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "ASC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
