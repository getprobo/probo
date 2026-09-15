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

package update

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/service-account/shared"
)

const updateMutation = `
mutation($input: UpdateServiceAccountInput!) {
  updateServiceAccount(input: $input) {
    serviceAccount {
      id
      name
    }
  }
}
`

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName        string
		flagDescription string
		flagScopes      []string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := map[string]any{"serviceAccountId": args[0]}

			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}
			if cmd.Flags().Changed("description") {
				input["description"] = flagDescription
			}
			if cmd.Flags().Changed("scope") {
				input["scopes"] = flagScopes
			}
			if len(input) == 1 {
				return fmt.Errorf("at least one of --name, --description, or --scope must be specified")
			}

			cfg, err := f.Config()
			if err != nil {
				return err
			}

			host, hc, err := cfg.DefaultHost()
			if err != nil {
				return err
			}

			data, err := shared.NewClient(cfg, host, hc).Do(
				updateMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp struct {
				UpdateServiceAccount struct {
					ServiceAccount shared.ServiceAccount `json:"serviceAccount"`
				} `json:"updateServiceAccount"`
			}
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Updated service account %s\n",
				resp.UpdateServiceAccount.ServiceAccount.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Service account name")
	cmd.Flags().StringVar(&flagDescription, "description", "", "Service account description")
	cmd.Flags().StringSliceVar(&flagScopes, "scope", nil, "OAuth scope (repeat or separate with commas; replaces existing)")

	return cmd
}
