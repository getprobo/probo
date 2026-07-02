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
    ... on Document {
      id
      status
      trustCenterVisibility
      currentPublishedMajor
      currentPublishedMinor
      versions(first: 1) {
        edges {
          node {
            title
            documentType
            classification
            major
            minor
            status
          }
        }
      }
      createdAt
      updatedAt
    }
  }
}
`

type viewResponse struct {
	Node *struct {
		Typename              string `json:"__typename"`
		ID                    string `json:"id"`
		Status                string `json:"status"`
		TrustCenterVisibility string `json:"trustCenterVisibility"`
		CurrentPublishedMajor *int   `json:"currentPublishedMajor"`
		CurrentPublishedMinor *int   `json:"currentPublishedMinor"`
		Versions              struct {
			Edges []struct {
				Node struct {
					Title          string `json:"title"`
					DocumentType   string `json:"documentType"`
					Classification string `json:"classification"`
					Major          int    `json:"major"`
					Minor          int    `json:"minor"`
					Status         string `json:"status"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"versions"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"node"`
}

func NewCmdView(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <id>",
		Short: "View a document",
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
				return fmt.Errorf("document %s not found", args[0])
			}

			if resp.Node.Typename != "Document" {
				return fmt.Errorf("expected Document node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			doc := resp.Node
			out := f.IOStreams.Out

			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			title := doc.ID
			if len(doc.Versions.Edges) > 0 {
				title = doc.Versions.Edges[0].Node.Title
			}

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(title))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), doc.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Status:"), doc.Status)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Visibility:"), doc.TrustCenterVisibility)

			if len(doc.Versions.Edges) > 0 {
				v := doc.Versions.Edges[0].Node
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Type:"), v.DocumentType)
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Classification:"), v.Classification)
				_, _ = fmt.Fprintf(out, "%s%d.%d (%s)\n", label.Render("Latest Version:"), v.Major, v.Minor, v.Status)
			}

			if doc.CurrentPublishedMajor != nil && doc.CurrentPublishedMinor != nil {
				_, _ = fmt.Fprintf(out, "%s%d.%d\n", label.Render("Published Version:"), *doc.CurrentPublishedMajor, *doc.CurrentPublishedMinor)
			}

			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(doc.CreatedAt))
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Updated:"), cmdutil.FormatTime(doc.UpdatedAt))

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
