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
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

func TestDocumentContentResolver_Prosemirror(t *testing.T) {
	t.Parallel()

	r := &documentContentResolver{
		Resolver: &Resolver{logger: log.NewLogger(log.WithOutput(io.Discard))},
	}

	source := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello"}]}]}`
	got, err := r.Prosemirror(context.Background(), &types.DocumentContent{Source: source})
	require.NoError(t, err)
	assert.Equal(t, source, got)
}

func TestDocumentContentResolver_Markdown(t *testing.T) {
	t.Parallel()

	r := &documentContentResolver{
		Resolver: &Resolver{logger: log.NewLogger(log.WithOutput(io.Discard))},
	}

	t.Run(
		"renders paragraph",
		func(t *testing.T) {
			t.Parallel()

			source := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Hello world"}]}]}`
			got, err := r.Markdown(context.Background(), &types.DocumentContent{Source: source})
			require.NoError(t, err)
			assert.Contains(t, got, "Hello world")
		},
	)

	t.Run(
		"empty source",
		func(t *testing.T) {
			t.Parallel()

			got, err := r.Markdown(context.Background(), &types.DocumentContent{Source: ""})
			require.NoError(t, err)
			assert.Equal(t, "", got)
		},
	)

	t.Run(
		"invalid json is internal error",
		func(t *testing.T) {
			t.Parallel()

			_, err := r.Markdown(context.Background(), &types.DocumentContent{Source: "not-json"})
			require.Error(t, err)

			var gqlErr *gqlerror.Error
			require.ErrorAs(t, err, &gqlErr)
			assert.Equal(t, "INTERNAL", gqlErr.Extensions["code"])
		},
	)
}
