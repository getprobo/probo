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

package setflag

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const flagMutation = `
mutation($input: FlagAccessReviewEntryInput!) {
  flagAccessReviewEntry(input: $input) {
    accessReviewEntry {
      id
      email
      fullName
      flags
      flagReasons
      decision
    }
  }
}
`

type flagResponse struct {
	FlagAccessReviewEntry struct {
		AccessReviewEntry struct {
			ID          string   `json:"id"`
			Email       string   `json:"email"`
			FullName    string   `json:"fullName"`
			Flags       []string `json:"flags"`
			FlagReasons []string `json:"flagReasons"`
			Decision    string   `json:"decision"`
		} `json:"accessReviewEntry"`
	} `json:"flagAccessReviewEntry"`
}

func NewCmdFlag(f *cmdutil.Factory) *cobra.Command {
	var (
		flagFlags  []string
		flagReason string
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "flag <entry-id>",
		Short: "Flag an access entry",
		Args:  cobra.ExactArgs(1),
		Example: `  # Flag an entry as orphaned
  prb access-review entry flag <entry-id> --flags ORPHANED --reason "No matching identity"

  # Flag an entry with multiple flags
  prb access-review entry flag <entry-id> --flags ORPHANED,INACTIVE --reason "No login in 90 days"

  # Clear all flags
  prb access-review entry flag <entry-id> --flags ""`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			validFlags := []string{
				"NONE", "ORPHANED", "INACTIVE", "EXCESSIVE", "ROLE_MISMATCH", "NEW",
				"DORMANT", "TERMINATED_USER", "CONTRACTOR_EXPIRED", "SOD_CONFLICT",
				"PRIVILEGED_ACCESS", "ROLE_CREEP", "NO_BUSINESS_JUSTIFICATION",
				"OUT_OF_DEPARTMENT", "SHARED_ACCOUNT",
			}
			for _, f := range flagFlags {
				if err := cmdutil.ValidateEnum("flags", f, validFlags); err != nil {
					return err
				}
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
				"flags":               flagFlags,
			}
			if flagReason != "" {
				input["flagReasons"] = []string{flagReason}
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

			e := resp.FlagAccessReviewEntry.AccessReviewEntry

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, e)
			}

			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Flagged entry %s (%s) as %s\n",
				e.ID,
				e.Email,
				strings.Join(e.Flags, ", "),
			)

			return nil
		},
	}

	cmd.Flags().StringSliceVar(&flagFlags, "flags", nil, "Flags to set (ORPHANED, INACTIVE, EXCESSIVE, ROLE_MISMATCH, NEW, etc.)")
	_ = cmd.MarkFlagRequired("flags")
	cmd.Flags().StringVar(&flagReason, "reason", "", "Reason for flagging")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
