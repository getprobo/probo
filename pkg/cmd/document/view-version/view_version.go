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

package viewversion

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
    ... on DocumentVersion {
      id
      title
      major
      minor
      status
      content
      changelog
      documentType
      classification
      publishedAt
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
		Title          string  `json:"title"`
		Major          int     `json:"major"`
		Minor          int     `json:"minor"`
		Status         string  `json:"status"`
		Content        string  `json:"content"`
		Changelog      string  `json:"changelog"`
		DocumentType   string  `json:"documentType"`
		Classification string  `json:"classification"`
		PublishedAt    *string `json:"publishedAt"`
		CreatedAt      string  `json:"createdAt"`
		UpdatedAt      string  `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdViewVersion(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view-version <id>",
		Short: "View a document version",
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
				return fmt.Errorf("document version %s not found", args[0])
			}

			if resp.Node.Typename != "DocumentVersion" {
				return fmt.Errorf("expected DocumentVersion node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			v := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(v.Title))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), v.ID)
			_, _ = fmt.Fprintf(out, "%s%d.%d\n", label.Render("Version:"), v.Major, v.Minor)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Status:"), v.Status)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Type:"), v.DocumentType)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Classification:"), v.Classification)

			if v.Changelog != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Changelog:"), v.Changelog)
			}

			_, _ = fmt.Fprintln(out)

			if v.PublishedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Published:"), cmdutil.FormatTime(*v.PublishedAt))
			}
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(v.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(v.UpdatedAt))

			if v.Content != "" {
				_, _ = fmt.Fprintf(out, "\n%s\n%s\n", bold.Render("Content:"), v.Content)
			}

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
