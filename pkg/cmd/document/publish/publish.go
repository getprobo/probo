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

package publish

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const publishMutation = `
mutation($input: PublishDocumentInput!) {
  publishDocument(input: $input) {
    documentVersion {
      id
      title
      major
      minor
      status
    }
    approvalQuorum {
      id
      status
    }
  }
}
`

type publishResponse struct {
	PublishDocument struct {
		DocumentVersion struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Major  int    `json:"major"`
			Minor  int    `json:"minor"`
			Status string `json:"status"`
		} `json:"documentVersion"`
		ApprovalQuorum *struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"approvalQuorum"`
	} `json:"publishDocument"`
}

func NewCmdPublish(f *cmdutil.Factory) *cobra.Command {
	var (
		flagMinor     bool
		flagApprover  []string
		flagChangelog string
	)

	cmd := &cobra.Command{
		Use:   "publish <document-id>",
		Short: "Publish a document",
		Long: `Publish the latest draft of a document.

By default, the draft is published as a new major version. Pass --minor to
publish as a minor version (the document must already have a major version).
When --approver is set (one or more profile IDs), an approval is requested
instead of publishing immediately. Approvers are ignored with --minor.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagChangelog == "" {
				return fmt.Errorf("--changelog is required")
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

			input := map[string]any{
				"documentId": args[0],
				"minor":      flagMinor,
				"changelog":  flagChangelog,
			}

			if len(flagApprover) > 0 {
				input["approverIds"] = flagApprover
			}

			data, err := client.Do(
				publishMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp publishResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			v := resp.PublishDocument.DocumentVersion
			if resp.PublishDocument.ApprovalQuorum != nil {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"Requested approval for %s (%s v%d.%d, status %s)\n",
					v.ID,
					v.Title,
					v.Major,
					v.Minor,
					v.Status,
				)
				return nil
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Published %s (%s v%d.%d)\n",
				v.ID,
				v.Title,
				v.Major,
				v.Minor,
			)

			return nil
		},
	}

	cmd.Flags().BoolVar(&flagMinor, "minor", false, "Publish as a minor version (no approval flow)")
	cmd.Flags().StringArrayVar(&flagApprover, "approver", nil, "Approver profile ID (can be repeated; ignored with --minor)")
	cmd.Flags().StringVar(&flagChangelog, "changelog", "", "Changelog for this version (required)")

	return cmd
}
