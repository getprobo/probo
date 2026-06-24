// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package bootstrap

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSecretsManagerClient struct {
	secrets map[string]string
	called  []string
}

func (m *mockSecretsManagerClient) GetSecretValue(
	_ context.Context,
	params *secretsmanager.GetSecretValueInput,
	_ ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	secretID := aws.ToString(params.SecretId)
	m.called = append(m.called, secretID)

	value, ok := m.secrets[secretID]
	if !ok {
		return nil, fmt.Errorf("secret %q not found", secretID)
	}

	return &secretsmanager.GetSecretValueOutput{
		SecretString: aws.String(value),
	}, nil
}

func testResolver(env map[string]string, client SecretsManagerClient) *Resolver {
	r := NewResolver(mockEnv(env))
	r.secretsManagerClient = client

	return r
}

func TestResolver_ResolveLiteralEnv(t *testing.T) {
	t.Parallel()

	r := NewResolver(mockEnv(map[string]string{"PROBOD_FOO": "bar"}))
	assert.Equal(t, "bar", r.getEnv("PROBOD_FOO"))
	require.NoError(t, r.Err())
}

func TestBuilder_Build_ResolvesAWSSecretRefs(t *testing.T) {
	t.Parallel()

	client := &mockSecretsManagerClient{
		secrets: map[string]string{
			"probo/sandbox/probod/encryption_key":     "test-encryption-key-32-bytes-long",
			"probo/sandbox/probod/cookie_secret":      "test-cookie-secret-32-bytes-long!",
			"probo/sandbox/probod/password_pepper":    "test-password-pepper-32-bytes-lo",
			"probo/sandbox/probod/oauth2_signing_key": "test-oauth2-signing-key",
		},
	}

	env := map[string]string{
		"PROBOD_ENCRYPTION_KEY":            "aws://probo/sandbox/probod/encryption_key",
		"PROBOD_AUTH_COOKIE_SECRET":        "aws://probo/sandbox/probod/cookie_secret",
		"PROBOD_AUTH_PASSWORD_PEPPER":      "aws://probo/sandbox/probod/password_pepper",
		"PROBOD_OAUTH2_SERVER_SIGNING_KEY": "aws://probo/sandbox/probod/oauth2_signing_key",
		"PROBOD_BASE_URL":                  "https://app.example.com",
	}
	b := NewBuilder(testResolver(env, client))

	cfg, err := b.Build()
	require.NoError(t, err)
	require.Len(t, client.called, 4)

	assert.Equal(t, "test-encryption-key-32-bytes-long", cfg.Probod.EncryptionKey)
	assert.Equal(t, "https://app.example.com", cfg.Probod.BaseURL)
}

func TestResolver_ResolveAWSSecretRefsCachesBySecretID(t *testing.T) {
	t.Parallel()

	client := &mockSecretsManagerClient{
		secrets: map[string]string{
			"probo/sandbox/probod/encryption_key": "secret-value",
		},
	}
	r := testResolver(map[string]string{
		"PROBOD_ENCRYPTION_KEY":            "aws://probo/sandbox/probod/encryption_key",
		"PROBOD_AUTH_COOKIE_SECRET":        "aws://probo/sandbox/probod/encryption_key",
		"PROBOD_AUTH_PASSWORD_PEPPER":      "plaintext-pepper",
		"PROBOD_OAUTH2_SERVER_SIGNING_KEY": "plaintext-oauth2",
	}, client)

	assert.Equal(t, "secret-value", r.getEnv("PROBOD_ENCRYPTION_KEY"))
	assert.Equal(t, "secret-value", r.getEnv("PROBOD_AUTH_COOKIE_SECRET"))
	require.NoError(t, r.Err())
	require.Len(t, client.called, 1)
}

func TestBuilder_Build_MixedPlaintextAndAWSSecrets(t *testing.T) {
	t.Parallel()

	client := &mockSecretsManagerClient{
		secrets: map[string]string{
			"probo/sandbox/probod/encryption_key": "test-encryption-key-32-bytes-long",
		},
	}
	b := NewBuilder(testResolver(map[string]string{
		"PROBOD_ENCRYPTION_KEY":            "aws://probo/sandbox/probod/encryption_key",
		"PROBOD_AUTH_COOKIE_SECRET":        "test-cookie-secret-32-bytes-long!",
		"PROBOD_AUTH_PASSWORD_PEPPER":      "test-password-pepper-32-bytes-lo",
		"PROBOD_OAUTH2_SERVER_SIGNING_KEY": "test-oauth2-signing-key",
	}, client))

	cfg, err := b.Build()
	require.NoError(t, err)
	assert.Equal(t, "test-encryption-key-32-bytes-long", cfg.Probod.EncryptionKey)
	assert.Equal(t, "test-cookie-secret-32-bytes-long!", cfg.Probod.Auth.Cookie.Secret)
}

func TestResolver_ResolveAWSSecretRefMissingSecret(t *testing.T) {
	t.Parallel()

	client := &mockSecretsManagerClient{secrets: map[string]string{}}
	r := testResolver(map[string]string{
		"PROBOD_ENCRYPTION_KEY": "aws://probo/sandbox/probod/encryption_key",
	}, client)

	assert.Empty(t, r.getEnv("PROBOD_ENCRYPTION_KEY"))
	require.Error(t, r.Err())
	assert.Contains(t, r.Err().Error(), "cannot resolve PROBOD_ENCRYPTION_KEY")
}

func TestResolver_ResolveAWSSecretRefEmptySecretString(t *testing.T) {
	t.Parallel()

	client := &mockSecretsManagerClient{
		secrets: map[string]string{
			"probo/sandbox/probod/encryption_key": "",
		},
	}
	r := testResolver(map[string]string{
		"PROBOD_ENCRYPTION_KEY": "aws://probo/sandbox/probod/encryption_key",
	}, client)

	assert.Empty(t, r.getEnv("PROBOD_ENCRYPTION_KEY"))
	require.Error(t, r.Err())
	assert.Contains(t, r.Err().Error(), "empty SecretString")
}

func TestParseAWSSecretRef(t *testing.T) {
	t.Parallel()

	secretID, ok := parseAWSSecretRef("aws://probo/sandbox/probod/encryption_key")
	require.True(t, ok)
	assert.Equal(t, "probo/sandbox/probod/encryption_key", secretID)

	_, ok = parseAWSSecretRef("literal-value")
	assert.False(t, ok)

	_, ok = parseAWSSecretRef("aws://")
	assert.False(t, ok)
}
