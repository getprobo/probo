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

package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListProjects_QueryFiltersByTeam(t *testing.T) {
	t.Parallel()

	var (
		mu       sync.Mutex
		requests []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()

		requests = append(requests, r.URL.RequestURI())

		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(r.URL.Path, "/typeahead"):
			_, _ = w.Write([]byte(`{"data":[{"gid":"p1","name":"Work"},{"gid":"p2","name":"Other work"}]}`))
		case strings.HasSuffix(r.URL.Path, "/projects/p1"):
			_, _ = w.Write([]byte(`{"data":{"gid":"p1","team":{"gid":"t1"}}}`))
		case strings.HasSuffix(r.URL.Path, "/projects/p2"):
			_, _ = w.Write([]byte(`{"data":{"gid":"p2","team":{"gid":"t2"}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	page, err := NewClient(server.Client(), server.URL).ListProjects(
		t.Context(),
		"w1",
		"t1",
		"work",
		20,
		"ignored-cursor",
	)
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, "p1", page.Items[0].GID)
	assert.Equal(t, "Work", page.Items[0].Name)
	assert.Empty(t, page.EndCursor)
	assert.False(t, page.HasNextPage)

	mu.Lock()

	got := append([]string(nil), requests...)

	mu.Unlock()

	require.NotEmpty(t, got)
	assert.Contains(t, got[0], "/typeahead")
	assert.Contains(t, got[0], "resource_type=project")
	assert.Contains(t, got[0], "query=work")
	assert.NotContains(t, got[0], "offset=")
	assert.NotContains(t, got[0], "team=")
	assert.Len(t, got, 3)
}

func TestListProjects_QueryWithoutTeamKeepsWorkspaceMatches(t *testing.T) {
	t.Parallel()

	var projectReads atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/projects/") {
			projectReads.Add(1)

			http.NotFound(w, r)

			return
		}

		_, _ = w.Write([]byte(`{"data":[{"gid":"p1","name":"Work"},{"gid":"p2","name":"Other work"}]}`))
	}))
	t.Cleanup(server.Close)

	page, err := NewClient(server.Client(), server.URL).ListProjects(t.Context(), "w1", "", "work", 20, "")
	require.NoError(t, err)
	require.Len(t, page.Items, 2)
	assert.Equal(t, "p1", page.Items[0].GID)
	assert.Equal(t, "p2", page.Items[1].GID)
	assert.False(t, page.HasNextPage)
	assert.Equal(t, int32(0), projectReads.Load())
}

func TestListProjects_QuerySkipsMissingTeamLookup(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case strings.Contains(r.URL.Path, "/typeahead"):
			_, _ = w.Write([]byte(`{"data":[{"gid":"gone","name":"Gone"},{"gid":"p1","name":"Work"}]}`))
		case strings.HasSuffix(r.URL.Path, "/projects/gone"):
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"errors":[{"message":"missing"}]}`))
		case strings.HasSuffix(r.URL.Path, "/projects/p1"):
			_, _ = w.Write([]byte(`{"data":{"gid":"p1","team":{"gid":"t1"}}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	page, err := NewClient(server.Client(), server.URL).ListProjects(t.Context(), "w1", "t1", "work", 20, "")
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, "p1", page.Items[0].GID)
}

func TestListProjects_QueryNamesFailedTeamLookup(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "/typeahead") {
			_, _ = w.Write([]byte(`{"data":[{"gid":"p1","name":"Work"}]}`))

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"unavailable"}]}`))
	}))
	t.Cleanup(server.Close)

	_, err := NewClient(server.Client(), server.URL).ListProjects(t.Context(), "w1", "t1", "work", 20, "")
	require.Error(t, err)
	assert.ErrorContains(t, err, "p1")
}

func TestJoin_RejectsDotSegments(t *testing.T) {
	t.Parallel()

	client := &Client{baseURL: "https://app.asana.com/api/1.0"}

	_, err := client.join("projects", "..")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid path segment")

	_, err = client.join("projects", ".")
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid path segment")

	_, err = client.join("projects", "")
	require.Error(t, err)
	assert.ErrorContains(t, err, "empty path segment")

	joined, err := client.join("projects", "123")
	require.NoError(t, err)
	assert.Equal(t, "https://app.asana.com/api/1.0/projects/123", joined)
}

