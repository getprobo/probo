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

package drivers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
)

func TestGitHubDriver(t *testing.T) {
	t.Parallel()

	rec := newRecorder(t, "testdata/github", "GITHUB_TOKEN")
	client := newVCRClient(rec, bearerAuth(os.Getenv("GITHUB_TOKEN")))

	org := os.Getenv("GITHUB_ORG")
	if org == "" {
		org = "acme-corp"
	}

	driver := NewGitHubDriver(client, org, log.NewLogger(log.WithName("test")), "https://api.github.com")
	records, err := driver.ListAccounts(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, records)

	r := records[0]
	assert.Equal(t, "jane.doe@example.com", r.Email)
	assert.NotEmpty(t, r.FullName)
	assert.NotEmpty(t, r.ExternalID)
	assert.NotEmpty(t, r.Roles)
	require.NotNil(t, r.LastLogin)
	assert.Equal(t, time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC), r.LastLogin.UTC())
}

func TestGitHubDriver_PrefersVerifiedDomainEmail(t *testing.T) {
	t.Parallel()

	srv := newGitHubListAccountsServer(
		t,
		"jdoe@personal.example.com",
		`{"data":{"organization":{"membersWithRole":{"nodes":[{"login":"jdoe","organizationVerifiedDomainEmails":["jdoe@acme.example.com","jdoe@mail.example.com"]}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`,
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "jdoe@acme.example.com", records[0].Email)
}

func TestGitHubDriver_FallsBackToPublicEmail(t *testing.T) {
	t.Parallel()

	srv := newGitHubListAccountsServer(
		t,
		"jdoe@personal.example.com",
		`{"data":{"organization":{"membersWithRole":{"nodes":[{"login":"jdoe","organizationVerifiedDomainEmails":[]}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`,
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "jdoe@personal.example.com", records[0].Email)
}

func TestGitHubDriver_ContinuesWhenVerifiedDomainEmailsUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == "/graphql" {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"Resource not accessible by integration"}`))

				return
			}

			writeGitHubListAccountsREST(t, w, r, "jdoe@personal.example.com")
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "jdoe@personal.example.com", records[0].Email)
}

func TestGitHubDriver_LastLoginFromAuditLog(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/audit-log" {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[
					{"@timestamp":1774018800000,"created_at":1774018800000,"actor":"jdoe","actor_id":100001,"action":"pull_request.create"},
					{"@timestamp":1771336800000,"created_at":1771336800000,"actor":"jdoe","actor_id":100001,"action":"pull_request.merge"},
					{"@timestamp":1774105200000,"created_at":1774105200000,"actor":"someone-else","actor_id":9,"action":"pull_request.create"}
				]`))

				return
			}

			writeGitHubListAccountsREST(t, w, r, "jdoe@personal.example.com")
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.NotNil(t, records[0].LastLogin)
	assert.Equal(t, time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC), records[0].LastLogin.UTC())
}

func TestGitHubDriver_ContinuesWhenAuditLogUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/audit-log" {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"Must have admin rights to Repository."}`))

				return
			}

			writeGitHubListAccountsREST(t, w, r, "jdoe@personal.example.com")
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "jdoe@personal.example.com", records[0].Email)
	assert.Nil(t, records[0].LastLogin)
}

func TestGitHubDriver_LastLoginPaginatesAuditLog(t *testing.T) {
	t.Parallel()

	var srv *httptest.Server
	srv = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/audit-log" {
				w.Header().Set("Content-Type", "application/json")

				if r.URL.Query().Get("after") == "" {
					w.Header().Set(
						"Link",
						`<`+srv.URL+`/orgs/acme/audit-log?after=page2&order=desc&per_page=100>; rel="next"`,
					)
					_, _ = w.Write([]byte(`[{"@timestamp":1771336800000,"actor":"bot","actor_id":1,"action":"git.push"}]`))

					return
				}

				_, _ = w.Write([]byte(`[{"@timestamp":1774018800000,"actor":"jdoe","actor_id":100001,"action":"pull_request.create"}]`))

				return
			}

			writeGitHubListAccountsREST(t, w, r, "jdoe@personal.example.com")
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.NotNil(t, records[0].LastLogin)
	assert.Equal(t, time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC), records[0].LastLogin.UTC())
}

