// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package link

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const linkMutation = `
mutation($input: CreateThirdPartyThirdPartyMappingInput!) {
  createThirdPartyThirdPartyMapping(input: $input) {
    thirdPartyEdge {
      node {
        id
        name
      }
    }
  }
}
`

type linkResponse struct {
	CreateThirdPartyThirdPartyMapping struct {
		ThirdPartyEdge struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"node"`
		} `json:"thirdPartyEdge"`
	} `json:"createThirdPartyThirdPartyMapping"`
}

func NewCmdLink(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link <parent-id> <child-id>",
		Short: "Link a child thirdParty to a parent thirdParty",
		Example: `  # Link a child third_party to a parent
  prb thirdParty link <parent-id> <child-id>`,
		Args: cobra.ExactArgs(2),
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

			data, err := client.Do(
				linkMutation,
				map[string]any{
					"input": map[string]any{
						"parentThirdPartyId": args[0],
						"childThirdPartyId":  args[1],
					},
				},
			)
			if err != nil {
				return err
			}

			var resp linkResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			v := resp.CreateThirdPartyThirdPartyMapping.ThirdPartyEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Linked thirdParty %s (%s) as child of %s\n",
				v.ID,
				v.Name,
				args[0],
			)

			return nil
		},
	}

	return cmd
}
