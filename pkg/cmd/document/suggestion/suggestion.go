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
	"strings"

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
	cmd.AddCommand(newAddCmd(f))
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
		flagAfter  string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list <document-id>",
		Aliases: []string{"ls"},
		Short:   "List suggestions on a document",
		Args:    cobra.ExactArgs(1),
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

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			suggestions, totalCount, nextCursor, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				flagAfter,
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

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				suggestions,
				totalCount,
				cmdutil.ListView[suggestion]{
					Empty:      "No suggestions found.",
					Noun:       "suggestions",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "STATUS", "AUTHOR", "CREATED"},
					Row: func(item suggestion) ([]string, error) {
						author := ""
						if item.Author != nil {
							author = item.Author.FullName
						}

						return []string{
							item.ID,
							item.Status,
							author,
							cmdutil.FormatTime(item.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagStatus, "status", "OPEN", "Suggestion status filter")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of suggestions")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
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

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
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
	var (
		flagOutput      *string
		flagQuote       string
		flagReplacement string
		flagComment     string
		flagOccurrence  int
	)

	cmd := &cobra.Command{
		Use:   "start <document-id>",
		Short: "Start a redline, or allocate an ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(flagQuote) == "" &&
				(cmd.Flags().Changed("replacement") ||
					cmd.Flags().Changed("occurrence") ||
					cmd.Flags().Changed("comment")) {
				return fmt.Errorf("quote is required with --replacement, --occurrence, or --comment")
			}

			input := map[string]any{"documentId": args[0]}
			if strings.TrimSpace(flagQuote) != "" {
				input["quote"] = flagQuote
				input["replacement"] = flagReplacement
			}

			if strings.TrimSpace(flagComment) != "" {
				input["comment"] = flagComment
			}

			if flagOccurrence != 0 {
				input["quoteOccurrence"] = flagOccurrence
			}

			return runMutation(
				f,
				flagOutput,
				startMutation,
				map[string]any{
					"input": input,
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
	cmd.Flags().StringVar(&flagQuote, "quote", "", "Text to replace")
	cmd.Flags().StringVar(&flagReplacement, "replacement", "", "Suggested text; empty deletes the quote")
	cmd.Flags().StringVar(&flagComment, "comment", "", "Plain-text note on the redline")
	cmd.Flags().IntVar(&flagOccurrence, "occurrence", 0, "Zero-based occurrence of the quote")

	return cmd
}

func newAddCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOutput      *string
		flagQuote       string
		flagReplacement string
		flagOccurrence  int
	)

	cmd := &cobra.Command{
		Use:   "add <suggestion-id>",
		Short: "Add a redline to a suggestion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(flagQuote) == "" {
				return fmt.Errorf("quote is required")
			}

			if !cmd.Flags().Changed("replacement") {
				return fmt.Errorf("replacement is required; pass an empty value to delete the quote")
			}

			input := map[string]any{
				"suggestionId": args[0],
				"quote":        flagQuote,
				"replacement":  flagReplacement,
			}
			if flagOccurrence != 0 {
				input["quoteOccurrence"] = flagOccurrence
			}

			return runMutation(
				f,
				flagOutput,
				addMutation,
				map[string]any{"input": input},
				func(data json.RawMessage) (string, error) {
					var resp struct {
						AddDocumentSuggestionChange struct {
							DocumentSuggestion struct {
								ID string `json:"id"`
							} `json:"documentSuggestion"`
						} `json:"addDocumentSuggestionChange"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return "", fmt.Errorf("cannot parse response: %w", err)
					}

					return fmt.Sprintf(
						"Added a redline to suggestion %s",
						resp.AddDocumentSuggestionChange.DocumentSuggestion.ID,
					), nil
				},
			)
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)
	cmd.Flags().StringVar(&flagQuote, "quote", "", "Text to replace")
	cmd.Flags().StringVar(&flagReplacement, "replacement", "", "Suggested text; empty deletes the quote")
	cmd.Flags().IntVar(&flagOccurrence, "occurrence", 0, "Zero-based occurrence of the quote")

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

	client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
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

const listQuery = `
query DocumentSuggestions($id: ID!, $first: Int, $after: CursorKey, $filter: DocumentSuggestionFilter) {
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
query DocumentSuggestion($id: ID!) {
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
mutation StartDocumentSuggestion($input: StartDocumentSuggestionInput!) {
  startDocumentSuggestion(input: $input) {
    documentSuggestionEdge { node { id status } }
  }
}
`

const addMutation = `
mutation AddDocumentSuggestionChange($input: AddDocumentSuggestionChangeInput!) {
  addDocumentSuggestionChange(input: $input) {
    documentSuggestion { id status }
  }
}
`

const acceptMutation = `
mutation AcceptDocumentSuggestion($input: AcceptDocumentSuggestionInput!) {
  acceptDocumentSuggestion(input: $input) {
    documentSuggestion { id status resolvedAt }
  }
}
`

const rejectMutation = `
mutation RejectDocumentSuggestion($input: RejectDocumentSuggestionInput!) {
  rejectDocumentSuggestion(input: $input) {
    documentSuggestion { id status resolvedAt }
  }
}
`
