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

package provider_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// allowedLiteral names one string literal that contains a provider's endpoint
// host but is not a request target, so the scan below must not fail on it. The
// exception is keyed by file AND by the exact literal — never by file alone —
// so allowing a scope identifier in a file cannot mask a genuine endpoint
// literal that appears elsewhere in the same file.
type allowedLiteral struct {
	file    string
	literal string
	reason  string
}

// literalAllowlist holds every literal that spells out a connector provider's
// endpoint host without being that connector's traffic. Each entry is keyed by
// file AND literal so it cannot mask a genuine endpoint pin elsewhere in the
// same file, and each carries the reason it is safe.
//
// Two shapes appear here, and nothing else should:
//
//   - OAuth2 scope plumbing. Microsoft mints Graph permissions as resource
//     URIs (`https://graph.microsoft.com/User.Read.All`), so scope handling
//     has to name the Graph origin without ever sending a request to it. The
//     individual scope STRINGS need no entry: knownOAuth2Scopes below derives
//     them from the registry, so a new scope needs no test change. Only the
//     bare prefix, which no registration declares, is listed.
//   - A different Probo subsystem that happens to share a vendor with a
//     connector. Signing in with Google is not the Google Analytics
//     connector; the agent's self-updater downloading getprobo/probo releases
//     is not the GitHub connector. Repointing the connector must NOT move
//     these, so threading the registration's Endpoints here would be the bug,
//     not the fix.
var literalAllowlist = []allowedLiteral{
	{
		file:    "pkg/connector/scopes.go",
		literal: "https://graph.microsoft.com/",
		reason:  "microsoftGraphScopePrefix: trims the resource-URI prefix off a granted scope when diffing requested against granted; never dialled",
	},
	{
		file:    "pkg/accessreview/errors.go",
		literal: "https://graph.microsoft.com/",
		reason:  "shortens Graph scope identifiers for a user-facing missing-scope message; never dialled",
	},
	{
		file:    "pkg/iam/oidc/service.go",
		literal: "https://accounts.google.com/o/oauth2/v2/auth",
		reason:  "Probo's own sign-in-with-Google, a separate OAuth client from the Google Analytics connector that shares the host",
	},
	{
		file:    "pkg/iam/oidc/service.go",
		literal: "https://oauth2.googleapis.com/token",
		reason:  "Probo's own sign-in-with-Google token exchange; see above",
	},
	{
		file:    "pkg/iam/oidc/service.go",
		literal: "https://accounts.google.com",
		reason:  "the expected iss claim of a Google ID token, a validation constant rather than a request target",
	},
	{
		file:    "pkg/deviceagent/update/update.go",
		literal: "https://api.github.com",
		reason:  "defaultAPIBaseURL: the agent's self-updater reads Probo's OWN getprobo/probo releases, not a customer's GitHub connector",
	},
	{
		file:    "pkg/deviceagent/update/update.go",
		literal: "https://github.com",
		reason:  "defaultAssetBaseURL: release asset downloads for the same self-updater; see above",
	},
	{
		file:    "pkg/proboctl/seed/common-tracker-patterns/common_tracker_patterns.go",
		literal: "https://github.com/jkwakman/Open-Cookie-Database.git",
		reason:  "the public dataset the seed command clones; a fixed upstream source, unrelated to any connector",
	},
	{
		file:    "pkg/proboctl/seed/common-tracker-patterns/common_tracker_patterns.go",
		literal: "(https://github.com/jkwakman/Open-Cookie-Database). ",
		reason:  "the same dataset named in that command's help text",
	},
	{
		file:    "internal/cmd/genmodels/main.go",
		literal: "https://openrouter.ai/api/v1/models",
		reason:  "a developer-run code generator fetching the public model catalog at build time; it never runs in probod and holds no connection credential",
	},
}

var generatedFileRe = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

