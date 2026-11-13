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

package probodconfig_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/probodconfig"
)

func TestLLMProviderConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     probodconfig.LLMProviderConfig
		wantErr string
	}{
		{
			name: "empty base url is the provider default",
			cfg:  probodconfig.LLMProviderConfig{Type: "openai", APIKey: "k"},
		},
		{
			name: "https is always allowed",
			cfg: probodconfig.LLMProviderConfig{
				Type:    "anthropic",
				BaseURL: "https://api.anthropic.com",
			},
		},
		{
			name: "http to loopback is allowed",
			cfg: probodconfig.LLMProviderConfig{
				Type:    "openai_compatible",
				BaseURL: "http://127.0.0.1:8080/v1",
			},
		},
		{
			name: "http to localhost is allowed",
			cfg: probodconfig.LLMProviderConfig{
				Type:    "openai_compatible",
				BaseURL: "http://localhost:11434/v1",
			},
		},
		{
			name: "http to a private address is allowed",
			cfg: probodconfig.LLMProviderConfig{
				Type:    "openai_compatible",
				BaseURL: "http://10.1.2.3:8000/v1",
			},
		},
		{
			name: "http to a single-label service name is allowed",
			cfg: probodconfig.LLMProviderConfig{
				Type:    "openai_compatible",
				BaseURL: "http://llama-server:8080/v1",
			},
		},
		{
			name: "http to a public host is rejected",
			cfg: probodconfig.LLMProviderConfig{
				Type:    "openai_compatible",
				BaseURL: "http://api.example.com/v1",
			},
			wantErr: "must use https unless the host is local",
		},
		{
			name: "openai_compatible requires a base url",
			cfg: probodconfig.LLMProviderConfig{
				Type:   "openai_compatible",
				APIKey: "k",
			},
			wantErr: "requires a base URL",
		},
		{
			name: "a scheme that is neither http nor https is rejected",
			cfg: probodconfig.LLMProviderConfig{
				Type:    "openai_compatible",
				BaseURL: "ftp://example.com/v1",
			},
			wantErr: "must be http or https",
		},
		{
			name: "a relative url is rejected",
			cfg: probodconfig.LLMProviderConfig{
				Type:    "openai",
				BaseURL: "/v1",
			},
			wantErr: "must be an absolute URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.Validate()

			if tt.wantErr == "" {
				assert.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestLLMProviderConfig_IsZero(t *testing.T) {
	t.Parallel()

	t.Run("fully empty is zero", func(t *testing.T) {
		t.Parallel()

		assert.True(t, probodconfig.LLMProviderConfig{Type: "openai"}.IsZero())
	})

	t.Run("an api key alone is configured", func(t *testing.T) {
		t.Parallel()

		assert.False(t, probodconfig.LLMProviderConfig{Type: "openai", APIKey: "k"}.IsZero())
	})

	t.Run("a base url alone is configured", func(t *testing.T) {
		t.Parallel()

		// The self-hosted case: a local server with no authentication. Keying
		// this on APIKey alone dropped such a provider from the registry, and
		// it surfaced only as "unknown LLM provider" at startup.
		cfg := probodconfig.LLMProviderConfig{
			Type:    "openai_compatible",
			BaseURL: "http://llama-server:8080/v1",
		}

		assert.False(t, cfg.IsZero())
	})
}
