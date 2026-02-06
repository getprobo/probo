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

	"go.gearno.de/kit/log"
)

var (
	testEnv   *TestEnv
	setupOnce sync.Once
)

type TestEnv struct {
	BaseURL string
	cmd     *exec.Cmd
	done    chan error
}

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

		cmd := exec.Command(binaryPath, "-cfg-file", configPath, "-format", log.FormatPretty)
		if coverDir != "" {
			cmd.Env = append(os.Environ(), "GOCOVERDIR="+coverDir)
		} else {
			cmd.Env = os.Environ()
		}
		if os.Getenv("PROBO_E2E_VERBOSE") != "" {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = io.Discard
			cmd.Stderr = io.Discard
		}

		testEnv.cmd = cmd

		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "e2etest: cannot start binary: %v\n", err)
			os.Exit(1)
		}

		go func() {
			err := cmd.Wait()
			testEnv.done <- err
		}()

		testEnv.BaseURL = "http://localhost:18080"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := waitForServer(ctx, testEnv.BaseURL, 30*time.Second); err != nil {
			fmt.Fprintf(os.Stderr, "e2etest: server failed to start: %v\n", err)
			_ = testEnv.cmd.Process.Kill()
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

		req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/api/console/v1/graphql", nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			// Any response means server is up
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server did not become ready within %v", timeout)
}

func Teardown() {
	if testEnv == nil {
		return
	}

	if testEnv.cmd != nil && testEnv.cmd.Process != nil {
		_ = testEnv.cmd.Process.Signal(syscall.SIGTERM)

		select {
		case <-testEnv.done:
		case <-time.After(10 * time.Second):
			_ = testEnv.cmd.Process.Kill()
			<-testEnv.done
		}
	}
}

func GetBaseURL() string {
	if testEnv == nil {
		return "http://localhost:8080"
	}
	return testEnv.BaseURL
}
