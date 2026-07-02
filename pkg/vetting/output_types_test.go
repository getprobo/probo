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

package vetting_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/vetting"
)

func TestOutputType_SchemaGeneration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"CrawlerOutput", assertSchema[vetting.CrawlerOutput]},
		{"SecurityOutput", assertSchema[vetting.SecurityOutput]},
		{"DocumentAnalysisOutput", assertSchema[vetting.DocumentAnalysisOutput]},
		{"ComplianceOutput", assertSchema[vetting.ComplianceOutput]},
		{"MarketOutput", assertSchema[vetting.MarketOutput]},
		{"DataProcessingOutput", assertSchema[vetting.DataProcessingOutput]},
		{"SubprocessorOutput", assertSchema[vetting.SubprocessorOutput]},
		{"IncidentResponseOutput", assertSchema[vetting.IncidentResponseOutput]},
		{"BusinessContinuityOutput", assertSchema[vetting.BusinessContinuityOutput]},
		{"ProfessionalStandingOutput", assertSchema[vetting.ProfessionalStandingOutput]},
		{"AIRiskOutput", assertSchema[vetting.AIRiskOutput]},
		{"RegulatoryComplianceOutput", assertSchema[vetting.RegulatoryComplianceOutput]},
		{"WebSearchOutput", assertSchema[vetting.WebSearchOutput]},
		{"FinancialStabilityOutput", assertSchema[vetting.FinancialStabilityOutput]},
		{"CodeSecurityOutput", assertSchema[vetting.CodeSecurityOutput]},
		{"ThirdPartyComparisonOutput", assertSchema[vetting.ThirdPartyComparisonOutput]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.fn(t)
		})
	}
}

// assertSchema creates an OutputType for T and verifies that the
// generated JSON Schema has the expected shape: an object type with a
// non-empty properties map. This catches struct tags that silently
// produce empty or malformed schemas.
func assertSchema[T any](t *testing.T) {
	t.Helper()

	outputType, err := agent.NewOutputType[T]("test")
	require.NoError(t, err)
	require.NotNil(t, outputType)
	require.NotEmpty(t, outputType.Schema)

	var schema map[string]any
	require.NoError(t, json.Unmarshal(outputType.Schema, &schema))

	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "schema must expose a properties map")
	assert.NotEmpty(t, properties, "schema must declare at least one property")
}
