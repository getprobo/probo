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

package accessreview_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

type probeFunc func(context.Context, *http.Client, *coredata.Connector, provider.Endpoints) error

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

func newCheckService(t *testing.T, checkSettings bool, logs io.Writer, probe probeFunc) *accessreview.Service {
	t.Helper()

	registry := provider.NewRegistry()
	require.NoError(
		t,
		registry.Register(
			&provider.Registration{
				Provider:    coredata.ConnectorProviderBetterStack,
				DisplayName: "Better Stack",
				APIKey: &provider.APIKeyConfig{
					ExtraSettings: []provider.ExtraSetting{
						{Key: "teamName", Label: "Team Name", Required: true},
					},
					CheckSettings: checkSettings,
				},
				Probe: probe,
			},
		),
	)

	return accessreview.NewService(
		nil,
		cipher.EncryptionKey{},
		connector.NewConnectorRegistry(),
		registry,
		log.NewLogger(log.WithOutput(logs)),
	)
}

func checkBetterStack(t *testing.T, svc *accessreview.Service) *drivers.SettingRejectedError {
	t.Helper()

	return svc.CheckNewAPIKeyConnector(
		context.Background(),
		coredata.ConnectorProviderBetterStack,
		&connector.APIKeyConnection{APIKey: "key"},
		json.RawMessage(`{"team_name":"Acme"}`),
	)
}

func TestCheckNewAPIKeyConnector(t *testing.T) {
	t.Parallel()

	t.Run("returns the refused setting", func(t *testing.T) {
		t.Parallel()

		want := &drivers.SettingRejectedError{Code: drivers.BetterStackTeamNotFound, Setting: "teamName"}

		svc := newCheckService(
			t,
			true,
			io.Discard,
			func(context.Context, *http.Client, *coredata.Connector, provider.Endpoints) error {
				return want
			},
		)

		assert.Same(t, want, checkBetterStack(t, svc))
	})

	t.Run("probes the unsaved connector", func(t *testing.T) {
		t.Parallel()

		var got *coredata.Connector

		svc := newCheckService(
			t,
			true,
			io.Discard,
			func(ctx context.Context, _ *http.Client, conn *coredata.Connector, _ provider.Endpoints) error {
				_, hasDeadline := ctx.Deadline()
				assert.True(t, hasDeadline)

				got = conn

				return nil
			},
		)

		require.Nil(t, checkBetterStack(t, svc))
		require.NotNil(t, got)
		assert.Equal(t, coredata.ConnectorProtocolAPIKey, got.Protocol)

		settings, err := coredata.ConnectorSettings[coredata.BetterStackConnectorSettings](got)
		require.NoError(t, err)
		assert.Equal(t, "Acme", settings.TeamName)
	})

	t.Run("refused key is left to the connection status", func(t *testing.T) {
		t.Parallel()

		svc := newCheckService(
			t,
			true,
			io.Discard,
			func(context.Context, *http.Client, *coredata.Connector, provider.Endpoints) error {
				return &provider.CredentialRejectedError{StatusCode: http.StatusUnauthorized}
			},
		)

		assert.Nil(t, checkBetterStack(t, svc))
	})

	t.Run("unanswered check logs no settings", func(t *testing.T) {
		t.Parallel()

		var logs syncBuffer

		svc := newCheckService(
			t,
			true,
			&logs,
			func(context.Context, *http.Client, *coredata.Connector, provider.Endpoints) error {
				return &url.Error{
					Op:  "Get",
					URL: "https://betterstack.com/api/v2/team-members?team_name=Acme",
					Err: context.DeadlineExceeded,
				}
			},
		)

		assert.Nil(t, checkBetterStack(t, svc))
		assert.Contains(t, logs.String(), "cannot check connector before save")
		assert.NotContains(t, logs.String(), "Acme")
	})

	t.Run("provider without a settings check is not probed", func(t *testing.T) {
		t.Parallel()

		svc := newCheckService(
			t,
			false,
			io.Discard,
			func(context.Context, *http.Client, *coredata.Connector, provider.Endpoints) error {
				t.Error("the probe must not run")

				return nil
			},
		)

		assert.Nil(t, checkBetterStack(t, svc))
	})
}
