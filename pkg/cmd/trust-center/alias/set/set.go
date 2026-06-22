// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
mutation($input: SetTrustCenterAliasInput!) {
    setTrustCenterAlias(input: $input) {
      alias {
        resourceId
        alias
      }
    }
}
`

type setResponse struct {
		SetTrustCenterAlias struct {
			Alias struct {
				ResourceID string `json:"resourceId"`
				Alias      string `json:"alias"`
			} `json:"alias"`
		} `json:"setTrustCenterAlias"`
}

func NewCmdSet(f *cmdutil.Factory) *cobra.Command {
	var (
		flagResourceID string
		flagAlias      string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a trust center alias",
		Example: `  # Set an alias interactively
  prb trust-center alias set --resource-id prbdoc_... 

  # Set an alias non-interactively
  prb trust-center alias set --resource-id prbdoc_... --alias privacy-policy`,
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
						Title("Resource ID (document, file, or audit)").
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

			a := resp.SetTrustCenterAlias.Alias
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Set alias %q for resource %s\n",
				a.Alias,
				a.ResourceID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagResourceID, "resource-id", "", "Document, trust center file, or audit ID")
	cmd.Flags().StringVar(&flagAlias, "alias", "", "Human-readable alias slug")

	return cmd
}
