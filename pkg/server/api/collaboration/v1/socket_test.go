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

package collaboration_v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMessages_CanceledContext(t *testing.T) {
	t.Parallel()

	serverConn, _ := acceptWebSocket(t)
	ctx, cancel := context.WithCancel(context.Background())
	incoming := make(chan Incoming, 1)
	done := make(chan struct{})

	go func() {
		_ = ReadMessages(ctx, serverConn, incoming)

		close(done)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		require.Fail(t, "ReadMessages did not return after cancel")
	}

	select {
	case message := <-incoming:
		assert.Fail(t, "canceled read was forwarded", message.Err)
	default:
	}
}

func TestReadMessages_ClosedConnection(t *testing.T) {
	t.Parallel()

	serverConn, client := acceptWebSocket(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	incoming := make(chan Incoming, 1)
	done := make(chan struct{})

	go func() {
		_ = ReadMessages(ctx, serverConn, incoming)

		close(done)
	}()

	require.NoError(t, client.Close(websocket.StatusNormalClosure, ""))

	select {
	case message := <-incoming:
		require.Error(t, message.Err)
		assert.NoError(t, ctx.Err())
	case <-ctx.Done():
		require.Fail(t, "timed out waiting for the close error")
	}

	select {
	case <-done:
	case <-ctx.Done():
		require.Fail(t, "ReadMessages did not return after close")
	}
}

func acceptWebSocket(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	accepted := make(chan *websocket.Conn, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}

		accepted <- connection

		<-release

		_ = connection.Close(websocket.StatusNormalClosure, "")
	}))

	t.Cleanup(func() {
		close(release)
		server.Close()
	})

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(dialCancel)

	client, _, err := websocket.Dial(
		dialCtx,
		"ws"+strings.TrimPrefix(server.URL, "http"),
		nil,
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = client.Close(websocket.StatusNormalClosure, "")
	})

	select {
	case connection := <-accepted:
		return connection, client
	case <-dialCtx.Done():
		require.NoError(t, dialCtx.Err())
		return nil, nil
	}
}

func TestRepoReadResult(t *testing.T) {
	t.Parallel()

	t.Run("empty close frame", func(t *testing.T) {
		t.Parallel()

		err := fmt.Errorf(
			"failed to get reader: %w",
			websocket.CloseError{Code: websocket.StatusNoStatusRcvd},
		)

		assert.NoError(t, repoReadResult(err))
	})

	t.Run("normal closure", func(t *testing.T) {
		t.Parallel()

		err := websocket.CloseError{Code: websocket.StatusNormalClosure}

		assert.NoError(t, repoReadResult(err))
	})

	t.Run("protocol error", func(t *testing.T) {
		t.Parallel()

		err := websocket.CloseError{
			Code:   websocket.StatusProtocolError,
			Reason: "bad",
		}

		assert.ErrorContains(t, repoReadResult(err), "repo collaboration read failed")
	})
}

func TestRequestCanceled(t *testing.T) {
	t.Parallel()

	rollback := fmt.Errorf(
		"cannot rollback transaction: failed to deallocate cached statement(s): conn closed",
	)

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "commit on a canceled request",
			err: fmt.Errorf(
				"cannot open collaboration room: %w",
				fmt.Errorf("cannot commit transaction: %w", context.Canceled),
			),
			want: true,
		},
		{
			name: "load canceled and rollback lost the connection",
			err: fmt.Errorf(
				"cannot open collaboration room: %w",
				errors.Join(
					fmt.Errorf(
						"cannot load collaboration document: %w",
						fmt.Errorf("cannot collect collaboration document: %w", context.Canceled),
					),
					rollback,
				),
			),
			want: true,
		},
		{
			name: "database failure with a canceled rollback",
			err: fmt.Errorf(
				"cannot open collaboration room: %w",
				errors.Join(
					fmt.Errorf("cannot load collaboration document: %w", errors.New("connection reset")),
					fmt.Errorf("cannot rollback transaction: %w", context.Canceled),
				),
			),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, requestCanceled(tt.err))
		})
	}
}
