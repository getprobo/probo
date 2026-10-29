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

package linkasana

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const linkMutation = `
mutation($input: LinkTaskToAsanaInput!) {
  linkTaskToAsana(input: $input) {
    task {
      id
      externalLink {
        identifier
        url
      }
    }
  }
}
`

type linkResponse struct {
	LinkTaskToAsana struct {
		Task struct {
			ID           string `json:"id"`
			ExternalLink *struct {
				Identifier string `json:"identifier"`
				URL        string `json:"url"`
			} `json:"externalLink"`
		} `json:"task"`
	} `json:"linkTaskToAsana"`
}

func NewCmdLinkAsana(f *cmdutil.Factory) *cobra.Command {
	var (
		flagProjectID string
		flagTaskGID   string
	)

	cmd := &cobra.Command{
		Use:   "link-asana <id>",
		Short: "Link a task to an existing Asana task",
		Long:  "Link a task to an existing Asana task. The Asana task's name, content, state, deadline, priority, and assignee replace the Probo task.",
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

			data, err := client.Do(
				linkMutation,
				map[string]any{
					"input": map[string]any{
						"taskId":    args[0],
						"projectId": flagProjectID,
						"taskGid":   flagTaskGID,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp linkResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			task := resp.LinkTaskToAsana.Task
			if task.ExternalLink != nil {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"%s %s\n",
					task.ExternalLink.Identifier,
					task.ExternalLink.URL,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagProjectID, "project-id", "", "Asana project GID")
	cmd.Flags().StringVar(&flagTaskGID, "task-gid", "", "Asana task GID")
	_ = cmd.MarkFlagRequired("project-id")
	_ = cmd.MarkFlagRequired("task-gid")

	return cmd
}
