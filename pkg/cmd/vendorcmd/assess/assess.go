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

package assess

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const assessMutation = `
mutation($input: AssessVendorInput!) {
  assessVendor(input: $input) {
    report
    subprocessors {
      name
      country
      purpose
    }
    vendor {
      id
      name
    }
  }
}
`

type assessResponse struct {
	AssessVendor struct {
		Report        string `json:"report"`
		Subprocessors []struct {
			Name    string `json:"name"`
			Country string `json:"country"`
			Purpose string `json:"purpose"`
		} `json:"subprocessors"`
		Vendor struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"vendor"`
	} `json:"assessVendor"`
}

func NewCmdAssess(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOutput *string
	)

	cmd := &cobra.Command{
		Use:   "assess <vendor-id> --url <website-url>",
		Short: "Run AI assessment on a vendor from its website",
		Long:  "Analyze a vendor's website using AI agents to extract security, compliance, and business information.",
		Example: `  # Assess a vendor by website URL
  prb vendor assess VND_123 --url https://example.com

  # Assess with a custom procedure file
  prb vendor assess VND_123 --url https://example.com --procedure-file ./my-procedure.txt

  # Output as JSON
  prb vendor assess VND_123 --url https://example.com -o json`,
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

			// The CLI timeout must outlast the server-side assessment
			// timeout (vetting.AssessmentTimeout = 20m) plus HTTP overhead.
			client := api.NewClient(
				host,
				hc.Token,
				"/api/console/v1/graphql",
				22*time.Minute,
			)

			_, _ = fmt.Fprintf(f.IOStreams.ErrOut, "Assessing vendor from %s (this may take a few minutes)...\n", flagURL)

			data, err := client.Do(
				assessMutation,
				map[string]any{
					"input": input,
				},
			)
			if err != nil {
				return err
			}

			var resp assessResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, resp.AssessVendor)
			}

			_, _ = fmt.Fprintln(f.IOStreams.Out, resp.AssessVendor.Report)

			return nil
		},
	}

	cmd.Flags().String("url", "", "Vendor website URL to assess (required)")
	_ = cmd.MarkFlagRequired("url")
	cmd.Flags().String("procedure-file", "", "Path to a custom assessment procedure file")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
