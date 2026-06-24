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
    ... on TrackerResource {
      id
      type
      origin
      path
      displayName
      description
      excluded
      lastDetectedAt
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename       string  `json:"__typename"`
		ID             string  `json:"id"`
		Type           string  `json:"type"`
		Origin         string  `json:"origin"`
		Path           string  `json:"path"`
		DisplayName    string  `json:"displayName"`
		Description    string  `json:"description"`
		Excluded       bool    `json:"excluded"`
		LastDetectedAt *string `json:"lastDetectedAt"`
		CreatedAt      string  `json:"createdAt"`
		UpdatedAt      string  `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a tracker resource",
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

			data, err := client.Do(viewQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil || resp.Node.Typename != "TrackerResource" {
				return fmt.Errorf("tracker resource %s not found", args[0])
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			v := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(v.DisplayName))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), v.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Type:"), v.Type)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Origin:"), v.Origin)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Path:"), v.Path)

			_, _ = fmt.Fprintf(out, "%s%v\n", label.Render("Excluded:"), v.Excluded)
			if v.Description != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), v.Description)
			}

			if v.LastDetectedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Last Detected:"), cmdutil.FormatTime(*v.LastDetectedAt))
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(v.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(v.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
