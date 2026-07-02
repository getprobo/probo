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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_ResolveLiteralEnv(t *testing.T) {
	t.Parallel()

	r := NewResolver(mockEnv(map[string]string{"PROBOD_FOO": "bar"}))
	assert.Equal(t, "bar", r.getEnv("PROBOD_FOO"))
	require.NoError(t, r.Err())
}

func TestResolver_ResolveEmptyAWSSecretsManagerRef(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"awssm://", "aws://"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			r := NewResolver(mockEnv(map[string]string{
				"PROBOD_ENCRYPTION_KEY": value,
			}))

			assert.Empty(t, r.getEnv("PROBOD_ENCRYPTION_KEY"))
			require.Error(t, r.Err())
			assert.Contains(t, r.Err().Error(), "cannot resolve PROBOD_ENCRYPTION_KEY")
			assert.Contains(t, r.Err().Error(), "empty AWS reference")
		})
	}
}

func TestResolver_ResolveEmptyAWSParameterStoreRef(t *testing.T) {
	t.Parallel()

	r := NewResolver(mockEnv(map[string]string{
		"PROBOD_ENCRYPTION_KEY": "awsps://",
	}))

	assert.Empty(t, r.getEnv("PROBOD_ENCRYPTION_KEY"))
	require.Error(t, r.Err())
	assert.Contains(t, r.Err().Error(), "cannot resolve PROBOD_ENCRYPTION_KEY")
	assert.Contains(t, r.Err().Error(), "empty AWS reference")
}

func TestParseAWSSecretsManagerRef(t *testing.T) {
	t.Parallel()

	secretID, ok := parseAWSSecretsManagerRef("awssm://probo/probod/encryption_key")
	require.True(t, ok)
	assert.Equal(t, "probo/probod/encryption_key", secretID)

	_, ok = parseAWSSecretsManagerRef("literal-value")
	assert.False(t, ok)

	_, ok = parseAWSSecretsManagerRef("awssm://")
	assert.False(t, ok)

	legacySecretID, ok := parseAWSSecretsManagerRef("aws://probo/probod/encryption_key")
	require.True(t, ok)
	assert.Equal(t, "probo/probod/encryption_key", legacySecretID)
}

func TestParseAWSParameterStoreRef(t *testing.T) {
	t.Parallel()

	paramName, ok := parseAWSParameterStoreRef("awsps:///probo/probod/encryption_key")
	require.True(t, ok)
	assert.Equal(t, "/probo/probod/encryption_key", paramName)

	_, ok = parseAWSParameterStoreRef("literal-value")
	assert.False(t, ok)

	_, ok = parseAWSParameterStoreRef("awsps://")
	assert.False(t, ok)
}

func TestEmptyAWSRefPrefix(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"awssm://", "aws://", "awsps://"} {
		prefix, empty := emptyAWSRefPrefix(value)
		require.True(t, empty)
		assert.Equal(t, value, prefix)
	}

	_, empty := emptyAWSRefPrefix("awssm://probo/probod/encryption_key")
	assert.False(t, empty)
}
