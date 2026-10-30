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

package suggestion

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdSuggestion(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suggestion <command>",
		Short: "Manage document redlines",
	}

	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newStartCmd(f))
	cmd.AddCommand(newAcceptCmd(f))
	cmd.AddCommand(newRejectCmd(f))

	return cmd
}

type suggestion struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Author *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"author"`
	CreatedAt string `json:"createdAt"`
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagStatus string
		flagLimit  int
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "list <document-id>",
		Short: "List suggestions on a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			variables := map[string]any{"id": args[0]}
			if flagStatus != "" {
				if err := cmdutil.ValidateEnum("status", flagStatus, []string{"OPEN", "ACCEPTED", "REJECTED"}); err != nil {
					return err
				}
				variables["filter"] = map[string]any{"status": flagStatus}
			}

			client, err := newClient(f)
			if err != nil {
				return err
			}

			suggestions, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[suggestion], error) {
					var resp struct {
						Node *struct {
							Typename    string                     `json:"__typename"`
							Suggestions api.Connection[suggestion] `json:"suggestions"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("document %s not found", args[0])
					}

					if resp.Node.Typename != "Document" {
						return nil, fmt.Errorf("expected Document node, got %s", resp.Node.Typename)
					}

					return &resp.Node.Suggestions, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, suggestions)
			}

			if len(suggestions) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No suggestions found.")
				return nil
			}

			rows := make([][]string, 0, len(suggestions))
			for _, s := range suggestions {
				author := ""
				if s.Author != nil {
					author = s.Author.FullName
				}

				rows = append(rows, []string{
					s.ID,
					s.Status,
					author,
					cmdutil.FormatTime(s.CreatedAt),
				})
			}

			tbl := cmdutil.NewTable("ID", "STATUS", "AUTHOR", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, tbl)

			if totalCount > len(suggestions) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d suggestions\n",
					len(suggestions),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagStatus, "status", "OPEN", "Suggestion status filter")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of suggestions")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

type viewResponse struct {
	Node *struct {
		Typename   string `json:"__typename"`
		ID         string `json:"id"`
		Status     string `json:"status"`
		CreatedAt  string `json:"createdAt"`
		ResolvedAt string `json:"resolvedAt"`
		Author     *struct {
			ID       string `json:"id"`
			FullName string `json:"fullName"`
		} `json:"author"`
		Document *struct {
			ID string `json:"id"`
		} `json:"document"`
	} `json:"node"`
}

func newViewCmd(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <suggestion-id>",
		Short: "View a suggestion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, err := newClient(f)
			if err != nil {
				return err
			}

			data, err := client.Do(viewQuery, map[string]any{"id": args[0]})
			if err != nil {
				return err
			}

			var resp viewResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if resp.Node == nil {
				return fmt.Errorf("suggestion %s not found", args[0])
			}

			if resp.Node.Typename != "DocumentSuggestion" {
				return fmt.Errorf("expected DocumentSuggestion node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			n := resp.Node
			out := f.IOStreams.Out

			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			author := ""
			if n.Author != nil {
				author = n.Author.FullName
			}

			document := ""
			if n.Document != nil {
				document = n.Document.ID
			}

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), n.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Status:"), n.Status)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Author:"), author)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Document:"), document)
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(n.CreatedAt))

			if n.ResolvedAt != "" {
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Resolved:"), cmdutil.FormatTime(n.ResolvedAt))
			}

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newStartCmd(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "start <document-id>",
		Short: "Allocate a suggestion ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMutation(
				f,
				flagOutput,
				startMutation,
				map[string]any{
					"input": map[string]any{"documentId": args[0]},
				},
				func(data json.RawMessage) (string, error) {
					var resp struct {
						StartDocumentSuggestion struct {
							DocumentSuggestionEdge struct {
								Node struct {
									ID string `json:"id"`
								} `json:"node"`
							} `json:"documentSuggestionEdge"`
						} `json:"startDocumentSuggestion"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return "", fmt.Errorf("cannot parse response: %w", err)
					}

					return fmt.Sprintf(
						"Started suggestion %s",
						resp.StartDocumentSuggestion.DocumentSuggestionEdge.Node.ID,
					), nil
				},
			)
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newAcceptCmd(f *cmdutil.Factory) *cobra.Command {
	return newDecisionCmd(f, "accept", "Accept a suggestion", acceptMutation, "acceptDocumentSuggestion", "Accepted")
}

func newRejectCmd(f *cmdutil.Factory) *cobra.Command {
	return newDecisionCmd(f, "reject", "Reject a suggestion", rejectMutation, "rejectDocumentSuggestion", "Rejected")
}

func newDecisionCmd(
	f *cmdutil.Factory,
	use string,
	short string,
	query string,
	field string,
	past string,
) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   use + " <suggestion-id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMutation(
				f,
				flagOutput,
				query,
				map[string]any{
					"input": map[string]any{"suggestionId": args[0]},
				},
				func(data json.RawMessage) (string, error) {
					var resp map[string]struct {
						DocumentSuggestion struct {
							ID string `json:"id"`
						} `json:"documentSuggestion"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return "", fmt.Errorf("cannot parse response: %w", err)
					}

					return fmt.Sprintf("%s suggestion %s", past, resp[field].DocumentSuggestion.ID), nil
				},
			)
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func runMutation(
	f *cmdutil.Factory,
	flagOutput *string,
	query string,
	variables map[string]any,
	confirm func(json.RawMessage) (string, error),
) error {
	if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
		return err
	}

	client, err := newClient(f)
	if err != nil {
		return err
	}

	data, err := client.Do(query, variables)
	if err != nil {
		return err
	}

	if *flagOutput == cmdutil.OutputJSON {
		return cmdutil.PrintJSON(f.IOStreams.Out, json.RawMessage(data))
	}

	msg, err := confirm(data)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintln(f.IOStreams.Out, msg)

	return nil
}

func newClient(f *cmdutil.Factory) (*api.Client, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}

	host, hc, err := cfg.DefaultHost()
	if err != nil {
		return nil, err
	}

	return api.NewClient(
		host,
		hc.Token,
		"/api/console/v1/graphql",
		cfg.HTTPTimeoutDuration(),
		cmdutil.TokenRefreshOption(cfg, host, hc),
	), nil
}

const listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $filter: DocumentSuggestionFilter) {
  node(id: $id) {
    __typename
    ... on Document {
      suggestions(first: $first, after: $after, filter: $filter) {
        totalCount
        edges { node { id status author { id fullName } createdAt } }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
`

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on DocumentSuggestion {
      id
      status
      createdAt
      resolvedAt
      author { id fullName }
      document { id }
    }
  }
}
`

const startMutation = `
mutation($input: StartDocumentSuggestionInput!) {
  startDocumentSuggestion(input: $input) {
    documentSuggestionEdge { node { id status } }
  }
}
`

const acceptMutation = `
mutation($input: AcceptDocumentSuggestionInput!) {
  acceptDocumentSuggestion(input: $input) {
    documentSuggestion { id status resolvedAt }
  }
}
`

const rejectMutation = `
mutation($input: RejectDocumentSuggestionInput!) {
  rejectDocumentSuggestion(input: $input) {
    documentSuggestion { id status resolvedAt }
  }
}
`
