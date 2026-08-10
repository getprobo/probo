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
    ... on AiSystem {
      id
      name
      version
      companyRoles
      status
      source
      purpose
      intendedUseCases
      autonomyLevel
      humanOversightMechanism
      riskClassification
      keyStakeholders
      dataSourcesAndType
      deploymentDate
      lastReviewDate
      nextReviewDate
      notes
      owner {
        id
        fullName
      }
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename                string   `json:"__typename"`
		ID                      string   `json:"id"`
		Name                    string   `json:"name"`
		Version                 *string  `json:"version"`
		CompanyRoles            []string `json:"companyRoles"`
		Status                  string   `json:"status"`
		Source                  *string  `json:"source"`
		Purpose                 *string  `json:"purpose"`
		IntendedUseCases        *string  `json:"intendedUseCases"`
		AutonomyLevel           *string  `json:"autonomyLevel"`
		HumanOversightMechanism *string  `json:"humanOversightMechanism"`
		RiskClassification      *string  `json:"riskClassification"`
		KeyStakeholders         *string  `json:"keyStakeholders"`
		DataSourcesAndType      *string  `json:"dataSourcesAndType"`
		DeploymentDate          *string  `json:"deploymentDate"`
		LastReviewDate          *string  `json:"lastReviewDate"`
		NextReviewDate          *string  `json:"nextReviewDate"`
		Notes                   *string  `json:"notes"`
		Owner                   *struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"owner"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View an AI system",
		Args:  cobra.ExactArgs(1),
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
				map[string]any{"id": args[0]},
			)
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("ai system %s not found", args[0])
			}

			if resp.Node.Typename != "AiSystem" {
				return fmt.Errorf("expected AiSystem node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			system := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(28)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(system.Name))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), system.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Status:"), system.Status)
			_, _ = fmt.Fprintf(out, "%s%v\n", label.Render("Company roles:"), system.CompanyRoles)

			printOptionalString(out, label, "Version:", system.Version)
			printOptionalString(out, label, "Risk classification:", system.RiskClassification)
			printOptionalString(out, label, "Source:", system.Source)
			printOptionalString(out, label, "Purpose:", system.Purpose)
			printOptionalString(out, label, "Intended use cases:", system.IntendedUseCases)
			printOptionalString(out, label, "Autonomy level:", system.AutonomyLevel)
			printOptionalString(out, label, "Human oversight:", system.HumanOversightMechanism)
			printOptionalString(out, label, "Key stakeholders:", system.KeyStakeholders)
			printOptionalString(out, label, "Data sources and type:", system.DataSourcesAndType)
			printOptionalString(out, label, "Deployment date:", system.DeploymentDate)
			printOptionalString(out, label, "Last review date:", system.LastReviewDate)
			printOptionalString(out, label, "Next review date:", system.NextReviewDate)
			printOptionalString(out, label, "Notes:", system.Notes)

			if system.Owner != nil {
				_, _ = fmt.Fprintf(
					out,
					"%s%s (%s)\n",
					label.Render("Owner:"),
					system.Owner.FullName,
					system.Owner.ID,
				)
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(system.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(system.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func printOptionalString(out interface{ Write([]byte) (int, error) }, label lipgloss.Style, name string, value *string) {
	if value == nil || *value == "" {
		return
	}

	_, _ = fmt.Fprintf(out, "%s%s\n", label.Render(name), *value)
}
