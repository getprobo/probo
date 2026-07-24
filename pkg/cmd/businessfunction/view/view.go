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
    ... on BusinessFunction {
      id
      referenceId
      name
      classification
      mtdMinutes
      rtoMinutes
      rpoMinutes
      impactTolerance
      notes
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename        string  `json:"__typename"`
		ID              string  `json:"id"`
		ReferenceID     string  `json:"referenceId"`
		Name            string  `json:"name"`
		Classification  string  `json:"classification"`
		MTDMinutes      int     `json:"mtdMinutes"`
		RTOMinutes      int     `json:"rtoMinutes"`
		RPOMinutes      int     `json:"rpoMinutes"`
		ImpactTolerance *string `json:"impactTolerance"`
		Notes           *string `json:"notes"`
		CreatedAt       string  `json:"createdAt"`
		UpdatedAt       string  `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a business function",
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
				return fmt.Errorf("business function %s not found", args[0])
			}

			if resp.Node.Typename != "BusinessFunction" {
				return fmt.Errorf("expected BusinessFunction node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			bf := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(bf.Name))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), bf.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Reference ID:"), bf.ReferenceID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Classification:"), bf.Classification)
			_, _ = fmt.Fprintf(out, "%s%d min\n", label.Render("MTD:"), bf.MTDMinutes)
			_, _ = fmt.Fprintf(out, "%s%d min\n", label.Render("RTO:"), bf.RTOMinutes)
			_, _ = fmt.Fprintf(out, "%s%d min\n", label.Render("RPO:"), bf.RPOMinutes)

			if bf.ImpactTolerance != nil && *bf.ImpactTolerance != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Impact Tolerance:"), *bf.ImpactTolerance)
			}

			if bf.Notes != nil && *bf.Notes != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Notes:"), *bf.Notes)
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(bf.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(bf.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
