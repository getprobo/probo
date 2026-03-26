// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Evidence {
      id
      size
      state
      type
      url
      description
      file {
        id
        filename
        contentType
      }
      measure {
        id
      }
      task {
        id
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
		Size        int     `json:"size"`
		State       string  `json:"state"`
		Type        string  `json:"type"`
		URL         string  `json:"url"`
		Description *string `json:"description"`
		File        *struct {
			ID          string `json:"id"`
			Filename    string `json:"filename"`
			ContentType string `json:"contentType"`
		} `json:"file"`
		Measure struct {
			ID string `json:"id"`
		} `json:"measure"`
		Task *struct {
			ID string `json:"id"`
		} `json:"task"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View an evidence",
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
				return fmt.Errorf("evidence %s not found", args[0])
			}

			if resp.Node.Typename != "Evidence" {
				return fmt.Errorf("expected Evidence node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			n := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(n.ID))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Type:"), n.Type)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("State:"), n.State)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Measure:"), n.Measure.ID)

			if n.File != nil {
				_, _ = fmt.Fprintf(out, "%s%s (%s)\n", label.Render("File:"), n.File.Filename, n.File.ContentType)
			}
			if n.URL != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("URL:"), n.URL)
			}
			if n.Task != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Task:"), n.Task.ID)
			}
			if n.Description != nil && *n.Description != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), *n.Description)
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
