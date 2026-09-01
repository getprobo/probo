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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
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
            description
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
	ID          string `json:"id"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	Owner       struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"owner"`
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
				func(data json.RawMessage) (*api.Connection[comment], error) {
					var resp struct {
						Node *struct {
							Typename string                  `json:"__typename"`
							Comments api.Connection[comment] `json:"comments"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("task %s not found", flagTask)
					}

					if resp.Node.Typename != "Task" {
						return nil, fmt.Errorf("expected Task node, got %s", resp.Node.Typename)
					}

					return &resp.Node.Comments, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, comments)
			}

			if len(comments) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No comments found.")
				return nil
			}

			rows := make([][]string, 0, len(comments))
			for _, c := range comments {
				rows = append(rows, []string{
					c.ID,
					c.Owner.FullName,
					c.Description,
					cmdutil.FormatTime(c.CreatedAt),
				})
			}

			t := cmdutil.NewTable("ID", "OWNER", "DESCRIPTION", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(comments) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d comments\n",
					len(comments),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTask, "task", "", "Task ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of comments to list")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "ASC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
