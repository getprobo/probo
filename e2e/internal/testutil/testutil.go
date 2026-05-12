// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bootstrap"
)

var (
	testEnv   *TestEnv
	setupOnce sync.Once
)

type TestEnv struct {
	MailpitBaseURL string
	BaseURL        string
	cmd            *exec.Cmd
	done           chan error
	outputBuf      *bytes.Buffer
	outputWriter   *switchableWriter
}

type switchableWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (s *switchableWriter) Write(p []byte) (int, error) {
	s.mu.Lock()
	w := s.w
	s.mu.Unlock()
	return w.Write(p)
}

func (s *switchableWriter) switchTo(w io.Writer) {
	s.mu.Lock()
	s.w = w
	s.mu.Unlock()
}

func Setup() {
	setupOnce.Do(func() {
		binaryPath := os.Getenv("PROBO_E2E_BINARY")
		coverDir := os.Getenv("PROBO_E2E_COVERDIR")

		if binaryPath == "" {
			fmt.Fprintf(os.Stderr, "e2etest: PROBO_E2E_BINARY is required\n")
			os.Exit(1)
		}

		// Create coverage directory if specified
		if coverDir != "" {
			if err := os.MkdirAll(coverDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "e2etest: cannot create coverage directory: %v\n", err)
				os.Exit(1)
			}
		}

		configPath, err := generateConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "e2etest: cannot generate config: %v\n", err)
			os.Exit(1)
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

		verbose := os.Getenv("PROBO_E2E_VERBOSE") != ""
		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		} else {
			var buf bytes.Buffer
			testEnv.outputBuf = &buf
			sw := &switchableWriter{w: &buf}
			testEnv.outputWriter = sw
			cmd.Stdout = sw
			cmd.Stderr = sw
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
		testEnv.MailpitBaseURL = "http://localhost:8025"

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := waitForServer(ctx, testEnv.BaseURL+"/api/console/v1/graphql", 30*time.Second); err != nil {
			testEnv.dumpOutputOnFailure("API server failed to start", err)
			_ = testEnv.cmd.Process.Kill()
			os.Exit(1)
		}
		if err := waitForServer(ctx, testEnv.MailpitBaseURL+"/api/v1/messages", 30*time.Second); err != nil {
			testEnv.dumpOutputOnFailure("MailPit server failed to start", err)
			_ = testEnv.cmd.Process.Kill()
			os.Exit(1)
		}

		if !verbose {
			testEnv.outputWriter.switchTo(io.Discard)
		}
	})
}

func (e *TestEnv) dumpOutputOnFailure(context string, err error) {
	fmt.Fprintf(os.Stderr, "\n=== e2etest: %s: %v ===\n", context, err)

	select {
	case waitErr := <-e.done:
		if waitErr != nil {
			fmt.Fprintf(os.Stderr, "e2etest: process exited with error: %v\n", waitErr)
		} else {
			fmt.Fprintf(os.Stderr, "e2etest: process exited cleanly (unexpected)\n")
		}
	default:
		fmt.Fprintf(os.Stderr, "e2etest: process is still running\n")
	}

	if e.outputBuf != nil && e.outputBuf.Len() > 0 {
		output := e.outputBuf.Bytes()
		const maxTail = 10_000
		if len(output) > maxTail {
			fmt.Fprintf(os.Stderr, "e2etest: (showing last %d bytes of output)\n", maxTail)
			output = output[len(output)-maxTail:]
		}
		fmt.Fprintf(os.Stderr, "--- probod output start ---\n%s\n--- probod output end ---\n", output)
	} else {
		fmt.Fprintf(os.Stderr, "e2etest: no captured output available\n")
	}
}

func waitForServer(ctx context.Context, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-testEnv.done:
			testEnv.done <- err
			return fmt.Errorf("process exited before becoming ready: %v", err)
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("server at %s did not become ready within %v", url, timeout)
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

func GetMailpitBaseURL() string {
	if testEnv == nil {
		return "http://localhost:8025"
	}
	return testEnv.MailpitBaseURL
}

// generateConfig builds a probod config for the e2e suite via the
// bootstrap package (which auto-generates SAML credentials) and
// writes it to a temp file. A fresh OAuth2 signing key is minted
// here and injected via env. Returns the path.
func generateConfig() (string, error) {
	oauth2SigningKey, err := bootstrap.GenerateOAuth2SigningKey()
	if err != nil {
		return "", fmt.Errorf("generate oauth2 signing key: %w", err)
	}

	env := map[string]string{
		// Required.
		"PROBOD_ENCRYPTION_KEY":     "thisisnotasecretAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
		"AUTH_COOKIE_SECRET":        "this-is-a-secure-secret-for-cookie-signing-at-least-32-bytes",
		"AUTH_PASSWORD_PEPPER":      "this-is-a-secure-pepper-for-password-hashing-at-least-32-bytes",
		"OAUTH2_SERVER_SIGNING_KEY": oauth2SigningKey,

		// Unit.
		"METRICS_ADDR": "localhost:19081",
		"TRACING_ADDR": "localhost:14317",

		// Probod base.
		"PROBOD_BASE_URL": "http://localhost:18080",

		// API.
		"API_ADDR":                 "localhost:18080",
		"API_CORS_ALLOWED_ORIGINS": "http://localhost:18080",

		// PG.
		"PG_DATABASE":      "probod_test",
		"PG_POOL_SIZE":     "10",
		"PG_MIN_POOL_SIZE": "1",

		// Auth.
		"AUTH_COOKIE_SECURE":       "false",
		"AUTH_PASSWORD_ITERATIONS": "600000",

		// OAuth2 server durations kept small for faster e2e flows.
		"OAUTH2_SERVER_ACCESS_TOKEN_DURATION":       "10",
		"OAUTH2_SERVER_REFRESH_TOKEN_DURATION":      "10",
		"OAUTH2_SERVER_AUTHORIZATION_CODE_DURATION": "5",
		"OAUTH2_SERVER_DEVICE_CODE_DURATION":        "15",

		// Trust center.
		"TRUST_CENTER_HTTP_ADDR":  ":10080",
		"TRUST_CENTER_HTTPS_ADDR": ":10443",

		// AWS / S3 (SeaweedFS).
		"AWS_BUCKET":            "probod-test",
		"AWS_ACCESS_KEY_ID":     "probod",
		"AWS_SECRET_ACCESS_KEY": "thisisnotasecret",
		"AWS_ENDPOINT":          "http://127.0.0.1:8333",

		// Mailer.
		"MAILER_SENDER_NAME":  "Probo Test",
		"MAILER_SENDER_EMAIL": "no-reply@test.getprobo.com",
		"MAILER_INTERVAL":     "1",

		// LLM.
		"OPENAI_API_KEY": "thisisnotasecret",

		// Custom domains.
		"CUSTOM_DOMAINS_CNAME_TARGET": "custom.test.getprobo.com",
		"ACME_DIRECTORY":              "https://localhost:14000/dir",
		"ACME_EMAIL":                  "admin@test.getprobo.com",
	}

	builder := bootstrap.NewBuilder(func(key string) string {
		if v, ok := env[key]; ok {
			return v
		}
		return os.Getenv(key)
	})

	cfg, err := builder.Build()
	if err != nil {
		return "", fmt.Errorf("build config: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "probo-e2e-")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	path := filepath.Join(tmpDir, "probod.yml")

	if err := bootstrap.WriteConfig(cfg, path); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}

	return path, nil
}
