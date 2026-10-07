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

package view

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on EmployeePortal {
      id
      name
      active
      capabilities {
        deviceAgent
      }
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename     string `json:"__typename"`
		ID           string `json:"id"`
		Name         string `json:"name"`
		Active       bool   `json:"active"`
		Capabilities struct {
			DeviceAgent bool `json:"deviceAgent"`
		} `json:"capabilities"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var (
		flagPortal string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View employee portal settings",
		Example: `  # View an employee portal
  prb employee-portal view --portal <employee-portal-id>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
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

			data, err := client.Do(
				viewQuery,
				map[string]any{"id": flagPortal},
			)
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("employee portal %s not found", flagPortal)
			}

			if resp.Node.Typename != "EmployeePortal" {
				return fmt.Errorf("expected EmployeePortal node, got %s", resp.Node.Typename)
			}

			portal := resp.Node

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, portal)
			}

			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(16)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Employee Portal"))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), portal.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Name:"), portal.Name)
			_, _ = fmt.Fprintf(out, "%s%v\n", label.Render("Active:"), portal.Active)
			_, _ = fmt.Fprintf(out, "%s%v\n", label.Render("Device Agent:"), portal.Capabilities.DeviceAgent)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(portal.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(portal.UpdatedAt))

			return nil
		},
	}

	cmd.Flags().StringVar(&flagPortal, "portal", "", "Employee portal ID")
	_ = cmd.MarkFlagRequired("portal")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
