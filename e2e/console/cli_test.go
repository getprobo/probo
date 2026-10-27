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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/agentexecution"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/coredata"
	employeeportalmgmt "go.probo.inc/probo/pkg/employeeportal/management"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
	"go.probo.inc/probo/pkg/itam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/riskmanagement"
	"go.probo.inc/probo/pkg/task"
)

var (
	prbOnce sync.Once
	prbBin  string
	prbErr  error
)

func TestCLI_AuthAndOutput(t *testing.T) {
	t.Parallel()

	session := newPRBSession(t)

	help := session.must("--help")
	assert.Contains(t, help, "third-party")
	assert.NotContains(t, help, "thirdParty")

	thirdPartyHelp := session.must("third-party", "--help")
	assert.Contains(t, thirdPartyHelp, "Manage third parties")
	assert.NotContains(t, thirdPartyHelp, "third_party")
	assert.NotContains(t, thirdPartyHelp, "thirdParty")

	status := session.must("auth", "status")
	assert.Contains(t, status, session.host)
	assert.Contains(t, status, "(active)")
	assert.Contains(t, status, "Token: set")

	withJSON := session.must("org", "list", "--json")
	withOutput := session.must("org", "list", "--output", "json")
	assert.JSONEq(t, withJSON, withOutput)
	assert.Contains(t, withJSON, session.orgID)

	users := session.must("user", "list", "--org", session.orgID, "--json")
	assert.Contains(t, users, "e2e.probo.test")

	byOrganization := session.must("user", "list", "--organization", session.orgID, "--json")
	assert.JSONEq(t, users, byOrganization)

	_, stderr, err := session.run("org", "list", "--json", "--output", "table")
	require.Error(t, err)
	assert.Contains(t, stderr, "--json cannot be combined with --output table")

	_, stderr, err = session.run("skills", "--help")
	require.Error(t, err)
	assert.Contains(t, stderr, "unknown command")
}

