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

package updateversion

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateDocumentVersionInput!) {
  updateDocumentVersion(input: $input) {
    documentVersion {
      id
      title
      major
      minor
      status
      documentType
      classification
    }
  }
}
`

type updateResponse struct {
	UpdateDocumentVersion struct {
		DocumentVersion struct {
			ID             string `json:"id"`
			Title          string `json:"title"`
			Major          int    `json:"major"`
			Minor          int    `json:"minor"`
			Status         string `json:"status"`
			DocumentType   string `json:"documentType"`
			Classification string `json:"classification"`
		} `json:"documentVersion"`
	} `json:"updateDocumentVersion"`
}

func NewCmdUpdateVersion(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTitle          string
		flagContent        string
		flagDocumentType   string
		flagClassification string
	)

	cmd := &cobra.Command{
		Use:   "update-version <document-version-id>",
		Short: "Update a document version",
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
				"documentVersionId": args[0],
			}

			if cmd.Flags().Changed("title") {
				input["title"] = flagTitle
			}
			if cmd.Flags().Changed("content") {
				input["content"] = flagContent
			}
			if cmd.Flags().Changed("document-type") {
				if err := cmdutil.ValidateEnum(
					"document-type",
					flagDocumentType,
					[]string{"OTHER", "GOVERNANCE", "POLICY", "PROCEDURE", "PLAN", "REGISTER", "RECORD", "REPORT", "TEMPLATE"},
				); err != nil {
					return err
				}
				input["documentType"] = flagDocumentType
			}
			if cmd.Flags().Changed("classification") {
				if err := cmdutil.ValidateEnum(
					"classification",
					flagClassification,
					[]string{"PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"},
				); err != nil {
					return err
				}
				input["classification"] = flagClassification
			}

			if len(input) == 1 {
				return fmt.Errorf("at least one field must be specified for update")
			}

			data, err := client.Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp updateResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			v := resp.UpdateDocumentVersion.DocumentVersion
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated version %s (%s v%d.%d)\n",
				v.ID,
				v.Title,
				v.Major,
				v.Minor,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTitle, "title", "", "Version title")
	cmd.Flags().StringVar(&flagContent, "content", "", "Version content")
	cmd.Flags().StringVar(&flagDocumentType, "document-type", "", "Document type: OTHER, GOVERNANCE, POLICY, PROCEDURE, PLAN, REGISTER, RECORD, REPORT, TEMPLATE")
	cmd.Flags().StringVar(&flagClassification, "classification", "", "Classification: PUBLIC, INTERNAL, CONFIDENTIAL, SECRET")

	return cmd
}
