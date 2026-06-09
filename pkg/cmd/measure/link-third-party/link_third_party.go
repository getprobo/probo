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

package linkthirdparty

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const linkThirdPartyMutation = `
mutation($input: CreateMeasureThirdPartyMappingInput!) {
  createMeasureThirdPartyMapping(input: $input) {
    measureEdge {
      node { id }
    }
    thirdPartyEdge {
      node { id }
    }
  }
}
`

func NewCmdLinkThirdParty(f *cmdutil.Factory) *cobra.Command {
	var (
		flagMeasureID    string
		flagThirdPartyID string
	)

	cmd := &cobra.Command{
		Use:   "link-third-party",
		Short: "Link a third party to a measure",
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

			_, err = client.Do(
				linkThirdPartyMutation,
				map[string]any{
					"input": map[string]any{
						"measureId":    flagMeasureID,
						"thirdPartyId": flagThirdPartyID,
					},
				},
			)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Linked third party %s to measure %s\n",
				flagThirdPartyID,
				flagMeasureID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagMeasureID, "measure-id", "", "Measure ID (required)")
	cmd.Flags().StringVar(&flagThirdPartyID, "third-party-id", "", "Third party ID (required)")

	_ = cmd.MarkFlagRequired("measure-id")
	_ = cmd.MarkFlagRequired("third-party-id")

	return cmd
}
