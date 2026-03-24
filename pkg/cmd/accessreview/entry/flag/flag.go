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

package flag

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const flagMutation = `
mutation($input: FlagAccessEntryInput!) {
  flagAccessEntry(input: $input) {
    accessEntry {
      id
      email
      fullName
      flag
      flagReason
      decision
    }
  }
}
`

type flagResponse struct {
	FlagAccessEntry struct {
		AccessEntry struct {
			ID         string  `json:"id"`
			Email      string  `json:"email"`
			FullName   string  `json:"fullName"`
			Flag       string  `json:"flag"`
			FlagReason *string `json:"flagReason"`
			Decision   string  `json:"decision"`
		} `json:"accessEntry"`
	} `json:"flagAccessEntry"`
}

func NewCmdFlag(f *cmdutil.Factory) *cobra.Command {
	var (
		flagFlag   string
		flagReason string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "flag <entry-id>",
		Short: "Flag an access entry",
		Args:  cobra.ExactArgs(1),
		Example: `  # Flag an entry as orphaned
  prb access-review entry flag <entry-id> --flag ORPHANED --reason "No matching identity"

  # Flag an entry as inactive
  prb access-review entry flag <entry-id> --flag INACTIVE --reason "No login in 90 days"

  # Clear a flag
  prb access-review entry flag <entry-id> --flag NONE`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateEnum(
				"flag",
				flagFlag,
				[]string{"NONE", "ORPHANED", "INACTIVE", "EXCESSIVE", "ROLE_MISMATCH", "NEW"},
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
			)

			input := map[string]any{
				"accessEntryId": args[0],
				"flag":          flagFlag,
			}
			if flagReason != "" {
				input["flagReason"] = flagReason
			}

			data, err := client.Do(
				flagMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp flagResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			e := resp.FlagAccessEntry.AccessEntry

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, e)
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Flagged entry %s (%s) as %s\n", e.ID, e.Email, e.Flag)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagFlag, "flag", "", "Flag to set (NONE, ORPHANED, INACTIVE, EXCESSIVE, ROLE_MISMATCH, NEW)")
	_ = cmd.MarkFlagRequired("flag")
	cmd.Flags().StringVar(&flagReason, "reason", "", "Reason for flagging")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
