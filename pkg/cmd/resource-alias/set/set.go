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

package set

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const setMutation = `
mutation($input: SetResourceAliasInput!) {
    setResourceAlias(input: $input) {
      resourceAlias {
        resourceId
        alias
      }
    }
}
`

type setResponse struct {
	SetResourceAlias struct {
		ResourceAlias struct {
			ResourceID string `json:"resourceId"`
			Alias      string `json:"alias"`
		} `json:"resourceAlias"`
	} `json:"setResourceAlias"`
}

func NewCmdSet(f *cmdutil.Factory) *cobra.Command {
	var (
		flagResourceID string
		flagAlias      string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a resource alias",
		Example: `  # Set an alias interactively
  prb resource-alias set --resource-id prbdoc_...

  # Set an alias non-interactively
  prb resource-alias set --resource-id prbdoc_... --alias privacy-policy`,
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
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			if f.IOStreams.IsInteractive() {
				if flagResourceID == "" {
					err := huh.NewInput().
						Title("Resource ID").
						Value(&flagResourceID).
						Run()
					if err != nil {
						return err
					}
				}

				if flagAlias == "" {
					err := huh.NewInput().
						Title("Alias").
						Value(&flagAlias).
						Run()
					if err != nil {
						return err
					}
				}
			}

			if flagResourceID == "" {
				return fmt.Errorf("resource ID is required; pass --resource-id or run interactively")
			}

			if flagAlias == "" {
				return fmt.Errorf("alias is required; pass --alias or run interactively")
			}

			data, err := client.Do(
				setMutation,
				map[string]any{
					"input": map[string]any{
						"resourceId": flagResourceID,
						"alias":      flagAlias,
					},
				},
			)
			if err != nil {
				return err
			}

			var resp setResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			a := resp.SetResourceAlias.ResourceAlias
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Set alias %q for resource %s\n",
				a.Alias,
				a.ResourceID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagResourceID, "resource-id", "", "ID of the resource to alias")
	cmd.Flags().StringVar(&flagAlias, "alias", "", "Human-readable alias slug")

	return cmd
}
