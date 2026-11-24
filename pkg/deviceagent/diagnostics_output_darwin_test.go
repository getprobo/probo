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

//go:build darwin

package deviceagent_test

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/deviceagent"
)

func TestWriteDiagnosticsOutput_RoundTrip(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("/tmp", "probo-diag-")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	path := filepath.Join(dir, "s")
	listener, err := net.Listen("unix", path)
	require.NoError(t, err)

	t.Cleanup(func() { _ = listener.Close() })

	got := make(chan string, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			got <- ""

			return
		}

		defer func() { _ = conn.Close() }()

		data, err := io.ReadAll(conn)
		if err != nil {
			got <- ""

			return
		}

		got <- string(data)
	}()

	require.NoError(t, deviceagent.WriteDiagnosticsOutput(path, "hello\n", os.Getpid()))
	assert.Equal(t, "hello\n", <-got)
}

func TestWriteDiagnosticsOutput_RejectsRegularFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "report.txt")
	require.NoError(t, os.WriteFile(path, []byte("x"), 0o600))

	err := deviceagent.WriteDiagnosticsOutput(path, "report", os.Getpid())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "socket")
}

func TestWriteDiagnosticsOutput_RejectsOtherListener(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("/tmp", "probo-diag-")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	path := filepath.Join(dir, "s")
	listener, err := net.Listen("unix", path)
	require.NoError(t, err)

	t.Cleanup(func() { _ = listener.Close() })

	got := make(chan string, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			got <- "accept-failed"

			return
		}

		defer func() { _ = conn.Close() }()

		data, err := io.ReadAll(conn)
		if err != nil {
			got <- "read-failed"

			return
		}

		got <- string(data)
	}()

	err = deviceagent.WriteDiagnosticsOutput(path, "secret", os.Getpid()+1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pid")
	assert.Empty(t, <-got)
}
