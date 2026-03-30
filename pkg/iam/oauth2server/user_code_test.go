// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package oauth2server_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2server"
)

func TestGenerateUserCode(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns 8 character code",
		func(t *testing.T) {
			t.Parallel()

			code, err := oauth2server.GenerateUserCode()
			require.NoError(t, err)
			assert.Len(t, string(code), 8)
		},
	)

	t.Run(
		"only contains unambiguous characters",
		func(t *testing.T) {
			t.Parallel()

			allowed := "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

			for range 100 {
				code, err := oauth2server.GenerateUserCode()
				require.NoError(t, err)

				for _, c := range string(code) {
					assert.True(
						t,
						strings.ContainsRune(allowed, c),
						"unexpected character %q in code %q", string(c), code,
					)
				}
			}
		},
	)

	t.Run(
		"excludes ambiguous characters",
		func(t *testing.T) {
			t.Parallel()

			ambiguous := "01OIL"

			for range 100 {
				code, err := oauth2server.GenerateUserCode()
				require.NoError(t, err)

				for _, c := range string(code) {
					assert.False(
						t,
						strings.ContainsRune(ambiguous, c),
						"ambiguous character %q found in code %q", string(c), code,
					)
				}
			}
		},
	)

	t.Run(
		"generates unique codes",
		func(t *testing.T) {
			t.Parallel()

			seen := make(map[coredata.OAuth2UserCode]struct{})

			for range 100 {
				code, err := oauth2server.GenerateUserCode()
				require.NoError(t, err)

				_, duplicate := seen[code]
				assert.False(t, duplicate, "duplicate code generated: %q", code)
				seen[code] = struct{}{}
			}
		},
	)
}