func TestGitHubDriver_IncludesMemberWhenMembershipUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/memberships/jdoe" {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"Resource not accessible by integration"}`))

				return
			}

			writeGitHubListAccountsREST(t, w, r, "jdoe@personal.example.com")
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "100001", records[0].ExternalID)
	assert.Equal(t, "Jane Doe", records[0].FullName)
	assert.Equal(t, "jdoe@personal.example.com", records[0].Email)
	assert.Empty(t, records[0].Roles)
	assert.Nil(t, records[0].Active)
	assert.Nil(t, records[0].IsAdmin)
}

func TestGitHubDriver_AuditLogStopsAfterMaxPages(t *testing.T) {
	t.Parallel()

	var srv *httptest.Server
	auditPages := 0
	srv = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/audit-log" {
				auditPages++
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set(
					"Link",
					`<`+srv.URL+`/orgs/acme/audit-log?after=more&order=desc&per_page=100>; rel="next"`,
				)
				_, _ = w.Write([]byte(`[{"@timestamp":1771336800000,"actor":"bot","actor_id":1,"action":"git.push"}]`))

				return
			}

			writeGitHubListAccountsREST(t, w, r, "jdoe@personal.example.com")
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, githubAuditLogMaxPages, auditPages)
	assert.Nil(t, records[0].LastLogin)
}

func TestGitHubApplyLastActivity(t *testing.T) {
	t.Parallel()

	byLogin := map[string]time.Time{
		"jdoe": time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	byID := map[int64]time.Time{
		100001: time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC),
	}

	record := AccountRecord{}
	githubApplyLastActivity(&record, githubMember{Login: "jdoe", ID: 100001}, byLogin, byID)
	require.NotNil(t, record.LastLogin)
	assert.Equal(t, time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC), record.LastLogin.UTC())

	loginOnly := AccountRecord{}
	githubApplyLastActivity(&loginOnly, githubMember{Login: "JDOE", ID: 9}, byLogin, nil)
	require.NotNil(t, loginOnly.LastLogin)
	assert.Equal(t, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), loginOnly.LastLogin.UTC())

	missing := AccountRecord{}
	githubApplyLastActivity(&missing, githubMember{Login: "nobody", ID: 2}, byLogin, byID)
	assert.Nil(t, missing.LastLogin)
}

func TestGitHubAuditEventTime(t *testing.T) {
	t.Parallel()

	milli, ok := githubAuditEventTime(githubAuditEvent{Timestamp: 1774018800000})
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC), milli)

	seconds, ok := githubAuditEventTime(githubAuditEvent{CreatedAt: 1774018800})
	require.True(t, ok)
	assert.Equal(t, time.Date(2026, 3, 20, 15, 0, 0, 0, time.UTC), seconds)

	_, ok = githubAuditEventTime(githubAuditEvent{})
	assert.False(t, ok)
}

func newGitHubListAccountsServer(t *testing.T, publicEmail, graphqlBody string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && r.URL.Path == "/graphql" {
				writeGitHubListAccountsGraphQL(t, w, r, graphqlBody)

				return
			}

			writeGitHubListAccountsREST(t, w, r, publicEmail)
		}),
	)
}

