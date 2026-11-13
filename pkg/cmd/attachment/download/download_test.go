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

package download

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cli/config"
	"go.probo.inc/probo/pkg/cmd/cmdutil"
	"go.probo.inc/probo/pkg/cmd/iostreams"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestSaveDownload_WritesIntoDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	path, size, err := saveDownload(dir, "evidence.pdf", bytes.NewBufferString("pdf-bytes"), nil)
	require.NoError(t, err)
	assert.Equal(t, int64(len("pdf-bytes")), size)
	assert.Equal(t, filepath.Join(dir, "evidence.pdf"), path)

	payload, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "pdf-bytes", string(payload))
}

func TestSaveDownload_RefusesToReplaceExistingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dest := filepath.Join(dir, "evidence.pdf")
	require.NoError(t, os.WriteFile(dest, []byte("keep"), 0o644))

	_, _, err := saveDownload(dest, "other.pdf", bytes.NewBufferString("new"), nil)
	require.Error(t, err)

	payload, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "keep", string(payload))
}

func TestSaveDownload_WritesStdout(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer

	path, size, err := saveDownload("-", "evidence.pdf", bytes.NewBufferString("pdf-bytes"), &out)
	require.NoError(t, err)
	assert.Equal(t, "-", path)
	assert.Equal(t, int64(len("pdf-bytes")), size)
	assert.Equal(t, "pdf-bytes", out.String())
}

func TestDownloadCommand_WritesFileFromDescriptionLink(t *testing.T) {
	t.Parallel()

	fileID := gid.New(gid.NewTenantID(), coredata.FileEntityType).String()

	storage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Empty(t, r.Header.Get("Authorization"))
		w.Header().Set("Content-Disposition", `attachment; filename="evidence.pdf"`)
		_, _ = w.Write([]byte("pdf-bytes"))
	}))
	t.Cleanup(storage.Close)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		http.Redirect(w, r, storage.URL+"/object", http.StatusTemporaryRedirect)
	}))
	t.Cleanup(apiServer.Close)

	streams, out, _ := iostreams.Test()
	factory := &cmdutil.Factory{
		IOStreams: streams,
		Config: func() (*config.Config, error) {
			return &config.Config{
				ActiveHost: apiServer.URL,
				Hosts: map[string]*config.HostConfig{
					apiServer.URL: {Token: "token"},
				},
			}, nil
		},
	}

	dest := filepath.Join(t.TempDir(), "saved.pdf")
	cmd := NewCmdDownload(factory)
	cmd.SetArgs([]string{fileID, dest})

	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Downloaded "+dest)

	payload, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "pdf-bytes", string(payload))
}
