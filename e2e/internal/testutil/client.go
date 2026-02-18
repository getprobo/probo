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
	"net/url"
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
	mailpitBaseURL string
	role           TestRole
	userID         gid.GID
	organizationID gid.GID
}

func NewClient(t testing.TB, role TestRole) *Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "cannot create cookie jar")

	client := &Client{
		T:              t,
		baseURL:        GetBaseURL(),
		mailpitBaseURL: GetMailpitBaseURL(),
		role:           role,
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
		mailpitBaseURL: GetMailpitBaseURL(),
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
	c.organizationID = c.createOrganization(orgName)

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

	// Owner invites user to organization
	profileID, identityID := ownerClient.createUser(email, fullName, coredata.MembershipRole(c.role))
	c.userID = identityID
	ownerClient.inviteUser(profileID)
	token := c.getActivationToken(email)
	passwordToken := c.activateUser(token)
	c.resetPassword(password, passwordToken)
	c.signIn(email, password)

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

func (c *Client) signIn(email string, password string) {
	const query = `
		mutation($input: SignInInput!) {
			signIn(input: $input) {
				identity { id }
			}
		}
	`

	var result struct {
		SignIn struct {
			Identity struct {
				ID string `json:"id"`
			} `json:"identity"`
		} `json:"signIn"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"email":    email,
			"password": password,
		},
	}, &result)
	require.NoError(c.T, err, "signIn mutation failed")
}

func (c *Client) createOrganization(name string) gid.GID {
	const query = `
		mutation($input: CreateOrganizationInput!) {
			createOrganization(input: $input) {
				organization { id }
				profile { id }
			}
		}
	`

	var result struct {
		CreateOrganization struct {
			Organization struct {
				ID string `json:"id"`
			} `json:"organization"`
		} `json:"createOrganization"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{"name": name},
	}, &result)
	require.NoError(c.T, err, "createOrganization mutation failed")

	orgID, err := gid.ParseGID(result.CreateOrganization.Organization.ID)
	require.NoError(c.T, err, "cannot parse organization ID")

	return orgID
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

func (c *Client) createUser(email, fullName string, role coredata.MembershipRole) (gid.GID, gid.GID) {
	const query = `
		mutation($input: CreateUserInput!) {
			createUser(input: $input) {
				profileEdge {
					node {
						id
						identity {
							id
						}
					}
				}
			}
		}
	`

	var result struct {
		CreateUser struct {
			ProfileEdge struct {
				Node struct {
					ID       string `json:"id"`
					Identity struct {
						ID string `json:"id"`
					} `json:"identity"`
				} `json:"node"`
			} `json:"profileEdge"`
		} `json:"createUser"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"organizationId":           c.organizationID.String(),
			"emailAddress":             email,
			"fullName":                 fullName,
			"role":                     string(role),
			"kind":                     coredata.MembershipProfileKindEmployee,
			"additionalEmailAddresses": []string{},
		},
	}, &result)
	require.NoError(c.T, err, "createUser mutation failed")

	profileID, err := gid.ParseGID(result.CreateUser.ProfileEdge.Node.ID)
	require.NoError(c.T, err, "cannot parse profile ID")

	identityID, err := gid.ParseGID(result.CreateUser.ProfileEdge.Node.Identity.ID)
	require.NoError(c.T, err, "cannot parse identity ID")

	return profileID, identityID
}

func (c *Client) inviteUser(profileID gid.GID) {
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

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"organizationId": c.organizationID.String(),
			"profileId":      profileID.String(),
		},
	}, &result)
	require.NoError(c.T, err, "inviteUser mutation failed")
}

func (c *Client) getActivationToken(email string) string {
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		searchMails, err := c.SearchMails(fmt.Sprintf("to:%s subject:\"Invitation to join\"", email))
		require.NoError(c.T, err, "mailpit messages search failed")

		for _, msg := range searchMails.Messages {
			linksCheck, err := c.CheckMessageLinks(msg.ID)
			require.NoError(c.T, err, "mailpit link check failed")

			for _, link := range linksCheck.Links {
				linkURL, err := url.Parse(link.URL)
				require.NoError(c.T, err, "mailpit link invalid URL")

				query := linkURL.Query()
				if query.Get("token") != "" {
					return query.Get("token")
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	c.T.Logf("activation token not found")
	c.T.FailNow()
	return ""
}

func (c *Client) activateUser(token string) string {
	const query = `
		mutation($input: ActivateAccountInput!) {
			activateAccount(input: $input) {
				createPasswordToken
			}
		}
	`

	var result struct {
		ActivateAccount struct {
			CreatePasswordToken string `json:"createPasswordToken"`
		} `json:"activateAccount"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"token": token,
		},
	}, &result)
	require.NoError(c.T, err, "activateAccount mutation failed")

	return result.ActivateAccount.CreatePasswordToken
}

func (c *Client) resetPassword(password string, token string) {
	const query = `
		mutation($input: ResetPasswordInput!) {
			resetPassword(input: $input) {
				success
			}
		}
	`

	var result struct {
		ResetPassword struct {
			Success bool `json:"success"`
		} `json:"resetPassword"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"token":    token,
			"password": password,
		},
	}, &result)
	require.NoError(c.T, err, "resetPassword mutation failed")
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

func (c *Client) GetOrganizationID() gid.GID {
	return c.organizationID
}

func (c *Client) GetRole() TestRole {
	return c.role
}
