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
    ... on TreatmentPlan {
      id
      treatment
      category
      inherentLikelihood
      inherentImpact
      inherentRiskScore
      residualLikelihood
      residualImpact
      residualRiskScore
      netLikelihood
      netImpact
      netRiskScore
      measures(first: 0) {
        totalCount
      }
      implementedMeasures: measures(first: 0, filter: { state: IMPLEMENTED }) {
        totalCount
      }
      inProgressMeasures: measures(first: 0, filter: { state: IN_PROGRESS }) {
        totalCount
      }
      notImplementedMeasures: measures(first: 0, filter: { state: NOT_IMPLEMENTED }) {
        totalCount
      }
      createdAt
      updatedAt
      owner {
        id
        fullName
      }
      risk {
        id
      }
      riskAnalysis {
        id
      }
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename           string `json:"__typename"`
		ID                 string `json:"id"`
		Treatment          string `json:"treatment"`
		Category           string `json:"category"`
		InherentLikelihood int    `json:"inherentLikelihood"`
		InherentImpact     int    `json:"inherentImpact"`
		InherentRiskScore  int    `json:"inherentRiskScore"`
		ResidualLikelihood int    `json:"residualLikelihood"`
		ResidualImpact     int    `json:"residualImpact"`
		ResidualRiskScore  int    `json:"residualRiskScore"`
		NetLikelihood      int    `json:"netLikelihood"`
		NetImpact          int    `json:"netImpact"`
		NetRiskScore       int    `json:"netRiskScore"`
		Measures           struct {
			TotalCount int `json:"totalCount"`
		} `json:"measures"`
		ImplementedMeasures struct {
			TotalCount int `json:"totalCount"`
		} `json:"implementedMeasures"`
		InProgressMeasures struct {
			TotalCount int `json:"totalCount"`
		} `json:"inProgressMeasures"`
		NotImplementedMeasures struct {
			TotalCount int `json:"totalCount"`
		} `json:"notImplementedMeasures"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
		Owner     struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"owner"`
		Risk struct {
			ID string `json:"id"`
		} `json:"risk"`
		RiskAnalysis struct {
			ID string `json:"id"`
		} `json:"riskAnalysis"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a treatment plan",
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
				return fmt.Errorf("treatment plan %s not found", args[0])
			}

			if resp.Node.Typename != "TreatmentPlan" {
				return fmt.Errorf("expected TreatmentPlan node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			p := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(p.Treatment))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), p.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Risk:"), p.Risk.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Risk Analysis:"), p.RiskAnalysis.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Treatment:"), p.Treatment)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Category:"), p.Category)
			_, _ = fmt.Fprintf(out, "%s%s (%s)\n", label.Render("Owner:"), p.Owner.FullName, p.Owner.ID)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(
				out,
				"%s%d (likelihood: %d, impact: %d)\n",
				label.Render("Inherent:"),
				p.InherentRiskScore,
				p.InherentLikelihood,
				p.InherentImpact,
			)
			_, _ = fmt.Fprintf(
				out,
				"%s%d (likelihood: %d, impact: %d)\n",
				label.Render("Residual:"),
				p.ResidualRiskScore,
				p.ResidualLikelihood,
				p.ResidualImpact,
			)
			_, _ = fmt.Fprintf(
				out,
				"%s%d (likelihood: %d, impact: %d)\n",
				label.Render("Net:"),
				p.NetRiskScore,
				p.NetLikelihood,
				p.NetImpact,
			)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(
				out,
				"%s%d done, %d in progress, %d not implemented, %d total\n",
				label.Render("Measures:"),
				p.ImplementedMeasures.TotalCount,
				p.InProgressMeasures.TotalCount,
				p.NotImplementedMeasures.TotalCount,
				p.Measures.TotalCount,
			)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(p.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(p.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
