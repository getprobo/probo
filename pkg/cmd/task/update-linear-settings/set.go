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

package updatelinearsettings

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateLinearSettingsInput!) {
  updateLinearSettings(input: $input) {
    organization {
      linearSettings {
        defaultTeam {
          id
          name
          key
        }
        issueImport
      }
    }
  }
}
`

type updateResponse struct {
	UpdateLinearSettings struct {
		Organization struct {
			LinearSettings struct {
				DefaultTeam *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
					Key  string `json:"key"`
				} `json:"defaultTeam"`
				IssueImport string `json:"issueImport"`
			} `json:"linearSettings"`
		} `json:"organization"`
	} `json:"updateLinearSettings"`
}

func NewCmdUpdateLinearSettings(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTeam        string
		flagClearTeam   bool
		flagIssueImport string
	)

	cmd := &cobra.Command{
		Use:   "update-linear-settings <organization-id>",
		Short: "Update the organization's Linear settings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagClearTeam && flagTeam != "" {
				return fmt.Errorf("--clear-team and --team cannot be used together")
			}

			status, err := linearIssueImportStatus(flagIssueImport)
			if err != nil {
				return err
			}

			if !flagClearTeam && flagTeam == "" && status == "" {
				return fmt.Errorf("pass --team, --clear-team, or --issue-import")
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
			if flagClearTeam {
				input["defaultTeamId"] = nil
			} else if flagTeam != "" {
				input["defaultTeamId"] = flagTeam
			}

			if status != "" {
				input["issueImport"] = status
			}

			data, err := client.Do(updateMutation, map[string]any{"input": input})
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			settings := resp.UpdateLinearSettings.Organization.LinearSettings
			if settings.DefaultTeam == nil {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "Cleared the default Linear team")
			} else {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"%s\t%s\t%s\n",
					settings.DefaultTeam.ID,
					settings.DefaultTeam.Key,
					settings.DefaultTeam.Name,
				)
			}

			_, _ = fmt.Fprintln(f.IOStreams.Out, strings.ToLower(settings.IssueImport))

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTeam, "team", "", "Linear team ID")
	cmd.Flags().BoolVar(&flagClearTeam, "clear-team", false, "Clear the default Linear team")
	cmd.Flags().StringVar(&flagIssueImport, "issue-import", "", "Issue import: active or inactive")

	return cmd
}

func linearIssueImportStatus(status string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "":
		return "", nil
	case "active":
		return "ACTIVE", nil
	case "inactive":
		return "INACTIVE", nil
	default:
		return "", fmt.Errorf("--issue-import must be active or inactive")
	}
}
