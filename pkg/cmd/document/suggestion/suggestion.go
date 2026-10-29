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
			variables := map[string]any{
				"id":    args[0],
				"first": flagLimit,
			}
			if flagStatus != "" {
				if err := cmdutil.ValidateEnum("status", flagStatus, []string{"OPEN", "ACCEPTED", "REJECTED"}); err != nil {
					return err
				}
				variables["filter"] = map[string]any{"status": flagStatus}
			}

			return printQuery(f, flagOutput, listQuery, variables)
		},
	}

	cmd.Flags().StringVar(&flagStatus, "status", "OPEN", "Suggestion status filter")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of suggestions")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newViewCmd(f *cmdutil.Factory) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   "view <suggestion-id>",
		Short: "View a suggestion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printQuery(f, flagOutput, viewQuery, map[string]any{"id": args[0]})
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
			return printQuery(f, flagOutput, startMutation, map[string]any{
				"input": map[string]any{"documentId": args[0]},
			})
		},
	}

	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}

func newAcceptCmd(f *cmdutil.Factory) *cobra.Command {
	return newDecisionCmd(f, "accept", "Accept a suggestion", acceptMutation)
}

func newRejectCmd(f *cmdutil.Factory) *cobra.Command {
	return newDecisionCmd(f, "reject", "Reject a suggestion", rejectMutation)
}

func newDecisionCmd(f *cmdutil.Factory, use string, short string, query string) *cobra.Command {
	var flagOutput *string

	cmd := &cobra.Command{
		Use:   use + " <suggestion-id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return printQuery(f, flagOutput, query, map[string]any{
				"input": map[string]any{"suggestionId": args[0]},
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

	cfg, err := f.Config()
	if err != nil {
		return err
	}

	host, hc, err := cfg.DefaultHost()
	if err != nil {
		return err
	}

	client := api.NewClient(host, hc.Token, "/api/console/v1/graphql", cfg.HTTPTimeoutDuration())

	data, err := client.Do(query, variables)
	if err != nil {
		return err
	}

	if *flagOutput == cmdutil.OutputJSON {
		return cmdutil.PrintJSON(f.IOStreams.Out, json.RawMessage(data))
	}

	_, _ = fmt.Fprintln(f.IOStreams.Out, string(data))

	return nil
}

const listQuery = `
query($id: ID!, $first: Int, $filter: DocumentSuggestionFilter) {
  node(id: $id) {
    __typename
    ... on Document {
      suggestions(first: $first, filter: $filter) {
        edges { node { id status author { id fullName } createdAt } }
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
