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

package update

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const updateMutation = `
mutation ThirdPartyAgreementUpdate($input: UpdateThirdPartyAgreementInput!) {
  updateThirdPartyAgreement(input: $input) {
    thirdPartyAgreement {
      id
      name
      type
    }
  }
}
`

func NewCmdUpdate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagName       string
		flagType       string
		flagValidFrom  string
		flagValidUntil string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a third-party agreement",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("name") &&
				!cmd.Flags().Changed("type") &&
				!cmd.Flags().Changed("valid-from") &&
				!cmd.Flags().Changed("valid-until") {
				return fmt.Errorf("nothing to update: set --name, --type, --valid-from, or --valid-until")
			}

			if cmd.Flags().Changed("type") {
				if err := cmdutil.ValidateEnum("type", flagType, []string{
					"BUSINESS_ASSOCIATE",
					"DATA_PRIVACY",
					"TERMS",
					"MASTER_SERVICE",
					"AMENDMENT",
					"ANNEX",
					"OTHER",
				}); err != nil {
					return err
				}
			}

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			input := map[string]any{
				"id": args[0],
			}
			if cmd.Flags().Changed("name") {
				input["name"] = flagName
			}
			if cmd.Flags().Changed("type") {
				input["type"] = flagType
			}

			period := map[string]any{}
			if cmd.Flags().Changed("valid-from") {
				if flagValidFrom == "" {
					period["start"] = nil
				} else {
					validFrom, err := normalizeDatetime(flagValidFrom)
					if err != nil {
						return fmt.Errorf("invalid --valid-from: %w", err)
					}

					period["start"] = validFrom
				}
			}
			if cmd.Flags().Changed("valid-until") {
				if flagValidUntil == "" {
					period["end"] = nil
				} else {
					validUntil, err := normalizeDatetime(flagValidUntil)
					if err != nil {
						return fmt.Errorf("invalid --valid-until: %w", err)
					}

					period["end"] = validUntil
				}
			}
			if len(period) > 0 {
				input["validity"] = period
			}

			if _, err := client.Do(updateMutation, map[string]any{"input": input}); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(f.IOStreams.Out, "Updated agreement %s\n", args[0])

			return nil
		},
	}

	cmd.Flags().StringVar(&flagName, "name", "", "Agreement name")
	cmd.Flags().StringVar(&flagType, "type", "", "Agreement type (TERMS, MASTER_SERVICE, BUSINESS_ASSOCIATE, DATA_PRIVACY, AMENDMENT, ANNEX, OTHER)")
	cmd.Flags().StringVar(&flagValidFrom, "valid-from", "", "Validity start (YYYY-MM-DD or RFC3339). Empty value clears it.")
	cmd.Flags().StringVar(&flagValidUntil, "valid-until", "", "Validity end (YYYY-MM-DD or RFC3339). Empty value clears it.")

	return cmd
}

func normalizeDatetime(value string) (string, error) {
	if _, err := time.Parse(time.RFC3339, value); err == nil {
		return value, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return "", fmt.Errorf("use YYYY-MM-DD or RFC3339")
	}

	return parsed.UTC().Format(time.RFC3339), nil
}
