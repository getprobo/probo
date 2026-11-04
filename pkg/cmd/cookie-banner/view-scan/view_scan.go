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

package viewscan

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewScanQuery = `
query CookieBannerViewScan($id: ID!) {
  node(id: $id) {
    __typename
    ... on CookieBannerScan {
      id
      state
      failureReason
      truncated
      pagesScanned
      cookiesFound
      storageFound
      resourcesFound
      startedAt
      completedAt
      createdAt
      updatedAt
      checkpoint {
        visitedTemplates
        queue { url }
        coveredHosts
        pagesScanned
        elapsedMs
        truncated
      }
    }
  }
}
`

type scanNode struct {
	Typename       string  `json:"__typename"`
	ID             string  `json:"id"`
	State          string  `json:"state"`
	FailureReason  *string `json:"failureReason"`
	Truncated      bool    `json:"truncated"`
	PagesScanned   int     `json:"pagesScanned"`
	CookiesFound   int     `json:"cookiesFound"`
	StorageFound   int     `json:"storageFound"`
	ResourcesFound int     `json:"resourcesFound"`
	StartedAt      *string `json:"startedAt"`
	CompletedAt    *string `json:"completedAt"`
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
	Checkpoint     *struct {
		VisitedTemplates []string `json:"visitedTemplates"`
		Queue            []struct {
			URL string `json:"url"`
		} `json:"queue"`
		CoveredHosts []string `json:"coveredHosts"`
		PagesScanned int      `json:"pagesScanned"`
		ElapsedMs    int      `json:"elapsedMs"`
		Truncated    bool     `json:"truncated"`
	} `json:"checkpoint"`
}

func NewCmdViewScan(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view-scan <id>",
		Short: "View a cookie banner scan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			data, err := client.Do(viewScanQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp struct {
				Node *scanNode `json:"node"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return err
			}

			if resp.Node == nil || resp.Node.Typename != "CookieBannerScan" {
				return fmt.Errorf("cookie banner scan %s not found", args[0])
			}

			if *flagOutput == "json" {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			scan := resp.Node
			out := f.IOStreams.Out
			bold := lipgloss.NewStyle().Bold(true)
			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(18)

			_, _ = fmt.Fprintf(out, "%s\n\n", bold.Render(scan.ID))

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("State:"), scan.State)
			if scan.FailureReason != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Failure:"), *scan.FailureReason)
			}

			_, _ = fmt.Fprintf(out, "%s%t\n", label.Render("Truncated:"), scan.Truncated)
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Pages:"), scan.PagesScanned)
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Cookies:"), scan.CookiesFound)
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Storage:"), scan.StorageFound)
			_, _ = fmt.Fprintf(out, "%s%d\n", label.Render("Resources:"), scan.ResourcesFound)

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Requested:"), cmdutil.FormatTime(scan.CreatedAt))
			if scan.StartedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Started:"), cmdutil.FormatTime(*scan.StartedAt))
			}

			if scan.CompletedAt != nil {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Finished:"), cmdutil.FormatTime(*scan.CompletedAt))
			}

			if scan.Checkpoint != nil {
				_, _ = fmt.Fprintf(out, "%s%d visited, %d queued, %d hosts\n", label.Render("Checkpoint:"), len(scan.Checkpoint.VisitedTemplates), len(scan.Checkpoint.Queue), len(scan.Checkpoint.CoveredHosts))
			}

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
