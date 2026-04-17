// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
		{"VendorComparisonOutput", assertSchema[vetting.VendorComparisonOutput]},
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
