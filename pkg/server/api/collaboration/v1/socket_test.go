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
		ReadMessages(ctx, serverConn, incoming)
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
	go ReadMessages(ctx, serverConn, incoming)

	require.NoError(t, client.Close(websocket.StatusNormalClosure, ""))

	select {
	case message := <-incoming:
		require.Error(t, message.Err)
		assert.NoError(t, ctx.Err())
	case <-ctx.Done():
		require.Fail(t, "timed out waiting for the close error")
	}
}

func acceptWebSocket(t *testing.T) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	accepted := make(chan *websocket.Conn, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
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
