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

package unlinkrisk

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const unlinkRiskMutation = `
mutation($input: UnlinkRiskAssessmentScenarioRiskInput!) {
  unlinkRiskAssessmentScenarioRisk(input: $input) {
    riskAssessmentScenario {
      id
    }
  }
}
`

func NewCmdUnlinkRisk(f *cmdutil.Factory) *cobra.Command {
	var (
		flagScenarioId string
		flagRiskId     string
	)

	cmd := &cobra.Command{
		Use:   "unlink-risk",
		Short: "Unlink a risk from a risk assessment scenario",
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
				unlinkRiskMutation,
				map[string]any{
					"input": map[string]any{
						"riskAssessmentScenarioId": flagScenarioId,
						"riskId":                   flagRiskId,
					},
				},
			)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Unlinked risk %s from scenario %s\n",
				flagRiskId,
				flagScenarioId,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagScenarioId, "scenario-id", "", "Risk assessment scenario ID (required)")
	cmd.Flags().StringVar(&flagRiskId, "risk-id", "", "Risk ID (required)")

	_ = cmd.MarkFlagRequired("scenario-id")
	_ = cmd.MarkFlagRequired("risk-id")

	return cmd
}
