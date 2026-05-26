// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package update

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
)

func TestParseTag(t *testing.T) {
	t.Parallel()

	cases := []struct {
		tag    string
		prefix string
		want   string
		ok     bool
	}{
		{"probo-agent/v0.1.0", "probo-agent/v", "0.1.0", true},
		{"probo-agent/v1.2.3", "probo-agent/v", "1.2.3", true},
		{"v1.2.3", "probo-agent/v", "", false},
		{"probo-agent/vlatest", "probo-agent/v", "", false},
		{"probo-agent/v", "probo-agent/v", "", false},
		{"unrelated/v0.1.0", "probo-agent/v", "", false},
	}

	for _, tc := range cases {
		got, ok := parseTag(tc.tag, tc.prefix)
		assert.Equal(t, tc.ok, ok, tc.tag)
		assert.Equal(t, tc.want, got, tc.tag)
	}
}

func TestNormalizeSemver(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "v0.1.0", normalizeSemver("0.1.0"))
	assert.Equal(t, "v1.2.3", normalizeSemver("v1.2.3"))
	assert.Equal(t, "v1.2.3-alpha.1", normalizeSemver("1.2.3-alpha.1"))
	assert.Equal(t, "", normalizeSemver(""))
	assert.Equal(t, "", normalizeSemver("not-a-version"))
}

func TestReadChecksum(t *testing.T) {
	t.Parallel()

	t.Run(
		"plain sha256sum output",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			file := filepath.Join(dir, "checksums.txt")
			content := "" +
				"deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef  probo-agent_Linux_x86_64.tar.gz\n" +
				"abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd  probo-agent_Darwin_arm64.tar.gz\n"
			require.NoError(t, os.WriteFile(file, []byte(content), 0o600))

			got, err := readChecksum(file, "probo-agent_Darwin_arm64.tar.gz")
			require.NoError(t, err)
			assert.Equal(t, "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abcd", got)
		},
	)

	t.Run(
		"binary-mode flag is stripped",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			file := filepath.Join(dir, "checksums.txt")
			content := "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef *probo-agent_Linux_x86_64.tar.gz\n"
			require.NoError(t, os.WriteFile(file, []byte(content), 0o600))

			got, err := readChecksum(file, "probo-agent_Linux_x86_64.tar.gz")
			require.NoError(t, err)
			assert.Equal(t, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", got)
		},
	)

	t.Run(
		"missing entry returns error",
		func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			file := filepath.Join(dir, "checksums.txt")
			require.NoError(t, os.WriteFile(file, []byte("deadbeef  other.tar.gz\n"), 0o600))

			_, err := readChecksum(file, "probo-agent_Linux_x86_64.tar.gz")
			require.Error(t, err)
		},
	)
}

// fakeReleaseServer simulates the GitHub releases API and the
// browser_download_url asset endpoints.
type fakeReleaseServer struct {
	t      *testing.T
	server *httptest.Server

	// release plumbing
	tag        string
	prerelease bool
	draft      bool

	// archive plumbing
	binaryContent []byte
	archiveBytes  []byte
	checksumLine  string
	bundleBytes   []byte

	// when true, the release does not advertise a checksums.txt.bundle asset
	omitBundle bool
}

func newFakeReleaseServer(t *testing.T, tag, version string, layout AssetLayout, binary []byte) *fakeReleaseServer {
	t.Helper()

	archive := buildArchive(t, layout, binary)
	sum := sha256.Sum256(archive)
	checksum := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), layout.ArchiveName)

	frs := &fakeReleaseServer{
		t:             t,
		tag:           tag,
		binaryContent: binary,
		archiveBytes:  archive,
		checksumLine:  checksum,
		bundleBytes:   []byte("dummy-sigstore-bundle"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/repos/getprobo/probo/releases", func(w http.ResponseWriter, r *http.Request) {
		base := "http://" + r.Host

		assets := []map[string]any{
			{
				"name":                 layout.ArchiveName,
				"browser_download_url": base + "/download/" + layout.ArchiveName,
			},
			{
				"name":                 checksumFileName,
				"browser_download_url": base + "/download/" + checksumFileName,
			},
		}
		if !frs.omitBundle {
			assets = append(assets, map[string]any{
				"name":                 checksumBundleFileName,
				"browser_download_url": base + "/download/" + checksumBundleFileName,
			})
		}

		body := []map[string]any{
			{
				"tag_name":   frs.tag,
				"draft":      frs.draft,
				"prerelease": frs.prerelease,
				"assets":     assets,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
		_ = version
	})
	mux.HandleFunc("/download/"+layout.ArchiveName, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(frs.archiveBytes)
	})
	mux.HandleFunc("/download/"+checksumFileName, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(frs.checksumLine))
	})
	mux.HandleFunc("/download/"+checksumBundleFileName, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(frs.bundleBytes)
	})

	frs.server = httptest.NewServer(mux)
	t.Cleanup(frs.server.Close)

	return frs
}

