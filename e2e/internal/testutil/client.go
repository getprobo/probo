// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package testutil

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
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
	RoleOwner    TestRole = "OWNER"
	RoleAdmin    TestRole = "ADMIN"
	RoleViewer   TestRole = "VIEWER"
	RoleEmployee TestRole = "EMPLOYEE"
	RoleAuditor  TestRole = "AUDITOR"
)

type Client struct {
	T               testing.TB
	httpClient      *http.Client
	proboHTTPClient *http.Client
	trustClient     *http.Client
	trustHost       string
	baseURL         string
	mailpitBaseURL  string
	role            TestRole
	userID          gid.GID
	profileID       gid.GID
	organizationID  gid.GID
	email           string
	password        string
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

	c.email = email
	c.password = password

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

	c.email = email
	c.password = password

	// Owner invites user to organization
	profileID, identityID := ownerClient.createUser(email, fullName, coredata.MembershipRole(c.role))
	c.userID = identityID
	c.profileID = profileID
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
			Profile struct {
				ID string `json:"id"`
			} `json:"profile"`
		} `json:"createOrganization"`
	}

	err := c.ExecuteConnect(query, map[string]any{
		"input": map[string]any{"name": name},
	}, &result)
	require.NoError(c.T, err, "createOrganization mutation failed")

	orgID, err := gid.ParseGID(result.CreateOrganization.Organization.ID)
	require.NoError(c.T, err, "cannot parse organization ID")

	profileID, err := gid.ParseGID(result.CreateOrganization.Profile.ID)
	require.NoError(c.T, err, "cannot parse profile ID")

	c.profileID = profileID

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
			"kind":                     "EMPLOYEE",
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
	return c.pollForLinkToken(fmt.Sprintf("to:%s subject:\"Invitation to join\"", email))
}