func TestCLI_ResourceLifecycles(t *testing.T) {
	t.Parallel()

	session := newPRBSession(t)
	name := factory.SafeName("cli")

	session.roundTrip(
		"framework",
		[]string{"--name", name + "-framework"},
		[]string{"--name", name + "-framework-updated"},
		name+"-framework-updated",
		nil,
	)
	session.roundTrip(
		"task",
		[]string{"--name", name + "-task", "--priority", "MEDIUM"},
		[]string{"--name", name + "-task-updated"},
		name+"-task-updated",
		nil,
	)
	session.roundTrip(
		"document",
		[]string{"--title", name + "-document", "--document-type", "POLICY", "--classification", "INTERNAL"},
		[]string{"--title", name + "-document-updated"},
		name+"-document-updated",
		nil,
	)
	session.roundTrip(
		"asset",
		[]string{
			"--name", name + "-asset",
			"--asset-type", "VIRTUAL",
			"--amount", "1",
			"--owner", session.profileID,
			"--data-types-stored", "Customer records",
		},
		[]string{"--name", name + "-asset-updated"},
		name+"-asset-updated",
		nil,
	)
	session.roundTrip(
		"datum",
		[]string{
			"--name", name + "-datum",
			"--data-classification", "INTERNAL",
			"--owner", session.profileID,
		},
		[]string{"--name", name + "-datum-updated"},
		name+"-datum-updated",
		nil,
	)
	session.roundTrip(
		"risk",
		[]string{"--name", name + "-risk", "--category", "Security"},
		[]string{"--name", name + "-risk-updated"},
		name+"-risk-updated",
		nil,
	)
	session.roundTrip(
		"processing-activity",
		[]string{
			"--name", name + "-processing",
			"--role", "CONTROLLER",
			"--lawful-basis", "CONSENT",
			"--special-or-criminal-data", "NO",
			"--data-protection-impact-assessment", "NOT_NEEDED",
			"--transfer-impact-assessment", "NOT_NEEDED",
		},
		[]string{"--name", name + "-processing-updated"},
		name+"-processing-updated",
		nil,
	)
	session.roundTrip(
		"obligation",
		[]string{
			"--area", name + "-area",
			"--source", "GDPR",
			"--status", "COMPLIANT",
			"--type", "LEGAL",
			"--owner", session.profileID,
		},
		[]string{"--area", name + "-area-updated"},
		name+"-area-updated",
		nil,
	)
	session.roundTrip(
		"rights-request",
		[]string{"--data-subject", name + "-subject", "--type", "ACCESS", "--state", "TODO"},
		[]string{"--state", "IN_PROGRESS"},
		"IN_PROGRESS",
		nil,
	)
	session.roundTrip(
		"internal-control",
		[]string{"--name", name + "-control", "--category", "Security"},
		[]string{"--name", name + "-control-updated"},
		name+"-control-updated",
		nil,
	)
	session.roundTrip(
		"business-function",
		[]string{
			"--name", name + "-function",
			"--classification", "STANDARD",
			"--mtd-minutes", "60",
			"--rto-minutes", "30",
			"--rpo-minutes", "15",
		},
		[]string{"--name", name + "-function-updated"},
		name+"-function-updated",
		nil,
	)
	session.roundTrip(
		"ai-system",
		[]string{"--name", name + "-ai", "--status", "ACTIVE", "--risk-classification", "MINIMAL"},
		[]string{"--status", "IN_DEVELOPMENT"},
		"IN_DEVELOPMENT",
		nil,
	)
	auditFrameworkID := session.mustID(
		"framework", "create",
		"--org", session.orgID,
		"--name", name+"-audit-framework",
	)
	session.roundTrip(
		"audit",
		[]string{"--name", name + "-audit", "--framework", auditFrameworkID},
		[]string{"--name", name + "-audit-updated"},
		name+"-audit-updated",
		nil,
	)
	session.must("framework", "delete", auditFrameworkID, "--yes")
	session.roundTrip(
		"finding",
		[]string{"--kind", "OBSERVATION", "--status", "OPEN", "--priority", "LOW"},
		[]string{"--priority", "HIGH"},
		"HIGH",
		nil,
	)
	session.roundTrip(
		"webhook",
		[]string{"--url", "https://example.com/e2e-webhook", "--event", "DOCUMENT_CREATED"},
		[]string{"--url", "https://example.com/e2e-webhook-updated"},
		"https://example.com/e2e-webhook-updated",
		nil,
	)

	thirdPartyID := session.mustID(
		"third-party", "create",
		"--org", session.orgID,
		"--name", name+"-vendor",
		"--category", "IT",
	)
	for _, group := range []string{"third-party", "tp"} {
		listed := session.must(group, "list", "--organization", session.orgID, "--json")
		assert.Contains(t, listed, thirdPartyID)
	}

	_, stderr, err := session.run("third_party", "list", "--org", session.orgID)
	require.Error(t, err)
	assert.Contains(t, stderr, "unknown command")
	_, stderr, err = session.run("thirdParty", "list", "--org", session.orgID)
	require.Error(t, err)
	assert.Contains(t, stderr, "unknown command")
	session.must("third-party", "update", thirdPartyID, "--name", name+"-vendor-updated")
	viewed := session.must("third-party", "view", thirdPartyID, "--json")
	assert.Contains(t, viewed, name+"-vendor-updated")

	_, stderr, err = session.run("third-party", "delete", thirdPartyID)
	require.Error(t, err)
	assert.Contains(t, stderr, "confirmation required")
	session.must("third-party", "delete", thirdPartyID, "--yes")

	frameworkID := session.mustID(
		"framework", "create",
		"--org", session.orgID,
		"--name", name+"-control-framework",
	)
	session.roundTrip(
		"control",
		[]string{
			"--framework", frameworkID,
			"--section-title", "Access",
			"--name", name + "-control",
			"--description", name + "-control description",
		},
		[]string{"--name", name + "-control-updated"},
		name+"-control-updated",
		[]string{"--framework", frameworkID},
	)
	session.must("framework", "delete", frameworkID, "--yes")

	_, stderr, err = session.run(
		"framework", "create",
		"--name", name+"-missing-org",
	)
	require.Error(t, err)
	assert.Contains(t, stderr, "organization is required")

	session.roundTrip(
		"soa",
		[]string{"--name", name + "-soa"},
		[]string{"--name", name + "-soa-updated"},
		name+"-soa-updated",
		nil,
	)
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	session.roundTrip(
		"cookie-banner",
		[]string{
			"--name", name + "-banner",
			"--origin", "https://" + slug + ".example.com",
			"--cookie-policy-url", "https://" + slug + ".example.com/cookies",
		},
		[]string{"--name", name + "-banner-updated"},
		name+"-banner-updated",
		nil,
	)
	session.roundTrip(
		"risk-analysis",
		[]string{"--name", name + "-analysis"},
		[]string{"--name", name + "-analysis-updated"},
		name+"-analysis-updated",
		nil,
	)

	activityID := session.mustID(
		"processing-activity", "create",
		"--org", session.orgID,
		"--name", name+"-activity-for-assessments",
		"--role", "CONTROLLER",
		"--lawful-basis", "CONSENT",
		"--special-or-criminal-data", "NO",
		"--data-protection-impact-assessment", "NEEDED",
		"--transfer-impact-assessment", "NEEDED",
	)
	session.recordLifecycle(
		[]string{"dpia"},
		[]string{"--processing-activity", activityID, "--description", name + "-dpia"},
		[]string{"--description", name + "-dpia-updated"},
		name+"-dpia-updated",
		nil,
		false,
	)
	session.recordLifecycle(
		[]string{"tia"},
		[]string{"--processing-activity", activityID, "--data-subjects", name + "-subjects"},
		[]string{"--data-subjects", name + "-subjects-updated"},
		name+"-subjects-updated",
		nil,
		false,
	)
	session.must("processing-activity", "delete", activityID, "--yes")

	bannerID := session.mustID(
		"cookie-banner", "create",
		"--org", session.orgID,
		"--name", name+"-banner-parent",
		"--origin", "https://"+slug+"-parent.example.com",
		"--cookie-policy-url", "https://"+slug+"-parent.example.com/cookies",
	)
	session.recordLifecycle(
		[]string{"cookie-category"},
		[]string{
			"--banner-id", bannerID,
			"--name", name + "-category",
			"--slug", slug + "-category",
			"--description", name + "-category description",
		},
		[]string{"--name", name + "-category-updated"},
		name+"-category-updated",
		[]string{"--banner-id", bannerID},
		false,
	)

	categoryID := session.mustID(
		"cookie-category", "create",
		"--banner-id", bannerID,
		"--name", name+"-tracker-category",
		"--slug", slug+"-trackers",
		"--description", name+"-tracker-category description",
	)
	session.recordLifecycle(
		[]string{"tracker-pattern"},
		[]string{
			"--category-id", categoryID,
			"--pattern", slug + "-cookie",
			"--match-type", "EXACT",
			"--description", name + "-pattern",
		},
		[]string{"--description", name + "-pattern-updated"},
		name+"-pattern-updated",
		[]string{"--category-id", categoryID},
		false,
	)
	session.recordLifecycle(
		[]string{"tracker-resource"},
		[]string{
			"--category-id", categoryID,
			"--resource-type", "SCRIPT",
			"--origin", "https://" + slug + "-parent.example.com",
			"--path", "/e2e.js",
			"--display-name", name + "-script",
			"--description", name + "-script description",
		},
		[]string{"--display-name", name + "-script-updated"},
		name+"-script-updated",
		[]string{"--category-id", categoryID},
		false,
	)
	session.must("cookie-category", "delete", categoryID, "--yes")
	session.must("cookie-banner", "delete", bannerID, "--yes")

	riskID := session.mustID(
		"risk", "create",
		"--org", session.orgID,
		"--name", name+"-plan-risk",
		"--category", "Security",
	)
	analysisID := session.mustID(
		"risk-analysis", "create",
		"--org", session.orgID,
		"--name", name+"-plan-analysis",
	)
	diagramID := session.mustID(
		"risk-analysis", "diagram", "create",
		"--risk-analysis-id", analysisID,
		"--name", name+"-diagram",
	)
	scenarioID := session.mustID(
		"risk-analysis", "scenario", "create",
		"--diagram-id", diagramID,
		"--name", name+"-scenario",
		"--description", name+"-scenario description",
	)
	session.must(
		"risk-analysis", "scenario", "link-risk",
		"--scenario-id", scenarioID,
		"--risk-id", riskID,
	)
	session.recordLifecycle(
		[]string{"treatment-plan"},
		[]string{
			"--risk", riskID,
			"--risk-analysis", analysisID,
			"--treatment", "MITIGATED",
			"--owner", session.profileID,
			"--inherent-likelihood", "2",
			"--inherent-impact", "3",
		},
		[]string{"--treatment", "AVOIDED"},
		"AVOIDED",
		nil,
		false,
	)
	session.must("risk-analysis", "scenario", "delete", scenarioID, "--yes")
	session.must("risk-analysis", "diagram", "delete", diagramID, "--yes")
	session.must("risk-analysis", "delete", analysisID, "--yes")
	session.must("risk", "delete", riskID, "--yes")
}

