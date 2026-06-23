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

package coredata_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestOAuth2Scope_IsRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		scope coredata.OAuth2Scope
		want  bool
	}{
		{scope: "v1:org:read", want: true},
		{scope: "v1:document:read", want: true},
		{scope: "v1:org", want: false},
		{scope: "v1:privacy", want: false},
		{scope: "openid", want: false},
		{scope: "offline_access", want: false},
	}

	for _, tt := range tests {
		t.Run(
			tt.scope.String(),
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.want, tt.scope.IsRead())
			},
		)
	}
}