func (f *fakeReleaseServer) URL() string { return f.server.URL }

func buildArchive(t *testing.T, layout AssetLayout, binary []byte) []byte {
	t.Helper()

	if layout.IsZip {
		return buildZip(t, layout, binary)
	}

	return buildTarGz(t, layout, binary)
}

func buildTarGz(t *testing.T, layout AssetLayout, binary []byte) []byte {
	t.Helper()

	dir := t.TempDir()
	out := filepath.Join(dir, layout.ArchiveName)

	f, err := os.Create(out)
	require.NoError(t, err)

	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	require.NoError(t, tw.WriteHeader(&tar.Header{
		Name:     path.Join(layout.ArchiveDir, layout.BinaryName),
		Mode:     0o755,
		Size:     int64(len(binary)),
		Typeflag: tar.TypeReg,
	}))
	_, err = tw.Write(binary)
	require.NoError(t, err)

	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	require.NoError(t, f.Close())

	data, err := os.ReadFile(out)
	require.NoError(t, err)

	return data
}

func buildZip(t *testing.T, layout AssetLayout, binary []byte) []byte {
	t.Helper()

	dir := t.TempDir()
	out := filepath.Join(dir, layout.ArchiveName)

	f, err := os.Create(out)
	require.NoError(t, err)

	zw := zip.NewWriter(f)
	w, err := zw.Create(path.Join(layout.ArchiveDir, layout.BinaryName))
	require.NoError(t, err)
	_, err = w.Write(binary)
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	require.NoError(t, f.Close())

	data, err := os.ReadFile(out)
	require.NoError(t, err)

	return data
}

func newTestUpdater(server *fakeReleaseServer, currentVersion, exePath, goos, goarch string) *Updater {
	return &Updater{
		Repo:           "getprobo/probo",
		TagPrefix:      DefaultTagPrefix,
		APIBaseURL:     server.URL(),
		AssetBaseURL:   server.URL(),
		CurrentVersion: currentVersion,
		ExePath:        exePath,
		UserAgent:      "probo-agent-test/0.0.0",
		Logger:         log.NewLogger(log.WithName("update-test")),
		HTTP: &http.Client{
			Transport: httpclient.DefaultPooledTransport(
				httpclient.WithSSRFProtection(),
				httpclient.WithSSRFAllowLoopback(),
			),
		},
		// Tests bypass the cosign verifier; production code wires
		// CosignVerifier in via Updater.SigstoreCacheDir.
		Verifier: AllowAllVerifier{},
		GOOS:     goos,
		GOARCH:   goarch,
	}
}

func TestUpdater_CheckLatest(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns release when newer version is available",
		func(t *testing.T) {
			t.Parallel()

			layout, err := LayoutFor("linux", "amd64")
			require.NoError(t, err)
			fake := newFakeReleaseServer(t, "probo-agent/v0.2.0", "0.2.0", layout, []byte("new"))

			u := newTestUpdater(fake, "0.1.0", filepath.Join(t.TempDir(), "probo-agent"), "linux", "amd64")
			rel, err := u.CheckLatest(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "0.2.0", rel.Version)
			assert.Equal(t, layout.ArchiveName, rel.AssetName)
		},
	)

	t.Run(
		"returns ErrNoUpdateAvailable when running latest",
		func(t *testing.T) {
			t.Parallel()

			layout, err := LayoutFor("darwin", "arm64")
			require.NoError(t, err)
			fake := newFakeReleaseServer(t, "probo-agent/v0.1.0", "0.1.0", layout, []byte("same"))

			u := newTestUpdater(fake, "0.1.0", filepath.Join(t.TempDir(), "probo-agent"), "darwin", "arm64")
			_, err = u.CheckLatest(context.Background())
			assert.ErrorIs(t, err, ErrNoUpdateAvailable)
		},
	)

	t.Run(
		"skips draft and prerelease tags",
		func(t *testing.T) {
			t.Parallel()

			layout, err := LayoutFor("linux", "amd64")
			require.NoError(t, err)
			fake := newFakeReleaseServer(t, "probo-agent/v0.2.0-rc.1", "0.2.0-rc.1", layout, []byte("rc"))
			fake.prerelease = true

			u := newTestUpdater(fake, "0.1.0", filepath.Join(t.TempDir(), "probo-agent"), "linux", "amd64")
			_, err = u.CheckLatest(context.Background())
			assert.ErrorIs(t, err, ErrNoUpdateAvailable)
		},
	)

	t.Run(
		"dev build always sees update available",
		func(t *testing.T) {
			t.Parallel()

			layout, err := LayoutFor("linux", "amd64")
			require.NoError(t, err)
			fake := newFakeReleaseServer(t, "probo-agent/v0.1.0", "0.1.0", layout, []byte("rel"))

			u := newTestUpdater(fake, "dev", filepath.Join(t.TempDir(), "probo-agent"), "linux", "amd64")
			rel, err := u.CheckLatest(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "0.1.0", rel.Version)
		},
	)
}

