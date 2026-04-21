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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateTaskInput!) {
  updateTask(input: $input) {
    task {
      id
      name
      state
      priority
    }
  }
}
`

type updateResponse struct {
	UpdateTask struct {
		Task struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			State    string `json:"state"`
			Priority string `json:"priority"`
		} `json:"task"`
	} `json:"updateTask"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName         string
		flagDescription  string
		flagState        string
		flagPriority     string
		flagTimeEstimate string
		flagDeadline     string
		flagAssignedTo   string
		flagMeasure      string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a task",
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

			input := map[string]any{
				"taskId": args[0],
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}
			if cmd.Flags().Changed("description") {
				input["description"] = flagDescription
			}
			if cmd.Flags().Changed("state") {
				input["state"] = flagState
			}
			if cmd.Flags().Changed("priority") {
				input["priority"] = flagPriority
			}
			if cmd.Flags().Changed("time-estimate") {
				input["timeEstimate"] = flagTimeEstimate
			}
			if cmd.Flags().Changed("deadline") {
				input["deadline"] = flagDeadline
			}
			if cmd.Flags().Changed("assigned-to") {
				if flagAssignedTo == "" {
					input["assignedToId"] = nil
				} else {
					input["assignedToId"] = flagAssignedTo
				}
			}
			if cmd.Flags().Changed("measure") {
				if flagMeasure == "" {
					input["measureId"] = nil
				} else {
					input["measureId"] = flagMeasure
				}
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
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

			t := resp.UpdateTask.Task
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated task %s (%s)\n",
				t.ID,
				t.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Task name")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Task description")
	cmd.Flags().StringVar(&flagState, "state", "", "Task state: TODO, IN_PROGRESS, DONE")
	cmd.Flags().StringVar(&flagPriority, "priority", "", "Task priority: URGENT, HIGH, MEDIUM, LOW")
	cmd.Flags().StringVar(&flagTimeEstimate, "time-estimate", "", "Time estimate")
	cmd.Flags().StringVar(&flagDeadline, "deadline", "", "Deadline")
	cmd.Flags().StringVar(&flagAssignedTo, "assigned-to", "", "Assigned profile ID")
	cmd.Flags().StringVar(&flagMeasure, "measure", "", "Measure ID")

	return cmd
}
