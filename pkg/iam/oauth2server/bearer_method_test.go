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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/iam/oauth2server"
)

func TestBearerMethod(t *testing.T) {
	t.Parallel()

	t.Run(
		"valid RFC values",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, oauth2server.BearerMethodHeader.IsValid())
			assert.True(t, oauth2server.BearerMethodBody.IsValid())
			assert.True(t, oauth2server.BearerMethodQuery.IsValid())
			assert.False(t, oauth2server.BearerMethod("cookie").IsValid())
		},
	)

	t.Run(
		"JSON marshals as string array",
		func(t *testing.T) {
			t.Parallel()

			methods := oauth2server.BearerMethods{oauth2server.BearerMethodHeader}
			data, err := json.Marshal(methods)
			require.NoError(t, err)

			assert.JSONEq(t, `["header"]`, string(data))
		},
	)
}
