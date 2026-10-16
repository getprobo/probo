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

package restorehistory

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const restoreMutation = `
mutation($input: RestoreDocumentCollaborationInput!) {
  restoreDocumentCollaboration(input: $input) {
    documentVersion {
      id
      title
      major
      minor
      status
      updatedAt
    }
  }
}
`

type restoreResponse struct {
	RestoreDocumentCollaboration *struct {
		DocumentVersion struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			Major     int    `json:"major"`
			Minor     int    `json:"minor"`
			Status    string `json:"status"`
			UpdatedAt string `json:"updatedAt"`
		} `json:"documentVersion"`
	} `json:"restoreDocumentCollaboration"`
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

func NewCmdRestoreHistory(f *cmdutil.Factory) *cobra.Command {
	var (
		flagHeads  string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "restore-history <document-version-id>",
		Short: "Restore a draft document version to a collaboration history point",
		Example: `  # Restore to a session's heads
  prb document restore-history <document-version-id> --heads <hex>[,<hex>...]`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			heads, err := parseHeadsFlag(flagHeads)
			if err != nil {
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
				restoreMutation,
				map[string]any{
					"input": map[string]any{
						"documentVersionId": args[0],
						"heads":             heads,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp restoreResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.RestoreDocumentCollaboration == nil {
				return fmt.Errorf("restore document collaboration returned no payload")
			}

			version := resp.RestoreDocumentCollaboration.DocumentVersion

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, version)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Restored document version %s (%s v%d.%d, %s)\n",
				version.ID,
				version.Title,
				version.Major,
				version.Minor,
				version.Status,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagHeads, "heads", "", "Automerge heads to restore (comma-separated lowercase hex)")
	_ = cmd.MarkFlagRequired("heads")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
