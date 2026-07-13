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

import "strings"

// discoveryGlobQueries maps vfs discovery glob patterns to GitHub code search queries.
var discoveryGlobQueries = map[string]string{
	"*/.github/workflows/*.yml":   "path:.github/workflows extension:yml",
	"*/.github/workflows/*.yaml":  "path:.github/workflows extension:yaml",
	"*/SECURITY.md":               "filename:SECURITY.md",
	"*/CONTRIBUTING.md":           "filename:CONTRIBUTING.md",
	"*/.github/dependabot.yml":    "path:.github filename:dependabot.yml",
	"*/renovate.json":             "filename:renovate.json",
	"*/.github/renovate.json":     "path:.github filename:renovate.json",
	"*/package-lock.json":         "filename:package-lock.json",
	"*/yarn.lock":                 "filename:yarn.lock",
	"*/pnpm-lock.yaml":            "filename:pnpm-lock.yaml",
	"*/go.sum":                    "filename:go.sum",
	"*/Gemfile.lock":              "filename:Gemfile.lock",
	"*/poetry.lock":               "filename:poetry.lock",
	"*/Cargo.lock":                "filename:Cargo.lock",
	"*/.env":                      "filename:.env",
	"*/.env.production":           "filename:.env.production",
	"*/.env.local":                "filename:.env.local",
	"*/DEVELOPMENT.md":            "filename:DEVELOPMENT.md",
	"*/docs/development.md":       "path:docs filename:development.md",
	"*/docs/code-review.md":       "path:docs filename:code-review.md",
	"*/docs/incident-response.md": "path:docs filename:incident-response.md",
	"*/docs/security/README.md":   "path:docs/security filename:README.md",
	"*/SECURITY_GUIDELINES.md":    "filename:SECURITY_GUIDELINES.md",
	"*/.github/ISSUE_TEMPLATE/*":  "path:.github/ISSUE_TEMPLATE",
	"*/Jenkinsfile":               "filename:Jenkinsfile",
	"*/.circleci/config.yml":      "path:.circleci filename:config.yml",
	"*/.gitlab-ci.yml":            "filename:.gitlab-ci.yml",
}

func globToCodeSearch(org, pattern string) (string, bool) {
	query, ok := discoveryGlobQueries[pattern]
	if !ok {
		return "", false
	}

	return "org:" + org + " " + strings.ReplaceAll(query, " ", "+"), true
}
