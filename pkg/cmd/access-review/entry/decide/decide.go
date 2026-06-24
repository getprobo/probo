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

package decide

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const decideMutation = `
mutation($input: RecordAccessReviewEntryDecisionInput!) {
  recordAccessReviewEntryDecision(input: $input) {
    accessReviewEntry {
      id
      email
      fullName
      decision
      decisionNote
      decidedAt
    }
  }
}
`

type decideResponse struct {
	RecordAccessReviewEntryDecision struct {
		AccessReviewEntry struct {
			ID           string  `json:"id"`
			Email        string  `json:"email"`
			FullName     string  `json:"fullName"`
			Decision     string  `json:"decision"`
			DecisionNote *string `json:"decisionNote"`
			DecidedAt    *string `json:"decidedAt"`
		} `json:"accessReviewEntry"`
	} `json:"recordAccessReviewEntryDecision"`
}

func NewCmdDecide(f *cmdutil.Factory) *cobra.Command {
	var (
		flagDecision string
		flagNote     string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "decide <entry-id>",
		Short: "Record a decision on an access entry",
		Args:  cobra.ExactArgs(1),
		Example: `  # Approve an access entry
  prb access-review entry decide <entry-id> --decision APPROVED

  # Revoke with a note
  prb access-review entry decide <entry-id> --decision REVOKE --note "User left the company"

  # Defer a decision
  prb access-review entry decide <entry-id> --decision DEFER --note "Need more context"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateEnum(
				"decision",
				flagDecision,
				[]string{"APPROVED", "REVOKE", "DEFER", "ESCALATE"},
			); err != nil {
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

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				cfg.HTTPTimeoutDuration(),
				cmdutil.TokenRefreshOption(cfg, host, hc),
			)

			input := map[string]any{
				"accessReviewEntryId": args[0],
				"decision":            flagDecision,
			}
			if flagNote != "" {
				input["decisionNote"] = flagNote
			}

			data, err := client.Do(
				decideMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp decideResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			e := resp.RecordAccessReviewEntryDecision.AccessReviewEntry

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, e)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Recorded decision %s on entry %s\n",
				e.Decision,
				e.ID,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(
		&flagDecision,
		"decision",
		"",
		"Decision to record (APPROVED, REVOKE, DEFER, ESCALATE)",
	)
	_ = cmd.MarkFlagRequired("decision")
	cmd.Flags().StringVar(&flagNote, "note", "", "Decision note")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
