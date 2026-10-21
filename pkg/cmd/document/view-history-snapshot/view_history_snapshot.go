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

package viewhistorysnapshot

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!, $revisionId: ID!) {
  node(id: $id) {
    __typename
    ... on Document {
      revision(id: $revisionId) {
        snapshot {
          content {
            prosemirror
          }
          highlights {
            kind
            from
            to
            text
          }
        }
      }
    }
  }
}
`

type snapshotResponse struct {
	Node *struct {
		Typename string `json:"__typename"`
		Revision *struct {
			Snapshot struct {
				Content struct {
					Prosemirror string `json:"prosemirror"`
				} `json:"content"`
				Highlights []struct {
					Kind string `json:"kind"`
					From int    `json:"from"`
					To   int    `json:"to"`
					Text string `json:"text"`
				} `json:"highlights"`
			} `json:"snapshot"`
		} `json:"revision"`
	} `json:"node"`
}

func NewCmdViewHistorySnapshot(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view-history-snapshot <document-id> <revision-id>",
		Short: "View the content and edit highlights of a revision",
		Example: `  # View the snapshot of one revision
  prb document view-history-snapshot <document-id> <revision-id>`,
		Args: cobra.ExactArgs(2),
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

			data, err := client.Do(viewQuery, map[string]any{
				"id":         args[0],
				"revisionId": args[1],
			})
			if err != nil {
				return err
			}

			var resp snapshotResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("document %s not found", args[0])
			}

			if resp.Node.Typename != "Document" {
				return fmt.Errorf("expected Document node, got %s", resp.Node.Typename)
			}

			if resp.Node.Revision == nil {
				return fmt.Errorf("revision %s not found", args[1])
			}

			snapshot := resp.Node.Revision.Snapshot

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, snapshot)
			}

			_, _ = fmt.Fprintln(f.IOStreams.Out, snapshot.Content.Prosemirror)

			if len(snapshot.Highlights) > 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "\nHighlights:")
				for _, highlight := range snapshot.Highlights {
					_, _ = fmt.Fprintf(
						f.IOStreams.Out,
						"  %s [%d:%d] %q\n",
						highlight.Kind,
						highlight.From,
						highlight.To,
						highlight.Text,
					)
				}
			}

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