func TestUpdater_Apply(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("apply test exercises the unix swap path; windows has its own .old shuffle")
	}

	dir := t.TempDir()
	exePath := filepath.Join(dir, "probo-agent")
	require.NoError(t, os.WriteFile(exePath, []byte("old-binary"), 0o755))

	layout, err := LayoutFor("linux", "amd64")
	require.NoError(t, err)
	fake := newFakeReleaseServer(t, "probo-agent/v0.2.0", "0.2.0", layout, []byte("new-binary"))

	u := newTestUpdater(fake, "0.1.0", exePath, "linux", "amd64")
	rel, err := u.CheckLatest(context.Background())
	require.NoError(t, err)

	require.NoError(t, u.Apply(context.Background(), rel))

	got, err := os.ReadFile(exePath)
	require.NoError(t, err)
	assert.Equal(t, []byte("new-binary"), got)

	stat, err := os.Stat(exePath)
	require.NoError(t, err)
	assert.NotZero(t, stat.Mode().Perm()&0o100, "new binary should be executable")
}

func TestUpdater_CheckLatest_SkipsUnsignedRelease(t *testing.T) {
	t.Parallel()

	layout, err := LayoutFor("linux", "amd64")
	require.NoError(t, err)
	fake := newFakeReleaseServer(t, "probo-agent/v0.2.0", "0.2.0", layout, []byte("new"))
	fake.omitBundle = true

	u := newTestUpdater(fake, "0.1.0", filepath.Join(t.TempDir(), "probo-agent"), "linux", "amd64")
	_, err = u.CheckLatest(context.Background())
	assert.ErrorIs(t, err, ErrNoUpdateAvailable, "release without a sigstore bundle must be ignored")
}

func TestUpdater_Apply_RejectsBadSignature(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	exePath := filepath.Join(dir, "probo-agent")
	require.NoError(t, os.WriteFile(exePath, []byte("old-binary"), 0o755))

	layout, err := LayoutFor("linux", "amd64")
	require.NoError(t, err)
	fake := newFakeReleaseServer(t, "probo-agent/v0.2.0", "0.2.0", layout, []byte("new-binary"))

	u := newTestUpdater(fake, "0.1.0", exePath, "linux", "amd64")
	u.Verifier = rejectAllVerifier{err: fmt.Errorf("test: signer identity mismatch")}

	rel, err := u.CheckLatest(context.Background())
	require.NoError(t, err)

	err = u.Apply(context.Background(), rel)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sigstore")

	got, err := os.ReadFile(exePath)
	require.NoError(t, err)
	assert.Equal(t, []byte("old-binary"), got, "rejected signature must not touch the running binary")
}

func TestUpdater_Apply_RejectsCorruptedArchive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	exePath := filepath.Join(dir, "probo-agent")
	require.NoError(t, os.WriteFile(exePath, []byte("old-binary"), 0o755))

	layout, err := LayoutFor("linux", "amd64")
	require.NoError(t, err)
	fake := newFakeReleaseServer(t, "probo-agent/v0.2.0", "0.2.0", layout, []byte("new-binary"))

	// Corrupt the archive without updating checksums.
	fake.archiveBytes = append(fake.archiveBytes, 0xff)

	u := newTestUpdater(fake, "0.1.0", exePath, "linux", "amd64")
	rel, err := u.CheckLatest(context.Background())
	require.NoError(t, err)

	err = u.Apply(context.Background(), rel)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")

	got, err := os.ReadFile(exePath)
	require.NoError(t, err)
	assert.Equal(t, []byte("old-binary"), got, "corrupted update must not touch the running binary")
}
