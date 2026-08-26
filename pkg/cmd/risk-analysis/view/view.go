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
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!, $asOf: Datetime) {
  node(id: $id) {
    __typename
    ... on RiskAnalysis {
      id
      name
      description
      period {
        start
        end
      }
      matrixSize {
        rows
        cols
      }
      matrixCells(asOf: $asOf) {
        type
        likelihood
        impact
        count
      }
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename    string  `json:"__typename"`
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Period      *struct {
			Start *string `json:"start"`
			End   *string `json:"end"`
		} `json:"period"`
		MatrixSize struct {
			Rows int `json:"rows"`
			Cols int `json:"cols"`
		} `json:"matrixSize"`
		MatrixCells []struct {
			Type       string `json:"type"`
			Likelihood int    `json:"likelihood"`
			Impact     int    `json:"impact"`
			Count      int    `json:"count"`
		} `json:"matrixCells"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var (
		flagAsOf   string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a risk analysis",
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

			var asOf *string

			if flagAsOf != "" {
				if _, err := time.Parse(time.RFC3339, flagAsOf); err != nil {
					return fmt.Errorf(
						"--as-of must be RFC3339 (e.g. 2026-01-15T23:59:59Z): %w",
						err,
					)
				}

				asOf = &flagAsOf
			}

			data, err := client.Do(
				viewQuery,
				map[string]any{"id": args[0], "asOf": asOf},
			)
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("risk analysis %s not found", args[0])
			}

			if resp.Node.Typename != "RiskAnalysis" {
				return fmt.Errorf("expected RiskAnalysis node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			r := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(r.Name))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), r.ID)

			if r.Description != nil && *r.Description != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), *r.Description)
			}

			if r.Period != nil {
				if r.Period.Start != nil {
					_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Period start:"), cmdutil.FormatTime(*r.Period.Start))
				}

				if r.Period.End != nil {
					_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Period end:"), cmdutil.FormatTime(*r.Period.End))
				}
			}

			_, _ = fmt.Fprintf(out, "%s%d×%d\n", label.Render("Matrix size:"), r.MatrixSize.Rows, r.MatrixSize.Cols)

			if len(r.MatrixCells) > 0 {
				_, _ = fmt.Fprintf(out, "%s%d occupied cells\n", label.Render("Matrix cells:"), len(r.MatrixCells))
				for _, cell := range r.MatrixCells {
					_, _ = fmt.Fprintf(
						out,
						"%s%s %d×%d (%d)\n",
						label.Render(""),
						cell.Type,
						cell.Likelihood,
						cell.Impact,
						cell.Count,
					)
				}
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(r.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(r.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)
	cmd.Flags().StringVar(
		&flagAsOf,
		"as-of",
		"",
		"Reconstruct matrix cells as of this RFC3339 instant (omit for live tables)",
	)

	return cmd
}
