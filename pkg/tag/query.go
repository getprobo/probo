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

package tag

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/searchquery"
	"go.probo.inc/probo/pkg/validator"
)

type ParsedQuery struct {
	Tags        []coredata.TagSelector
	EntityTypes []uint16
	Text        string
}

// entityTypeByName is the query-string name for each allowlisted type.
// Names are the GraphQL type in kebab case.
var entityTypeByName = map[string]uint16{
	"ai-system":                         coredata.AiSystemEntityType,
	"asset":                             coredata.AssetEntityType,
	"audit":                             coredata.AuditEntityType,
	"business-function":                 coredata.BusinessFunctionEntityType,
	"control":                           coredata.ControlEntityType,
	"data-protection-impact-assessment": coredata.DataProtectionImpactAssessmentEntityType,
	"datum":                             coredata.DatumEntityType,
	"device":                            coredata.DeviceEntityType,
	"document":                          coredata.DocumentEntityType,
	"evidence":                          coredata.EvidenceEntityType,
	"finding":                           coredata.FindingEntityType,
	"framework":                         coredata.FrameworkEntityType,
	"internal-control":                  coredata.InternalControlEntityType,
	"obligation":                        coredata.ObligationEntityType,
	"processing-activity":               coredata.ProcessingActivityEntityType,
	"rights-request":                    coredata.RightsRequestEntityType,
	"risk":                              coredata.RiskEntityType,
	"risk-analysis":                     coredata.RiskAnalysisEntityType,
	"statement-of-applicability":        coredata.StatementOfApplicabilityEntityType,
	"task":                              coredata.TaskEntityType,
	"third-party":                       coredata.ThirdPartyEntityType,
	"transfer-impact-assessment":        coredata.TransferImpactAssessmentEntityType,
}

// ParseQuery reads a search query. Bare words are required text prefixes.
// Every tag:name must match. type:name is optional and repeatable.
func ParseQuery(raw string) (ParsedQuery, error) {
	if strings.TrimSpace(raw) == "" {
		v := validator.New()
		v.Check(raw, "query", validator.Required())

		return ParsedQuery{}, v.Error()
	}

	ast, err := searchquery.Parse(raw)
	if err != nil {
		return ParsedQuery{}, queryError(err.Error())
	}

	tags := make([]coredata.TagSelector, 0)
	seenTags := make(map[string]struct{})
	entityTypes := make([]uint16, 0)
	seenTypes := make(map[uint16]struct{})
	textTerms := make([]string, 0)

	for _, clause := range ast.Clauses {
		switch c := clause.(type) {
		case searchquery.TagClause:
			if _, ok := seenTags[c.Name]; ok {
				continue
			}

			seenTags[c.Name] = struct{}{}
			tags = append(tags, coredata.TagSelector{Name: c.Name})
		case searchquery.TypeClause:
			entityType, ok := entityTypeByName[c.Name]
			if !ok {
				return ParsedQuery{}, queryError(fmt.Sprintf("unknown type %q", c.Name))
			}

			if _, ok := seenTypes[entityType]; ok {
				continue
			}

			seenTypes[entityType] = struct{}{}
			entityTypes = append(entityTypes, entityType)
		case searchquery.TextClause:
			if c.Value == "" {
				continue
			}

			textTerms = append(textTerms, c.Value)
		}
	}

	if len(tags) == 0 && len(textTerms) == 0 {
		return ParsedQuery{}, queryError("at least one word or tag:name is required")
	}

	return ParsedQuery{
		Tags:        tags,
		EntityTypes: entityTypes,
		Text:        strings.Join(textTerms, " "),
	}, nil
}

func queryError(message string) error {
	v := validator.New()
	v.Check("", "query", func(any) *validator.ValidationError {
		return &validator.ValidationError{
			Code:    validator.ErrorCodeInvalidFormat,
			Message: message,
		}
	})

	return v.Error()
}
