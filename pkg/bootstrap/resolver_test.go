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
