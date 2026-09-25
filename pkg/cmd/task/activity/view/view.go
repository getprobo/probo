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
    ... on TaskActivity {
      id
      activityType
      field
      oldValue
      newValue
      createdAt
      actor {
        id
        fullName
      }
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename     string  `json:"__typename"`
		ID           string  `json:"id"`
		ActivityType string  `json:"activityType"`
		Field        *string `json:"field"`
		OldValue     *string `json:"oldValue"`
		NewValue     *string `json:"newValue"`
		CreatedAt    string  `json:"createdAt"`
		Actor        *struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"actor"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a task activity",
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
				return fmt.Errorf("activity %s not found", args[0])
			}

			if resp.Node.Typename != "TaskActivity" {
				return fmt.Errorf("expected TaskActivity node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			e := resp.Node
			out := f.IOStreams.Out
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			actorName := ""
			if e.Actor != nil {
				actorName = e.Actor.FullName
			}

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), e.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Type:"), e.ActivityType)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Field:"), stringOrEmpty(e.Field))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Old value:"), stringOrEmpty(e.OldValue))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("New value:"), stringOrEmpty(e.NewValue))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Actor:"), actorName)
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(e.CreatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func stringOrEmpty(v *string) string {
	if v == nil {
		return ""
	}

	return *v
}
