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

package publishmajor

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const publishMajorMutation = `
mutation($input: PublishMajorDocumentVersionInput!) {
  publishMajorDocumentVersion(input: $input) {
    documentVersion {
      id
      title
      major
      minor
      status
    }
  }
}
`

type publishResponse struct {
	PublishMajorDocumentVersion struct {
		DocumentVersion struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Major  int    `json:"major"`
			Minor  int    `json:"minor"`
			Status string `json:"status"`
		} `json:"documentVersion"`
	} `json:"publishMajorDocumentVersion"`
}

func NewCmdPublishMajor(f *cmdutil.Factory) *cobra.Command {
	var flagChangelog string

	cmd := &cobra.Command{
		Use:   "publish-major <document-id>",
		Short: "Publish a major version of a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
			}

			if flagChangelog != "" {
				input["changelog"] = flagChangelog
			}

			data, err := client.Do(
				publishMajorMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp publishResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			v := resp.PublishMajorDocumentVersion.DocumentVersion
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Published major version %s (%s v%d.%d)\n",
				v.ID,
				v.Title,
				v.Major,
				v.Minor,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagChangelog, "changelog", "", "Changelog for this version")

	return cmd
}
