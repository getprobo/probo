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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestSlackReads_ViewerSoftFail(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	portalID := factory.CreateCompliancePortal(
		owner,
		factory.Attrs{"entityName": factory.SafeName("Slack read RBAC portal")},
	)

	const query = `
		query SlackReadRBAC($organizationId: ID!, $compliancePortalId: ID!) {
			organization: node(id: $organizationId) {
				... on Organization {
					canConnectSlack: permission(action: "core:connector:initiate")
					canUninstallSlack: permission(action: "core:connector:delete")
					slackbotAvailable
					slackbotInstallation {
						active
					}
					slackbotChannels {
						channels {
							id
							name
						}
					}
				}
			}
			compliancePortal: node(id: $compliancePortalId) {
				... on CompliancePortal {
					slackbotNotificationChannel {
						channelId
					}
					organization {
						slackbotInstallation {
							active
						}
						slackbotChannels {
							channels {
								id
								name
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Organization *struct {
			CanConnectSlack      bool `json:"canConnectSlack"`
			CanUninstallSlack    bool `json:"canUninstallSlack"`
			SlackbotInstallation *struct {
				Active bool `json:"active"`
			} `json:"slackbotInstallation"`
			SlackbotChannels struct {
				Channels []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"channels"`
			} `json:"slackbotChannels"`
		} `json:"organization"`
		CompliancePortal *struct {
			SlackbotNotificationChannel *struct {
				ChannelID string `json:"channelId"`
			} `json:"slackbotNotificationChannel"`
			Organization struct {
				SlackbotInstallation *struct {
					Active bool `json:"active"`
				} `json:"slackbotInstallation"`
				SlackbotChannels struct {
					Channels []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"channels"`
				} `json:"slackbotChannels"`
			} `json:"organization"`
		} `json:"compliancePortal"`
	}

	err := viewer.Execute(
		query,
		map[string]any{
			"organizationId":     viewer.GetOrganizationID().String(),
			"compliancePortalId": portalID,
		},
		&result,
	)
	require.NoError(t, err)
	require.NotNil(t, result.Organization)
	require.NotNil(t, result.CompliancePortal)
	assert.False(t, result.Organization.CanConnectSlack)
	assert.False(t, result.Organization.CanUninstallSlack)
	assert.Nil(t, result.Organization.SlackbotInstallation)
	assert.Empty(t, result.Organization.SlackbotChannels.Channels)
	assert.Nil(t, result.CompliancePortal.SlackbotNotificationChannel)
	assert.Nil(t, result.CompliancePortal.Organization.SlackbotInstallation)
	assert.Empty(t, result.CompliancePortal.Organization.SlackbotChannels.Channels)
}