// pollForLinkToken polls mailpit for a message matching searchQuery and
// returns the first "token" query parameter found among its links.
func (c *Client) pollForLinkToken(searchQuery string) string {
	deadline := time.Now().Add(10 * time.Second)

	for time.Now().Before(deadline) {
		searchMails, err := c.SearchMails(searchQuery)
		require.NoError(c.T, err, "mailpit messages search failed")

		for _, msg := range searchMails.Messages {
			linksCheck, err := c.CheckMessageLinks(msg.ID)
			require.NoError(c.T, err, "mailpit link check failed")

			for _, link := range linksCheck.Links {
				linkURL, err := url.Parse(link.URL)
				require.NoError(c.T, err, "mailpit link invalid URL")

				if token := linkURL.Query().Get("token"); token != "" {
					return token
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	c.T.Logf("link token not found for query %q", searchQuery)
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

// NewClientWithNewSession creates a new Client that signs in as the same
// identity but with a fresh HTTP session (new cookie jar). This is useful for
// testing session-scoped authorization.
func NewClientWithNewSession(t testing.TB, from *Client) *Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "cannot create cookie jar")

	client := &Client{
		T:              t,
		baseURL:        from.baseURL,
		mailpitBaseURL: from.mailpitBaseURL,
		role:           from.role,
		userID:         from.userID,
		organizationID: from.organizationID,
		email:          from.email,
		password:       from.password,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}

	client.signIn(client.email, client.password)

	return client
}

func SelfProvisionTrustCenterVisitor(t testing.TB, trustHost string) *Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err, "cannot create cookie jar")

	email := fmt.Sprintf("visitor-%s@e2e.probo.test", generateUniqueID())

	visitor := &Client{
		T:              t,
		baseURL:        GetBaseURL(),
		mailpitBaseURL: GetMailpitBaseURL(),
		email:          email,
		trustHost:      trustHost,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
		trustClient: trustHTTPClientWithJar(trustHost, jar),
		proboHTTPClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}

	visitor.connectViaCIMD(email)

	return visitor
}

func (c *Client) connectViaCIMD(email string) {
	c.T.Helper()

	WaitForTrustCenterHTTPS(c.T, c.trustHost)

	initiateURL := fmt.Sprintf(
		"https://%s/initiate?continue=/overview",
		c.trustHost,
	)

	authorizeURL := c.redirectLocation(c.trustClient, initiateURL)
	require.NotEmpty(c.T, authorizeURL, "oauth initiate must redirect to authorize")

	loginURL := c.redirectLocation(c.proboHTTPClient, authorizeURL)
	require.Contains(c.T, loginURL, "/auth/login", "unauthenticated authorize must redirect to login")
	require.Contains(c.T, loginURL, "continue=", "login redirect must preserve continue URL")

	continueURL := extractContinueQueryParam(loginURL)
	require.NotEmpty(c.T, continueURL)
	require.Contains(c.T, continueURL, "/api/connect/v1/oauth2/authorize")

	c.postConnectMagicLink(email, continueURL)

	token := c.pollForLinkToken(fmt.Sprintf("to:%s", email))
	verifyURL := c.baseURL + "/api/connect/v1/magic-link/verify?token=" + url.QueryEscape(token)

	resumeAuthorizeURL := c.redirectLocation(c.proboHTTPClient, verifyURL)
	require.Contains(c.T, resumeAuthorizeURL, "/api/connect/v1/oauth2/authorize")

	authorizeResp := c.redirectHTTPResponse(c.proboHTTPClient, resumeAuthorizeURL)
	require.False(
		c.T,
		IsConsentRedirect(authorizeResp),
		"compliance portal CIMD must skip oauth consent screen",
	)
	require.Equal(c.T, http.StatusFound, authorizeResp.StatusCode)

	_, err := OAuth2AuthorizeCodeFromRedirect(authorizeResp)
	require.NoError(c.T, err, "compliance portal CIMD must issue authorization code without consent")

	callbackURL := resolveRedirectURL(resumeAuthorizeURL, authorizeResp.Header.Get("Location"))
	require.Contains(c.T, callbackURL, "/callback")

	finalURL := c.redirectLocation(c.trustClient, callbackURL)
	require.True(
		c.T,
		strings.HasSuffix(finalURL, "/overview") || strings.Contains(finalURL, "/overview"),
		"oauth callback must redirect to continue URL, got %q",
		finalURL,
	)
}

func (c *Client) postConnectMagicLink(email, continueURL string) {
	c.T.Helper()

	body := url.Values{}
	body.Set("email", email)
	body.Set("continue", continueURL)

	req, err := http.NewRequest(
		"POST",
		c.baseURL+"/api/connect/v1/magic-link/send",
		strings.NewReader(body.Encode()),
	)
	require.NoError(c.T, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.proboHTTPClient.Do(req)
	require.NoError(c.T, err, "magic-link send request failed")

	defer func() { _ = resp.Body.Close() }()

	require.Equal(c.T, http.StatusNoContent, resp.StatusCode, "magic-link send must return 204")
}

func (c *Client) redirectLocation(client *http.Client, rawURL string) string {
	c.T.Helper()

	resp := c.redirectHTTPResponse(client, rawURL)
	require.True(
		c.T,
		resp.StatusCode >= http.StatusMultipleChoices && resp.StatusCode < http.StatusBadRequest,
		"expected redirect from %s, got %d",
		rawURL,
		resp.StatusCode,
	)

	location := resp.Header.Get("Location")
	require.NotEmpty(c.T, location, "redirect from %s missing Location header", rawURL)

	return resolveRedirectURL(rawURL, location)
}

func (c *Client) redirectHTTPResponse(client *http.Client, rawURL string) *OAuth2HTTPResponse {
	c.T.Helper()

	noRedirectClient := &http.Client{
		Jar:       client.Jar,
		Timeout:   client.Timeout,
		Transport: client.Transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	require.NoError(c.T, err)

	resp, err := noRedirectClient.Do(req)
	require.NoError(c.T, err, "request to %s failed", rawURL)

	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(c.T, err)

	return &OAuth2HTTPResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       body,
	}
}

func resolveRedirectURL(baseURL, location string) string {
	locURL, err := url.Parse(location)
	if err != nil {
		return location
	}

	if locURL.IsAbs() {
		return locURL.String()
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return location
	}

	return base.ResolveReference(locURL).String()
}

func extractContinueQueryParam(loginURL string) string {
	parsed, err := url.Parse(loginURL)
	if err != nil {
		return ""
	}

	return parsed.Query().Get("continue")
}

func (c *Client) GetEmail() string {
	return c.email
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
