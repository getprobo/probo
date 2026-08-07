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
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
)

// IsolatedEnv is a short-lived probod process with its own listen addresses.
// It shares Postgres and Mailpit with the main e2e suite so identities created
// against the default server remain visible, which is how a disable-signup
// instance is bootstrapped in production (orgs exist before the flag flips).
type IsolatedEnv struct {
	BaseURL        string
	MailpitBaseURL string

	cmd       *exec.Cmd
	done      chan error
	outputBuf *bytes.Buffer
}

type IsolatedEnvOptions struct {
	DisableSignup bool
}

// StartIsolatedEnv launches a second probod for gates that depend on process
// config (for example PROBOD_AUTH_DISABLE_SIGNUP). The process is stopped via
// t.Cleanup.
func StartIsolatedEnv(t testing.TB, opts IsolatedEnvOptions) *IsolatedEnv {
	t.Helper()

	binaryPath := os.Getenv("PROBO_E2E_BINARY")
	require.NotEmpty(t, binaryPath, "PROBO_E2E_BINARY is required")

	apiPort := freeTCPPort(t)
	metricsPort := freeTCPPort(t)
	trustHTTPPort := freeTCPPort(t)
	trustHTTPSPort := freeTCPPort(t)

	apiAddr := "localhost:" + apiPort
	baseURL := "http://" + apiAddr

	configPath, err := generateConfig(configOptions{
		DisableSignup:  opts.DisableSignup,
		APIAddr:        apiAddr,
		BaseURL:        baseURL,
		MetricsAddr:    "localhost:" + metricsPort,
		TrustHTTPAddr:  ":" + trustHTTPPort,
		TrustHTTPSAddr: ":" + trustHTTPSPort,
	})
	require.NoError(t, err, "cannot generate isolated env config")

	env := &IsolatedEnv{
		BaseURL:        baseURL,
		MailpitBaseURL: GetMailpitBaseURL(),
		done:           make(chan error, 1),
	}

	cmd := exec.Command(binaryPath, "-cfg-file", configPath, "-format", log.FormatPretty)
	cmd.Env = os.Environ()

	var buf bytes.Buffer

	env.outputBuf = &buf
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	env.cmd = cmd

	require.NoError(t, cmd.Start(), "cannot start isolated probod")

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

func freeTCPPort(t testing.TB) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "cannot allocate free TCP port")

	defer func() { _ = listener.Close() }()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err, "cannot parse allocated TCP port")

	return port
}
