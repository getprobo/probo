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

package testutil

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

var isolatedEnvStartMu sync.Mutex

type (
	// IsolatedEnv is a short-lived probod process with its own listen addresses.
	// It shares Postgres and Mailpit with the main e2e suite so identities created
	// against the default server remain visible, which is how a disable-signup
	// instance is bootstrapped in production (orgs exist before the flag flips).
	IsolatedEnv struct {
		BaseURL        string
		MailpitBaseURL string

		cmd       *exec.Cmd
		done      chan error
		outputBuf *lockedBuffer
	}

	IsolatedEnvOptions struct {
		DisableSignup bool
	}

	lockedBuffer struct {
		mu  sync.Mutex
		buf bytes.Buffer
	}

	reservedTCPPort struct {
		port     string
		listener net.Listener
	}
)

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Len()
}

func (b *lockedBuffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := make([]byte, b.buf.Len())
	copy(out, b.buf.Bytes())

	return out
}

// StartIsolatedEnv launches a second probod for gates that depend on process
// config (for example PROBOD_AUTH_DISABLE_SIGNUP). The process is stopped via
// t.Cleanup.
func StartIsolatedEnv(t testing.TB, opts IsolatedEnvOptions) *IsolatedEnv {
	t.Helper()

	binaryPath := os.Getenv("PROBO_E2E_BINARY")
	require.NotEmpty(t, binaryPath, "PROBO_E2E_BINARY is required")

	// Serialize allocate → config → release → start so parallel isolated envs
	// cannot steal each other's briefly-freed ports.
	isolatedEnvStartMu.Lock()
	startUnlocked := false
	defer func() {
		if !startUnlocked {
			isolatedEnvStartMu.Unlock()
		}
	}()

	apiPort := reserveTCPPort(t)
	metricsPort := reserveTCPPort(t)
	trustHTTPPort := reserveTCPPort(t)
	trustHTTPSPort := reserveTCPPort(t)

	ports := []reservedTCPPort{apiPort, metricsPort, trustHTTPPort, trustHTTPSPort}
	defer closeReservedPorts(ports)

	apiAddr := "localhost:" + apiPort.port
	baseURL := "http://" + apiAddr

	configPath, err := generateConfig(
		configOptions{
			DisableSignup:  opts.DisableSignup,
			APIAddr:        apiAddr,
			BaseURL:        baseURL,
			MetricsAddr:    "localhost:" + metricsPort.port,
			TrustHTTPAddr:  ":" + trustHTTPPort.port,
			TrustHTTPSAddr: ":" + trustHTTPSPort.port,
		},
	)
	require.NoError(t, err, "cannot generate isolated env config")

	env := &IsolatedEnv{
		BaseURL:        baseURL,
		MailpitBaseURL: GetMailpitBaseURL(),
		done:           make(chan error, 1),
		outputBuf:      &lockedBuffer{},
	}

	cmd := exec.Command(binaryPath, "-cfg-file", configPath, "-format", log.FormatPretty)
	cmd.Env = os.Environ()
	cmd.Stdout = env.outputBuf
	cmd.Stderr = env.outputBuf
	env.cmd = cmd

	closeReservedPorts(ports)

	require.NoError(t, cmd.Start(), "cannot start isolated probod")

	startUnlocked = true
	isolatedEnvStartMu.Unlock()

	go func() {
		env.done <- cmd.Wait()
	}()

	t.Cleanup(func() {
		env.stop()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := waitForServer(ctx, env.done, env.BaseURL+"/api/console/v1/graphql", 30*time.Second); err != nil {
		env.dumpOutput("isolated API server failed to start", err)
		require.NoError(t, err)
	}

	return env
}

func (e *IsolatedEnv) stop() {
	if e == nil || e.cmd == nil || e.cmd.Process == nil {
		return
	}

	_ = e.cmd.Process.Signal(syscall.SIGTERM)

	select {
	case <-e.done:
	case <-time.After(10 * time.Second):
		_ = e.cmd.Process.Kill()
		<-e.done
	}
}

func (e *IsolatedEnv) dumpOutput(contextMsg string, err error) {
	fmt.Fprintf(os.Stderr, "\n=== e2etest isolated env: %s: %v ===\n", contextMsg, err)

	if e.outputBuf == nil || e.outputBuf.Len() == 0 {
		fmt.Fprintf(os.Stderr, "e2etest: no captured isolated output available\n")
		return
	}

	output := e.outputBuf.Bytes()

	const maxTail = 10_000
	if len(output) > maxTail {
		fmt.Fprintf(os.Stderr, "e2etest: (showing last %d bytes of isolated output)\n", maxTail)
		output = output[len(output)-maxTail:]
	}

	fmt.Fprintf(os.Stderr, "--- isolated probod output start ---\n%s\n--- isolated probod output end ---\n", output)
}

func reserveTCPPort(t testing.TB) reservedTCPPort {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "cannot allocate free TCP port")

	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err, "cannot parse allocated TCP port")

	return reservedTCPPort{port: port, listener: listener}
}

func closeReservedPorts(ports []reservedTCPPort) {
	for i := range ports {
		if ports[i].listener != nil {
			_ = ports[i].listener.Close()
			ports[i].listener = nil
		}
	}
}