func TestCLI_FilesAndContext(t *testing.T) {
	t.Parallel()

	session := newPRBSession(t)
	name := factory.SafeName("file")

	deviceID := session.mustID(
		"device", "create",
		"--org", session.orgID,
		"--owner", session.profileID,
	)
	assert.Contains(t, session.must("device", "list", "--org", session.orgID, "--json"), deviceID)
	assert.Contains(t, session.must("device", "view", deviceID, "--json"), deviceID)

	_, stderr, err := session.run("device", "revoke", deviceID)
	require.Error(t, err)
	assert.Contains(t, stderr, "confirmation required")
	session.must("device", "revoke", deviceID, "--yes")
	assert.Contains(t, session.must("device", "view", deviceID, "--json"), "REVOKED")

	_, stderr, err = session.run("device", "delete", deviceID)
	require.Error(t, err)
	assert.Contains(t, stderr, "confirmation required")
	session.must("device", "delete", deviceID, "--yes")
	assert.NotContains(t, session.must("device", "list", "--org", session.orgID, "--json"), deviceID)

	controlID := session.mustID(
		"internal-control", "create",
		"--org", session.orgID,
		"--name", name+"-evidence-control",
		"--category", "Security",
	)
	evidencePath := filepath.Join(t.TempDir(), "evidence.txt")
	require.NoError(t, os.WriteFile(evidencePath, []byte("e2e evidence"), 0o600))

	evidenceID := session.mustID(
		"evidence", "upload", evidencePath,
		"--internal-control", controlID,
	)
	assert.Contains(t, session.must("evidence", "list", "--internal-control", controlID, "--json"), evidenceID)
	assert.Contains(t, session.must("evidence", "view", evidenceID, "--json"), evidenceID)

	_, stderr, err = session.run("evidence", "delete", evidenceID)
	require.Error(t, err)
	assert.Contains(t, stderr, "confirmation required")
	session.must("evidence", "delete", evidenceID, "--yes")
	assert.NotContains(t, session.must("evidence", "list", "--internal-control", controlID, "--json"), evidenceID)
	session.must("internal-control", "delete", controlID, "--yes")

	_, stderr, err = session.run("context", "update", "--org", session.orgID)
	require.Error(t, err)
	assert.Contains(t, stderr, "at least one section flag is required")

	product := name + " product"
	session.must("context", "update", "--org", session.orgID, "--product", product)
	assert.Contains(t, session.must("context", "get", "--org", session.orgID, "--json"), product)

	logged := session.must("audit-log", "list", "--org", session.orgID, "--json")

	var entries []json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(logged), &entries), "stdout:\n%s", logged)
}

