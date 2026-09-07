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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/prosemirror"
)

const updateMutation = `
mutation($input: UpdateTaskCommentInput!) {
  updateTaskComment(input: $input) {
    taskComment {
      id
      content
    }
  }
}
`

type updateResponse struct {
	UpdateTaskComment struct {
		TaskComment struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		} `json:"taskComment"`
	} `json:"updateTaskComment"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOwner   string
		flagContent string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a task comment",
		Args:  cobra.ExactArgs(1),
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

			if !cmd.Flags().Changed("content") && !cmd.Flags().Changed("owner") {
				return fmt.Errorf("content or owner is required; pass --content or --owner")
			}

			input := map[string]any{
				"taskCommentId": args[0],
			}

			if cmd.Flags().Changed("content") {
				if flagContent == "" {
					input["content"] = nil
				} else {
					input["content"] = prosemirror.FromPlainText(flagContent)
				}
			}

			if cmd.Flags().Changed("owner") {
				input["ownerId"] = flagOwner
			}

			data, err := client.Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated comment %s\n",
				resp.UpdateTaskComment.TaskComment.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringVar(&flagContent, "content", "", "Comment content")

	return cmd
}
