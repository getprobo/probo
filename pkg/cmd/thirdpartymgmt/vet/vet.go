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

package vet

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const vetMutation = `
mutation($input: VetThirdPartyInput!) {
  vetThirdParty(input: $input) {
    thirdParty {
      id
      name
    }
  }
}
`

type vetResponse struct {
	VetThirdParty struct {
		ThirdParty struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"thirdParty"`
	} `json:"vetThirdParty"`
}

func NewCmdVet(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "vet <thirdParty-id> --url <website-url>",
		Short: "Start AI vetting of a third party from its website",
		Long:  "Queue a vetting job that crawls a third party's website using AI agents to extract security, compliance, and business information.",
		Example: `  # Vet a third party by website URL
  prb third-party vet VND_123 --url https://example.com

  # Vet with a custom procedure file
  prb third-party vet VND_123 --url https://example.com --procedure-file ./my-procedure.txt

  # Output as JSON
  prb third-party vet VND_123 --url https://example.com -o json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
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

			flagURL, _ := cmd.Flags().GetString("url")
			flagProcedureFile, _ := cmd.Flags().GetString("procedure-file")

			input := map[string]any{
				"id":         args[0],
				"websiteUrl": flagURL,
			}

			if flagProcedureFile != "" {
				data, err := os.ReadFile(flagProcedureFile)
				if err != nil {
					return fmt.Errorf("cannot read procedure file: %w", err)
				}

				input["procedure"] = string(data)
			}

			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				30*time.Second,
			)

			_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "Starting vetting for %s...\n", flagURL)

			data, err := client.Do(
				vetMutation,
				map[string]any{
					"input": input,
				},
			)
			if err != nil {
				return err
			}

			var resp vetResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.VetThirdParty)
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Vetting started for %s\n", resp.VetThirdParty.ThirdParty.Name)

			return nil
		},
	}

	cmd.Flags().String("url", "", "Third party website URL to vet (required)")
	_ = cmd.MarkFlagRequired("url")
	cmd.Flags().String("procedure-file", "", "Path to a custom vetting procedure file")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