func TestCLI_ListWindow(t *testing.T) {
	t.Parallel()

	session := newPRBSession(t)
	name := factory.SafeName("page")

	for i := 1; i <= 3; i++ {
		session.mustID(
			"framework", "create",
			"--org", session.orgID,
			"--name", fmt.Sprintf("%s-%d", name, i),
		)
	}

	pageArgs := []string{
		"framework", "list",
		"--org", session.orgID,
		"--limit", "1",
		"--order-by", "CREATED_AT",
		"--order-direction", "ASC",
		"--json",
	}
	stdout, stderr, err := session.run(pageArgs...)
	require.NoError(t, err, "stderr:\n%s", stderr)
	assert.Contains(t, stderr, "Showing 1 of")
	assert.Contains(t, stderr, "frameworks")

	first := decodeListIDs(t, stdout)
	require.Len(t, first, 1)

	cursor := nextCursor(t, stderr)
	stdout, stderr, err = session.run(append(pageArgs, "--after", cursor)...)
	require.NoError(t, err, "stderr:\n%s", stderr)
	assert.Contains(t, stderr, "Next cursor:")

	second := decodeListIDs(t, stdout)
	require.Len(t, second, 1)
	assert.NotEqual(t, first[0], second[0])

	listed := session.must(pageArgs...)
	aliased := session.must(append([]string{"framework", "ls"}, pageArgs[2:]...)...)
	assert.JSONEq(t, listed, aliased)

	_, stderr, err = session.run("framework", "list", "--org", session.orgID, "--limit", "0")
	require.Error(t, err)
	assert.Contains(t, stderr, "invalid --limit value 0")

	_, stderr, err = session.run("framework", "list", "--org", session.orgID, "--order-direction", "SIDEWAYS")
	require.Error(t, err)
	assert.Contains(t, stderr, `invalid --order-direction value "SIDEWAYS"`)

	assert.Contains(t, session.must("framework", "list", "--help"), "--after")
	assert.NotContains(t, session.must("connector", "list", "--help"), "--after")
}

