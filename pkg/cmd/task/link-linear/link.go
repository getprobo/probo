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

package linklinear

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const linkMutation = `
mutation($input: LinkTaskToLinearInput!) {
  linkTaskToLinear(input: $input) {
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
	LinkTaskToLinear struct {
		Task struct {
			ID           string `json:"id"`
			ExternalLink *struct {
				Identifier string `json:"identifier"`
				URL        string `json:"url"`
			} `json:"externalLink"`
		} `json:"task"`
	} `json:"linkTaskToLinear"`
}

func NewCmdLinkLinear(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTeamID  string
		flagIssueID string
		flagYes     bool
	)

	cmd := &cobra.Command{
		Use:   "link-linear <id>",
		Short: "Link a task to an existing Linear issue",
		Long:  "Link a task to an existing Linear issue. This updates the Probo task with the Linear issue.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !flagYes {
				if !f.IOStreams.IsInteractive() {
					return fmt.Errorf("cannot link task: confirmation required, use --yes to confirm")
				}

				var confirmed bool

				err := huh.NewConfirm().
					Title("Linking will update the Probo task with the Linear issue. Continue?").
					Value(&confirmed).
					Run()
				if err != nil {
					return fmt.Errorf("cannot confirm Linear link: %w", err)
				}

				if !confirmed {
					return nil
				}
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

			data, err := client.Do(
				linkMutation,
				map[string]any{
					"input": map[string]any{
						"taskId":  args[0],
						"teamId":  flagTeamID,
						"issueId": flagIssueID,
					},
				},
			)
			if err != nil {
				return fmt.Errorf("cannot link task to Linear: %w", err)
			}

			var resp linkResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			task := resp.LinkTaskToLinear.Task
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

	cmd.Flags().StringVar(&flagTeamID, "team-id", "", "Linear team ID")
	cmd.Flags().StringVar(&flagIssueID, "issue-id", "", "Linear issue ID")
	cmd.Flags().BoolVar(&flagYes, "yes", false, "Confirm that linking updates the Probo task")
	_ = cmd.MarkFlagRequired("team-id")
	_ = cmd.MarkFlagRequired("issue-id")

	return cmd
}
