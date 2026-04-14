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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation($input: UpdateDocumentInput!) {
  updateDocument(input: $input) {
    document {
      id
      trustCenterVisibility
    }
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
	UpdateDocument struct {
		Document struct {
			ID                    string `json:"id"`
			TrustCenterVisibility string `json:"trustCenterVisibility"`
		} `json:"document"`
		DocumentVersion *struct {
			ID             string `json:"id"`
			Title          string `json:"title"`
			Major          int    `json:"major"`
			Minor          int    `json:"minor"`
			Status         string `json:"status"`
			DocumentType   string `json:"documentType"`
			Classification string `json:"classification"`
		} `json:"documentVersion"`
	} `json:"updateDocument"`
}

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagTitle                 string
		flagContent               string
		flagDocumentType          string
		flagClassification        string
		flagTrustCenterVisibility string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a document",
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
				"id": args[0],
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
			if cmd.Flags().Changed("trust-center-visibility") {
				if err := cmdutil.ValidateEnum(
					"trust-center-visibility",
					flagTrustCenterVisibility,
					[]string{"NONE", "PRIVATE", "PUBLIC"},
				); err != nil {
					return err
				}
				input["trustCenterVisibility"] = flagTrustCenterVisibility
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

			doc := resp.UpdateDocument.Document
			if v := resp.UpdateDocument.DocumentVersion; v != nil {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"Updated document %s (%s v%d.%d)\n",
					doc.ID,
					v.Title,
					v.Major,
					v.Minor,
				)
			} else {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"Updated document %s\n",
					doc.ID,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagTitle, "title", "", "Document title")
	cmd.Flags().StringVar(&flagContent, "content", "", "Document content")
	cmd.Flags().StringVar(&flagDocumentType, "document-type", "", "Document type: OTHER, GOVERNANCE, POLICY, PROCEDURE, PLAN, REGISTER, RECORD, REPORT, TEMPLATE")
	cmd.Flags().StringVar(&flagClassification, "classification", "", "Classification: PUBLIC, INTERNAL, CONFIDENTIAL, SECRET")
	cmd.Flags().StringVar(&flagTrustCenterVisibility, "trust-center-visibility", "", "Trust center visibility: NONE, PRIVATE, PUBLIC")

	return cmd
}