type prbSession struct {
	t         *testing.T
	bin       string
	host      string
	token     string
	orgID     string
	profileID string
	home      string
}

func newPRBSession(t *testing.T) *prbSession {
	t.Helper()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	scopes := cliTokenScopes()

	return &prbSession{
		t:         t,
		bin:       prbBinary(t),
		host:      testutil.GetBaseURL(),
		token:     owner.CreateOAuth2AccessToken("e2e-prb", scopes),
		orgID:     owner.GetOrganizationID().String(),
		profileID: owner.GetProfileID().String(),
		home:      t.TempDir(),
	}
}

func (s *prbSession) roundTrip(group string, createArgs, updateArgs []string, needle string, listArgs []string) {
	s.t.Helper()

	s.recordLifecycle(strings.Fields(group), createArgs, updateArgs, needle, listArgs, group != "control")
}

func (s *prbSession) recordLifecycle(path, createArgs, updateArgs []string, needle string, listArgs []string, withOrg bool) {
	s.t.Helper()

	label := strings.Join(path, " ")

	if listArgs == nil {
		listArgs = []string{"--org", s.orgID}
	}

	create := append(append([]string{}, path...), "create")
	if withOrg {
		create = append(create, "--org", s.orgID)
	}

	create = append(create, createArgs...)
	id := s.mustID(create...)

	listed := s.must(append(append(append([]string{}, path...), "list"), append(listArgs, "--json")...)...)
	require.Contains(s.t, listed, id, "%s list did not include the created resource", label)

	viewed := s.must(append(append([]string{}, path...), "view", id, "--json")...)
	require.Contains(s.t, viewed, id, "%s view did not include the created resource", label)

	_, stderr, err := s.run(append(append([]string{}, path...), "update", id)...)
	require.Error(s.t, err, "%s update without fields succeeded: %s", label, stderr)
	assert.Contains(s.t, stderr, "at least one field must be specified")

	s.must(append(append(append([]string{}, path...), "update", id), updateArgs...)...)
	viewed = s.must(append(append([]string{}, path...), "view", id, "--json")...)
	assert.Contains(s.t, viewed, needle, "%s view did not show the update", label)

	_, stderr, err = s.run(append(append([]string{}, path...), "delete", id)...)
	require.Error(s.t, err, "%s delete without confirmation succeeded", label)
	assert.Contains(s.t, stderr, "confirmation required")

	s.must(append(append([]string{}, path...), "delete", id, "--yes")...)

	listed = s.must(append(append(append([]string{}, path...), "list"), append(listArgs, "--json")...)...)
	assert.NotContains(s.t, listed, id, "%s list still includes a deleted resource", label)
}

func decodeListIDs(t *testing.T, stdout string) []string {
	t.Helper()

	var rows []struct {
		ID string `json:"id"`
	}

	require.NoError(t, json.Unmarshal([]byte(stdout), &rows), "stdout:\n%s", stdout)

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	return ids
}

