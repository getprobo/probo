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

package create

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
)

const createMutation = `
mutation($input: CreateBusinessFunctionInput!) {
  createBusinessFunction(input: $input) {
    businessFunctionEdge {
      node {
        id
        referenceId
        name
        classification
        mtdMinutes
        rtoMinutes
        rpoMinutes
      }
    }
  }
}
`

type createResponse struct {
	CreateBusinessFunction struct {
		BusinessFunctionEdge struct {
			Node struct {
				ID             string `json:"id"`
				ReferenceID    string `json:"referenceId"`
				Name           string `json:"name"`
				Classification string `json:"classification"`
				MTDMinutes     int    `json:"mtdMinutes"`
				RTOMinutes     int    `json:"rtoMinutes"`
				RPOMinutes     int    `json:"rpoMinutes"`
			} `json:"node"`
		} `json:"businessFunctionEdge"`
	} `json:"createBusinessFunction"`
}

func NewCmdCreate(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg             string
		flagReferenceID     string
		flagName            string
		flagClassification  string
		flagMTDMinutes      int
		flagRTOMinutes      int
		flagRPOMinutes      int
		flagImpactTolerance string
		flagNotes           string
		flagOwner           string
		flagAssetIDs        []string
		flagThirdPartyIDs   []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new business function",
		Example: `  # Create a business function interactively
  prb business-function create

  # Create a business function non-interactively
  prb business-function create --reference-id BF-001 --name "Payments" --classification CRITICAL --mtd-minutes 60 --rto-minutes 30 --rpo-minutes 15`,
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			if f.IOStreams.IsInteractive() {
				if flagReferenceID == "" {
					err := huh.NewInput().
						Title("Reference ID").
						Value(&flagReferenceID).
						Run()
					if err != nil {
						return err
					}
				}

				if flagName == "" {
					err := huh.NewInput().
						Title("Name").
						Value(&flagName).
						Run()
					if err != nil {
						return err
					}
				}

				if flagClassification == "" {
					err := huh.NewSelect[string]().
						Title("Classification").
						Options(
							huh.NewOption("Critical", "CRITICAL"),
							huh.NewOption("Important", "IMPORTANT"),
							huh.NewOption("Secondary", "SECONDARY"),
							huh.NewOption("Standard", "STANDARD"),
						).
						Value(&flagClassification).
						Run()
					if err != nil {
						return err
					}
				}

				if !cmd.Flags().Changed("mtd-minutes") {
					var mtd string

					err := huh.NewInput().
						Title("MTD (minutes)").
						Value(&mtd).
						Run()
					if err != nil {
						return err
					}

					flagMTDMinutes, err = strconv.Atoi(mtd)
					if err != nil {
						return fmt.Errorf("mtd-minutes must be an integer")
					}
				}

				if !cmd.Flags().Changed("rto-minutes") {
					var rto string

					err := huh.NewInput().
						Title("RTO (minutes)").
						Value(&rto).
						Run()
					if err != nil {
						return err
					}

					flagRTOMinutes, err = strconv.Atoi(rto)
					if err != nil {
						return fmt.Errorf("rto-minutes must be an integer")
					}
				}

				if !cmd.Flags().Changed("rpo-minutes") {
					var rpo string

					err := huh.NewInput().
						Title("RPO (minutes)").
						Value(&rpo).
						Run()
					if err != nil {
						return err
					}

					flagRPOMinutes, err = strconv.Atoi(rpo)
					if err != nil {
						return fmt.Errorf("rpo-minutes must be an integer")
					}
				}
			}

			if flagReferenceID == "" {
				return fmt.Errorf("reference ID is required; pass --reference-id or run interactively")
			}

			if flagName == "" {
				return fmt.Errorf("name is required; pass --name or run interactively")
			}

			if flagClassification == "" {
				return fmt.Errorf("classification is required; pass --classification or run interactively")
			}

			if err := cmdutil.ValidateEnum(
				"classification",
				flagClassification,
				[]string{"CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"},
			); err != nil {
				return err
			}

			if !f.IOStreams.IsInteractive() {
				if !cmd.Flags().Changed("mtd-minutes") {
					return fmt.Errorf("mtd-minutes is required")
				}

				if !cmd.Flags().Changed("rto-minutes") {
					return fmt.Errorf("rto-minutes is required")
				}

				if !cmd.Flags().Changed("rpo-minutes") {
					return fmt.Errorf("rpo-minutes is required")
				}
			}

			input := map[string]any{
				"organizationId": flagOrg,
				"referenceId":    flagReferenceID,
				"name":           flagName,
				"classification": flagClassification,
				"mtdMinutes":     flagMTDMinutes,
				"rtoMinutes":     flagRTOMinutes,
				"rpoMinutes":     flagRPOMinutes,
			}

			if flagImpactTolerance != "" {
				input["impactTolerance"] = flagImpactTolerance
			}

			if flagNotes != "" {
				input["notes"] = flagNotes
			}

			if flagOwner != "" {
				input["ownerId"] = flagOwner
			}

			if len(flagAssetIDs) > 0 {
				input["assetIds"] = flagAssetIDs
			}

			if len(flagThirdPartyIDs) > 0 {
				input["thirdPartyIds"] = flagThirdPartyIDs
			}

			data, err := client.Do(
				createMutation,
				map[string]any{"input": input},
			)
			if err != nil {
				return err
			}

			var resp createResponse
			if err := json.Unmarshal(data, &resp); err != nil {
				return fmt.Errorf("cannot parse response: %w", err)
			}

			bf := resp.CreateBusinessFunction.BusinessFunctionEdge.Node
			_, _ = fmt.Fprintf(
				f.IOStreams.Out,
				"Created business function %s (%s)\n",
				bf.ID,
				bf.Name,
			)

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().StringVar(&flagReferenceID, "reference-id", "", "Reference ID (required)")
	cmd.Flags().StringVar(&flagName, "name", "", "Business function name (required)")
	cmd.Flags().StringVar(&flagClassification, "classification", "", "Classification: CRITICAL, IMPORTANT, SECONDARY, STANDARD (required)")
	cmd.Flags().IntVar(&flagMTDMinutes, "mtd-minutes", 0, "Maximum tolerable downtime in minutes (required)")
	cmd.Flags().IntVar(&flagRTOMinutes, "rto-minutes", 0, "Recovery time objective in minutes (required)")
	cmd.Flags().IntVar(&flagRPOMinutes, "rpo-minutes", 0, "Recovery point objective in minutes (required)")
	cmd.Flags().StringVar(&flagImpactTolerance, "impact-tolerance", "", "Impact tolerance")
	cmd.Flags().StringVar(&flagNotes, "notes", "", "Notes")
	cmd.Flags().StringVar(&flagOwner, "owner", "", "Owner profile ID")
	cmd.Flags().StringArrayVar(&flagAssetIDs, "asset", nil, "Asset ID (can be repeated)")
	cmd.Flags().StringArrayVar(&flagThirdPartyIDs, "third-party", nil, "Third party ID (can be repeated)")

	return cmd
}
