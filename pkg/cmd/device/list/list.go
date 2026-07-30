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

package list

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/device/shared"
)

const (
	listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DeviceOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      devices(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            state
            hostname
            platform
            osVersion
            agentVersion
            serialNumber
            lastSeenAt
            enrolledAt
            owner {
              id
            }
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
`

	listWithPosturesQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DeviceOrder) {
  node(id: $id) {
    __typename
    ... on Organization {
      devices(first: $first, after: $after, orderBy: $orderBy) {
        totalCount
        edges {
          node {
            id
            state
            hostname
            platform
            osVersion
            agentVersion
            serialNumber
            lastSeenAt
            enrolledAt
            owner {
              id
            }
            latestPostures {
              id
              checkKey
              status
              value {
                kind
                text
                number
              }
              observedAt
            }
          }
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }
}
`
)

type device struct {
	ID           string  `json:"id"`
	State        string  `json:"state"`
	Hostname     *string `json:"hostname"`
	Platform     *string `json:"platform"`
	OsVersion    *string `json:"osVersion"`
	AgentVersion *string `json:"agentVersion"`
	SerialNumber *string `json:"serialNumber"`
	LastSeenAt   *string `json:"lastSeenAt"`
	EnrolledAt   *string `json:"enrolledAt"`
	Owner        *struct {
		ID string `json:"id"`
	} `json:"owner"`
	LatestPostures []shared.Posture `json:"latestPostures"`
}

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
	var (
		flagOrg      string
		flagLimit    int
		flagOrderBy  string
		flagOrderDir string
		flagPostures bool
		flagOutput   *string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List ITAM devices in an organization",
		Aliases: []string{"ls"},
		Example: `  # List devices in the default organization
  prb device list

  # List devices with latest posture check results
  prb device ls --postures

  # List devices sorted by last seen time
  prb device ls --order-by LAST_SEEN_AT --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
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

			if flagOrg == "" {
				flagOrg = hc.Organization
			}

			if flagOrg == "" {
				return fmt.Errorf("organization is required; pass --org or set a default with 'prb auth login'")
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			if flagOrderBy != "" {
				if err := cmdutil.ValidateEnum(
					"order-by",
					flagOrderBy,
					[]string{"CREATED_AT", "UPDATED_AT", "HOSTNAME", "LAST_SEEN_AT"},
				); err != nil {
					return err
				}

				if err := cmdutil.ValidateEnum(
					"order-direction",
					flagOrderDir,
					[]string{"ASC", "DESC"},
				); err != nil {
					return err
				}

				variables["orderBy"] = map[string]any{
					"field":     flagOrderBy,
					"direction": flagOrderDir,
				}
			}

			query := listQuery
			if flagPostures {
				query = listWithPosturesQuery
			}

			devices, totalCount, err := api.Paginate(
				client,
				query,
				variables,
				flagLimit,
				func(data json.RawMessage) (*api.Connection[device], error) {
					var resp struct {
						Node *struct {
							Typename string                 `json:"__typename"`
							Devices  api.Connection[device] `json:"devices"`
						} `json:"node"`
					}
					if err := json.Unmarshal(data, &resp); err != nil {
						return nil, err
					}

					if resp.Node == nil {
						return nil, fmt.Errorf("organization %s not found", flagOrg)
					}

					if resp.Node.Typename != "Organization" {
						return nil, fmt.Errorf("expected Organization node, got %s", resp.Node.Typename)
					}

					return &resp.Node.Devices, nil
				},
			)
			if err != nil {
				return err
			}

			if *flagOutput == cmdutil.OutputJSON {
				return cmdutil.PrintJSON(f.IOStreams.Out, devices)
			}

			if len(devices) == 0 {
				_, _ = fmt.Fprintln(f.IOStreams.Out, "No devices found.")
				return nil
			}

			headers := []string{"ID", "STATE", "HOSTNAME", "PLATFORM"}
			if flagPostures {
				headers = append(headers, "POSTURES")
			}

			rows := make([][]string, 0, len(devices))
			for _, d := range devices {
				row := []string{
					d.ID,
					d.State,
					ref.UnrefOrZero(d.Hostname),
					ref.UnrefOrZero(d.Platform),
				}
				if flagPostures {
					row = append(row, fmt.Sprintf("%d", len(d.LatestPostures)))
				}

				rows = append(rows, row)
			}

			t := cmdutil.NewTable(headers...).Rows(rows...)

			_, _ = fmt.Fprintln(f.IOStreams.Out, t)

			if totalCount > len(devices) {
				_, _ = fmt.Fprintf(
					f.IOStreams.ErrOut,
					"\nShowing %d of %d devices\n",
					len(devices),
					totalCount,
				)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&flagOrg, "org", "", "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of devices to list")
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, UPDATED_AT, HOSTNAME, LAST_SEEN_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().BoolVar(&flagPostures, "postures", false, "Include latest posture check results")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
