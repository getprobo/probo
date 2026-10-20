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

package create

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFillSecret_TrimsProvidedValue(t *testing.T) {
	t.Parallel()

	t.Run("drops surrounding space", func(t *testing.T) {
		t.Parallel()

		value := "  secret\n"
		err := fillSecret(false, &value, envConnectorClientSecret, "Client secret")

		require.NoError(t, err)
		assert.Equal(t, "secret", value)
	})

	t.Run("blank value is empty", func(t *testing.T) {
		t.Parallel()

		value := " \n\t"
		err := fillSecret(false, &value, "PROBO_CONNECTOR_SECRET_UNSET", "Client secret")

		require.NoError(t, err)
		assert.Empty(t, value)
	})
}

func TestFillSecret_TrimsEnvValue(t *testing.T) {
	t.Setenv(envConnectorClientSecret, " secret\n")

	t.Run("empty flag", func(t *testing.T) {
		value := ""
		err := fillSecret(false, &value, envConnectorClientSecret, "Client secret")

		require.NoError(t, err)
		assert.Equal(t, "secret", value)
	})

	t.Run("whitespace flag", func(t *testing.T) {
		value := " \n"
		err := fillSecret(false, &value, envConnectorClientSecret, "Client secret")

		require.NoError(t, err)
		assert.Equal(t, "secret", value)
	})
}

func TestFillSecret_WhitespaceEnvIsEmpty(t *testing.T) {
	t.Setenv(envConnectorAPIKey, " \n\t")

	value := ""
	err := fillSecret(false, &value, envConnectorAPIKey, "API key")

	require.NoError(t, err)
	assert.Empty(t, value)
}
