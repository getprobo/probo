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

package decideall

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const decideAllMutation = `
mutation($input: RecordAccessEntryDecisionsInput!) {
  recordAccessEntryDecisions(input: $input) {
    accessEntries {
      id
      email
      decision
    }
  }
}
`

type decideAllResponse struct {
	RecordAccessEntryDecisions struct {
		AccessEntries []struct {
			ID       string `json:"id"`
			Email    string `json:"email"`
			Decision string `json:"decision"`
		} `json:"accessEntries"`
	} `json:"recordAccessEntryDecisions"`
}

func NewCmdDecideAll(f *cmdutil.Factory) *cobra.Command {
	var (
		flagEntryIDs []string
		flagDecision string
		flagNote     string
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:   "decide-all",
		Short: "Record decisions on multiple access entries",
		Args:  cobra.NoArgs,
		Example: `  # Approve multiple entries
  prb access-review entry decide-all --entry-id <id1> --entry-id <id2> --decision APPROVED

  # Revoke multiple entries with a note
  prb access-review entry decide-all --entry-id <id1> --entry-id <id2> --decision REVOKE --note "Batch cleanup"`,
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

			decisions := make([]map[string]any, len(flagEntryIDs))
			for i, id := range flagEntryIDs {
				d := map[string]any{
					"accessEntryId": id,
					"decision":      flagDecision,
				}
				if flagNote != "" {
					d["decisionNote"] = flagNote
				}
				decisions[i] = d
			}

			data, err := client.Do(
				decideAllMutation,
				map[string]any{"input": map[string]any{"decisions": decisions}},
			)
			if err != nil {
				return err
			}

			var resp decideAllResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			entries := resp.RecordAccessEntryDecisions.AccessEntries

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, entries)
			}

			for _, e := range entries {
				_, _ = fmt.Fprintf(
					f.IOStreams.Out,
					"Recorded decision %s on entry %s\n",
					e.Decision,
					e.ID,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringSliceVar(
		&flagEntryIDs,
		"entry-id",
		nil,
		"Access entry IDs (can be repeated)",
	)
	_ = cmd.MarkFlagRequired("entry-id")
	cmd.Flags().StringVar(
		&flagDecision,
		"decision",
		"",
		"Decision to record (APPROVED, REVOKE, DEFER, ESCALATE)",
	)
	_ = cmd.MarkFlagRequired("decision")
	cmd.Flags().StringVar(&flagNote, "note", "", "Decision note (applied to all entries)")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
