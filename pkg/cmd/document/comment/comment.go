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

			variables := map[string]any{
				"id":    args[0],
				"first": flagLimit,
			}
			if flagResolved != "" {
				if err := cmdutil.ValidateEnum("resolved", flagResolved, []string{"true", "false"}); err != nil {
					return err
				}
				variables["filter"] = map[string]any{"resolved": flagResolved == "true"}
			}

			data, err := do(f, listQuery, variables)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, json.RawMessage(data))
			}

			_, _ = fmt.Fprintln(f.IOStreams.Out, string(data))

			return nil
		},
	}

	cmd.Flags().StringVar(&flagResolved, "resolved", "", "Filter by resolved state (true or false)")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of threads")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newViewCmd(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <thread-id>",
		Short: "View a comment thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printQuery(f, flagOutput, viewQuery, map[string]any{"id": args[0]})
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

			return printQuery(f, flagOutput, createMutation, map[string]any{"input": input})
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

			return printQuery(f, flagOutput, replyMutation, map[string]any{
				"input": map[string]any{
					"threadId": args[0],
					"content":  flagContent,
				},
			})
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

			return printQuery(f, flagOutput, updateMutation, map[string]any{
				"input": map[string]any{
					"commentId": args[0],
					"content":   flagContent,
				},
			})
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
			return printQuery(f, flagOutput, deleteMutation, map[string]any{
				"input": map[string]any{"commentId": args[0]},
			})
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
			return printQuery(f, flagOutput, resolveMutation, map[string]any{
				"input": map[string]any{"threadId": args[0]},
			})
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
			return printQuery(f, flagOutput, reopenMutation, map[string]any{
				"input": map[string]any{"threadId": args[0]},
			})
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func printQuery(f *cmdutil.Factory, flagOutput *string, query string, variables map[string]any) error {
	if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
		return err
	}

	data, err := do(f, query, variables)
	if err != nil {
		return err
	}

	if *flagOutput == cmdutil.OutputJSON {
		return cmdutil.PrintJSON(f.IOStreams.Out, json.RawMessage(data))
	}

	_, _ = fmt.Fprintln(f.IOStreams.Out, string(data))

	return nil
}

func do(f *cmdutil.Factory, query string, variables map[string]any) (json.RawMessage, error) {
	cfg, err := f.Config()
	if err != nil {
		return nil, err
	}

	host, hc, err := cfg.DefaultHost()
	if err != nil {
		return nil, err
	}

	client := api.NewClient(host, hc.Token, "/api/console/v1/graphql", cfg.HTTPTimeoutDuration())

	return client.Do(query, variables)
}

const listQuery = `
query($id: ID!, $first: Int, $filter: DocumentCommentThreadFilter) {
  node(id: $id) {
    __typename
    ... on Document {
      commentThreads(first: $first, filter: $filter) {
        edges {
          node {
            id
            anchorQuote
            resolvedAt
            suggestion { id }
            createdAt
          }
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
