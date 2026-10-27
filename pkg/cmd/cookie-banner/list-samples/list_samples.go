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

package listsamples

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listSamplesQuery = `
query($id: ID!, $first: Int, $after: CursorKey) {
  node(id: $id) {
    __typename
    ... on CookieBanner {
      samples(first: $first, after: $after) {
        totalCount
        edges {
          node {
            id
            createdAt
            frozenAt
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

type sample struct {
	ID        string  `json:"id"`
	CreatedAt string  `json:"createdAt"`
	FrozenAt  *string `json:"frozenAt"`
}

func NewCmdListSamples(f *cmdutil.Factory) *cobra.Command {
	var (
		flagLimit  int
		flagAfter  string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list-samples <id>",
		Short:   "List cookie banner samples",
		Aliases: []string{"ls-samples"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			variables := map[string]any{"id": args[0]}

			samples, totalCount, nextCursor, err := api.Paginate(
				client,
				listSamplesQuery,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnectionOrMissing[sample]("samples", "CookieBanner", fmt.Sprintf("cookie banner %s not found", args[0])),
			)
			if err != nil {
				return err
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				samples,
				totalCount,
				cmdutil.ListView[sample]{
					Empty:      "No cookie banner samples found.",
					Noun:       "samples",
					NextCursor: nextCursor,
					Headers:    []string{"ID", "CREATED", "FROZEN"},
					Row: func(s sample) ([]string, error) {
						frozen := "open"
						if s.FrozenAt != nil {
							frozen = cmdutil.FormatTime(*s.FrozenAt)
						}

						return []string{s.ID, cmdutil.FormatTime(s.CreatedAt), frozen}, nil
					},
				},
			)
		},
	}

	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of items")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
