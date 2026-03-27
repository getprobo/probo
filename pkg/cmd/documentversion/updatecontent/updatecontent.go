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

package updatecontent

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/prosemirror"
)

const updateContentMutation = `
mutation($input: UpdateDocumentVersionContentInput!) {
  updateDocumentVersionContent(input: $input) {
    content
  }
}
`

type updateContentResponse struct {
	UpdateDocumentVersionContent struct {
		Content string `json:"content"`
	} `json:"updateDocumentVersionContent"`
}

func NewCmdUpdateContent(f *cmdutil.Factory) *cobra.Command {
	var (
		flagID           string
		flagContent      string
		flagFromMarkdown string
	)

	cmd := &cobra.Command{
		Use:   "update-content",
		Short: "Update document version content",
		Example: `  # Update with ProseMirror JSON
  prb document-version update-content --id <version-id> --content '{"type":"doc",...}'

  # Update from markdown
  prb document-version update-content --id <version-id> --from-markdown "# Hello"

  # Update from stdin
  cat content.json | prb document-version update-content --id <version-id>`,
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

			var content string
			switch {
			case flagFromMarkdown != "":
				doc, err := prosemirror.ParseMarkdown(flagFromMarkdown)
				if err != nil {
					return err
				}
				out, err := json.Marshal(doc)
				if err != nil {
					return fmt.Errorf("cannot marshal prosemirror document: %w", err)
				}
				content = string(out)
			case flagContent != "":
				content = flagContent
			default:
				data, err := io.ReadAll(f.IOStreams.In)
				if err != nil {
					return fmt.Errorf("cannot read from stdin: %w", err)
				}
				content = string(data)
			}

			input := map[string]any{
				"id":      flagID,
				"content": content,
			}

			data, err := client.Do(
				updateContentMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateContentResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated document version content %s\n",
				flagID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagID, "id", "", "Document version ID (required)")
	cmd.Flags().StringVar(&flagContent, "content", "", "ProseMirror JSON content")
	cmd.Flags().StringVar(
		&flagFromMarkdown,
		"from-markdown",
		"",
		"Markdown content to convert and upload",
	)

	_ = cmd.MarkFlagRequired("id")

	return cmd
}
