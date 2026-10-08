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
		flagName               string
		flagContent            string
		flagState              string
		flagPriority           string
		flagTimeEstimate       string
		flagDeadline           string
		flagAssignedTo         string
		flagInternalControls   []string
		flagRecurrenceInterval string
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

			if cmd.Flags().Changed("content") {
				if flagContent == "" {
					input["content"] = nil
				} else {
					content, err := cmdutil.CLIContent(flagContent)
					if err != nil {
						return err
					}

					input["content"] = content
				}
			}

			if cmd.Flags().Changed("state") {
				if err := cmdutil.ValidateEnum("state", flagState, cmdutil.TaskStates()); err != nil {
					return err
				}

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

			if cmd.Flags().Changed("internal-control") {
				internalControlIDs := make([]string, 0, len(flagInternalControls))
				for _, internalControlID := range flagInternalControls {
					if internalControlID != "" {
						internalControlIDs = append(internalControlIDs, internalControlID)
					}
				}

				input["internalControlIds"] = internalControlIDs
			}

			if cmd.Flags().Changed("recurrence-interval") {
				if flagRecurrenceInterval == "" {
					input["recurrenceInterval"] = nil
				} else {
					input["recurrenceInterval"] = flagRecurrenceInterval
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
	cmd.Flags().StringVar(&flagContent, "content", "", "Task content as markdown or ProseMirror JSON. Images and links to /api/files/v1/attachments/{id} are kept")
	cmd.Flags().StringVar(&flagState, "state", "", cmdutil.TaskStateFlagUsage())
	cmd.Flags().StringVar(&flagPriority, "priority", "", "Task priority: URGENT, HIGH, MEDIUM, LOW")
	cmd.Flags().StringVar(&flagTimeEstimate, "time-estimate", "", "Time estimate")
	cmd.Flags().StringVar(&flagDeadline, "deadline", "", "Deadline")
	cmd.Flags().StringVar(&flagAssignedTo, "assigned-to", "", "Assigned profile ID")
	cmd.Flags().StringArrayVar(&flagInternalControls, "internal-control", nil, "Internal control ID (repeatable, empty value clears all)")
	cmd.Flags().StringVar(&flagRecurrenceInterval, "recurrence-interval", "", "Recurrence interval as an ISO-8601 duration, e.g. P7D, P1M or P1Y (empty clears recurrence). The next task is created when the deadline passes")

	return cmd
}
