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

package asanaprojectsetup

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const setupQuery = `
query($id: ID!, $projectId: String!) {
  node(id: $id) {
    __typename
    ... on Organization {
      asanaProjectSetup(projectId: $projectId) {
        project { id name }
        sections { id name }
        enumFields {
          id
          name
          options { id name }
        }
        mapping {
          states { state sectionId inbound }
          priorityFieldId
          priorities { priority optionId inbound }
        }
      }
    }
  }
}
`

type setupResponse struct {
	Node *struct {
		Typename          string `json:"__typename"`
		AsanaProjectSetup struct {
			Project struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"project"`
			Sections []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"sections"`
			EnumFields []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Options []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"options"`
			} `json:"enumFields"`
			Mapping *struct {
				States []struct {
					State     string `json:"state"`
					SectionID string `json:"sectionId"`
					Inbound   bool   `json:"inbound"`
				} `json:"states"`
				PriorityFieldID *string `json:"priorityFieldId"`
				Priorities      []struct {
					Priority string `json:"priority"`
					OptionID string `json:"optionId"`
					Inbound  bool   `json:"inbound"`
				} `json:"priorities"`
			} `json:"mapping"`
		} `json:"asanaProjectSetup"`
	} `json:"node"`
}

func NewCmdAsanaProjectSetup(f *cmdutil.Factory) *cobra.Command {
	var flagProjectID string

	cmd := &cobra.Command{
		Use:   "asana-project-setup <organization-id>",
		Short: "Show an Asana project's sections, priority fields, and mapping",
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

			data, err := client.Do(setupQuery, map[string]any{
				"id":        args[0],
				"projectId": flagProjectID,
			})
			if err != nil {
				return err
			}

			var resp setupResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse Asana project setup: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("organization %s not found", args[0])
			}

			if resp.Node.Typename != "Organization" {
				return fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
			}

			setup := resp.Node.AsanaProjectSetup
			_, _ = fmt.Fprintf(f.IOStreams.Out, "%s\t%s\n", setup.Project.ID, setup.Project.Name)

			for _, section := range setup.Sections {
				_, _ = fmt.Fprintf(f.IOStreams.Out, "section\t%s\t%s\n", section.ID, section.Name)
			}

			for _, field := range setup.EnumFields {
				for _, option := range field.Options {
					_, _ = fmt.Fprintf(
						f.IOStreams.Out,
						"option\t%s\t%s\t%s\t%s\n",
						field.ID,
						field.Name,
						option.ID,
						option.Name,
					)
				}
			}

			if setup.Mapping == nil {
				return nil
			}

			for _, state := range setup.Mapping.States {
				inbound := "outbound"
				if state.Inbound {
					inbound = "inbound"
				}

				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"state\t%s\t%s\t%s\n",
					state.State,
					state.SectionID,
					inbound,
				)
			}

			if setup.Mapping.PriorityFieldID != nil {
				_, _ = fmt.Fprintf(f.IOStreams.Out, "priority-field\t%s\n", *setup.Mapping.PriorityFieldID)
			}

			for _, priority := range setup.Mapping.Priorities {
				inbound := "outbound"
				if priority.Inbound {
					inbound = "inbound"
				}

				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"priority\t%s\t%s\t%s\n",
					priority.Priority,
					priority.OptionID,
					inbound,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagProjectID, "project-id", "", "Asana project GID")
	_ = cmd.MarkFlagRequired("project-id")

	return cmd
}
