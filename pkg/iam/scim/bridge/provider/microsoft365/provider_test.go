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

package microsoft365

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_ListUsers_DoesNotRequestPhotos(t *testing.T) {
	t.Parallel()

	var (
		mu    sync.Mutex
		paths []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths = append(paths, r.URL.Path)
		mu.Unlock()

		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/users") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{
				"value": [
					{"id":"user-1","mail":"ada@example.com","displayName":"Ada","accountEnabled":true}
				]
			}`)

			return
		}

		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	provider := New(server.Client(), nil)
	provider.graphBaseURL = server.URL

	users, err := provider.ListUsers(context.Background())

	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "ada@example.com", users[0].UserName)
	assert.Empty(t, users[0].PhotoETag)
	assert.False(t, users[0].PhotoETagKnown)
	mu.Lock()
	recorded := append([]string(nil), paths...)
	mu.Unlock()

	require.Len(t, recorded, 1)
	assert.True(t, strings.HasSuffix(recorded[0], "/users"))
}