// TestNoEndpointHostLiteralsOutsideOwningFiles is the invariant that makes the
// bug class this package's Endpoints machinery exists to prevent unrepresentable
// by audit fatigue.
//
// A provider WITHOUT Registration.EndpointOverrideUnsupported promises that a
// deployment can repoint every host it reaches by overriding Endpoints. That
// promise is only true while no OTHER package pins the same host as a literal:
// Slack shipped exactly that bug — pkg/slack/client.go dialled
// https://slack.com/api/chat.postMessage with a token minted by the OVERRIDDEN
// Endpoints.Token, so a sandbox deployment sent a sandbox-issued token to the
// real slack.com — and the SCIM bridge shipped the Microsoft 365 half of it.
// Both were found by reading pkg/accessreview/drivers, which is precisely the
// scope that misses them: the offending code was in neither the registry nor
// the drivers.
//
// So the check is repository-wide. For every overridable provider, every host
// its Endpoints declare must appear only in the two files that own it:
//
//   - pkg/connector/provider/<provider>.go — the registration itself, the one
//     place the host is meant to be written down.
//   - pkg/accessreview/drivers/<provider>.go — the driver's path constants and
//     its <provider>DefaultBaseURL, the documented production fallback used by
//     the exported picker listers (ListSentryOrganizations and friends) that
//     the console calls with no registration in scope. Those defaults are in
//     the provider's own driver file, so this rule already permits them and
//     they need no exception.
//
// Anything else is a host the override cannot move, and the provider must
// either thread the endpoint through to that call site or declare
// EndpointOverrideUnsupported with the reason.
func TestNoEndpointHostLiteralsOutsideOwningFiles(t *testing.T) {
	t.Parallel()

	root, ok := repoRoot(t)
	if !ok {
		t.Skip("repository source tree not available; nothing to scan")
	}

	r := provider.NewBuiltinRegistry()

	declarers, overridable := endpointHosts(t, r)
	require.NotEmpty(t, overridable, "registry declared no overridable endpoint hosts, the scan would be vacuous")

	scopes := knownOAuth2Scopes(r)

	var findings []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if skipDir(d.Name()) {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		rel = filepath.ToSlash(rel)

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if isGenerated(src) {
			return nil
		}

		findings = append(findings, scanFile(t, rel, src, declarers, overridable, scopes)...)

		return nil
	})
	require.NoError(t, err)

	sort.Strings(findings)

	require.Empty(
		t,
		findings,
		"a provider endpoint host is pinned outside the files that own it, so an endpoint override would move only part of the provider's traffic.\n"+
			"Fix by threading the registration's Endpoints value to the call site, or by declaring Registration.EndpointOverrideUnsupported with the reason.\n%s",
		strings.Join(findings, "\n"),
	)
}

// TestEndpointHostLiteralAllowlistIsLive keeps the exception list honest: an
// entry whose file or literal no longer exists is a permission nobody revoked,
// and the next literal added to that file inherits an argument that was made
// about a different one. Every entry must name a real literal and say why.
func TestEndpointHostLiteralAllowlistIsLive(t *testing.T) {
	t.Parallel()

	root, ok := repoRoot(t)
	if !ok {
		t.Skip("repository source tree not available; nothing to scan")
	}

	for _, a := range literalAllowlist {
		t.Run(a.file+" "+a.literal, func(t *testing.T) {
			t.Parallel()

			require.NotEmpty(t, a.reason, "allowlist entry must explain why the literal is not connector traffic")

			src, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(a.file)))
			require.NoError(t, err, "allowlisted file no longer exists")

			fset := token.NewFileSet()

			f, err := parser.ParseFile(fset, a.file, src, parser.SkipObjectResolution)
			require.NoError(t, err)

			var found bool

			ast.Inspect(f, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}

				if value, err := strconv.Unquote(lit.Value); err == nil && value == a.literal {
					found = true
				}

				return !found
			})

			require.True(t, found, "allowlisted literal is gone; drop the entry")
		})
	}
}

// scanFile reports every string literal in src that pins one of the overridable
// hosts outside the files that own it. It parses rather than greps so a host
// named in a doc comment — pkg/accessreview/drivers/microsoft_365.go documents
// the Graph origin in prose — is not mistaken for a request target.
func scanFile(t *testing.T, rel string, src []byte, declarers, overridable map[string][]coredata.ConnectorProvider, scopes map[string]bool) []string {
	t.Helper()

	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, rel, src, parser.SkipObjectResolution)
	if err != nil {
		// A file this package cannot parse is not a file it can clear, so
		// surface it rather than skip it.
		t.Fatalf("cannot parse %s: %v", rel, err)
	}

	var findings []string

	ast.Inspect(f, func(n ast.Node) bool {
		lit, ok := n.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			return true
		}

		value, err := strconv.Unquote(lit.Value)
		if err != nil {
			return true
		}

		// An exact OAuth2 scope identifier is a permission name that happens
		// to be spelled as a URI (Microsoft Graph, Google), not a host the
		// connector dials. The set comes from the registry itself, so a new
		// scope needs no test change.
		if scopes[value] {
			return true
		}

		for host, providers := range overridable {
			if !literalPinsHost(value, host) {
				continue
			}

			if allowedLiteralFor(rel, value) {
				continue
			}

			// Ownership is checked against every provider that declares the
			// host, not just the overridable ones: two providers can sit on one
			// vendor origin (Google Analytics and Google Workspace both
			// authorize on accounts.google.com), each writes it down in its own
			// registration and driver, and an override moves one provider
			// without touching the other.
			if slices.ContainsFunc(declarers[host], func(p coredata.ConnectorProvider) bool {
				return ownsHostLiteral(rel, p)
			}) {
				continue
			}

			names := make([]string, 0, len(providers))
			for _, p := range providers {
				names = append(names, string(p))
			}

			sort.Strings(names)

			pos := fset.Position(lit.Pos())

			findings = append(findings, "  "+rel+":"+strconv.Itoa(pos.Line)+": host "+host+" (declared by "+strings.Join(names, ", ")+") pinned as "+lit.Value)
		}

		return true
	})

	return findings
}

