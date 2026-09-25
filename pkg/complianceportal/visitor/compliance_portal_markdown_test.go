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

package visitor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComplianceTemplate_Execute(t *testing.T) {
	t.Parallel()

	t.Run("frameworks only", func(t *testing.T) {
		t.Parallel()

		data := &compliancePageData{
			OrgName:     "Tinfoil",
			Description: "Private AI inference inside secure hardware enclaves.",
			Details: []compliancePageDetail{
				{Label: "Website", Value: "https://tinfoil.sh"},
				{Label: "Email", Value: "security@tinfoil.sh"},
			},
			Frameworks: []compliancePageFramework{
				{Name: "SOC 2"},
				{Name: "HIPAA"},
				{Name: "GDPR"},
			},
		}

		rendered := renderComplianceTemplate(t, data)

		assert.Contains(t, rendered, "# Tinfoil — Compliance")
		assert.Contains(t, rendered, "- SOC 2\n- HIPAA\n- GDPR\n")
		assert.NotContains(t, rendered, "## Subprocessors")
		assert.NotContains(t, rendered, "## Documents")
	})

	t.Run("all sections", func(t *testing.T) {
		t.Parallel()

		data := &compliancePageData{
			OrgName: "Tinfoil",
			Frameworks: []compliancePageFramework{
				{Name: "SOC 2", Description: "Type II"},
			},
			Documents: []compliancePageDocument{
				{Title: "Security | Policy", Type: "POLICY"},
			},
			Audits: []compliancePageAudit{
				{
					Name:       "2025 review",
					Framework:  "SOC 2",
					ValidFrom:  "2025-01-01",
					ValidUntil: "2026-01-01",
				},
			},
			ThirdParties: []compliancePageThirdParty{
				{
					Name:      "Acme Cloud",
					Category:  "CLOUD_PROVIDER",
					Countries: "US, DE",
					Website:   "https://acme.example",
				},
			},
			References: []compliancePageReference{
				{
					Name:        "Whitepaper",
					Description: "Overview",
					Website:     "https://example.com/wp",
				},
			},
			CustomLinks: []compliancePageCustomLink{
				{Name: "Status", URL: "https://status.example"},
			},
		}

		rendered := renderComplianceTemplate(t, data)

		assert.Contains(t, rendered, "- SOC 2: Type II")
		assert.Contains(t, rendered, "| Security \\| Policy | POLICY |")
		assert.Contains(t, rendered, "| 2025 review | SOC 2 | 2025-01-01 | 2026-01-01 |")
		assert.Contains(t, rendered, "## Subprocessors")
		assert.Contains(
			t,
			rendered,
			"| Acme Cloud | CLOUD_PROVIDER | US, DE | https://acme.example |",
		)
		assert.Contains(t, rendered, "| Whitepaper | Overview | https://example.com/wp |")
		assert.Contains(t, rendered, "| Status | https://status.example |")
	})
}

func renderComplianceTemplate(t *testing.T, data *compliancePageData) string {
	t.Helper()

	var buf bytes.Buffer

	err := complianceTmpl.Execute(&buf, data)
	require.NoError(t, err)

	return buf.String()
}
