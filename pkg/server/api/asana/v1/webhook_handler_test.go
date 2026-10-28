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

package asana_v1_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	asanav1 "go.probo.inc/probo/pkg/server/api/asana/v1"
	syncasana "go.probo.inc/probo/pkg/task/sync/asana"
	asana "go.probo.inc/probo/pkg/task/sync/asana/client"
)

type stubWebhookEnqueuer struct {
	secret string
	err    error
	called bool
}

func (s *stubWebhookEnqueuer) StoreAsanaHookSecret(context.Context, gid.GID, string, string) error {
	return nil
}

func (s *stubWebhookEnqueuer) AsanaHookSecret(context.Context, gid.GID, string) (string, error) {
	return s.secret, nil
}

func (s *stubWebhookEnqueuer) EnqueueAsanaWebhook(context.Context, gid.GID, []byte) (bool, error) {
	s.called = true
	if s.err != nil {
		return false, s.err
	}

	return true, nil
}

func TestWebhookHandler_EnqueueErrorStatus(t *testing.T) {
	t.Parallel()

	connectorID := gid.New(gid.NewTenantID(), 1)
	secret := "hook-secret"
	body := []byte(`{"events":[]}`)

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{
			name:   "invalid json",
			err:    fmt.Errorf("cannot parse Asana webhook events: %w", syncasana.ErrInvalidAsanaWebhookBody),
			status: http.StatusBadRequest,
		},
		{
			name:   "database failure",
			err:    errors.New("cannot insert Asana webhook event: db is down"),
			status: http.StatusInternalServerError,
		},
		{
			name:   "persisted",
			status: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				queue := &stubWebhookEnqueuer{secret: secret, err: tt.err}
				mux := asanav1.NewMux(log.NewLogger(), queue)
				req := httptest.NewRequest(
					http.MethodPost,
					"/webhooks/"+connectorID.String()+"/nonce-1",
					bytes.NewReader(body),
				)
				req.Header.Set(asana.HeaderHookSignature, asana.Sign(secret, body))

				rec := httptest.NewRecorder()
				mux.ServeHTTP(rec, req)

				assert.Equal(t, tt.status, rec.Code)
				assert.True(t, queue.called)

				if tt.status == http.StatusBadRequest {
					assert.Contains(t, rec.Body.String(), "invalid json")
				}
			},
		)
	}
}

func TestWebhookHandler_InvalidJSONFromEnqueue(t *testing.T) {
	t.Parallel()

	connectorID := gid.New(gid.NewTenantID(), 1)
	secret := "hook-secret"
	body := []byte(`not-json`)

	queue := &stubWebhookEnqueuer{
		secret: secret,
		err: fmt.Errorf(
			"cannot parse Asana webhook events: %w: %w",
			syncasana.ErrInvalidAsanaWebhookBody,
			errors.New("unexpected end of JSON input"),
		),
	}
	mux := asanav1.NewMux(log.NewLogger(), queue)
	req := httptest.NewRequest(
		http.MethodPost,
		"/webhooks/"+connectorID.String()+"/nonce-1",
		bytes.NewReader(body),
	)
	req.Header.Set(asana.HeaderHookSignature, asana.Sign(secret, body))

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid json")
	assert.True(t, queue.called)
}
