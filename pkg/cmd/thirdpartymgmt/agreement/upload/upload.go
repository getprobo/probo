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

package upload

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const uploadMutation = `
mutation ThirdPartyAgreementUpload($input: UploadThirdPartyAgreementInput!) {
  uploadThirdPartyAgreement(input: $input) {
    thirdPartyAgreementEdge {
      node {
        id
        name
        type
      }
    }
  }
}
`

var agreementTypes = []string{
	"BUSINESS_ASSOCIATE",
	"DATA_PRIVACY",
	"TERMS",
	"MASTER_SERVICE",
	"AMENDMENT",
	"ANNEX",
	"OTHER",
}

type uploadResponse struct {
	UploadThirdPartyAgreement struct {
		ThirdPartyAgreementEdge struct {
			Node struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"node"`
		} `json:"thirdPartyAgreementEdge"`
	} `json:"uploadThirdPartyAgreement"`
}

func NewCmdUpload(f *cmdutil.Factory) *cobra.Command {
	var (
		flagThirdParty string
		flagName       string
		flagType       string
		flagValidFrom  string
		flagValidUntil string
	)

	cmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload a third-party agreement",
		Example: `  # Upload a named MSA
  prb third-party agreement upload ./msa.pdf \
    --third-party <third-party-id> \
    --name "Master service agreement" \
    --type MASTER_SERVICE`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateEnum("type", flagType, agreementTypes); err != nil {
				return err
			}

			filePath := args[0]

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("cannot open file: %w", err)
			}
			defer func() { _ = file.Close() }()

			client, _, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			input := map[string]any{
				"thirdPartyId": flagThirdParty,
				"name":         flagName,
				"type":         flagType,
				"file":         nil,
			}
			validFrom, err := normalizeDatetime(flagValidFrom)
			if err != nil {
				return fmt.Errorf("invalid --valid-from: %w", err)
			}

			validUntil, err := normalizeDatetime(flagValidUntil)
			if err != nil {
				return fmt.Errorf("invalid --valid-until: %w", err)
			}

			if period := validityPeriod(validFrom, validUntil); period != nil {
				input["validity"] = period
			}

			data, err := client.DoUpload(
				uploadMutation,
				map[string]any{"input": input},
				"variables.input.file",
				filepath.Base(filePath),
				file,
			)
			if err != nil {
				return err
			}

			var resp uploadResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			node := resp.UploadThirdPartyAgreement.ThirdPartyAgreementEdge.Node
			_, _ = fmt.Fprintf(f.IOStreams.Out, "Uploaded agreement %s (%s)\n", node.ID, node.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagThirdParty, "third-party", "", "Third party ID")
	cmd.Flags().StringVar(&flagName, "name", "", "Agreement name")
	cmd.Flags().StringVar(&flagType, "type", "", "Agreement type (TERMS, MASTER_SERVICE, BUSINESS_ASSOCIATE, DATA_PRIVACY, AMENDMENT, ANNEX, OTHER)")
	cmd.Flags().StringVar(&flagValidFrom, "valid-from", "", "Validity start (YYYY-MM-DD or RFC3339)")
	cmd.Flags().StringVar(&flagValidUntil, "valid-until", "", "Validity end (YYYY-MM-DD or RFC3339)")
	_ = cmd.MarkFlagRequired("third-party")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

func normalizeDatetime(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	if _, err := time.Parse(time.RFC3339, value); err == nil {
		return value, nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return "", fmt.Errorf("use YYYY-MM-DD or RFC3339")
	}

	return parsed.UTC().Format(time.RFC3339), nil
}

func validityPeriod(validFrom, validUntil string) map[string]any {
	period := map[string]any{}
	if validFrom != "" {
		period["start"] = validFrom
	}
	if validUntil != "" {
		period["end"] = validUntil
	}
	if len(period) == 0 {
		return nil
	}

	return period
}