func writeGitHubListAccountsREST(t *testing.T, w http.ResponseWriter, r *http.Request, publicEmail string) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/members":
		if r.URL.Query().Get("filter") == "2fa_disabled" {
			_, _ = w.Write([]byte(`[]`))

			return
		}

		_, _ = w.Write([]byte(`[{"login":"jdoe","id":100001,"type":"User"}]`))
	case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/memberships/jdoe":
		_, _ = w.Write([]byte(`{"role":"member","state":"active"}`))
	case r.Method == http.MethodGet && r.URL.Path == "/users/jdoe":
		profile, err := json.Marshal(
			map[string]any{
				"login":      "jdoe",
				"name":       "Jane Doe",
				"email":      publicEmail,
				"created_at": "2009-05-06T20:34:11Z",
				"type":       "User",
			},
		)
		require.NoError(t, err)
		_, _ = w.Write(profile)
	case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/audit-log":
		_, _ = w.Write([]byte(`[{"@timestamp":1774018800000,"created_at":1774018800000,"actor":"jdoe","actor_id":100001,"action":"pull_request.create"}]`))
	case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/installations":
		_, _ = w.Write([]byte(`{"total_count":0,"installations":[]}`))
	case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/personal-access-tokens":
		_, _ = w.Write([]byte(`[]`))
	case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/credential-authorizations":
		_, _ = w.Write([]byte(`[]`))
	case r.Method == http.MethodPost && r.URL.Path == "/graphql":
		writeGitHubListAccountsGraphQL(
			t,
			w,
			r,
			`{"data":{"organization":{"membersWithRole":{"nodes":[{"login":"jdoe","organizationVerifiedDomainEmails":[]}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`,
		)
	default:
		t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}
}

func writeGitHubListAccountsGraphQL(t *testing.T, w http.ResponseWriter, r *http.Request, emailsBody string) {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)

	var req struct {
		Query     string `json:"query"`
		Variables struct {
			Org string `json:"org"`
		} `json:"variables"`
	}
	require.NoError(t, json.Unmarshal(body, &req))
	assert.Equal(t, "acme", req.Variables.Org)

	w.Header().Set("Content-Type", "application/json")

	switch {
	case strings.Contains(req.Query, "organizationVerifiedDomainEmails"):
		_, _ = w.Write([]byte(emailsBody))
	case strings.Contains(req.Query, "deployKeys"):
		_, _ = w.Write([]byte(`{"data":{"organization":{"repositories":{"nodes":[],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`))
	default:
		t.Errorf("unexpected graphql query %s", req.Query)
		http.NotFound(w, r)
	}
}

func TestGitHubDriver_ListsAppsTokensAndDeployKeys(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			switch {
			case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/installations":
				_, _ = w.Write([]byte(`{"total_count":1,"installations":[{"id":42,"app_slug":"dependabot","permissions":{"administration":"write"},"created_at":"2024-01-02T03:04:05Z","suspended_at":null}]}`))
			case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/personal-access-tokens":
				_, _ = w.Write([]byte(`[{"id":11,"token_id":25381,"token_name":"terraform","token_expired":false,"token_last_used_at":"2026-03-01T00:00:00Z","access_granted_at":"2025-01-01T00:00:00Z","owner":{"login":"jdoe"},"permissions":{"organization":{},"repository":{"contents":"read"}}}]`))
			case r.Method == http.MethodGet && r.URL.Path == "/orgs/acme/credential-authorizations":
				_, _ = w.Write([]byte(`[
					{"login":"jdoe","credential_id":99,"credential_type":"personal access token","credential_authorized_at":"2024-06-01T00:00:00Z","authorized_credential_title":"ci-classic"},
					{"login":"jdoe","credential_id":100,"credential_type":"SSH key","credential_authorized_at":"2024-06-01T00:00:00Z","authorized_credential_title":"laptop"},
					{"login":"jdoe","credential_id":101,"credential_type":"OAuth app token","credential_authorized_at":"2024-07-01T00:00:00Z","authorized_credential_title":"n8n"}
				]`))
			case r.Method == http.MethodPost && r.URL.Path == "/graphql":
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				if strings.Contains(string(body), "deployKeys") {
					_, _ = w.Write([]byte(`{"data":{"organization":{"repositories":{"nodes":[{"name":"api","deployKeys":{"nodes":[{"id":"DK_1","title":"ci-deploy","createdAt":"2023-05-01T00:00:00Z","readOnly":true}]}}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`))

					return
				}

				_, _ = w.Write([]byte(`{"data":{"organization":{"membersWithRole":{"nodes":[{"login":"jdoe","organizationVerifiedDomainEmails":["jdoe@acme.example.com"]}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`))
			default:
				writeGitHubListAccountsREST(t, w, r, "jdoe@personal.example.com")
			}
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 6)

	byID := make(map[string]AccountRecord, len(records))
	for _, rec := range records {
		byID[rec.ExternalID] = rec
	}

	member := byID["100001"]
	assert.Equal(t, "jdoe@acme.example.com", member.Email)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, member.AccountType)

	app := byID["installation:42"]
	assert.Equal(t, "dependabot", app.FullName)
	assert.Equal(t, []string{"github-app"}, app.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, app.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodOAuth2, app.AuthMethod)
	require.NotNil(t, app.Active)
	assert.True(t, *app.Active)
	require.NotNil(t, app.IsAdmin)
	assert.True(t, *app.IsAdmin)

	pat := byID["pat:25381"]
	assert.Equal(t, "terraform", pat.FullName)
	assert.Equal(t, []string{"fine-grained PAT"}, pat.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, pat.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, pat.AuthMethod)
	require.NotNil(t, pat.LastLogin)
	assert.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), pat.LastLogin.UTC())

	classic := byID["credential:99"]
	assert.Equal(t, "ci-classic", classic.FullName)
	assert.Equal(t, []string{"personal access token"}, classic.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodAPIKey, classic.AuthMethod)

	_, sshListed := byID["credential:100"]
	assert.False(t, sshListed)

	oauth := byID["credential:101"]
	assert.Equal(t, "n8n", oauth.FullName)
	assert.Equal(t, []string{"OAuth app token"}, oauth.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodOAuth2, oauth.AuthMethod)

	key := byID["deploy-key:DK_1"]
	assert.Equal(t, "ci-deploy (api)", key.FullName)
	assert.Equal(t, []string{"deploy-key"}, key.Roles)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeServiceAccount, key.AccountType)
	assert.Equal(t, coredata.AccessReviewEntryAuthMethodSSH, key.AuthMethod)
}

