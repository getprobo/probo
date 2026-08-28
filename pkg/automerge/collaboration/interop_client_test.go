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

package collaboration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/collaboration"
)

// The document id both sides agree on: the real repo client requests this URL
// and the gateway serves its seeded document under the same id.
const (
	interopDocumentID  = "34YWzjYt5gPJpq5RfXAkPfPcUj1r"
	interopDocumentURL = "automerge:" + interopDocumentID
)

// TestInterop_RealRepoClientLoadsGoDocument stands up a Go WebSocket server
// backed by ServerConn and a seeded document, then runs a real
// @automerge/automerge-repo client against it and asserts the client
// materializes the server's document. This validates the whole gateway stack
// (handshake, framing, sync loop) against the actual JavaScript client rather
// than against our reading of its source.
//
// It is gated on AUTOMERGE_REPO_INTEROP_CLIENT (the path to the Node client
// script) so the default test run does not require Node; the
// test-automerge-repo-interop make target sets it.
func TestInterop_RealRepoClientLoadsGoDocument(t *testing.T) {
	script := os.Getenv("AUTOMERGE_REPO_INTEROP_CLIENT")
	if script == "" {
		t.Skip("AUTOMERGE_REPO_INTEROP_CLIENT is not set")
	}

	ctx := context.Background()

	server, err := automerge.New(actor(1))
	require.NoError(t, err)

	defer func() { _ = server.Close() }()

	text, err := server.CreateText("body")
	require.NoError(t, err)

	largeBody := "hello world" + strings.Repeat("x", 128<<10)
	require.NoError(t, text.Splice(0, 0, largeBody))

	_, err = server.Commit("seed", commitTime())
	require.NoError(t, err)

	httpServer := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				serveGateway(t, w, r, server)
			},
		),
	)
	defer httpServer.Close()

	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	runCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	command := exec.CommandContext(runCtx, "node", script, wsURL, interopDocumentURL)

	output, err := command.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("interop client failed: %v\nstderr:\n%s", err, exitErr.Stderr)
		}

		t.Fatalf("cannot run interop client: %v", err)
	}

	var document map[string]any
	require.NoError(t, json.Unmarshal(output, &document))
	assert.Equal(t, largeBody, document["body"])
}

// serveGateway is a minimal repo gateway for the interop test: it accepts a
// binary WebSocket, runs the ServerConn handshake and sync loop for the seeded
// document, and echoes each peer's own ephemerals (there is a single client, so
// fan-out is a no-op).
func serveGateway(t *testing.T, w http.ResponseWriter, r *http.Request, document *automerge.Document) {
	t.Helper()

	connection, err := websocket.Accept(
		w,
		r,
		&websocket.AcceptOptions{
			InsecureSkipVerify: true, // test-only: the Node client sends no Origin
		},
	)
	if err != nil {
		return
	}

	defer func() { _ = connection.Close(websocket.StatusNormalClosure, "") }()

	connection.SetReadLimit(collaboration.MaxWireFrameBytes)

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	syncState, err := document.NewSyncState()
	if err != nil {
		return
	}

	defer func() { _ = syncState.Close() }()

	conn, err := collaboration.NewServerConn(
		collaboration.ServerConfig{ServerPeerID: "probo-gateway"},
		interopDocumentID,
		syncState,
	)
	if err != nil {
		return
	}

	write := func(frames [][]byte) bool {
		for _, frame := range frames {
			if err := connection.Write(ctx, websocket.MessageBinary, frame); err != nil {
				return false
			}
		}

		return true
	}

	// First frame is the client's join.
	kind, join, err := connection.Read(ctx)
	if err != nil || kind != websocket.MessageBinary {
		return
	}

	out, accepted, err := conn.Start(join)
	if err != nil {
		return
	}

	if !write(out) || !accepted {
		return
	}

	for {
		kind, frame, err := connection.Read(ctx)
		if err != nil {
			return
		}

		if kind != websocket.MessageBinary {
			continue
		}

		reply, _, err := conn.Receive(frame)
		if err != nil {
			return
		}

		if !write(reply) {
			return
		}
	}
}
