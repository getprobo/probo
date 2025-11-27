// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

// Package testutil provides end-to-end testing infrastructure for the Probo API.
//
// This package runs an external probod binary for realistic e2e testing.
// It supports coverage collection when using a coverage-instrumented binary.
//
// Required environment variables:
//   - PROBO_E2E_BINARY: Path to the probod binary
//   - PROBO_E2E_CONFIG: Path to the config file
//
// Optional environment variables:
//   - PROBO_E2E_COVERDIR: Directory for coverage data (enables coverage collection)
//
// Example usage:
//
//	# Build the binary (with coverage)
//	go build -cover -o bin/probod-coverage ./cmd/probod
//
//	# Run e2e tests
//	PROBO_E2E_BINARY=./bin/probod-coverage \
//	PROBO_E2E_COVERDIR=./coverage/e2e \
//	PROBO_E2E_CONFIG=./e2e/console/testdata/config.yaml \
//	go test -v ./e2e/console/...
//
//	# Generate coverage report
//	go tool covdata textfmt -i=./coverage/e2e -o=coverage-e2e.out
//	go tool cover -html=coverage-e2e.out -o=coverage-e2e.html
package testutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var (
	testEnv   *TestEnv
	setupOnce sync.Once
)

// TestEnv holds the test environment state
type TestEnv struct {
	BaseURL string
	cmd     *exec.Cmd
	done    chan error
}

// Setup initializes the test environment. Call this from TestMain.
// It starts probod with the provided configuration and waits for it to be ready.
//
// Example:
//
//	func TestMain(m *testing.M) {
//	    testutil.Setup()
//	    code := m.Run()
//	    testutil.Teardown()
//	    os.Exit(code)
//	}
func Setup() {
	setupOnce.Do(func() {
		binaryPath := os.Getenv("PROBO_E2E_BINARY")
		configPath := os.Getenv("PROBO_E2E_CONFIG")
		coverDir := os.Getenv("PROBO_E2E_COVERDIR")

		if binaryPath == "" {
			fmt.Fprintf(os.Stderr, "e2etest: PROBO_E2E_BINARY is required\n")
			os.Exit(1)
		}

		if configPath == "" {
			fmt.Fprintf(os.Stderr, "e2etest: PROBO_E2E_CONFIG is required\n")
			os.Exit(1)
		}

		// Create coverage directory if specified
		if coverDir != "" {
			if err := os.MkdirAll(coverDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "e2etest: cannot create coverage directory: %v\n", err)
				os.Exit(1)
			}
		}

		testEnv = &TestEnv{
			done: make(chan error, 1),
		}

		// Start the external binary
		// Note: We use exec.Command instead of exec.CommandContext because
		// CommandContext sends SIGKILL on context cancel, which prevents the
		// binary from writing coverage data. We manage the process lifecycle
		// manually in Teardown() using SIGTERM for graceful shutdown.
		cmd := exec.Command(binaryPath, "-cfg-file", configPath)
		if coverDir != "" {
			cmd.Env = append(os.Environ(), "GOCOVERDIR="+coverDir)
		} else {
			cmd.Env = os.Environ()
		}
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard

		testEnv.cmd = cmd

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "e2etest: cannot start binary: %v\n", err)
			os.Exit(1)
		}

		// Wait for process to exit in background
		go func() {
			err := cmd.Wait()
			testEnv.done <- err
		}()

		// TODO: Parse config file to get actual port
		testEnv.BaseURL = "http://localhost:18080"

		// Wait for server to be ready
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := waitForServer(ctx, testEnv.BaseURL, 30*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "e2etest: server failed to start: %v\n", err)
			testEnv.cmd.Process.Kill()
			os.Exit(1)
		}
	})
}

func waitForServer(ctx context.Context, baseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/console/v1/query", nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			// Any response means server is up
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}

// Teardown shuts down the test environment. Call this after m.Run() in TestMain.
func Teardown() {
	if testEnv == nil {
		return
	}

	if testEnv.cmd != nil && testEnv.cmd.Process != nil {
		// Send SIGTERM for graceful shutdown (allows coverage data to be written)
		testEnv.cmd.Process.Signal(syscall.SIGTERM)

		// Wait for graceful shutdown with timeout
		select {
		case <-testEnv.done:
		case <-time.After(10 * time.Second):
			testEnv.cmd.Process.Kill()
			<-testEnv.done
		}
	}
}

// GetBaseURL returns the base URL of the test server
func GetBaseURL() string {
	if testEnv == nil {
		return "http://localhost:8080"
	}
	return testEnv.BaseURL
}
