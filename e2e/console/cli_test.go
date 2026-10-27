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

	if listArgs == nil {
		listArgs = []string{"--org", s.orgID}
	}

	create := []string{group, "create"}
	if group != "control" {
		create = append(create, "--org", s.orgID)
	}

	create = append(create, createArgs...)
	id := s.mustID(create...)

	listed := s.must(append([]string{group, "list"}, append(listArgs, "--json")...)...)
	require.Contains(s.t, listed, id, "%s list did not include the created resource", group)

	viewed := s.must(group, "view", id, "--json")
	require.Contains(s.t, viewed, id, "%s view did not include the created resource", group)

	_, stderr, err := s.run(append([]string{group, "update", id}, []string{}...)...)
	require.Error(s.t, err, "%s update without fields succeeded: %s", group, stderr)
	assert.Contains(s.t, stderr, "at least one field must be specified")

	s.must(append([]string{group, "update", id}, updateArgs...)...)
	viewed = s.must(group, "view", id, "--json")
	assert.Contains(s.t, viewed, needle, "%s view did not show the update", group)

	_, stderr, err = s.run(group, "delete", id)
	require.Error(s.t, err, "%s delete without confirmation succeeded", group)
	assert.Contains(s.t, stderr, "confirmation required")

	s.must(group, "delete", id, "--yes")

	listed = s.must(append([]string{group, "list"}, append(listArgs, "--json")...)...)
	assert.NotContains(s.t, listed, id, "%s list still includes a deleted resource", group)
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