// endpointHosts maps each host an Endpoints field names to the providers that
// name it. A host is shared when two providers sit on the same vendor origin,
// so the values are slices.
//
// declarers covers every registration; overridable covers only those without
// EndpointOverrideUnsupported. The scan enforces the rule on the second and
// resolves file ownership against the first — see scanFile.
func endpointHosts(t *testing.T, r *provider.Registry) (declarers, overridable map[string][]coredata.ConnectorProvider) {
	t.Helper()

	declarers = make(map[string][]coredata.ConnectorProvider)
	overridable = make(map[string][]coredata.ConnectorProvider)

	for _, reg := range r.All() {
		for _, raw := range []string{
			reg.Endpoints.Auth,
			reg.Endpoints.Token,
			reg.Endpoints.Probe,
			reg.Endpoints.Identity,
			reg.Endpoints.APIBase,
		} {
			if raw == "" {
				continue
			}

			u, err := url.Parse(raw)
			require.NoErrorf(t, err, "provider %s declares an unparseable endpoint %q", reg.Provider, raw)

			host := strings.ToLower(u.Host)
			if host == "" {
				continue
			}

			if !slices.Contains(declarers[host], reg.Provider) {
				declarers[host] = append(declarers[host], reg.Provider)
			}

			if reg.EndpointOverrideUnsupported != "" {
				continue
			}

			if !slices.Contains(overridable[host], reg.Provider) {
				overridable[host] = append(overridable[host], reg.Provider)
			}
		}
	}

	return declarers, overridable
}

// knownOAuth2Scopes collects every scope string the registry declares, so the
// scan can tell a Graph permission identifier from a Graph endpoint.
func knownOAuth2Scopes(r *provider.Registry) map[string]bool {
	scopes := make(map[string]bool)

	for _, reg := range r.All() {
		for _, s := range reg.OAuth2Scopes {
			scopes[s] = true
		}
	}

	return scopes
}

// literalPinsHost reports whether value targets host, i.e. contains
// "https://<host>" ending on a URL boundary. The boundary check is what keeps
// "https://slack.com" from matching "https://slack.community".
func literalPinsHost(value, host string) bool {
	needle := "https://" + host
	hay := strings.ToLower(value)

	for i := 0; ; {
		j := strings.Index(hay[i:], needle)
		if j < 0 {
			return false
		}

		end := i + j + len(needle)
		if end == len(hay) || strings.ContainsRune("/?#:", rune(hay[end])) {
			return true
		}

		i = end
	}
}

// ownsHostLiteral reports whether rel is one of the two files allowed to spell
// out provider p's host: its registration in pkg/connector/provider and its
// driver in pkg/accessreview/drivers. File names normalise the provider
// constant loosely (ONE_PASSWORD is one_password.go in one directory and
// onepassword.go in the other), and a driver may be split across companion
// files (onepassword_users_api.go), so both spellings and a suffixed variant
// of each are accepted.
func ownsHostLiteral(rel string, p coredata.ConnectorProvider) bool {
	dir, base := filepath.Split(rel)

	dir = strings.TrimSuffix(dir, "/")
	if dir != "pkg/connector/provider" && dir != "pkg/accessreview/drivers" {
		return false
	}

	name := strings.TrimSuffix(base, ".go")
	lower := strings.ToLower(string(p))

	for _, candidate := range []string{lower, strings.ReplaceAll(lower, "_", "")} {
		if name == candidate || strings.HasPrefix(name, candidate+"_") {
			return true
		}
	}

	return false
}

// allowedLiteralFor matches an exception on file AND literal, never on file
// alone, so an entry cannot cover a literal it was not argued about.
func allowedLiteralFor(rel, value string) bool {
	for _, a := range literalAllowlist {
		if a.file == rel && a.literal == value {
			return true
		}
	}

	return false
}

// skipDir excludes trees whose Go the deployment never runs: third-party code
// (vendor, node_modules) and fixtures (testdata, which Go itself excludes from
// the build).
func skipDir(name string) bool {
	switch name {
	case "vendor", "node_modules", "testdata":
		return true
	}

	// .git, .github and friends hold no Go the build ever sees.
	return strings.HasPrefix(name, ".") && name != "."
}

// isGenerated reports the standard machine-generated header. A generated file
// mirrors a schema or a spec, so a host in it is a symptom to fix upstream,
// not a call site anyone can thread.
func isGenerated(src []byte) bool {
	for i, line := range strings.Split(string(src), "\n") {
		if i > 40 {
			return false
		}

		if generatedFileRe.MatchString(strings.TrimRight(line, "\r")) {
			return true
		}
	}

	return false
}

// repoRoot walks up from the test's working directory to the module root, so
// the scan works from any package directory and skips cleanly when the source
// tree is not present (a packaged or source-less CI context).
func repoRoot(t *testing.T) (string, bool) {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}

		dir = parent
	}
}
