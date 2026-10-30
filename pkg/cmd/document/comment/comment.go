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

package comment

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

func NewCmdComment(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment <command>",
		Short: "Manage document comments",
	}

	cmd.AddCommand(newListCmd(f))
	cmd.AddCommand(newViewCmd(f))
	cmd.AddCommand(newCreateCmd(f))
	cmd.AddCommand(newReplyCmd(f))
	cmd.AddCommand(newUpdateCmd(f))
	cmd.AddCommand(newDeleteCmd(f))
	cmd.AddCommand(newResolveCmd(f))
	cmd.AddCommand(newReopenCmd(f))

	return cmd
}

type commentThread struct {
	ID          string `json:"id"`
	AnchorQuote string `json:"anchorQuote"`
	ResolvedAt  string `json:"resolvedAt"`
	Suggestion  *struct {
		ID string `json:"id"`
	} `json:"suggestion"`
	CreatedAt string `json:"createdAt"`
}

func newListCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagResolved string
		flagLimit    int
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "list <document-id>",
		Short: "List comment threads on a document",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			variables := map[string]any{"id": args[0]}
			if flagResolved != "" {
				if err := cmdutil.ValidateEnum("resolved", flagResolved, []string{"true", "false"}); err != nil {
					return err
				}
				variables["filter"] = map[string]any{"resolved": flagResolved == "true"}
			}

			client, err := newClient(f)
			if err != nil {
				return err
			}

			threads, totalCount, err := api.Paginate(
				client,
				listQuery,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[commentThread], error) {
					var resp struct {
						Node *struct {
							Typename       string                        `json:"__typename"`
							CommentThreads api.Connection[commentThread] `json:"commentThreads"`
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

					return &resp.Node.CommentThreads, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, threads)
			}

			if len(threads) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No comment threads found.")
				return nil
			}

			rows := make([][]string, 0, len(threads))
			for _, t := range threads {
				resolved := "no"
				if t.ResolvedAt != "" {
					resolved = "yes"
				}

				suggestion := ""
				if t.Suggestion != nil {
					suggestion = t.Suggestion.ID
				}

				rows = append(rows, []string{
					t.ID,
					t.AnchorQuote,
					resolved,
					suggestion,
					cmdutil.FormatTime(t.CreatedAt),
				})
			}

			tbl := cmdutil.NewTable("ID", "QUOTE", "RESOLVED", "SUGGESTION", "CREATED").Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, tbl)

			if totalCount > len(threads) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d comment threads\n",
					len(threads),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagResolved, "resolved", "", "Filter by resolved state (true or false)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of threads")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