func nextCursor(t *testing.T, stderr string) string {
	t.Helper()

	const prefix = "Next cursor: "

	for line := range strings.SplitSeq(stderr, "\n") {
		if after, ok := strings.CutPrefix(line, prefix); ok {
			return after
		}
	}

	t.Fatalf("stderr did not contain a next cursor:\n%s", stderr)

	return ""
}

func (s *prbSession) must(args ...string) string {
	s.t.Helper()

	stdout, stderr, err := s.run(args...)
	require.NoError(s.t, err, "prb %s\nstdout:\n%s\nstderr:\n%s", strings.Join(args, " "), stdout, stderr)

	return stdout
}

func (s *prbSession) mustID(args ...string) string {
	s.t.Helper()

	stdout := s.must(args...)
	for field := range strings.FieldsSeq(stdout) {
		field = strings.Trim(field, "()")
		if _, err := gid.ParseGID(field); err == nil {
			return field
		}
	}

	s.t.Fatalf("prb %s did not print a resource id:\n%s", strings.Join(args, " "), stdout)

	return ""
}

func (s *prbSession) run(args ...string) (string, string, error) {
	s.t.Helper()

	cmd := exec.Command(s.bin, args...)
	cmd.Env = prbEnv(map[string]string{
		"HOME":                 s.home,
		"XDG_CONFIG_HOME":      filepath.Join(s.home, "config"),
		"PROBO_HOST":           s.host,
		"PROBO_TOKEN":          s.token,
		"PROBO_NO_INTERACTIVE": "1",
		"NO_COLOR":             "1",
		"TERM":                 "dumb",
	})

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}

func prbEnv(overrides map[string]string) []string {
	env := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if _, ok := overrides[key]; ok {
			continue
		}

		env = append(env, entry)
	}

	for key, value := range overrides {
		env = append(env, key+"="+value)
	}

	return env
}

// cliTokenScopes is the subset of CLIClientScopes that manual access tokens
// accept. prb login also requests identity scopes and a few API scopes that
// are listed on the CLI OAuth client but have no action mapping, so the
// token endpoint rejects them.
func cliTokenScopes() []string {
	registry := oauth2scope.NewRegistry().
		Register(iam.IAMOAuth2ScopeMappings).
		Register(probo.OAuth2ScopeMappings).
		Register(management.OAuth2ScopeMappings).
		Register(employeeportalmgmt.OAuth2ScopeMappings).
		Register(agentexecution.OAuth2ScopeMappings).
		Register(accessreview.OAuth2ScopeMappings).
		Register(resourcealias.OAuth2ScopeMappings).
		Register(itam.OAuth2ScopeMappings).
		Register(riskmanagement.OAuth2ScopeMappings).
		Register(task.OAuth2ScopeMappings)

	var scopes []string

	for scope := range strings.FieldsSeq(config.CLIClientScopes) {
		if !strings.HasPrefix(scope, "v1:") {
			continue
		}

		if err := registry.ValidateScopes(coredata.OAuth2Scopes{coredata.OAuth2Scope(scope)}); err != nil {
			continue
		}

		scopes = append(scopes, scope)
	}

	return scopes
}

func prbBinary(t *testing.T) string {
	t.Helper()

	prbOnce.Do(func() {
		if path := os.Getenv("PROBO_E2E_PRB"); path != "" {
			if _, err := os.Stat(path); err == nil {
				prbBin = path

				return
			}
		}

		mod, err := exec.Command("go", "env", "GOMOD").Output()
		if err != nil {
			prbErr = fmt.Errorf("cannot locate module: %w", err)

			return
		}

		root := filepath.Dir(strings.TrimSpace(string(mod)))
		out := filepath.Join(os.TempDir(), "prb-e2e-test")
		cmd := exec.Command("go", "build", "-o", out, "./cmd/prb")
		cmd.Dir = root

		output, err := cmd.CombinedOutput()
		if err != nil {
			prbErr = fmt.Errorf("cannot build prb: %w\n%s", err, output)

			return
		}

		prbBin = out
	})

	require.NoError(t, prbErr)

	return prbBin
}
