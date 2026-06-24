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
    ... on TrackerPattern {
      id
      pattern
      matchType
      trackerType
      displayName
      maxAgeSeconds
      description
      source
      excluded
      lastMatchedAt
      commonTrackerPatternId
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename               string  `json:"__typename"`
		ID                     string  `json:"id"`
		Pattern                string  `json:"pattern"`
		MatchType              string  `json:"matchType"`
		TrackerType            string  `json:"trackerType"`
		DisplayName            string  `json:"displayName"`
		MaxAgeSeconds          *int    `json:"maxAgeSeconds"`
		Description            *string `json:"description"`
		Source                 string  `json:"source"`
		Excluded               bool    `json:"excluded"`
		LastMatchedAt          *string `json:"lastMatchedAt"`
		CommonTrackerPatternID *string `json:"commonTrackerPatternId"`
		CreatedAt              string  `json:"createdAt"`
		UpdatedAt              string  `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a tracker pattern",
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

			if resp.Node == nil || resp.Node.Typename != "TrackerPattern" {
				return fmt.Errorf("tracker pattern %s not found", args[0])
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
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Pattern:"), v.Pattern)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Match Type:"), v.MatchType)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Tracker Type:"), v.TrackerType)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Source:"), v.Source)

			_, _ = fmt.Fprintf(out, "%s%v\n", label.Render("Excluded:"), v.Excluded)
			if v.MaxAgeSeconds != nil {
				_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Max Age (seconds):"), *v.MaxAgeSeconds)
			}

			if v.Description != nil && *v.Description != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Description:"), *v.Description)
			}

			if v.LastMatchedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Last Matched:"), cmdutil.FormatTime(*v.LastMatchedAt))
			}

			if v.CommonTrackerPatternID != nil && *v.CommonTrackerPatternID != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Common Pattern:"), *v.CommonTrackerPatternID)
			} else {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Origin:"), "Manual (no catalog link)")
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
