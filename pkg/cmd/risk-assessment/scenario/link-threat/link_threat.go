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

package linkthreat

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const linkThreatMutation = `
mutation($input: LinkRiskAssessmentScenarioThreatInput!) {
  linkRiskAssessmentScenarioThreat(input: $input) {
    riskAssessmentScenario {
      id
    }
  }
}
`

func NewCmdLinkThreat(f *cmdutil.Factory) *cobra.Command {
	var (
		flagScenarioId string
		flagThreatId   string
	)

	cmd := &cobra.Command{
		Use:   "link-threat",
		Short: "Link a threat to a risk assessment scenario",
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
				linkThreatMutation,
				map[string]any{
					"input": map[string]any{
						"riskAssessmentScenarioId": flagScenarioId,
						"threatId":                 flagThreatId,
					},
				},
			)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Linked threat %s to scenario %s\n",
				flagThreatId,
				flagScenarioId,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagScenarioId, "scenario-id", "", "Risk assessment scenario ID (required)")
	cmd.Flags().StringVar(&flagThreatId, "threat-id", "", "Risk assessment threat ID (required)")

	_ = cmd.MarkFlagRequired("scenario-id")
	_ = cmd.MarkFlagRequired("threat-id")

	return cmd
}
