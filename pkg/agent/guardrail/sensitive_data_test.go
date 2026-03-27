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

package guardrail_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent/guardrail"
	"go.probo.inc/probo/pkg/llm"
)

func assistantMessage(text string) llm.Message {
	return llm.Message{
		Role:  llm.RoleAssistant,
		Parts: []llm.Part{llm.TextPart{Text: text}},
	}
}

func TestSensitiveDataGuardrail_Check(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		tripwire bool
	}{
		// Safe messages
		{"safe message", "Here is your compliance overview.", false},
		{"safe message with numbers", "You have 42 controls across 3 frameworks.", false},

		// Slack tokens
		{"slack bot token", "The token is xoxb-1234-abcd", true},
		{"slack user token", "Use xoxp-secret-token to authenticate", true},
		{"slack app token", "App token: xoxa-2-abc", true},
		{"slack session token", "Session: xoxs-abc123", true},
		{"slack app-level token", "Token: xapp-1-abc123", true},

		// GitHub tokens
		{"github personal access token", "Use ghp_abc123def456 for auth", true},
		{"github oauth token", "Token: gho_abc123", true},
		{"github user-to-server token", "Token: ghu_abc123", true},
		{"github server-to-server token", "Token: ghs_abc123", true},
		{"github refresh token", "Refresh: ghr_abc123", true},

		// Cloud provider keys
		{"aws access key", "AWS key: AKIAIOSFODNN7EXAMPLE", true},

		// Payment provider keys
		{"stripe live key", "Stripe key: sk_live_abc123", true},
		{"stripe test key", "Stripe key: sk_test_abc123", true},

		// LLM provider keys
		{"openai key", "The API key is sk-proj-abc123", true},

		// JWT tokens
		{"jwt token", "Token: eyJhbGciOiJIUzI1NiJ9.payload.sig", true},

		// Authorization headers
		{"bearer auth", "Authorization: Bearer abc123", true},
		{"basic auth", "Authorization: Basic dXNlcjpwYXNz", true},
		{"bearer auth uppercase", "BEARER token123", true},

		// PEM / certificates
		{"pem private key", "-----BEGIN RSA PRIVATE KEY-----\nMIIE...", true},
		{"pem certificate", "-----BEGIN CERTIFICATE-----\nMIIE...", true},

		// Connection strings
		{"postgres uri", "Connect to postgres://user:pass@host/db", true},
		{"postgresql uri", "Connect to postgresql://user:pass@host/db", true},
		{"mongodb uri", "Use mongodb://user:pass@host/db", true},
		{"mysql uri", "Use mysql://user:pass@host/db", true},
		{"redis uri", "Cache at redis://localhost:6379", true},
		{"amqp uri", "Queue at amqp://guest:guest@host/vhost", true},

		// Generic secret field names
		{"encryption_key", "The encryption_key is set in config", true},
		{"signing_secret", "Your signing_secret was rotated", true},
		{"secret_key", "The secret_key value is abc", true},
		{"private_key", "Set private_key in the env", true},
		{"client_secret", "The client_secret is abc123", true},
		{"access_token", "Use this access_token to call the API", true},
		{"api_key", "Your api_key is xyz", true},
		{"apikey", "Set apikey in headers", true},
		{"password", "Your password is hunter2", true},

		// Raw SQL
		{"sql select", "SELECT * FROM users WHERE id = 1", true},
		{"sql insert", "INSERT INTO users VALUES (1, 'admin')", true},
		{"sql update", "UPDATE users SET name = 'foo'", true},
		{"sql delete", "DELETE FROM users WHERE id = 1", true},
		{"sql drop", "DROP TABLE users", true},
	}

	g := guardrail.NewSensitiveDataGuardrail()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := g.Check(context.Background(), assistantMessage(tt.text))

			require.NoError(t, err)
			assert.Equal(t, tt.tripwire, result.Tripwire)
		})
	}
}
