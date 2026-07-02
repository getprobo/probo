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
    ... on Finding {
      id
      referenceId
      kind
      description
      source
      identifiedOn
      rootCause
      correctiveAction
      owner {
        id
        fullName
      }
      dueDate
      status
      priority
      risk {
        id
        name
      }
      effectivenessCheck
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename         string  `json:"__typename"`
		ID               string  `json:"id"`
		ReferenceID      string  `json:"referenceId"`
		Kind             string  `json:"kind"`
		Description      *string `json:"description"`
		Source           *string `json:"source"`
		IdentifiedOn     *string `json:"identifiedOn"`
		RootCause        *string `json:"rootCause"`
		CorrectiveAction *string `json:"correctiveAction"`
		Owner            struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"owner"`
		DueDate  *string `json:"dueDate"`
		Status   string  `json:"status"`
		Priority string  `json:"priority"`
		Risk     *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"risk"`
		EffectivenessCheck *string `json:"effectivenessCheck"`
		CreatedAt          string  `json:"createdAt"`
		UpdatedAt          string  `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a finding",
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
				return fmt.Errorf("finding %s not found", args[0])
			}

			if resp.Node.Typename != "Finding" {
				return fmt.Errorf("expected Finding node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			n := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(n.ReferenceID))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), n.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Kind:"), n.Kind)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Status:"), n.Status)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Priority:"), n.Priority)
			_, _ = fmt.Fprintf(out, "%s%s (%s)\n", label.Render("Owner:"), n.Owner.FullName, n.Owner.ID)

			if n.Description != nil && *n.Description != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), *n.Description)
			}

			if n.Source != nil && *n.Source != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Source:"), *n.Source)
			}

			if n.IdentifiedOn != nil && *n.IdentifiedOn != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Identified On:"), cmdutil.FormatTime(*n.IdentifiedOn))
			}

			if n.DueDate != nil && *n.DueDate != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Due Date:"), cmdutil.FormatTime(*n.DueDate))
			}

			if n.RootCause != nil && *n.RootCause != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Root Cause:"), *n.RootCause)
			}

			if n.CorrectiveAction != nil && *n.CorrectiveAction != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Corrective Action:"), *n.CorrectiveAction)
			}

			if n.Risk != nil {
				_, _ = fmt.Fprintf(out, "%s%s (%s)\n", label.Render("Risk:"), n.Risk.Name, n.Risk.ID)
			}

			if n.EffectivenessCheck != nil && *n.EffectivenessCheck != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Effectiveness Check:"), *n.EffectivenessCheck)
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(n.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(n.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
