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

package awsconfig_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/awsconfig"
)

// credentialsResponse mirrors the JSON structure returned by the AWS container
// credentials endpoint.
type credentialsResponse struct {
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Token           string `json:"Token"`
}

// newCredentialsServer starts an httptest server that returns the given
// credentials and records the Authorization header sent by the client.
func newCredentialsServer(t *testing.T, creds credentialsResponse) (serverURL string, authHeader *string) {
	t.Helper()

	captured := new(string)

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*captured = r.Header.Get("Authorization")
			w.Header().Set("Content-Type", "application/json")

			if err := json.NewEncoder(w).Encode(creds); err != nil {
				t.Errorf("failed to encode credentials response: %v", err)
			}
		}),
	)

	t.Cleanup(srv.Close)

	return srv.URL, captured
}

func TestNewConfig_StaticCredentials(t *testing.T) {
	t.Parallel()

	logger := log.NewLogger(log.WithOutput(io.Discard))

	cfg := awsconfig.NewConfig(
		logger,
		nil,
		awsconfig.Options{
			AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		},
	)

	creds, err := cfg.Credentials.Retrieve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", creds.AccessKeyID)
	assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", creds.SecretAccessKey)
}

// TestNewConfig_EKSPodIdentity tests all EKS Pod Identity credential paths.
// Tests in this group are sequential because they mutate process environment
// variables via t.Setenv, which is incompatible with t.Parallel.
func TestNewConfig_EKSPodIdentity(t *testing.T) {
	logger := log.NewLogger(log.WithOutput(io.Discard))

	t.Run(
		"full uri without auth token",
		func(t *testing.T) {
			want := credentialsResponse{
				AccessKeyID:     "ASIAQEKSEXAMPLE1111",
				SecretAccessKey: "eks+secret+key+example",
				Token:           "eks-session-token",
			}

			serverURL, authHeader := newCredentialsServer(t, want)

			t.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", serverURL)
			t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN", "")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE", "")

			cfg := awsconfig.NewConfig(logger, &http.Client{}, awsconfig.Options{})

			creds, err := cfg.Credentials.Retrieve(context.Background())
			require.NoError(t, err)
			assert.Equal(t, want.AccessKeyID, creds.AccessKeyID)
			assert.Equal(t, want.SecretAccessKey, creds.SecretAccessKey)
			assert.Equal(t, want.Token, creds.SessionToken)
			assert.Empty(t, *authHeader)
		},
	)

	t.Run(
		"full uri with static auth token",
		func(t *testing.T) {
			want := credentialsResponse{
				AccessKeyID:     "ASIAQEKSEXAMPLE2222",
				SecretAccessKey: "eks+secret+key+example",
				Token:           "eks-session-token",
			}

			serverURL, authHeader := newCredentialsServer(t, want)

			t.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", serverURL)
			t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN", "my-static-auth-token")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE", "")

			cfg := awsconfig.NewConfig(logger, &http.Client{}, awsconfig.Options{})

			creds, err := cfg.Credentials.Retrieve(context.Background())
			require.NoError(t, err)
			assert.Equal(t, want.AccessKeyID, creds.AccessKeyID)
			assert.Equal(t, "my-static-auth-token", *authHeader)
		},
	)

	t.Run(
		"full uri with file-based auth token",
		func(t *testing.T) {
			want := credentialsResponse{
				AccessKeyID:     "ASIAQEKSEXAMPLE3333",
				SecretAccessKey: "eks+secret+key+example",
				Token:           "eks-session-token",
			}

			serverURL, authHeader := newCredentialsServer(t, want)

			tokenFile := filepath.Join(t.TempDir(), "token")
			require.NoError(t, os.WriteFile(tokenFile, []byte("file-auth-token\n"), 0o600))

			t.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", serverURL)
			t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN", "")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE", tokenFile)

			cfg := awsconfig.NewConfig(logger, &http.Client{}, awsconfig.Options{})

			creds, err := cfg.Credentials.Retrieve(context.Background())
			require.NoError(t, err)
			assert.Equal(t, want.AccessKeyID, creds.AccessKeyID)
			// Token file content is trimmed before use.
			assert.Equal(t, "file-auth-token", *authHeader)
		},
	)

	t.Run(
		"file-based auth token takes precedence over static token",
		func(t *testing.T) {
			want := credentialsResponse{
				AccessKeyID:     "ASIAQEKSEXAMPLE4444",
				SecretAccessKey: "eks+secret+key+example",
				Token:           "eks-session-token",
			}

			serverURL, authHeader := newCredentialsServer(t, want)

			tokenFile := filepath.Join(t.TempDir(), "token")
			require.NoError(t, os.WriteFile(tokenFile, []byte("file-token"), 0o600))

			t.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", serverURL)
			t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN", "static-token")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE", tokenFile)

			cfg := awsconfig.NewConfig(logger, &http.Client{}, awsconfig.Options{})

			creds, err := cfg.Credentials.Retrieve(context.Background())
			require.NoError(t, err)
			assert.Equal(t, want.AccessKeyID, creds.AccessKeyID)
			// File-based token takes precedence over static token.
			assert.Equal(t, "file-token", *authHeader)
		},
	)

	t.Run(
		"full uri takes precedence over relative uri",
		func(t *testing.T) {
			wantFromFullURI := credentialsResponse{
				AccessKeyID:     "ASIAQEKSEXAMPLE5555",
				SecretAccessKey: "eks+secret+key+full",
				Token:           "eks-session-token",
			}

			// Server representing the EKS full-URI endpoint.
			fullURIServer, _ := newCredentialsServer(t, wantFromFullURI)

			// A second server that returns different credentials — reachable only
			// if the relative URI were (incorrectly) preferred.
			wrongCreds := credentialsResponse{
				AccessKeyID:     "WRONG",
				SecretAccessKey: "WRONG",
			}
			relativeServer, _ := newCredentialsServer(t, wrongCreds)

			t.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", fullURIServer)
			t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", relativeServer)
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN", "")
			t.Setenv("AWS_CONTAINER_AUTHORIZATION_TOKEN_FILE", "")

			cfg := awsconfig.NewConfig(logger, &http.Client{}, awsconfig.Options{})

			creds, err := cfg.Credentials.Retrieve(context.Background())
			require.NoError(t, err)
			assert.Equal(t, wantFromFullURI.AccessKeyID, creds.AccessKeyID)
		},
	)
}

// TestNewConfig_NoContainerEnvVars verifies that NewConfig does not panic when
// no container credential environment variables are set.  Credential retrieval
// will fail (no EC2 IMDS in the test environment), which is expected.
func TestNewConfig_NoContainerEnvVars(t *testing.T) {
	t.Setenv("AWS_CONTAINER_CREDENTIALS_FULL_URI", "")
	t.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "")

	logger := log.NewLogger(log.WithOutput(io.Discard))

	cfg := awsconfig.NewConfig(logger, &http.Client{}, awsconfig.Options{})

	require.NotNil(t, cfg.Credentials)
}
