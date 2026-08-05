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

package connect_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactSCIMCredentials(t *testing.T) {
	t.Parallel()

	t.Run("leaves a credential-free body byte-identical", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"userName":"a@example.com","name":{"givenName":"A"}}`)

		assert.Equal(t, body, redactSCIMCredentials(body))
	})

	t.Run("redacts a user password attribute", func(t *testing.T) {
		t.Parallel()

		got := redactSCIMCredentials([]byte(`{"userName":"a@example.com","password":"s3cret"}`))

		assert.JSONEq(t, `{"userName":"a@example.com","password":"[REDACTED]"}`, string(got))
	})

	t.Run("redacts a password regardless of attribute casing", func(t *testing.T) {
		t.Parallel()

		got := redactSCIMCredentials([]byte(`{"Password":"s3cret"}`))

		assert.JSONEq(t, `{"Password":"[REDACTED]"}`, string(got))
	})

	t.Run("redacts a password nested in an extension", func(t *testing.T) {
		t.Parallel()

		got := redactSCIMCredentials([]byte(`{"urn:example":{"password":"s3cret"}}`))

		assert.JSONEq(t, `{"urn:example":{"password":"[REDACTED]"}}`, string(got))
	})

	t.Run("redacts a patch operation targeting the password", func(t *testing.T) {
		t.Parallel()

		got := redactSCIMCredentials([]byte(`{"Operations":[{"op":"replace","path":"password","value":"s3cret"}]}`))

		assert.JSONEq(
			t,
			`{"Operations":[{"op":"replace","path":"password","value":"[REDACTED]"}]}`,
			string(got),
		)
	})

	t.Run("keeps other patch operations intact", func(t *testing.T) {
		t.Parallel()

		got := redactSCIMCredentials([]byte(`{"Operations":[{"op":"replace","path":"password","value":"s3cret"},{"op":"replace","path":"active","value":false}]}`))

		assert.JSONEq(
			t,
			`{"Operations":[{"op":"replace","path":"password","value":"[REDACTED]"},{"op":"replace","path":"active","value":false}]}`,
			string(got),
		)
	})

	t.Run("drops a credential-bearing body that cannot be parsed", func(t *testing.T) {
		t.Parallel()

		assert.Nil(t, redactSCIMCredentials([]byte(`{"password":"s3cret"`)))
	})

	t.Run("keeps a malformed body without a credential", func(t *testing.T) {
		t.Parallel()

		body := []byte(`{"userName":`)

		assert.Equal(t, body, redactSCIMCredentials(body))
	})

	t.Run("never leaks the credential", func(t *testing.T) {
		t.Parallel()

		bodies := []string{
			`{"password":"s3cret"}`,
			`{"PASSWORD":"s3cret"}`,
			`{"Operations":[{"op":"add","path":"Password","value":"s3cret"}]}`,
		}

		for _, body := range bodies {
			got := redactSCIMCredentials([]byte(body))
			require.NotNil(t, got, body)
			assert.NotContains(t, string(got), "s3cret", body)
		}
	})
}
