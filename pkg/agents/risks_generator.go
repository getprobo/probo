// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package agents

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
)

//go:embed data/risks.json
var risksJSONContent []byte

type (
	OrganizationInfo struct {
		Name                  *string `json:"name"`
		FoundingYear          *int    `json:"founding_year"`
		CompanyType           *string `json:"company_type"`
		PremarketFit          *bool   `json:"premarket_fit"`
		UsesCloudProviders    *bool   `json:"uses_cloud_providers"`
		AIFocused             *bool   `json:"ai_focused"`
		UsesAIGeneratedCode   *bool   `json:"uses_ai_generated_code"`
		VCBacked              *bool   `json:"vc_backed"`
		HasRaisedMoney        *bool   `json:"has_raised_money"`
		HasEnterpriseAccounts *bool   `json:"has_enterprise_accounts"`
		PeopleCount           *int    `json:"people_count"`
	}
)

const (
	risksGeneratorSystemPrompt = `
		# Role:
		You are an assistant that determines key risks required for addressing compliance, legal, and security risks in an organization.

		# Objective
		Given the organization's characteristics and a list of legal/compliance/security risks (in JSON format),
		identify among the given risks the most relevant ones for the organization.

		# Response Format
		Respond with a single comma-separated list of relevant **risk names only**. Do not include categories, descriptions, explanations, or any other metadata.

		# SOP
		- Consider the organization's structure, maturity, size, funding, and exposure to risk.
		- Include only risks clearly applicable to the organization's context.
		- Prioritize coverage of high-impact regulatory, legal, and operational concerns.
		- Only return risks that are in the JSON file.
		- Only use the risk names from the JSON file.

		# Example output format:
		Regulatory penalty with GDPR non-compliance, Ransomware locking critical systems, Revenue loss due to pricing strategy
	`
)

func (a *Agent) GenerateRisks(ctx context.Context, organizationInfo OrganizationInfo) ([]string, error) {
	orgJSON, err := json.Marshal(organizationInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal organization info: %w", err)
	}

	model := openai.ChatModel(a.cfg.ModelName)
	chatCompletion, err := a.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(risksGeneratorSystemPrompt),
			openai.UserMessage(fmt.Sprintf(`Company info JSON: %s`, orgJSON)),
			openai.UserMessage(fmt.Sprintf(`Risks JSON: %s`, string(risksJSONContent))),
		},
		Model:       model,
		Temperature: param.NewOpt(a.cfg.Temperature),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get completion: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned from API")
	}

	content := chatCompletion.Choices[0].Message.Content
	riskNames := strings.Split(content, ",")
	for i := range riskNames {
		riskNames[i] = strings.TrimSpace(riskNames[i])
	}
	return riskNames, nil
}
