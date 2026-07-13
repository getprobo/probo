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
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestSecurityContactInMarkdown(t *testing.T) {
	t.Parallel()

	assert.True(t, securityContactInMarkdown("Report vulnerabilities to security@example.com"))
	assert.True(t, securityContactInMarkdown("## Responsible disclosure\nmailto:sec@example.com"))
	assert.False(t, securityContactInMarkdown("# Security\nPlease open an issue."))
}

func TestIncidentResponseInMarkdown(t *testing.T) {
	t.Parallel()

	assert.True(t, incidentResponseInMarkdown("# Incident response\nFollow the on-call runbook."))
	assert.True(t, incidentResponseInMarkdown("Security incident escalation path"))
	assert.False(t, incidentResponseInMarkdown("# Development guide"))
}

func TestDetectCIProviders(t *testing.T) {
	t.Parallel()

	providers := detectCIProviders("ci/circleci", "https://circleci.com/gh/acme/api/123")
	assert.Contains(t, providers, "circleci")
	assert.NotContains(t, providers, "jenkins")

	providers = detectCIProviders("continuous-integration/jenkins", "Jenkins build")
	assert.Contains(t, providers, "jenkins")

	providers = detectCIProviders("github-actions", "check-suite")
	assert.Contains(t, providers, "github_actions")
}

func TestIsExternalCIProvider(t *testing.T) {
	t.Parallel()

	assert.False(t, isExternalCIProvider("github_actions"))
	assert.True(t, isExternalCIProvider("circleci"))
}

func TestEvaluatePRApprovalRate(t *testing.T) {
	t.Parallel()

	assert.Equal(t, coredata.MeasureStateImplemented, evaluatePRApprovalRate(map[string]int{
		"reviewed": 8,
		"sampled":  10,
	}))
	assert.Equal(t, coredata.MeasureStateNotImplemented, evaluatePRApprovalRate(map[string]any{
		"reviewed": 2,
		"sampled":  10,
	}))
}
