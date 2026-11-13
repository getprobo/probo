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
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
)

func TestSkipNoisyTLSHandshake_DropsKnownFragments(t *testing.T) {
	t.Parallel()

	for _, fragment := range noisyTLSHandshakeFragments {
		t.Run(fragment, func(t *testing.T) {
			t.Parallel()

			assert.True(t, skipNoisyTLSHandshake(log.LevelError, fragment, nil))
		})
	}
}

func TestSkipNoisyTLSHandshake_KeepsOperationalErrors(t *testing.T) {
	t.Parallel()

	kept := []string{
		"http: Accept error: accept tcp 127.0.0.1:443: too many open files",
		"http: TLS handshake error from 192.0.2.1:443: tls: no certificates available",
		"http: TLS handshake error from 192.0.2.1:30050: remote error: tls: bad certificate",
	}

	for _, msg := range kept {
		t.Run(msg, func(t *testing.T) {
			t.Parallel()

			assert.False(t, skipNoisyTLSHandshake(log.LevelError, msg, nil))
		})
	}
}

func TestHTTPSServerLogger_DropsScannerNoise(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := log.NewLogger(
		log.WithName("probod"),
		log.WithOutput(&buf),
		log.WithFormat(log.FormatJSON),
	)
	httpsServerLogger := logger.Named("https", log.SkipMatch(skipNoisyTLSHandshake))
	server := httpserver.NewServer(
		"127.0.0.1:0",
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}),
		httpserver.WithLogger(httpsServerLogger),
		httpserver.WithRegisterer(prometheus.NewRegistry()),
	)

	server.ErrorLog.Printf(
		"http: TLS handshake error from 192.0.2.1:30050: tls: no key exchanges supported by both client and server",
	)
	assert.Empty(t, buf.String())

	const kept = "http: Accept error: accept tcp 127.0.0.1:443: too many open files"

	server.ErrorLog.Print(kept)

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	assert.Equal(t, "probod.https.http.server", entry["name"])
	assert.Equal(t, kept, entry["msg"])
	assert.Equal(t, "ERROR", entry["level"])
}