type viewResponse struct {
	Node *struct {
		Typename    string `json:"__typename"`
		ID          string `json:"id"`
		AnchorQuote string `json:"anchorQuote"`
		ResolvedAt  string `json:"resolvedAt"`
		Suggestion  *struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"suggestion"`
		Comments struct {
			Edges []struct {
				Node struct {
					ID        string `json:"id"`
					Content   string `json:"content"`
					CreatedAt string `json:"createdAt"`
					Author    *struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
					} `json:"author"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"comments"`
	} `json:"node"`
}

func newViewCmd(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <thread-id>",
		Short: "View a comment thread",
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
				return fmt.Errorf("comment thread %s not found", args[0])
			}

			if resp.Node.Typename != "DocumentCommentThread" {
				return fmt.Errorf("expected DocumentCommentThread node, got %s", resp.Node.Typename)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.Node)
			}

			n := resp.Node
			out := f.IOStreams.Out

			label := lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Width(22)

			status := "Open"
			if n.ResolvedAt != "" {
				status = "Resolved " + cmdutil.FormatTime(n.ResolvedAt)
			}

			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("ID:"), n.ID)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Status:"), status)
			_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Quote:"), n.AnchorQuote)

			if n.Suggestion != nil {
				_, _ = fmt.Fprintf(out, "%s%s (%s)\n", label.Render("Suggestion:"), n.Suggestion.ID, n.Suggestion.Status)
			}

			for _, edge := range n.Comments.Edges {
				c := edge.Node

				author := ""
				if c.Author != nil {
					author = c.Author.FullName
				}

				content, err := cmdutil.FormatRichText(c.Content)
				if err != nil {
					return fmt.Errorf("cannot format comment content: %w", err)
				}

				_, _ = fmt.Fprintln(out)
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Comment:"), c.ID)
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Author:"), author)
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Created:"), cmdutil.FormatTime(c.CreatedAt))
				_, _ = fmt.Fprintf(out, "%s%s\n", label.Render("Content:"), content)
			}

			return nil
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newCreateCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagContent    string
		flagQuote      string
		flagOccurrence int
		flagSuggestion string
		flagOutput     *string
	)

	cmd := &cobra.Command{
		Use:   "create <document-id>",
		Short: "Create a comment thread from a quote, or on a suggestion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagContent == "" {
				return fmt.Errorf("content is required")
			}
			if flagQuote == "" && flagSuggestion == "" {
				return fmt.Errorf("quote or suggestion is required")
			}

			input := map[string]any{
				"documentId": args[0],
				"content":    flagContent,
			}
			if flagQuote != "" {
				input["quote"] = flagQuote
				input["quoteOccurrence"] = flagOccurrence
			}
			if flagSuggestion != "" {
				input["suggestionId"] = flagSuggestion
			}

			return runMutation(
				f,
				flagOutput,
				createMutation,
				map[string]any{"input": input},
				func(data json.RawMessage) (string, error) {
					var resp struct {
						CreateDocumentCommentThread struct {
							DocumentCommentThreadEdge struct {
								Node struct {
									ID string `json:"id"`
								} `json:"node"`
							} `json:"documentCommentThreadEdge"`
						} `json:"createDocumentCommentThread"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return "", fmt.Errorf("cannot parse response: %w", err)
					}

					return fmt.Sprintf(
						"Created comment thread %s",
						resp.CreateDocumentCommentThread.DocumentCommentThreadEdge.Node.ID,
					), nil
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagContent, "content", "", "Comment text")
	cmd.Flags().StringVar(&flagQuote, "quote", "", "Document text to anchor")
	cmd.Flags().IntVar(&flagOccurrence, "occurrence", 0, "Zero-based quote occurrence")
	cmd.Flags().StringVar(&flagSuggestion, "suggestion", "", "Suggestion ID to comment on")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newReplyCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagContent string
		flagOutput  *string
	)

	cmd := &cobra.Command{
		Use:   "reply <thread-id>",
		Short: "Reply on a comment thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagContent == "" {
				return fmt.Errorf("content is required")
			}

			return runMutation(
				f,
				flagOutput,
				replyMutation,
				map[string]any{
					"input": map[string]any{
						"threadId": args[0],
						"content":  flagContent,
					},
				},
				func(data json.RawMessage) (string, error) {
					var resp struct {
						CreateDocumentComment struct {
							DocumentCommentEdge struct {
								Node struct {
									ID string `json:"id"`
								} `json:"node"`
							} `json:"documentCommentEdge"`
						} `json:"createDocumentComment"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return "", fmt.Errorf("cannot parse response: %w", err)
					}

					return fmt.Sprintf(
						"Created comment %s",
						resp.CreateDocumentComment.DocumentCommentEdge.Node.ID,
					), nil
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagContent, "content", "", "Reply text")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		flagContent string
		flagOutput  *string
	)

	cmd := &cobra.Command{
		Use:   "update <comment-id>",
		Short: "Edit your own comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagContent == "" {
				return fmt.Errorf("content is required")
			}

			return runMutation(
				f,
				flagOutput,
				updateMutation,
				map[string]any{
					"input": map[string]any{
						"commentId": args[0],
						"content":   flagContent,
					},
				},
				func(data json.RawMessage) (string, error) {
					var resp struct {
						UpdateDocumentComment struct {
							DocumentComment struct {
								ID string `json:"id"`
							} `json:"documentComment"`
						} `json:"updateDocumentComment"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return "", fmt.Errorf("cannot parse response: %w", err)
					}

					return fmt.Sprintf(
						"Updated comment %s",
						resp.UpdateDocumentComment.DocumentComment.ID,
					), nil
				},
			)
		},
	}

	cmd.Flags().StringVar(&flagContent, "content", "", "Replacement comment text")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "delete <comment-id>",
		Short: "Delete your own comment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMutation(
				f,
				flagOutput,
				deleteMutation,
				map[string]any{
					"input": map[string]any{"commentId": args[0]},
				},
				func(data json.RawMessage) (string, error) {
					var resp struct {
						DeleteDocumentComment struct {
							DeletedDocumentCommentID string `json:"deletedDocumentCommentId"`
						} `json:"deleteDocumentComment"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return "", fmt.Errorf("cannot parse response: %w", err)
					}

					return fmt.Sprintf(
						"Deleted comment %s",
						resp.DeleteDocumentComment.DeletedDocumentCommentID,
					), nil
				},
			)
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newResolveCmd(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "resolve <thread-id>",
		Short: "Resolve a comment thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runThreadStateMutation(f, flagOutput, resolveMutation, args[0], "resolveDocumentCommentThread", "Resolved")
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newReopenCmd(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "reopen <thread-id>",
		Short: "Reopen a comment thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runThreadStateMutation(f, flagOutput, reopenMutation, args[0], "reopenDocumentCommentThread", "Reopened")
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func runThreadStateMutation(
	f *cmdutil.Factory,
	flagOutput *string,
	query string,
	threadID string,
	field string,
	past string,
) error {
	return runMutation(
		f,
		flagOutput,
		query,
		map[string]any{
			"input": map[string]any{"threadId": threadID},
		},
		func(data json.RawMessage) (string, error) {
			var resp map[string]struct {
				DocumentCommentThread struct {
					ID string `json:"id"`
				} `json:"documentCommentThread"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return "", fmt.Errorf("cannot parse response: %w", err)
			}

			return fmt.Sprintf("%s comment thread %s", past, resp[field].DocumentCommentThread.ID), nil
		},
	)
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
query($id: ID!, $first: Int, $after: CursorKey, $filter: DocumentCommentThreadFilter) {
  node(id: $id) {
    __typename
    ... on Document {
      commentThreads(first: $first, after: $after, filter: $filter) {
        totalCount
        edges {
          node {
            id
            anchorQuote
            resolvedAt
            suggestion { id }
            createdAt
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

const viewQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on DocumentCommentThread {
      id
      anchorQuote
      resolvedAt
      suggestion { id status }
      comments(first: 50) {
        edges { node { id content createdAt author { id fullName } } }
      }
    }
  }
}
`

const createMutation = `
mutation($input: CreateDocumentCommentThreadInput!) {
  createDocumentCommentThread(input: $input) {
    documentCommentThreadEdge { node { id anchorQuote suggestion { id } } }
  }
}
`

const replyMutation = `
mutation($input: CreateDocumentCommentInput!) {
  createDocumentComment(input: $input) {
    documentCommentEdge { node { id content } }
  }
}
`

const updateMutation = `
mutation($input: UpdateDocumentCommentInput!) {
  updateDocumentComment(input: $input) {
    documentComment { id content }
  }
}
`

const deleteMutation = `
mutation($input: DeleteDocumentCommentInput!) {
  deleteDocumentComment(input: $input) {
    deletedDocumentCommentId
    deletedDocumentCommentThreadId
  }
}
`

const resolveMutation = `
mutation($input: ResolveDocumentCommentThreadInput!) {
  resolveDocumentCommentThread(input: $input) {
    documentCommentThread { id resolvedAt }
  }
}
`

const reopenMutation = `
mutation($input: ReopenDocumentCommentThreadInput!) {
  reopenDocumentCommentThread(input: $input) {
    documentCommentThread { id resolvedAt }
  }
}
`
