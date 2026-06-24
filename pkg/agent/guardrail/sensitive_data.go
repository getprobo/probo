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

package guardrail

import (
	"context"
	"strings"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

type SensitiveDataGuardrail struct{}

func NewSensitiveDataGuardrail() *SensitiveDataGuardrail {
	return &SensitiveDataGuardrail{}
}

func (g *SensitiveDataGuardrail) Name() string {
	return "sensitive-data"
}

func (g *SensitiveDataGuardrail) Check(_ context.Context, message llm.Message) (*agent.GuardrailResult, error) {
	text := strings.ToLower(message.Text())

	sensitivePatterns := []string{
		// Slack tokens
		"xoxb-",
		"xoxp-",
		"xoxa-",
		"xoxs-",
		"xapp-",

		// GitHub tokens
		"ghp_",
		"gho_",
		"ghu_",
		"ghs_",
		"ghr_",

		// Cloud provider keys
		"akia", // AWS access key ID

		// Payment provider keys
		"sk_live_",
		"sk_test_",

		// LLM provider keys
		"sk-proj-", // OpenAI
		"sk-ant-",  // Anthropic

		// JWT tokens
		"eyj", // base64-encoded JSON header

		// Authorization headers
		"bearer ",
		"basic ",

		// PEM / certificates
		"-----begin",

		// Connection strings
		"postgres://",
		"postgresql://",
		"mongodb://",
		"mysql://",
		"redis://",
		"amqp://",

		// Generic secret field names
		"encryption_key",
		"signing_secret",
		"secret_key",
		"private_key",
		"client_secret",
		"access_token",
		"api_key",
		"apikey",
		"password",

		// Raw SQL
		"select ",
		"insert into",
		"update ",
		"delete from",
		"drop table",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(text, pattern) {
			return &agent.GuardrailResult{
				Tripwire: true,
				Message:  "response contains potentially sensitive data",
			}, nil
		}
	}

	return &agent.GuardrailResult{Tripwire: false}, nil
}
