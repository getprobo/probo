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
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query CorrectiveActionPlanView($id: ID!) {
  node(id: $id) {
    __typename
    ... on CorrectiveActionPlan {
      id
      title
      description
      dueDate
      owner {
        id
        fullName
      }
      finding {
        id
        referenceId
        title
      }
      internalControls(first: 100) {
        totalCount
        edges {
          node {
            id
            referenceId
            name
          }
        }
      }
      tasks(first: 0) {
        totalCount
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
		Title       string  `json:"title"`
		Description *string `json:"description"`
		DueDate     *string `json:"dueDate"`
		Owner       *struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"owner"`
		Finding struct {
			ID          string `json:"id"`
			ReferenceID string `json:"referenceId"`
			Title       string `json:"title"`
		} `json:"finding"`
		InternalControls struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node struct {
					ID          string `json:"id"`
					ReferenceID string `json:"referenceId"`
					Name        string `json:"name"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"internalControls"`
		Tasks struct {
			TotalCount int `json:"totalCount"`
		} `json:"tasks"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a corrective action plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
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
				return fmt.Errorf("corrective action plan %s not found", args[0])
			}

			if resp.Node.Typename != "CorrectiveActionPlan" {
				return fmt.Errorf("expected CorrectiveActionPlan node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			p := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(p.Title))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), p.ID)
			_, _ = fmt.Fprintf(out, "%s%s %s (%s)\n", label.Render("Finding:"), p.Finding.ReferenceID, p.Finding.Title, p.Finding.ID)

			if p.Owner != nil {
				_, _ = fmt.Fprintf(out, "%s%s (%s)\n", label.Render("Owner:"), p.Owner.FullName, p.Owner.ID)
			}

			if p.DueDate != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Due Date:"), cmdutil.FormatTime(*p.DueDate))
			}

			if p.Description != nil && *p.Description != "" {
				_, _ = fmt.Fprintf(out, "\n%s\n%s\n", label.Render("Description:"), *p.Description)
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Internal Controls:"), p.InternalControls.TotalCount)

			for _, e := range p.InternalControls.Edges {
				_, _ = fmt.Fprintf(out, "  %s %s (%s)\n", e.Node.ReferenceID, e.Node.Name, e.Node.ID)
			}

			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Tasks:"), p.Tasks.TotalCount)

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(p.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(p.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
