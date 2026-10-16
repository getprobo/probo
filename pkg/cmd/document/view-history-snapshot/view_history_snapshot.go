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
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const viewQuery = `
query($id: ID!, $heads: [String!]!, $previousHeads: [String!]) {
  node(id: $id) {
    __typename
    ... on DocumentVersion {
      collaborationHistorySnapshot(heads: $heads, previousHeads: $previousHeads) {
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
`

type snapshotResponse struct {
	Node *struct {
		Typename                     string `json:"__typename"`
		CollaborationHistorySnapshot *struct {
			Content struct {
				Prosemirror string `json:"prosemirror"`
			} `json:"content"`
			Highlights []struct {
				Kind string `json:"kind"`
				From int    `json:"from"`
				To   int    `json:"to"`
				Text string `json:"text"`
			} `json:"highlights"`
		} `json:"collaborationHistorySnapshot"`
	} `json:"node"`
}

func parseHeadsFlag(value string) ([]string, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("heads must not be empty")
	}

	parts := strings.Split(value, ",")
	heads := make([]string, 0, len(parts))
	for _, part := range parts {
		head := strings.TrimSpace(part)
		if head == "" {
			continue
		}

		heads = append(heads, head)
	}

	if len(heads) == 0 {
		return nil, fmt.Errorf("heads must not be empty")
	}

	return heads, nil
}

func NewCmdViewHistorySnapshot(f *cmdutil.Factory) *cobra.Command {
	var (
		flagHeads         string
		flagPreviousHeads string
		flagOutput        *string
	)

	cmd := &cobra.Command{
		Use:   "view-history-snapshot <document-version-id>",
		Short: "View document content and edit highlights at collaboration history heads",
		Example: `  # View snapshot at a session's heads
  prb document view-history-snapshot <document-version-id> --heads <hex>[,<hex>...]

  # Diff against previous heads
  prb document view-history-snapshot <document-version-id> --heads <hex> --previous-heads <hex>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			heads, err := parseHeadsFlag(flagHeads)
			if err != nil {
				return err
			}

			var previousHeads []string
			if flagPreviousHeads != "" {
				previousHeads, err = parseHeadsFlag(flagPreviousHeads)
				if err != nil {
					return fmt.Errorf("cannot parse previous-heads: %w", err)
				}
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

			variables := map[string]any{
				"id":    args[0],
				"heads": heads,
			}
			if previousHeads != nil {
				variables["previousHeads"] = previousHeads
			}

			data, err := client.Do(viewQuery, variables)
			if err != nil {
				return err
			}

			var resp snapshotResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("document version %s not found", args[0])
			}

			if resp.Node.Typename != "DocumentVersion" {
				return fmt.Errorf("expected DocumentVersion node, got %s", resp.Node.Typename)
			}

			snapshot := resp.Node.CollaborationHistorySnapshot
			if snapshot == nil {
				return fmt.Errorf("collaboration history snapshot not available")
			}

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

	cmd.Flags().StringVar(&flagHeads, "heads", "", "Automerge heads (comma-separated lowercase hex)")
	cmd.Flags().StringVar(&flagPreviousHeads, "previous-heads", "", "Previous Automerge heads for the diff (comma-separated)")
	_ = cmd.MarkFlagRequired("heads")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
