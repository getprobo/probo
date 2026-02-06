// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package testutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func generateUniqueID() string {
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes))
}

type TestRole string

const (
	RoleOwner  TestRole = "OWNER"
	RoleAdmin  TestRole = "ADMIN"
	RoleViewer TestRole = "VIEWER"
)

type Client struct {
	T              testing.TB
	httpClient     *http.Client
	baseURL        string
	role           TestRole
	userID         gid.GID
	profileID      gid.GID
	organizationID gid.GID
}

func NewClient(t testing.TB, role TestRole) *Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "cannot create cookie jar")

	client := &Client{
		T:       t,
		baseURL: GetBaseURL(),
		role:    role,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}

	client.setupTestUser()

	return client
}

func NewClientInOrg(t testing.TB, role TestRole, ownerClient *Client) *Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "cannot create cookie jar")

	client := &Client{
		T:              t,
		baseURL:        GetBaseURL(),
		role:           role,
		organizationID: ownerClient.organizationID,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}

	client.SetupTestUserInOrg(ownerClient)

	return client
}

func (c *Client) setupTestUser() {
	uniqueID := generateUniqueID()
	email := fmt.Sprintf("test-%s@e2e.probo.test", uniqueID)
	password := "TestPassword123!"
	fullName := fmt.Sprintf("Test User %s", uniqueID)

	// Sign up
	c.userID = c.signUp(email, password, fullName)

	// Create organization (this makes the user an OWNER)
	orgName := fmt.Sprintf("Test Org %s", uniqueID)
	c.organizationID, c.profileID = c.createOrganization(orgName)

	// Assume organization session to use console API
	c.assumeOrganizationSession()

	// If the role is not OWNER, we need to adjust the membership
	if c.role != RoleOwner {
		c.updateOwnMembershipRole(coredata.MembershipRole(c.role))
	}
}

func (c *Client) SetupTestUserInOrg(ownerClient *Client) {
	uniqueID := generateUniqueID()
	email := fmt.Sprintf("test-%s@e2e.probo.test", uniqueID)
	password := "TestPassword123!"
	fullName := fmt.Sprintf("Test User %s", uniqueID)

	// Sign up new user
	c.userID = c.signUp(email, password, fullName)

	// Owner invites user to organization
	invitationID := ownerClient.inviteMember(email, fullName, coredata.MembershipRole(c.role))

	// New user accepts invitation
	c.profileID = c.acceptInvitation(invitationID)

	// Assume organization session to use console API
	c.assumeOrganizationSession()
}

func (c *Client) signUp(email, password, fullName string) gid.GID {
	const query = `
		mutation($input: SignUpInput!) {
			signUp(input: $input) {
				identity { id }
			}
		}
	`

	var result struct {
		SignUp struct {
			Identity struct {
				ID string `json:"id"`
			} `json:"identity"`
		} `json:"signUp"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"email":    email,
			"password": password,
			"fullName": fullName,
		},
	}, &result)
	require.NoError(c.T, err, "signUp mutation failed")

	userID, err := gid.ParseGID(result.SignUp.Identity.ID)
	require.NoError(c.T, err, "cannot parse user ID")

	return userID
}

func (c *Client) createOrganization(name string) (gid.GID, gid.GID) {
	const query = `
		mutation($input: CreateOrganizationInput!) {
			createOrganization(input: $input) {
				organization { id }
				membershipEdge {
					node {
						profile { id }
					}
				}
			}
		}
	`

	var result struct {
		CreateOrganization struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
			MembershipEdge struct {
				Node struct {
					ID      string `json:"id"`
					Profile struct {
						ID string `json:"id"`
					} `json:"profile"`
				} `json:"node"`
			} `json:"membershipEdge"`
		} `json:"createOrganization"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{"name": name},
	}, &result)
	require.NoError(c.T, err, "createOrganization mutation failed")

	orgID, err := gid.ParseGID(result.CreateOrganization.Organization.ID)
	require.NoError(c.T, err, "cannot parse organization ID")

	profileID, err := gid.ParseGID(result.CreateOrganization.MembershipEdge.Node.Profile.ID)
	require.NoError(c.T, err, "cannot parse profile ID")

	return orgID, profileID
}

