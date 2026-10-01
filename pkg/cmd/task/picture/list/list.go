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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: TaskPictureOrder) {
  node(id: $id) {
    __typename
    ... on Task {
      pictures(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            createdAt
            linearAssetUrl
            file {
              fileName
              mimeType
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

type picture struct {
	ID             string  `json:"id"`
	CreatedAt      string  `json:"createdAt"`
	LinearAssetURL *string `json:"linearAssetUrl"`
	File           struct {
		FileName string `json:"fileName"`
		MimeType string `json:"mimeType"`
	} `json:"file"`
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
		Short:   "List pictures on a task",
		Aliases: []string{"ls"},
		Example: `  # List pictures on a task
  prb task picture list --task <task-id>`,
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

			pictures, totalCount, err := api.Paginate(
				client,
				listQuery,
				map[string]any{
					"id": flagTask,
					"orderBy": map[string]any{
						"field":     "CREATED_AT",
						"direction": flagOrderDir,
					},
				},
				flagLimit,
				func(data json.RawMessage) (*api.Connection[picture], error) {
					var resp struct {
						Node *struct {
							Typename string                  `json:"__typename"`
							Pictures api.Connection[picture] `json:"pictures"`
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

					return &resp.Node.Pictures, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, pictures)
			}

			if len(pictures) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No pictures found.")
				return nil
			}

			rows := make([][]string, 0, len(pictures))
			for _, picture := range pictures {
				attached := ""
				if picture.LinearAssetURL != nil {
					attached = "yes"
				}

				rows = append(rows, []string{
					picture.ID,
					picture.File.FileName,
					picture.File.MimeType,
					attached,
					cmdutil.FormatTime(picture.CreatedAt),
				})
			}

			table := cmdutil.NewTable("ID", "FILE", "TYPE", "LINEAR", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, table)

			if totalCount > len(pictures) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d pictures\n",
					len(pictures),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTask, "task", "", "Task ID (required)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of pictures to list")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "ASC", "Sort direction (ASC, DESC)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
