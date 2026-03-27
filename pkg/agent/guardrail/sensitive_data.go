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
		"sk-",

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
