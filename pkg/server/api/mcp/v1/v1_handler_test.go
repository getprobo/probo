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

package mcp_v1_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	cloudaws "go.probo.inc/probo/pkg/cloud/aws"
	cloudazure "go.probo.inc/probo/pkg/cloud/azure"
	cloudgcp "go.probo.inc/probo/pkg/cloud/gcp"
	mcp_v1 "go.probo.inc/probo/pkg/server/api/mcp/v1"
)

func TestNewMux_UnauthenticatedChallenge(t *testing.T) {
	t.Parallel()

	mux := mcp_v1.NewMux(
		log.NewLogger(log.WithOutput(io.Discard)),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		baseurl.MustParse("https://probo.example.com"),
		nil,
		cloudaws.ConnectorInstallConfig{},
		cloudgcp.ConnectorInstallConfig{},
		cloudazure.ConnectorInstallConfig{},
		"",
		nil,
		nil,
	)

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(
		t,
		`Bearer resource_metadata="https://probo.example.com/.well-known/oauth-protected-resource/api/mcp/v1"`,
		rec.Header().Get("WWW-Authenticate"),
	)
}
