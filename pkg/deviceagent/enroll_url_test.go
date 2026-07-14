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

package deviceagent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEnrollURL(t *testing.T) {
	t.Parallel()

	t.Run(
		"parses host enroll URL",
		func(t *testing.T) {
			t.Parallel()

			serverURL, enrollmentToken, err := ParseEnrollURL(
				"probo://enroll?server=https%3A%2F%2Fus.probo.com&token=secret-token",
			)
			require.NoError(t, err)
			assert.Equal(t, "https://us.probo.com", serverURL)
			assert.Equal(t, "secret-token", enrollmentToken)
		},
	)

	t.Run(
		"rejects non-enroll host",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := ParseEnrollURL(
				"probo:///enroll?server=https%3A%2F%2Feu.probo.com&token=abc123",
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "enrollment URL must be probo://enroll")
		},
	)

	t.Run(
		"rejects non probo scheme",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := ParseEnrollURL(
				"https://example.com/enroll?server=https%3A%2F%2Fus.probo.com&token=secret-token",
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "probo scheme")
		},
	)

	t.Run(
		"rejects missing token",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := ParseEnrollURL(
				"probo://enroll?server=https%3A%2F%2Fus.probo.com",
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "enrollment token is missing")
		},
	)

	t.Run(
		"rejects invalid server URL",
		func(t *testing.T) {
			t.Parallel()

			_, _, err := ParseEnrollURL(
				"probo://enroll?server=https%3A%2F%2Fus.probo.com%2Fextra&token=abc123",
			)
			require.Error(t, err)
			assert.ErrorContains(t, err, "invalid server")
		},
	)
}
