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

package probod

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.opentelemetry.io/otel/trace/noop"
	"go.probo.inc/probo/pkg/llm"
	"go.probo.inc/probo/pkg/probodconfig"
)

func TestBuildLLMClient_BaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		provider string
		suffix   string
		path     string
	}{
		{name: "openai", provider: "openai", suffix: "/v1", path: "/v1/responses"},
		{name: "anthropic", provider: "anthropic", suffix: "", path: "/v1/messages"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var hits atomic.Int32

			server := httptest.NewTLSServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						hits.Add(1)
						assert.Equal(t, tt.path, r.URL.Path)
						w.WriteHeader(http.StatusBadRequest)
					},
				),
			)
			t.Cleanup(server.Close)

			pool := x509.NewCertPool()
			pool.AddCert(server.Certificate())

			impl := &Implm{}
			client, err := impl.buildLLMClient(
				probodconfig.LLMProviderConfig{
					Type:    tt.provider,
					APIKey:  "test-key",
					BaseURL: server.URL + tt.suffix,
				},
				log.NewLogger(log.WithOutput(io.Discard)),
				noop.NewTracerProvider(),
				prometheus.NewRegistry(),
				httpclient.WithTLSConfig(
					&tls.Config{
						RootCAs: pool,
					},
				),
			)
			require.NoError(t, err)

			maxTokens := 16
			_, err = client.ChatCompletion(
				context.Background(),
				&llm.ChatCompletionRequest{
					Model:     "test-model",
					MaxTokens: &maxTokens,
					Messages: []llm.Message{
						{
							Role:  llm.RoleUser,
							Parts: []llm.Part{llm.TextPart{Text: "ping"}},
						},
					},
				},
			)
			require.Error(t, err)
			assert.Equal(t, int32(1), hits.Load())
		})
	}
}

func TestBuildLLMClient_InvalidBaseURL(t *testing.T) {
	t.Parallel()

	impl := &Implm{}
	_, err := impl.buildLLMClient(
		probodconfig.LLMProviderConfig{
			Type:    "openai",
			APIKey:  "test-key",
			BaseURL: "example.openai.azure.com/openai/v1/",
		},
		log.NewLogger(log.WithOutput(io.Discard)),
		noop.NewTracerProvider(),
		prometheus.NewRegistry(),
	)
	require.Error(t, err)
}
