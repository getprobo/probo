// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
mutation($input: CreateTaskInput!) {
  createTask(input: $input) {
    taskEdge {
      node {
        id
        name
        state
        priority
      }
    }
  }
}
`

type createResponse struct {
	CreateTask struct {
		TaskEdge struct {
			Node struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				State    string `json:"state"`
				Priority string `json:"priority"`
			} `json:"node"`
		} `json:"taskEdge"`
	} `json:"createTask"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg          string
		flagName         string
		flagDescription  string
		flagPriority     string
		flagMeasure      string
		flagTimeEstimate string
		flagAssignedTo   string
		flagDeadline     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new task",
		Example: `  # Create a task interactively
  prb task create

  # Create a task non-interactively
  prb task create --name "Review access controls" --priority HIGH`,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			if f.IOStreams.IsInteractive() {
				if flagName == "" {
					err := huh.NewInput().
						Title("Task name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagPriority == "" {
					err := huh.NewSelect[string]().
						Title("Task priority").
						Options(
							huh.NewOption("Urgent", "URGENT"),
							huh.NewOption("High", "HIGH"),
							huh.NewOption("Medium", "MEDIUM"),
							huh.NewOption("Low", "LOW"),
						).
						Value(&flagPriority).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
			}

			if flagDescription != "" {
				input["description"] = flagDescription
			}
			if flagPriority != "" {
				input["priority"] = flagPriority
			}
			if flagMeasure != "" {
				input["measureId"] = flagMeasure
			}
			if flagTimeEstimate != "" {
				input["timeEstimate"] = flagTimeEstimate
			}
			if flagAssignedTo != "" {
				input["assignedToId"] = flagAssignedTo
			}
			if flagDeadline != "" {
				input["deadline"] = flagDeadline
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

			t := resp.CreateTask.TaskEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created task %s (%s)\n",
				t.ID,
				t.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Task name (required)")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Task description")
	cmd.Flags().StringVar(&flagPriority, "priority", "", "Task priority: URGENT, HIGH, MEDIUM, LOW")
	cmd.Flags().StringVar(&flagMeasure, "measure", "", "Measure ID")
	cmd.Flags().StringVar(&flagTimeEstimate, "time-estimate", "", "Time estimate")
	cmd.Flags().StringVar(&flagAssignedTo, "assigned-to", "", "Assigned profile ID")
	cmd.Flags().StringVar(&flagDeadline, "deadline", "", "Deadline")

	return cmd
}
