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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateTaskCommentInput!) {
  createTaskComment(input: $input) {
    taskCommentEdge {
      node {
        id
        description
      }
    }
  }
}
`

type createResponse struct {
	CreateTaskComment struct {
		TaskCommentEdge struct {
			Node struct {
				ID          string `json:"id"`
				Description string `json:"description"`
			} `json:"node"`
		} `json:"taskCommentEdge"`
	} `json:"createTaskComment"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTask        string
		flagOwner       string
		flagDescription string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Add a comment to a task",
		Example: `  # Create a comment interactively
  prb task comment create

  # Create a comment non-interactively
  prb task comment create --task <task-id> --description "Looks good"`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if f.IOStreams.IsInteractive() {
				if flagTask == "" {
					err := huh.NewInput().
						Title("Task ID").
						Value(&flagTask).
						Run()
					if err != nil {
						return err
					}
				}

				if flagDescription == "" {
					err := huh.NewText().
						Title("Comment").
						Value(&flagDescription).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagTask == "" {
				return fmt.Errorf("task is required; pass --task or run interactively")
			}

			if flagDescription == "" {
				return fmt.Errorf("description is required; pass --description or run interactively")
			}

			input := map[string]any{
				"taskId":      flagTask,
				"description": flagDescription,
			}

			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created comment %s\n",
				resp.CreateTaskComment.TaskCommentEdge.Node.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTask, "task", "", "Task ID (required)")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID (defaults to the authenticated user)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Comment description (required)")

	return cmd
}
