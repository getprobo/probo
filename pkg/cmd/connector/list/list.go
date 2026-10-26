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

package list

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const listQuery = `
query($id: ID!) {
  node(id: $id) {
    __typename
    ... on Organization {
      connectors {
        id
        name
        provider
        protocol
        createdAt
      }
    }
  }
}
`

type connector struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	Protocol  string `json:"protocol"`
	CreatedAt string `json:"createdAt"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg    string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List connectors in an organization",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			client, hc, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
			}

			data, err := client.Do(listQuery, map[string]any{"id": flagOrg})
			if err != nil {
				return err
			}

			connectors, err := api.NodeSlice[connector](
				data,
				"connectors",
				"Organization",
				fmt.Sprintf("organization %s not found", flagOrg),
			)
			if err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				connectors,
				len(connectors),
				cmdutil.ListView[connector]{
					Empty:   "No connectors found.",
					Noun:    "connectors",
					Headers: []string{"ID", "NAME", "PROVIDER", "PROTOCOL", "CREATED"},
					Row: func(c connector) ([]string, error) {
						return []string{
							c.ID,
							c.Name,
							c.Provider,
							c.Protocol,
							cmdutil.FormatTime(c.CreatedAt),
						}, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
