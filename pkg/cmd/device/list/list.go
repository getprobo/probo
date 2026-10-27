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
	"fmt"

	"github.com/spf13/cobra"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/cli/api"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/device/shared"
)

const (
	listQuery = `
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DeviceOrder, $filter: DeviceFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      devices(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
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
query($id: ID!, $first: Int, $after: CursorKey, $orderBy: DeviceOrder, $filter: DeviceFilter) {
  node(id: $id) {
    __typename
    ... on Organization {
      devices(first: $first, after: $after, orderBy: $orderBy, filter: $filter) {
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
              version
              agentVersion
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
		flagOrg             string
		flagLimit           int
		flagAfter           string
		flagOrderBy         string
		flagOrderDir        string
		flagQuery           string
		flagState           []string
		flagPlatform        []string
		flagAgentVersion    []string
		flagPostureWarnings bool
		flagPostures        bool
		flagOutput          *string
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
  prb device ls --order-by LAST_SEEN_AT --json

  # Search devices by owner name
  prb device ls --query "Ada"

  # List active macOS devices
  prb device ls --state ACTIVE --platform DARWIN

  # List active and pending devices
  prb device ls --state ACTIVE --state PENDING`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidateOutputFlag(flagOutput); err != nil {
				return err
			}

			if err := cmdutil.ValidateLimit(flagLimit); err != nil {
				return err
			}

			client, hc, err := f.GraphQLClient(cmdutil.ConsoleEndpoint)
			if err != nil {
				return err
			}

			flagOrg, err = cmdutil.ResolveOrganization(flagOrg, hc.Organization)
			if err != nil {
				return err
			}

			variables := map[string]any{
				"id": flagOrg,
			}

			filter := map[string]any{}

			if flagQuery != "" {
				filter["query"] = flagQuery
			}

			if len(flagState) > 0 {
				for _, state := range flagState {
					if err := cmdutil.ValidateEnum(
						"state",
						state,
						[]string{"PENDING", "ACTIVE", "REVOKED"},
					); err != nil {
						return err
					}
				}

				filter["states"] = flagState
			}

			if len(flagPlatform) > 0 {
				for _, platform := range flagPlatform {
					if err := cmdutil.ValidateEnum(
						"platform",
						platform,
						[]string{"DARWIN", "LINUX", "FREEBSD", "WINDOWS"},
					); err != nil {
						return err
					}
				}

				filter["platforms"] = flagPlatform
			}

			if len(flagAgentVersion) > 0 {
				filter["agentVersions"] = flagAgentVersion
			}

			if flagPostureWarnings {
				filter["hasPostureWarning"] = true
			}

			if len(filter) > 0 {
				variables["filter"] = filter
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

			devices, totalCount, nextCursor, err := api.Paginate(
				client,
				query,
				variables,
				flagLimit,
				flagAfter,
				api.NodeConnection[device]("devices", "Organization", fmt.Sprintf("organization %s not found", flagOrg)),
			)
			if err != nil {
				return err
			}

			headers := []string{"ID", "STATE", "HOSTNAME", "PLATFORM"}
			if flagPostures {
				headers = append(headers, "POSTURES")
			}

			return cmdutil.PrintList(
				f.IOStreams,
				*flagOutput,
				devices,
				totalCount,
				cmdutil.ListView[device]{
					Empty:      "No devices found.",
					Noun:       "devices",
					NextCursor: nextCursor,
					Headers:    headers,
					Row: func(d device) ([]string, error) {
						row := []string{
							d.ID,
							d.State,
							ref.UnrefOrZero(d.Hostname),
							ref.UnrefOrZero(d.Platform),
						}
						if flagPostures {
							row = append(row, fmt.Sprintf("%d", len(d.LatestPostures)))
						}

						return row, nil
					},
				},
			)
		},
	}

	cmdutil.AddOrganizationFlag(cmd, &flagOrg, "Organization ID")
	cmd.Flags().IntVarP(&flagLimit, "limit", "L", 30, "Maximum number of devices to list")
	cmdutil.AddAfterFlag(cmd, &flagAfter)
	cmd.Flags().StringVar(&flagQuery, "query", "", "Search devices by owner name")
	cmd.Flags().StringSliceVar(&flagState, "state", nil, "Filter by state; repeat or comma-separate for multiple (PENDING, ACTIVE, REVOKED)")
	cmd.Flags().StringSliceVar(&flagPlatform, "platform", nil, "Filter by platform; repeat or comma-separate for multiple (DARWIN, LINUX, FREEBSD, WINDOWS)")
	cmd.Flags().StringSliceVar(&flagAgentVersion, "agent-version", nil, "Filter by agent version; repeat or comma-separate for multiple")
	cmd.Flags().BoolVar(
		&flagPostureWarnings,
		"posture-warnings",
		false,
		"Only devices whose latest posture for some check is FAIL or UNKNOWN",
	)
	cmd.Flags().StringVar(&flagOrderBy, "order-by", "", "Order by field (CREATED_AT, UPDATED_AT, HOSTNAME, LAST_SEEN_AT)")
	cmd.Flags().StringVar(&flagOrderDir, "order-direction", "DESC", "Sort direction (ASC, DESC)")
	cmd.Flags().BoolVar(&flagPostures, "postures", false, "Include latest posture check results")
	flagOutput = cmdutil.AddOutputFlag(cmd)

	return cmd
}
