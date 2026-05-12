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

package list

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DocumentOrder, $filter: DocumentFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      documents(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
        totalCount
        edges {
          node {
            id
            status
            trustCenterVisibility
            versions(first: 1) {
              edges {
                node {
                  title
                  documentType
                  classification
                }
              }
            }
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
`

type document struct {
	ID                    string `json:"id"`
	Status                string `json:"status"`
	TrustCenterVisibility string `json:"trustCenterVisibility"`
	Versions              struct {
		Edges []struct {
			Node struct {
				Title          string `json:"title"`
				DocumentType   string `json:"documentType"`
				Classification string `json:"classification"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"versions"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg            string
		flagLimit          int
		flagOrderBy        string
		flagOrderDir       string
		flagQuery          string
		flagWriteMode      string
		flagDocumentType   string
		flagClassification string
		flagStatus         string
		flagOutput         *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List documents in an organization",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"TITLE", "CREATED_AT", "UPDATED_AT", "DOCUMENT_TYPE"},
				); err != nil {
					return err
				}
				if err := cmdutil.ValidateEnum(
					"order-direction",
					flagOrderDir,
					[]string{"ASC", "DESC"},
				); err != nil {
					return err
				}
				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			filter := map[string]any{}
			if flagQuery != "" {
				filter["query"] = flagQuery
			}
			if flagWriteMode != "" {
				if err := cmdutil.ValidateEnum(
					"write-mode",
					flagWriteMode,
					[]string{"AUTHORED", "GENERATED"},
				); err != nil {
					return err
				}
				filter["writeModes"] = []string{flagWriteMode}
			}
			if flagDocumentType != "" {
				if err := cmdutil.ValidateEnum(
					"document-type",
					flagDocumentType,
					[]string{"OTHER", "GOVERNANCE", "POLICY", "PROCEDURE", "PLAN", "REGISTER", "RECORD", "REPORT", "TEMPLATE", "STATEMENT_OF_APPLICABILITY"},
				); err != nil {
					return err
				}
				filter["documentTypes"] = []string{flagDocumentType}
			}
			if flagClassification != "" {
				if err := cmdutil.ValidateEnum(
					"classification",
					flagClassification,
					[]string{"PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"},
				); err != nil {
					return err
				}
				filter["classifications"] = []string{flagClassification}
			}
			if flagStatus != "" {
				if err := cmdutil.ValidateEnum(
					"status",
					flagStatus,
					[]string{"ACTIVE", "ARCHIVED"},
				); err != nil {
					return err
				}
				filter["status"] = []string{flagStatus}
			}
			if len(filter) > 0 {
				variables["filter"] = filter
			}

			documents, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[document], error) {
					var resp struct {
						Node *struct {
							Typename  string                   `json:"__typename"`
							Documents api.Connection[document] `json:"documents"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}
					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}
					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}
					return &resp.Node.Documents, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, documents)
			}

			if len(documents) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No documents found.")
				return nil
			}

			rows := make([][]string, 0, len(documents))
			for _, doc := range documents {
				title := ""
				docType := ""
				classification := ""
				if len(doc.Versions.Edges) > 0 {
					v := doc.Versions.Edges[0].Node
					title = v.Title
					docType = v.DocumentType
					classification = v.Classification
				}
				rows = append(rows, []string{
					doc.ID,
					title,
					docType,
					classification,
					doc.Status,
				})
			}

			t := cmdutil.NewTable("ID", "TITLE", "TYPE", "CLASSIFICATION", "STATUS").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(documents) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d documents\n",
					len(documents),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of documents to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (TITLE, CREATED_AT, UPDATED_AT, DOCUMENT_TYPE)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().StringVarP(&flagQuery, "query", "q", "", "Search query")
	cmd.Flags().StringVar(&flagWriteMode, "write-mode", "", "Filter by write mode (AUTHORED, GENERATED)")
	cmd.Flags().StringVar(&flagDocumentType, "document-type", "", "Filter by document type (OTHER, GOVERNANCE, POLICY, PROCEDURE, PLAN, REGISTER, RECORD, REPORT, TEMPLATE, STATEMENT_OF_APPLICABILITY)")
	cmd.Flags().StringVar(&flagClassification, "classification", "", "Filter by classification (PUBLIC, INTERNAL, CONFIDENTIAL, SECRET)")
	cmd.Flags().StringVar(&flagStatus, "status", "", "Filter by status (ACTIVE, ARCHIVED)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
