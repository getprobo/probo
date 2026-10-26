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
	"strings"

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

type createEnvelope struct {
	Data struct {
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
	} `json:"data"`
	Errors []graphQLError `json:"errors"`
}

type graphQLError struct {
	Message string `json:"message"`
}

func graphQLErrorMessage(errs []graphQLError) string {
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		if err.Message != "" {
			messages = append(messages, err.Message)
		}
	}

	return strings.Join(messages, "; ")
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                string
		flagName               string
		flagContent            string
		flagState              string
		flagPriority           string
		flagInternalControls   []string
		flagTimeEstimate       string
		flagAssignedTo         string
		flagDeadline           string
		flagRecurrenceInterval string
		flagLinearTeam         string
		flagNoLinearTeam       bool
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

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
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

			if flagPriority == "" {
				return fmt.Errorf("priority is required; pass --priority or run interactively")
			}

			if err := cmdutil.ValidateEnum(
				"priority",
				flagPriority,
				[]string{"URGENT", "HIGH", "MEDIUM", "LOW"},
			); err != nil {
				return err
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"name":           flagName,
				"priority":       flagPriority,
			}

			if flagContent != "" {
				content, err := cmdutil.CLIContent(flagContent)
				if err != nil {
					return fmt.Errorf("cannot read content: %w", err)
				}

				input["content"] = content
			}

			if flagState != "" {
				if err := cmdutil.ValidateEnum("state", flagState, cmdutil.TaskStates()); err != nil {
					return err
				}

				input["state"] = flagState
			}

			if len(flagInternalControls) > 0 {
				internalControlIDs := make([]string, 0, len(flagInternalControls))
				for _, internalControlID := range flagInternalControls {
					if internalControlID != "" {
						internalControlIDs = append(internalControlIDs, internalControlID)
					}
				}

				if len(internalControlIDs) > 0 {
					input["internalControlIds"] = internalControlIDs
				}
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

			if flagRecurrenceInterval != "" {
				if flagDeadline == "" {
					return fmt.Errorf("--recurrence-interval requires --deadline")
				}

				input["recurrenceInterval"] = flagRecurrenceInterval
			}

			if flagNoLinearTeam && flagLinearTeam != "" {
				return fmt.Errorf("--no-linear-team and --linear-team cannot be used together")
			}

			if flagNoLinearTeam {
				input["linearTeamId"] = nil
			} else if flagLinearTeam != "" {
				input["linearTeamId"] = flagLinearTeam
			}

			raw, err := client.DoRaw(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createEnvelope
			if err := json.Unmarshal(raw, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			task := resp.Data.CreateTask.TaskEdge.Node
			if task.ID == "" {
				if len(resp.Errors) > 0 {
					return fmt.Errorf("GraphQL error: %s", graphQLErrorMessage(resp.Errors))
				}

				return fmt.Errorf("cannot parse response: task was not returned")
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created task %s (%s)\n",
				task.ID,
				task.Name,
			)

			if len(resp.Errors) > 0 {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"Linear publish failed: %s\n",
					graphQLErrorMessage(resp.Errors),
				)
			}

			return nil
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Task name (required)")
	cmd.Flags().StringVar(&flagContent, "content", "", "Task content as markdown or ProseMirror JSON. Images and links to /api/files/v1/{id} are kept")
	cmd.Flags().StringVar(&flagState, "state", "", cmdutil.TaskStateFlagUsage())
	cmd.Flags().StringVar(&flagPriority, "priority", "", "Task priority: URGENT, HIGH, MEDIUM, LOW (required)")
	cmd.Flags().StringArrayVar(&flagInternalControls, "internal-control", nil, "Internal control ID (repeatable)")
	cmd.Flags().StringVar(&flagTimeEstimate, "time-estimate", "", "Time estimate")
	cmd.Flags().StringVar(&flagAssignedTo, "assigned-to", "", "Assigned profile ID")
	cmd.Flags().StringVar(&flagDeadline, "deadline", "", "Deadline")
	cmd.Flags().StringVar(&flagRecurrenceInterval, "recurrence-interval", "", "Recurrence interval as an ISO-8601 duration, e.g. P7D, P1M or P1Y (requires --deadline). The next task is created when the deadline passes")
	cmd.Flags().StringVar(&flagLinearTeam, "linear-team", "", "Linear team ID. Omit to use the organization's default Linear team. If the Linear publish attempt fails after the task is created, the task still exists")
	cmd.Flags().BoolVar(&flagNoLinearTeam, "no-linear-team", false, "Create the task with no Linear team")

	return cmd
}