func TestGitHubDriver_ContinuesWhenServiceAccountsUnavailable(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/orgs/acme/installations",
				"/orgs/acme/personal-access-tokens",
				"/orgs/acme/credential-authorizations":
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"message":"Resource not accessible by integration"}`))

				return
			}

			if r.Method == http.MethodPost && r.URL.Path == "/graphql" {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				w.Header().Set("Content-Type", "application/json")
				if strings.Contains(string(body), "deployKeys") {
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"message":"Resource not accessible by integration"}`))

					return
				}

				_, _ = w.Write([]byte(`{"data":{"organization":{"membersWithRole":{"nodes":[{"login":"jdoe","organizationVerifiedDomainEmails":[]}],"pageInfo":{"hasNextPage":false,"endCursor":""}}}}}`))

				return
			}

			writeGitHubListAccountsREST(t, w, r, "jdoe@personal.example.com")
		}),
	)
	t.Cleanup(srv.Close)

	records, err := NewGitHubDriver(
		srv.Client(),
		"acme",
		log.NewLogger(log.WithName("test")),
		srv.URL,
	).ListAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, "100001", records[0].ExternalID)
	assert.Equal(t, coredata.AccessReviewEntryAccountTypeUser, records[0].AccountType)
}

func TestGitHubDriver_SkipsFineGrainedPATDuplicatedInSSO(t *testing.T) {
	t.Parallel()

	tokenID := int64(25381)
	pat := githubFineGrainedPAT{ID: 11, TokenID: tokenID, TokenName: "terraform"}
	seen := map[string]struct{}{}
	githubRememberPATIDs(seen, pat)
	rec, ok := githubFineGrainedPATRecord(pat)
	require.True(t, ok)
	seen[rec.ExternalID] = struct{}{}

	dup := githubCredentialAuthorization{
		CredentialID:           tokenID,
		CredentialType:         githubCredentialTypePAT,
		AuthorizedCredentialID: &tokenID,
	}
	assert.True(t, githubCredentialDuplicatesPAT(seen, dup))

	oauth := githubCredentialAuthorization{
		CredentialID:   101,
		CredentialType: githubCredentialTypeOAuthApp,
	}
	assert.False(t, githubCredentialDuplicatesPAT(seen, oauth))
}
