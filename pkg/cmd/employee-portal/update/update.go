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
mutation($input: UpdateEmployeePortalInput!) {
  updateEmployeePortal(input: $input) {
    employeePortal {
      id
      name
      active
      capabilities {
        deviceAgent
      }
    }
  }
}
`

type updateResponse struct {
	UpdateEmployeePortal struct {
		EmployeePortal struct {
			ID           string `json:"id"`
			Name         string `json:"name"`
			Active       bool   `json:"active"`
			Capabilities struct {
				DeviceAgent bool `json:"deviceAgent"`
			} `json:"capabilities"`
		} `json:"employeePortal"`
	} `json:"updateEmployeePortal"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagPortal      string
		flagName        string
		flagActive      bool
		flagDeviceAgent bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update employee portal settings",
		Example: `  # Rename an employee portal
  prb employee-portal update --name "Acme Employees"

  # Hide the device agent in the employee app
  prb employee-portal update --device-agent=false`,
		Args: cobra.NoArgs,
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
				"employeePortalId": flagPortal,
			}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}

			if cmd.Flags().Changed("active") {
				input["active"] = flagActive
			}

			if cmd.Flags().Changed("device-agent") {
				input["capabilities"] = map[string]any{
					"deviceAgent": flagDeviceAgent,
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

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated employee portal %s\n",
				resp.UpdateEmployeePortal.EmployeePortal.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagPortal, "portal", "", "Employee portal ID")
	_ = cmd.MarkFlagRequired("portal")
	cmd.Flags().StringVar(&flagName, "name", "", "Display name of the employee portal")
	cmd.Flags().BoolVar(&flagActive, "active", false, "Enable or disable the employee portal")
	cmd.Flags().BoolVar(&flagDeviceAgent, "device-agent", false, "Offer the device agent in the employee app")

	return cmd
}
