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

package updateasanasettings

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateAsanaSettingsInput!) {
  updateAsanaSettings(input: $input) {
    organization {
      asanaSettings {
        defaultProject {
          id
          name
        }
        taskImport
      }
    }
  }
}
`

type updateResponse struct {
	UpdateAsanaSettings struct {
		Organization struct {
			AsanaSettings struct {
				DefaultProject *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"defaultProject"`
				TaskImport string `json:"taskImport"`
			} `json:"asanaSettings"`
		} `json:"organization"`
	} `json:"updateAsanaSettings"`
}

func NewCmdUpdateAsanaSettings(f *cmdutil.Factory) *cobra.Command {
	var (
		flagProject      string
		flagClearProject bool
		flagTaskImport   string
	)

	cmd := &cobra.Command{
		Use:   "update-asana-settings <organization-id>",
		Short: "Update the organization's Asana settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagClearProject && flagProject != "" {
				return fmt.Errorf("--clear-project and --project cannot be used together")
			}

			status, err := asanaTaskImportStatus(flagTaskImport)
			if err != nil {
				return err
			}

			if !flagClearProject && flagProject == "" && status == "" {
				return fmt.Errorf("pass --project, --clear-project, or --task-import")
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
			}
			if flagClearProject {
				input["defaultProjectId"] = nil
			} else if flagProject != "" {
				input["defaultProjectId"] = flagProject
			}

			if status != "" {
				input["taskImport"] = status
			}

			data, err := client.Do(updateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			settings := resp.UpdateAsanaSettings.Organization.AsanaSettings
			if settings.DefaultProject == nil {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "Cleared the default Asana project")
			} else {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"%s\t%s\n",
					settings.DefaultProject.ID,
					settings.DefaultProject.Name,
				)
			}

			_, _ = fmt.Fprintln(f.IOStreams.Out, strings.ToLower(settings.TaskImport))

			return nil
		},
	}

	cmd.Flags().StringVar(&flagProject, "project", "", "Asana project GID")
	cmd.Flags().BoolVar(&flagClearProject, "clear-project", false, "Clear the default Asana project")
	cmd.Flags().StringVar(&flagTaskImport, "task-import", "", "Task import: active or inactive")

	return cmd
}

func asanaTaskImportStatus(status string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "":
		return "", nil
	case "active":
		return "ACTIVE", nil
	case "inactive":
		return "INACTIVE", nil
	default:
		return "", fmt.Errorf("--task-import must be active or inactive")
	}
}
