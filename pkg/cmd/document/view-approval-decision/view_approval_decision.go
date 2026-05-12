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

package viewapprovaldecision

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
    ... on DocumentVersionApprovalDecision {
      id
      quorum {
        id
      }
      approver {
        id
        fullName
      }
      state
      comment
      decidedAt
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename string `json:"__typename"`
		ID       string `json:"id"`
		Quorum   struct {
			ID string `json:"id"`
		} `json:"quorum"`
		Approver struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"approver"`
		State     string  `json:"state"`
		Comment   *string `json:"comment"`
		DecidedAt *string `json:"decidedAt"`
		CreatedAt string  `json:"createdAt"`
		UpdatedAt string  `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdViewApprovalDecision(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view-approval-decision <id>",
		Short: "View an approval decision",
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
				return fmt.Errorf("approval decision %s not found", args[0])
			}

			if resp.Node.Typename != "DocumentVersionApprovalDecision" {
				return fmt.Errorf("expected DocumentVersionApprovalDecision node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			d := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render("Approval Decision"))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), d.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Quorum ID:"), d.Quorum.ID)
			_, _ = fmt.Fprintf(out, "%s%s (%s)\n", label.Render("Approver:"), d.Approver.FullName, d.Approver.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("State:"), d.State)

			if d.Comment != nil && *d.Comment != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Comment:"), *d.Comment)
			}

			_, _ = fmt.Fprintln(out)

			if d.DecidedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Decided:"), cmdutil.FormatTime(*d.DecidedAt))
			}
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(d.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(d.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
