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

package updateasanaprojectmapping

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateAsanaProjectMappingInput!) {
  updateAsanaProjectMapping(input: $input) {
    setup {
      project { id name }
    }
  }
}
`

type stateRow struct {
	State     string `json:"state"`
	SectionID string `json:"sectionId"`
	Inbound   bool   `json:"inbound"`
}

type priorityRow struct {
	Priority string `json:"priority"`
	OptionID string `json:"optionId"`
	Inbound  bool   `json:"inbound"`
}

type updateResponse struct {
	UpdateAsanaProjectMapping struct {
		Setup struct {
			Project struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"project"`
		} `json:"setup"`
	} `json:"updateAsanaProjectMapping"`
}

func NewCmdUpdateAsanaProjectMapping(f *cmdutil.Factory) *cobra.Command {
	var (
		flagProjectID     string
		flagClear         bool
		flagStates        string
		flagPriorityField string
		flagClearPriority bool
		flagPriorities    string
	)

	cmd := &cobra.Command{
		Use:   "update-asana-project-mapping <organization-id>",
		Short: "Save how an Asana project maps onto Probo states and priorities",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			statesChanged := cmd.Flags().Changed("states")
			prioritiesChanged := cmd.Flags().Changed("priorities")
			fieldChanged := cmd.Flags().Changed("priority-field")

			if flagClearPriority && fieldChanged {
				return fmt.Errorf("--clear-priority and --priority-field cannot be used together")
			}

			if !flagClear && !statesChanged && !prioritiesChanged && !fieldChanged && !flagClearPriority {
				return fmt.Errorf("pass --clear, --states, --priorities, --priority-field, or --clear-priority")
			}

			states, err := decodeRows[stateRow](flagStates, "states")
			if err != nil {
				return err
			}

			priorities, err := decodeRows[priorityRow](flagPriorities, "priorities")
			if err != nil {
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

			input := map[string]any{
				"organizationId": args[0],
				"projectId":      flagProjectID,
			}
			if flagClear {
				input["clear"] = true
			}

			if statesChanged {
				input["states"] = states
			}

			if prioritiesChanged {
				input["priorities"] = priorities
			}

			if flagClearPriority {
				input["priorityFieldId"] = nil
			} else if fieldChanged {
				input["priorityFieldId"] = flagPriorityField
			}

			data, err := client.Do(updateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			project := resp.UpdateAsanaProjectMapping.Setup.Project
			_, _ = fmt.Fprintf(f.IOStreams.Out, "%s\t%s\n", project.ID, project.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagProjectID, "project-id", "", "Asana project GID")
	cmd.Flags().BoolVar(&flagClear, "clear", false, "Remove the saved mapping")
	cmd.Flags().StringVar(&flagStates, "states", "", "JSON array of {state, sectionId, inbound}")
	cmd.Flags().StringVar(&flagPriorityField, "priority-field", "", "Asana enum custom field GID")
	cmd.Flags().BoolVar(&flagClearPriority, "clear-priority", false, "Stop syncing priority")
	cmd.Flags().StringVar(&flagPriorities, "priorities", "", "JSON array of {priority, optionId, inbound}")
	_ = cmd.MarkFlagRequired("project-id")

	return cmd
}

func decodeRows[T any](raw string, name string) ([]T, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	var rows []T
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil, fmt.Errorf("--%s must be a JSON array: %w", name, err)
	}

	return rows, nil
}
