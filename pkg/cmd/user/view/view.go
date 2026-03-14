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
    ... on Profile {
      id
      fullName
      emailAddress
      state
      additionalEmailAddresses
      kind
      position
      contractStartDate
      contractEndDate
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node struct {
		ID                       string   `json:"id"`
		FullName                 string   `json:"fullName"`
		EmailAddress             string   `json:"emailAddress"`
		State                    string   `json:"state"`
		AdditionalEmailAddresses []string `json:"additionalEmailAddresses"`
		Kind                     *string  `json:"kind"`
		Position                 *string  `json:"position"`
		ContractStartDate        *string  `json:"contractStartDate"`
		ContractEndDate          *string  `json:"contractEndDate"`
		CreatedAt                string   `json:"createdAt"`
		UpdatedAt                string   `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <id>",
		Short: "View a user",
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

			p := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			fmt.Fprintf(out, "%s\n\n", bold.Render(p.FullName))

			fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), p.ID)
			fmt.Fprintf(out, "%s%s\n", label.Render("Email:"), p.EmailAddress)
			fmt.Fprintf(out, "%s%s\n", label.Render("State:"), p.State)

			if p.Kind != nil && *p.Kind != "" {
				fmt.Fprintf(out, "%s%s\n", label.Render("Kind:"), *p.Kind)
			}

			if p.Position != nil && *p.Position != "" {
				fmt.Fprintf(out, "%s%s\n", label.Render("Position:"), *p.Position)
			}

			if len(p.AdditionalEmailAddresses) > 0 {
				fmt.Fprintln(out)
				fmt.Fprintf(out, "%s\n", bold.Render("Additional Emails"))
				for _, email := range p.AdditionalEmailAddresses {
					fmt.Fprintf(out, "  %s\n", email)
				}
			}

			fmt.Fprintln(out)
			if p.ContractStartDate != nil {
				fmt.Fprintf(out, "%s%s\n", label.Render("Contract Start:"), *p.ContractStartDate)
			}
			if p.ContractEndDate != nil {
				fmt.Fprintf(out, "%s%s\n", label.Render("Contract End:"), *p.ContractEndDate)
			}

			fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), p.CreatedAt)
			fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), p.UpdatedAt)

			return nil
		},
	}
}