func (c *Client) updateOwnMembershipRole(role coredata.MembershipRole) {
	// First get the membership ID
	const queryMembers = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					members(first: 100) {
						edges {
							node {
								id
								identity { id }
								role
							}
						}
					}
				}
			}
		}
	`

	var qResult struct {
		Node struct {
			Members struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						Identity struct {
							ID string `json:"id"`
						} `json:"identity"`
						Role string `json:"role"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"members"`
		} `json:"node"`
	}

	err := c.ExecuteConnect(queryMembers, map[string]any{
		"id": c.organizationID.String(),
	}, &qResult)
	require.NoError(c.T, err, "cannot query organization members")

	var membershipID string
	for _, edge := range qResult.Node.Members.Edges {
		if edge.Node.Identity.ID == c.userID.String() {
			membershipID = edge.Node.ID
			break
		}
	}
	require.NotEmpty(c.T, membershipID, "membership not found for user")

	// Update the role
	const updateQuery = `
		mutation($input: UpdateMembershipInput!) {
			updateMembership(input: $input) {
				membership {
					id
					role
				}
			}
		}
	`

	err = c.ExecuteConnect(updateQuery, map[string]any{
		"input": map[string]any{
			"organizationId": c.organizationID.String(),
			"membershipId":   membershipID,
			"role":           string(role),
		},
	}, nil)
	require.NoError(c.T, err, "updateMembership mutation failed")
}

func (c *Client) inviteMember(email, fullName string, role coredata.MembershipRole) gid.GID {
	const query = `
		mutation($input: InviteMemberInput!) {
			inviteMember(input: $input) {
				invitationEdge {
					node { id }
				}
			}
		}
	`

	var result struct {
		InviteMember struct {
			InvitationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"invitationEdge"`
		} `json:"inviteMember"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"organizationId": c.organizationID.String(),
			"email":          email,
			"fullName":       fullName,
			"role":           string(role),
		},
	}, &result)
	require.NoError(c.T, err, "inviteMember mutation failed")

	invitationID, err := gid.ParseGID(result.InviteMember.InvitationEdge.Node.ID)
	require.NoError(c.T, err, "cannot parse invitation ID")

	return invitationID
}

func (c *Client) acceptInvitation(invitationID gid.GID) gid.GID {
	const query = `
		mutation($input: AcceptInvitationInput!) {
			acceptInvitation(input: $input) {
				membershipEdge {
					node {
						profile { id }
					}
				}
			}
		}
	`

	var result struct {
		AcceptInvitation struct {
			MembershipEdge struct {
				Node struct {
					Profile struct {
						ID string `json:"id"`
					} `json:"profile"`
				} `json:"node"`
			} `json:"membershipEdge"`
		} `json:"acceptInvitation"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"invitationId": invitationID.String(),
		},
	}, &result)
	require.NoError(c.T, err, "acceptInvitation mutation failed")

	profileID, err := gid.ParseGID(result.AcceptInvitation.MembershipEdge.Node.Profile.ID)
	require.NoError(c.T, err, "cannot parse profile ID")

	return profileID
}

func (c *Client) assumeOrganizationSession() {
	const query = `
		mutation($input: AssumeOrganizationSessionInput!) {
			assumeOrganizationSession(input: $input) {
				result {
					... on OrganizationSessionCreated {
						session { id }
					}
				}
			}
		}
	`

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"organizationId": c.organizationID.String(),
			"continue":       c.baseURL,
		},
	}, nil)
	require.NoError(c.T, err, "assumeOrganizationSession mutation failed")
}

func (c *Client) GetUserID() gid.GID {
	return c.userID
}

func (c *Client) GetProfileID() gid.GID {
	return c.profileID
}

func (c *Client) GetOrganizationID() gid.GID {
	return c.organizationID
}

func (c *Client) GetRole() TestRole {
	return c.role
}
