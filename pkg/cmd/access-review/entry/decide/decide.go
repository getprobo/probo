// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package decide

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const decideMutation = `
mutation($input: RecordAccessEntryDecisionInput!) {
  recordAccessEntryDecision(input: $input) {
    accessEntry {
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
	RecordAccessEntryDecision struct {
		AccessEntry struct {
			ID           string  `json:"id"`
			Email        string  `json:"email"`
			FullName     string  `json:"fullName"`
			Decision     string  `json:"decision"`
			DecisionNote *string `json:"decisionNote"`
			DecidedAt    *string `json:"decidedAt"`
		} `json:"accessEntry"`
	} `json:"recordAccessEntryDecision"`
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
				"accessEntryId": args[0],
				"decision":      flagDecision,
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

			e := resp.RecordAccessEntryDecision.AccessEntry

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
