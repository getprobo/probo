// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package github

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

func TestDiscoveryScanner_CollectsOrgAndRepoFacts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/orgs/acme":
			_, _ = w.Write([]byte(`{
				"two_factor_requirement_enabled": true,
				"default_repository_permission": "read",
				"members_can_create_public_repositories": false,
				"members_can_change_repo_visibility": false
			}`))
		case r.URL.Path == "/orgs/acme/outside_collaborators":
			_, _ = w.Write([]byte(`[]`))
		case r.URL.Path == "/orgs/acme/actions/permissions":
			_, _ = w.Write([]byte(`{
				"enabled_repositories": "all",
				"allowed_actions": "selected",
				"sha_pinning_required": true,
				"can_approve_pull_request_reviews": false
			}`))
		case r.URL.Path == "/orgs/acme/installations":
			_, _ = w.Write([]byte(`{"total_count": 2, "installations": [{"id": 1}, {"id": 2}]}`))
		case r.URL.Path == "/orgs/acme/repos":
			_, _ = w.Write([]byte(`[
				{
					"name": "api",
					"default_branch": "main",
					"private": true,
					"archived": false,
					"disabled": false
				}
			]`))
		case r.URL.Path == "/repos/acme/.github":
			_, _ = w.Write([]byte(`{"name": ".github", "default_branch": "main"}`))
		case strings.HasSuffix(r.URL.Path, "/branches/main"):
			_, _ = w.Write([]byte(`{"commit": {"sha": "abc123"}}`))
		case strings.HasSuffix(r.URL.Path, "/commits/abc123/status"):
			_, _ = w.Write([]byte(`{"statuses": [{"context": "ci/circleci", "target_url": "https://circleci.com/gh/acme/api/1"}]}`))
		case strings.HasSuffix(r.URL.Path, "/commits/abc123/check-runs"):
			_, _ = w.Write([]byte(`{"check_runs": []}`))
		case strings.HasSuffix(r.URL.Path, "/pulls") && !strings.Contains(r.URL.Path, "/pulls/"):
			_, _ = w.Write([]byte(`[
				{"number": 42, "merged_at": "2026-01-01T00:00:00Z"},
				{"number": 41, "merged_at": null}
			]`))
		case strings.HasSuffix(r.URL.Path, "/pulls/42/reviews"):
			_, _ = w.Write([]byte(`[{"state": "APPROVED"}]`))
		case strings.HasSuffix(r.URL.Path, "/branches/main/protection"):
			_, _ = w.Write([]byte(`{
				"required_pull_request_reviews": {"required_approving_review_count": 1},
				"required_signatures": {"enabled": true},
				"allow_force_pushes": {"enabled": false}
			}`))
		case strings.HasSuffix(r.URL.Path, "/actions/workflows"):
			_, _ = w.Write([]byte(`{"total_count": 1}`))
		case strings.HasSuffix(r.URL.Path, "/contents/SECURITY.md"):
			_, _ = w.Write([]byte(`{
				"name": "SECURITY.md",
				"encoding": "base64",
				"content": "UmVwb3J0IHZ1bG5lcmFiaWxpdGllcyB0byBzZWN1cml0eUBleGFtcGxlLmNvbQ=="
			}`))
		case strings.HasSuffix(r.URL.Path, "/repos/acme/.github/contents/SECURITY.md"):
			_, _ = w.Write([]byte(`{"name": "SECURITY.md"}`))
		case strings.HasSuffix(r.URL.Path, "/repos/acme/.github/contents/CONTRIBUTING.md"):
			_, _ = w.Write([]byte(`{"name": "CONTRIBUTING.md"}`))
		case strings.HasSuffix(r.URL.Path, "/contents/CONTRIBUTING.md"):
			http.NotFound(w, r)
		case strings.HasSuffix(r.URL.Path, "/contents/.github/dependabot.yml"):
			_, _ = w.Write([]byte(`{"name": "dependabot.yml"}`))
		case strings.Contains(r.URL.Path, "/dependabot/alerts"):
			_, _ = w.Write([]byte(`[]`))
		case strings.Contains(r.URL.Path, "/secret-scanning/alerts"):
			_, _ = w.Write([]byte(`[]`))
		case strings.Contains(r.URL.Path, "/code-scanning/alerts"):
			_, _ = w.Write([]byte(`[]`))
		case r.URL.Path == "/search/code":
			_, _ = w.Write([]byte(`{
				"total_count": 2,
				"incomplete_results": false,
				"items": [
					{
						"name": "ci.yml",
						"path": ".github/workflows/ci.yml",
						"repository": {"name": "api"}
					},
					{
						"name": "dependabot.yml",
						"path": ".github/dependabot.yml",
						"repository": {"name": "api"}
					}
				]
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client := server.Client()
	client.Transport = &rewriteTransport{
		base:    client.Transport,
		apiBase: server.URL,
	}

	logger := log.NewLogger(log.WithOutput(io.Discard))
	conn := &connector.OAuth2Connection{Scope: "read:org repo security_events"}

	sheet, err := newDiscoveryScanner(client, "acme", conn, logger).scan(context.Background())
	require.NoError(t, err)
	require.NotNil(t, sheet)

	keys := factKeys(sheet.Facts)

	assert.Contains(t, keys, "org_mfa_required")
	assert.Contains(t, keys, "org_outside_collaborators")
	assert.Contains(t, keys, "org_actions_restricted")
	assert.Contains(t, keys, "repo_branch_protection_coverage")
	assert.Contains(t, keys, "repo_dependabot_critical_open")
	assert.Contains(t, keys, "repo_commit_status_ci_coverage")
	assert.Contains(t, keys, "repo_ci_providers")
	assert.Contains(t, keys, "repo_de_facto_pr_review_coverage")
	assert.Contains(t, keys, "repo_security_contact_coverage")
	assert.Contains(t, keys, "org_profile_security_md")
	assert.Equal(t, 1, sheet.ReposScanned)
}

func TestMaterializeFromFacts_IncludesRepoCoverageMeasures(t *testing.T) {
	t.Parallel()

	sheet := &FactSheet{
		GitHubOrg: "acme",
		Facts: []Fact{
			{
				FactID:  "f-branch-protection-coverage",
				FactKey: "repo_branch_protection_coverage",
				Value: map[string]int{
					"matched": 2,
					"total":   2,
				},
			},
			{
				FactID:  "f-secret-scanning-open",
				FactKey: "repo_secret_scanning_alerts_open",
				Value: map[string]int{
					"open": 1,
				},
			},
		},
	}

	plan, err := MaterializeFromFacts(sheet, nil)
	require.NoError(t, err)

	names := make([]string, 0, len(plan.Creates))
	for _, create := range plan.Creates {
		names = append(names, create.Name)
	}

	assert.Contains(t, names, "Default branch protection")
	assert.Contains(t, names, "Secret scanning alerts resolved")
	assert.Equal(t, coredata.MeasureStateImplemented, findCreateState(plan, "Default branch protection"))
	assert.Equal(t, coredata.MeasureStateNotImplemented, findCreateState(plan, "Secret scanning alerts resolved"))
}

func factKeys(facts []Fact) []string {
	keys := make([]string, 0, len(facts))

	for _, fact := range facts {
		keys = append(keys, fact.FactKey)
	}

	return keys
}

func findCreateState(plan *MeasurePlan, name string) coredata.MeasureState {
	for _, create := range plan.Creates {
		if create.Name == name {
			return create.State
		}
	}

	return ""
}

type rewriteTransport struct {
	base    http.RoundTripper
	apiBase string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = "http"
	clone.URL.Host = strings.TrimPrefix(t.apiBase, "http://")

	if t.base == nil {
		return http.DefaultTransport.RoundTrip(clone)
	}

	return t.base.RoundTrip(clone)
}
