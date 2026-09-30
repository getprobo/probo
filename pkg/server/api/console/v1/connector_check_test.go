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

package console_v1

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.probo.inc/probo/pkg/accessreview/drivers"
)

func TestSettingRejectedError(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		setting string
		field   string
	}{
		"refused setting on its own field": {setting: "teamName", field: "teamName"},
		"unreachable target on the key":    {setting: "", field: "apiKey"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			rejected := &drivers.SettingRejectedError{
				Code:       "code",
				Setting:    tc.setting,
				StatusCode: http.StatusUnprocessableEntity,
				Message:    "Probo's message.",
			}

			gqlErr, ok := errors.AsType[*gqlerror.Error](settingRejectedError(context.Background(), rejected))
			require.True(t, ok)
			assert.Equal(t, "INVALID", gqlErr.Extensions["code"])
			assert.Equal(t, tc.field, gqlErr.Extensions["field"])
			assert.Equal(t, "code", gqlErr.Extensions["cause"])
			assert.Equal(t, "Probo's message.", gqlErr.Message)
		})
	}
}
