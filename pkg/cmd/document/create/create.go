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

package create

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateDocumentInput!) {
  createDocument(input: $input) {
    documentEdge {
      node {
        id
      }
    }
    documentVersionEdge {
      node {
        title
        documentType
        classification
      }
    }
  }
}
`

type createResponse struct {
	CreateDocument struct {
		DocumentEdge struct {
			Node struct {
				ID string `json:"id"`
			} `json:"node"`
		} `json:"documentEdge"`
		DocumentVersionEdge struct {
			Node struct {
				Title          string `json:"title"`
				DocumentType   string `json:"documentType"`
				Classification string `json:"classification"`
			} `json:"node"`
		} `json:"documentVersionEdge"`
	} `json:"createDocument"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg                   string
		flagTitle                 string
		flagContent               string
		flagDocumentType          string
		flagClassification        string
		flagTrustCenterVisibility string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new document",
		Example: `  # Create a policy document
  prb document create --title "Information Security Policy" --document-type POLICY --classification INTERNAL

  # Create a document interactively
  prb document create`,
		Args: cobra.NoArgs,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			if f.IOStreams.IsInteractive() {
				if flagTitle == "" {
					err := huh.NewInput().
						Title("Document title").
						Value(&flagTitle).
						Run()
					if err != nil {
						return err
					}
				}

				if flagDocumentType == "" {
					err := huh.NewSelect[string]().
						Title("Document type").
						Options(
							huh.NewOption("Policy", "POLICY"),
							huh.NewOption("Procedure", "PROCEDURE"),
							huh.NewOption("Governance", "GOVERNANCE"),
							huh.NewOption("Plan", "PLAN"),
							huh.NewOption("Register", "REGISTER"),
							huh.NewOption("Record", "RECORD"),
							huh.NewOption("Report", "REPORT"),
							huh.NewOption("Template", "TEMPLATE"),
							huh.NewOption("Other", "OTHER"),
						).
						Value(&flagDocumentType).
						Run()
					if err != nil {
						return err
					}
				}

				if flagClassification == "" {
					err := huh.NewSelect[string]().
						Title("Classification").
						Options(
							huh.NewOption("Public", "PUBLIC"),
							huh.NewOption("Internal", "INTERNAL"),
							huh.NewOption("Confidential", "CONFIDENTIAL"),
							huh.NewOption("Secret", "SECRET"),
						).
						Value(&flagClassification).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagTitle == "" {
				return fmt.Errorf("title is required; pass --title or run interactively")
			}
			if flagDocumentType == "" {
				return fmt.Errorf("document type is required; pass --document-type or run interactively")
			}
			if flagClassification == "" {
				return fmt.Errorf("classification is required; pass --classification or run interactively")
			}

			if err := cmdutil.ValidateEnum(
				"document-type",
				flagDocumentType,
				[]string{"OTHER", "GOVERNANCE", "POLICY", "PROCEDURE", "PLAN", "REGISTER", "RECORD", "REPORT", "TEMPLATE"},
			); err != nil {
				return err
			}

			if err := cmdutil.ValidateEnum(
				"classification",
				flagClassification,
				[]string{"PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"},
			); err != nil {
				return err
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"title":          flagTitle,
				"documentType":   flagDocumentType,
				"classification": flagClassification,
			}

			if flagContent != "" {
				input["content"] = flagContent
			}

			if flagTrustCenterVisibility != "" {
				if err := cmdutil.ValidateEnum(
					"trust-center-visibility",
					flagTrustCenterVisibility,
					[]string{"NONE", "PRIVATE", "PUBLIC"},
				); err != nil {
					return err
				}
				input["trustCenterVisibility"] = flagTrustCenterVisibility
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			doc := resp.CreateDocument.DocumentEdge.Node
			ver := resp.CreateDocument.DocumentVersionEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created document %s (%s)\n",
				doc.ID,
				ver.Title,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagTitle, "title", "", "Document title")
	cmd.Flags().StringVar(&flagContent, "content", "", "Document content")
	cmd.Flags().StringVar(&flagDocumentType, "document-type", "", "Document type: OTHER, GOVERNANCE, POLICY, PROCEDURE, PLAN, REGISTER, RECORD, REPORT, TEMPLATE")
	cmd.Flags().StringVar(&flagClassification, "classification", "", "Classification: PUBLIC, INTERNAL, CONFIDENTIAL, SECRET")
	cmd.Flags().StringVar(&flagTrustCenterVisibility, "trust-center-visibility", "", "Trust center visibility: NONE, PRIVATE, PUBLIC")

	return cmd
}