func TestListStories_PagesUntilNextPageIsEmpty(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Query().Get("offset") == "" {
			_, _ = w.Write([]byte(`{"data":[{"gid":"c1","resource_subtype":"comment_added","html_text":"<body>One</body>","text":"One","created_at":"2024-01-02T03:04:05Z","created_by":{"email":"a@example.com"}},{"gid":"s1","resource_subtype":"system","text":"assigned"}],"next_page":{"offset":"page-2"}}`))

			return
		}

		_, _ = w.Write([]byte(`{"data":[{"gid":"c2","resource_subtype":"comment_added","html_text":"<body>Two</body>","text":"Two","created_at":"2024-01-03T03:04:05Z"}]}`))
	}))
	t.Cleanup(server.Close)

	stories, err := NewClient(server.Client(), server.URL).ListStories(t.Context(), "task-1")
	require.NoError(t, err)
	require.Len(t, stories, 2)
	assert.Equal(t, "c1", stories[0].GID)
	assert.Equal(t, "a@example.com", stories[0].AuthorEmail)
	assert.Equal(t, "c2", stories[1].GID)
	assert.Equal(t, "<body>Two</body>", stories[1].HTMLText)
}

func TestListStories_PageLimitReturnsNoStories(t *testing.T) {
	t.Parallel()

	var pages atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := pages.Add(1)
		if page > 5 {
			w.WriteHeader(http.StatusInternalServerError)

			_, _ = w.Write([]byte(`{"errors":[{"message":"too many"}]}`))

			return
		}

		w.Header().Set("Content-Type", "application/json")

		_, _ = fmt.Fprintf(
			w,
			`{"data":[{"gid":"c%d","resource_subtype":"comment_added","text":"n"}],"next_page":{"offset":"page-%d"}}`,
			page,
			page+1,
		)
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.Client(), server.URL)
	client.storyPages = 2

	raw, err := collectStoryPages(t.Context(), client, "task-1", 2)
	require.ErrorIs(t, err, ErrStoryPageLimit)
	assert.Nil(t, raw)

	stories, err := client.ListStories(t.Context(), "task-1")
	require.ErrorIs(t, err, ErrStoryPageLimit)
	assert.Empty(t, stories)
	assert.Equal(t, int32(4), pages.Load())
}

func TestListStories_PageErrorHasOnePrefix(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"message":"unavailable"}]}`))
	}))
	t.Cleanup(server.Close)

	stories, err := NewClient(server.Client(), server.URL).ListStories(t.Context(), "task-1")
	require.Error(t, err)
	assert.Nil(t, stories)
	assert.ErrorContains(t, err, "cannot list asana stories")
	assert.ErrorContains(t, err, "cannot load asana page")
	assert.ErrorContains(t, err, "unavailable")
	assert.NotContains(t, err.Error(), "cannot list asana stories: cannot list asana stories")
	assert.NotContains(t, err.Error(), "cannot load asana page: cannot load asana page")
}

func TestCreateStory_RequestsHTMLText(t *testing.T) {
	t.Parallel()

	story := roundTripStory(
		t,
		http.MethodPost,
		"/tasks/task-1/stories",
		"<body>Hi</body>",
		func(client *Client) (Story, error) {
			return client.CreateStory(t.Context(), "task-1", "<body>Hi</body>")
		},
	)
	assert.Equal(t, "story-1", story.GID)
	assert.Equal(t, "<body>Hi</body>", story.HTMLText)
	assert.Equal(t, "a@example.com", story.AuthorEmail)
}

func TestUpdateStory_RequestsHTMLText(t *testing.T) {
	t.Parallel()

	story := roundTripStory(
		t,
		http.MethodPut,
		"/stories/story-1",
		"<body>Edited</body>",
		func(client *Client) (Story, error) {
			return client.UpdateStory(t.Context(), "story-1", "<body>Edited</body>")
		},
	)
	assert.Equal(t, "story-1", story.GID)
	assert.Equal(t, "<body>Edited</body>", story.HTMLText)
}

func roundTripStory(
	t *testing.T,
	method string,
	pathSuffix string,
	htmlText string,
	call func(*Client) (Story, error),
) Story {
	t.Helper()

	var (
		mu       sync.Mutex
		fields   string
		sentHTML string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("cannot read request body: %v", err)
		}

		var payload struct {
			Data struct {
				HTMLText string `json:"html_text"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Errorf("cannot decode request body: %v", err)
		}

		mu.Lock()

		fields = r.URL.Query().Get("opt_fields")
		sentHTML = payload.Data.HTMLText

		mu.Unlock()

		if r.Method != method || !strings.HasSuffix(r.URL.Path, pathSuffix) {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)

			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(
			w,
			`{"data":{"gid":"story-1","html_text":%q,"text":"Hi","created_at":"2024-01-02T03:04:05Z","created_by":{"email":"a@example.com"}}}`,
			htmlText,
		)
	}))
	t.Cleanup(server.Close)

	story, err := call(NewClient(server.Client(), server.URL))
	require.NoError(t, err)

	mu.Lock()

	got := fields
	gotHTML := sentHTML

	mu.Unlock()

	assert.Equal(t, "gid,html_text,text,created_at,created_by.email", got)
	assert.Equal(t, htmlText, gotHTML)

	return story
}
