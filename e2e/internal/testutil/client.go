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
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// generateUniqueID creates a unique identifier for test isolation.
// It combines a timestamp with random bytes to ensure uniqueness
// even when tests run in parallel.
func generateUniqueID() string {
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), hex.EncodeToString(randomBytes))
}

// TestRole represents the role a test user should have
type TestRole string

const (
	RoleOwner  TestRole = "OWNER"
	RoleAdmin  TestRole = "ADMIN"
	RoleViewer TestRole = "VIEWER"
)

// Client is an authenticated HTTP client for making API requests
type Client struct {
	T              testing.TB
	httpClient     *http.Client
	baseURL        string
	role           TestRole
	userID         gid.GID
	organizationID gid.GID
}

// NewClient creates a new authenticated test client with the specified role.
// It creates a new user, organization, and sets up the membership with the given role.
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

// NewClientInOrg creates a test client for a user in an existing organization.
// The ownerClient must be an OWNER of the organization to invite the new user.
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

	client.setupTestUserInOrg(ownerClient)

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
	c.organizationID = c.createOrganization(orgName)

	// If the role is not OWNER, we need to adjust the membership
	if c.role != RoleOwner {
		c.updateOwnMembershipRole(coredata.MembershipRole(c.role))
	}
}

func (c *Client) setupTestUserInOrg(ownerClient *Client) {
	uniqueID := generateUniqueID()
	email := fmt.Sprintf("test-%s@e2e.probo.test", uniqueID)
	password := "TestPassword123!"
	fullName := fmt.Sprintf("Test User %s", uniqueID)

	// Sign up new user
	c.userID = c.signUp(email, password, fullName)

	// Owner invites user to organization
	invitationID := ownerClient.inviteMember(email, fullName, coredata.MembershipRole(c.role))

	// New user accepts invitation
	c.acceptInvitation(invitationID)
}

func (c *Client) signUp(email, password, fullName string) gid.GID {
	payload := map[string]string{
		"email":    email,
		"password": password,
		"fullName": fullName,
	}

	body, err := json.Marshal(payload)
	require.NoError(c.T, err, "cannot marshal sign-up payload")

	req, err := http.NewRequest("POST", c.baseURL+"/connect/register", bytes.NewReader(body))
	require.NoError(c.T, err, "cannot create sign-up request")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	require.NoError(c.T, err, "sign-up request failed")
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	require.Equal(c.T, http.StatusOK, resp.StatusCode, "sign-up failed: %s", string(respBody))

	var result struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}

	err = json.Unmarshal(respBody, &result)
	require.NoError(c.T, err, "cannot decode sign-up response")

	userID, err := gid.ParseGID(result.User.ID)
	require.NoError(c.T, err, "cannot parse user ID")

	return userID
}

func (c *Client) createOrganization(name string) gid.GID {
	const query = `
		mutation($input: CreateOrganizationInput!) {
			createOrganization(input: $input) {
				organizationEdge {
					node { id }
				}
			}
		}
	`

	var result struct {
		CreateOrganization struct {
			OrganizationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"organizationEdge"`
		} `json:"createOrganization"`
	}

	err := c.Execute(query, map[string]any{
		"input": map[string]any{"name": name},
	}, &result)
	require.NoError(c.T, err, "createOrganization mutation failed")

	orgID, err := gid.ParseGID(result.CreateOrganization.OrganizationEdge.Node.ID)
	require.NoError(c.T, err, "cannot parse organization ID")

	return orgID
}

func (c *Client) updateOwnMembershipRole(role coredata.MembershipRole) {
	// First get the membership ID
	const queryMemberships = `
		query($id: ID!) {
			organization(id: $id) {
				memberships(first: 100) {
					edges {
						node {
							id
							userId
							role
						}
					}
				}
			}
		}
	`

	var qResult struct {
		Organization struct {
			Memberships struct {
				Edges []struct {
					Node struct {
						ID     string `json:"id"`
						UserID string `json:"userId"`
						Role   string `json:"role"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"memberships"`
		} `json:"organization"`
	}

	err := c.Execute(queryMemberships, map[string]any{
		"id": c.organizationID.String(),
	}, &qResult)
	require.NoError(c.T, err, "cannot query organization memberships")

	var membershipID string
	for _, edge := range qResult.Organization.Memberships.Edges {
		if edge.Node.UserID == c.userID.String() {
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

	err = c.Execute(updateQuery, map[string]any{
		"input": map[string]any{
			"organizationId": c.organizationID.String(),
			"memberId":       membershipID,
			"role":           string(role),
		},
	}, nil)
	require.NoError(c.T, err, "updateMembership mutation failed")
}

func (c *Client) inviteMember(email, fullName string, role coredata.MembershipRole) gid.GID {
	const query = `
		mutation($input: InviteUserInput!) {
			inviteUser(input: $input) {
				invitationEdge {
					node { id }
				}
			}
		}
	`

	var result struct {
		InviteUser struct {
			InvitationEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"invitationEdge"`
		} `json:"inviteUser"`
	}

	err := c.Execute(query, map[string]any{
		"input": map[string]any{
			"organizationId": c.organizationID.String(),
			"email":          email,
			"fullName":       fullName,
			"role":           string(role),
			"createPeople":   false,
		},
	}, &result)
	require.NoError(c.T, err, "inviteUser mutation failed")

	invitationID, err := gid.ParseGID(result.InviteUser.InvitationEdge.Node.ID)
	require.NoError(c.T, err, "cannot parse invitation ID")

	return invitationID
}

func (c *Client) acceptInvitation(invitationID gid.GID) {
	const query = `
		mutation($input: AcceptInvitationInput!) {
			acceptInvitation(input: $input) {
				invitation {
					id
					status
				}
			}
		}
	`

	err := c.Execute(query, map[string]any{
		"input": map[string]any{
			"invitationId": invitationID.String(),
		},
	}, nil)
	require.NoError(c.T, err, "acceptInvitation mutation failed")
}

// GetUserID returns the authenticated user's ID
func (c *Client) GetUserID() gid.GID {
	return c.userID
}

// GetOrganizationID returns the test organization's ID
func (c *Client) GetOrganizationID() gid.GID {
	return c.organizationID
}

// GetRole returns the client's role
func (c *Client) GetRole() TestRole {
	return c.role
}
