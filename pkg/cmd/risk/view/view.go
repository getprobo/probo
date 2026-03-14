// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package view

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    ... on Risk {
      id
      name
      description
      category
      treatment
      note
      inherentLikelihood
      inherentImpact
      inherentRiskScore
      residualLikelihood
      residualImpact
      residualRiskScore
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		ID                  string  `json:"id"`
		Name                string  `json:"name"`
		Description         *string `json:"description"`
		Category            string  `json:"category"`
		Treatment           string  `json:"treatment"`
		Note                string  `json:"note"`
		InherentLikelihood  int     `json:"inherentLikelihood"`
		InherentImpact      int     `json:"inherentImpact"`
		InherentRiskScore   int     `json:"inherentRiskScore"`
		ResidualLikelihood  int     `json:"residualLikelihood"`
		ResidualImpact      int     `json:"residualImpact"`
		ResidualRiskScore   int     `json:"residualRiskScore"`
		CreatedAt           string  `json:"createdAt"`
		UpdatedAt           string  `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View a risk",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.APIClient()
			if err != nil {
				return err
			}

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
				return fmt.Errorf("risk %s not found", args[0])
			}

			r := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			fmt.Fprintf(out, "%s\n\n", bold.Render(r.Name))

			fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), r.ID)
			fmt.Fprintf(out, "%s%s\n", label.Render("Category:"), r.Category)
			fmt.Fprintf(out, "%s%s\n", label.Render("Treatment:"), r.Treatment)

			if r.Description != nil && *r.Description != "" {
				fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), *r.Description)
			}

			if r.Note != "" {
				fmt.Fprintf(out, "%s%s\n", label.Render("Note:"), r.Note)
			}

			fmt.Fprintln(out)
			fmt.Fprintf(
				out,
				"%s%d (likelihood: %d, impact: %d)\n",
				label.Render("Inherent Risk Score:"),
				r.InherentRiskScore,
				r.InherentLikelihood,
				r.InherentImpact,
			)
			fmt.Fprintf(
				out,
				"%s%d (likelihood: %d, impact: %d)\n",
				label.Render("Residual Risk Score:"),
				r.ResidualRiskScore,
				r.ResidualLikelihood,
				r.ResidualImpact,
			)

			fmt.Fprintln(out)
			fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), r.CreatedAt)
			fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), r.UpdatedAt)

			return nil
		},
	}
}
