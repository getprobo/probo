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

// Package update self-updates the probo-agent binary from GitHub
// Releases. The update flow is:
//
//  1. List the latest releases for the configured repo, filtering on a
//     tag prefix (`probo-agent/v` by default).
//  2. Pick the highest semver newer than the agent's current version.
//  3. Download the matching archive plus checksums.txt, verify SHA-256.
//  4. Extract the archive and atomically replace the running binary.
//
// The caller is responsible for restarting the process so the OS
// service supervisor re-execs the new binary.
package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"golang.org/x/mod/semver"
)

const (
	// DefaultRepo is the GitHub repository hosting probo-agent
	// releases.
	DefaultRepo = "getprobo/probo"
	// DefaultTagPrefix is the tag prefix used by the agent's release
	// pipeline. Releases look like `probo-agent/v0.1.0`.
	DefaultTagPrefix = "probo-agent/v"

	defaultAPIBaseURL      = "https://api.github.com"
	defaultAssetBaseURL    = "https://github.com"
	defaultPageSize        = 30
	defaultDownloadLimit   = 200 * 1024 * 1024 // 200 MiB cap on archive size
	checksumFileName       = "checksums.txt"
	checksumBundleFileName = "checksums.txt.bundle"
)

// ErrNoUpdateAvailable is returned by CheckLatest when no release
// newer than the current version exists.
var ErrNoUpdateAvailable = errors.New("no update available")

type (
	// Updater self-updates the agent binary on the local host.
	Updater struct {
		Repo           string
		TagPrefix      string
		APIBaseURL     string
		AssetBaseURL   string
		CurrentVersion string
		ExePath        string
		UserAgent      string
		HTTP           *http.Client
		Logger         *log.Logger

		// SigstoreCacheDir is the on-disk directory used by the
		// default cosign Verifier to cache Sigstore TUF metadata.
		// Required when Verifier is nil.
		SigstoreCacheDir string

		// Verifier validates the Sigstore bundle that accompanies
		// every release. When nil, the default cosign Verifier is
		// constructed lazily on first Apply, pinned to the
		// probo-agent release workflow.
		Verifier Verifier

		// GOOS/GOARCH override the values used to compute the
		// archive name. They default to runtime.GOOS/GOARCH and
		// exist for tests.
		GOOS   string
		GOARCH string
	}

	// Release describes a candidate update.
	Release struct {
		Version           string
		Tag               string
		AssetName         string
		AssetURL          string
		ChecksumURL       string
		ChecksumBundleURL string
	}

	githubRelease struct {
		TagName     string        `json:"tag_name"`
		Draft       bool          `json:"draft"`
		Prerelease  bool          `json:"prerelease"`
		Assets      []githubAsset `json:"assets"`
		PublishedAt jsonTimestamp `json:"published_at"`
	}

	githubAsset struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}

	jsonTimestamp time.Time
)

func (j *jsonTimestamp) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		return nil
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}

	*j = jsonTimestamp(t)

	return nil
}

// New returns an Updater with sane defaults for production use.
//
// sigstoreCacheDir is the on-disk directory used to cache Sigstore
// TUF metadata for cosign bundle verification. It MUST be writable by
// the agent. A typical value is `<agent state dir>/sigstore-cache`.
func New(currentVersion, exePath, userAgent, sigstoreCacheDir string, logger *log.Logger) *Updater {
	if logger == nil {
		logger = log.NewLogger(log.WithName("agent-update"))
	}

	return &Updater{
		Repo:             DefaultRepo,
		TagPrefix:        DefaultTagPrefix,
		APIBaseURL:       defaultAPIBaseURL,
		AssetBaseURL:     defaultAssetBaseURL,
		CurrentVersion:   currentVersion,
		ExePath:          exePath,
		UserAgent:        userAgent,
		Logger:           logger,
		HTTP:             defaultHTTPClient(logger),
		SigstoreCacheDir: sigstoreCacheDir,
		GOOS:             runtime.GOOS,
		GOARCH:           runtime.GOARCH,
	}
}

func defaultHTTPClient(logger *log.Logger) *http.Client {
	return &http.Client{
		Transport: httpclient.DefaultPooledTransport(
			httpclient.WithLogger(logger),
			httpclient.WithSSRFProtection(),
		),
		Timeout: 5 * time.Minute,
	}
}

// CheckLatest queries GitHub for the highest semver release whose tag
// matches u.TagPrefix and is newer than u.CurrentVersion. It returns
// ErrNoUpdateAvailable when nothing newer exists.
func (u *Updater) CheckLatest(ctx context.Context) (*Release, error) {
	layout, err := LayoutFor(u.goos(), u.goarch())
	if err != nil {
		return nil, err
	}

	releases, err := u.listReleases(ctx)
	if err != nil {
		return nil, err
	}

	current := normalizeSemver(u.CurrentVersion)

	var best *Release

	for i := range releases {
		rel := &releases[i]
		if rel.Draft || rel.Prerelease {
			continue
		}

		ver, ok := parseTag(rel.TagName, u.TagPrefix)
		if !ok {
			continue
		}

		// Skip anything that is not strictly newer than the running
		// version. When running a dev build (`current` is empty)
		// every published release is considered newer.
		if current != "" && semver.Compare(normalizeSemver(ver), current) <= 0 {
			continue
		}

		if best != nil && semver.Compare(normalizeSemver(ver), normalizeSemver(best.Version)) <= 0 {
			continue
		}

		assetURL, ok := findAssetURL(rel.Assets, layout.ArchiveName)
		if !ok {
			continue
		}

		checksumURL, ok := findAssetURL(rel.Assets, checksumFileName)
		if !ok {
			continue
		}

		bundleURL, ok := findAssetURL(rel.Assets, checksumBundleFileName)
		if !ok {
			// Releases without a Sigstore bundle predate the
			// signed-release pipeline and cannot be verified.
			// Skip them so the agent never auto-installs an
			// unsigned artifact.
			continue
		}

		best = &Release{
			Version:           ver,
			Tag:               rel.TagName,
			AssetName:         layout.ArchiveName,
			AssetURL:          assetURL,
			ChecksumURL:       checksumURL,
			ChecksumBundleURL: bundleURL,
		}
	}

	if best == nil {
		return nil, ErrNoUpdateAvailable
	}

	return best, nil
}

// Apply downloads the release archive, verifies the Sigstore bundle
// covering checksums.txt, verifies the SHA-256 of the archive against
// the now-trusted checksums.txt, extracts the binary to a temp
// directory, and atomically replaces u.ExePath with the new binary.
//
// Failure at *any* verification step aborts the update without
// touching the running binary.
func (u *Updater) Apply(ctx context.Context, rel *Release) error {
	if rel == nil {
		return errors.New("nil release")
	}

	if u.ExePath == "" {
		return errors.New("agent executable path is empty")
	}

	if rel.ChecksumBundleURL == "" {
		return errors.New("release is not signed (no checksums.txt.bundle)")
	}

	layout, err := LayoutFor(u.goos(), u.goarch())
	if err != nil {
		return err
	}

	verifier, err := u.resolveVerifier()
	if err != nil {
		return err
	}

	workDir, err := os.MkdirTemp("", "probo-agent-update-")
	if err != nil {
		return fmt.Errorf("cannot create update workdir: %w", err)
	}

	defer func() { _ = os.RemoveAll(workDir) }()

	archivePath := filepath.Join(workDir, layout.ArchiveName)
	if err := u.downloadFile(ctx, rel.AssetURL, archivePath); err != nil {
		return fmt.Errorf("cannot download archive: %w", err)
	}

	checksumPath := filepath.Join(workDir, checksumFileName)
	if err := u.downloadFile(ctx, rel.ChecksumURL, checksumPath); err != nil {
		return fmt.Errorf("cannot download checksums: %w", err)
	}

	bundlePath := filepath.Join(workDir, checksumBundleFileName)
	if err := u.downloadFile(ctx, rel.ChecksumBundleURL, bundlePath); err != nil {
		return fmt.Errorf("cannot download sigstore bundle: %w", err)
	}

	// Anchor the trust chain: verify that the bundle attests
	// checksums.txt was signed by the pinned release workflow,
	// before reading anything from checksums.txt.
	if err := verifier.Verify(ctx, checksumPath, bundlePath); err != nil {
		return fmt.Errorf("cannot verify sigstore bundle: %w", err)
	}

	if err := verifyChecksum(archivePath, checksumPath, layout.ArchiveName); err != nil {
		return err
	}

	extractedBinary, err := extractBinary(archivePath, layout, workDir)
	if err != nil {
		return fmt.Errorf("cannot extract archive: %w", err)
	}

	if err := replaceBinary(u.ExePath, extractedBinary); err != nil {
		return fmt.Errorf("cannot replace agent binary: %w", err)
	}

	u.Logger.InfoCtx(
		ctx,
		"agent binary updated",
		log.String("version", rel.Version),
		log.String("tag", rel.Tag),
		log.String("asset", rel.AssetName),
	)

	return nil
}

// resolveVerifier returns a non-nil Verifier, lazily building the
// default cosign-backed verifier when none was injected.
func (u *Updater) resolveVerifier() (Verifier, error) {
	if u.Verifier != nil {
		return u.Verifier, nil
	}

	if u.SigstoreCacheDir == "" {
		return nil, errors.New("SigstoreCacheDir must be set for default cosign verifier")
	}

	v, err := NewCosignVerifier(
		CosignVerifierConfig{
			Repo:         u.Repo,
			WorkflowPath: expectedWorkflowPath,
			TagPrefix:    u.TagPrefix,
			CacheDir:     u.SigstoreCacheDir,
		},
	)
	if err != nil {
		return nil, err
	}

	u.Verifier = v

	return v, nil
}

// listReleases returns the most recent page of releases from GitHub.
func (u *Updater) listReleases(ctx context.Context) ([]githubRelease, error) {
	apiBase := u.APIBaseURL
	if apiBase == "" {
		apiBase = defaultAPIBaseURL
	}

	endpoint, err := url.JoinPath(apiBase, "repos", u.Repo, "releases")
	if err != nil {
		return nil, fmt.Errorf("cannot build releases URL: %w", err)
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("cannot parse releases URL: %w", err)
	}

	q := parsed.Query()
	q.Set("per_page", fmt.Sprintf("%d", defaultPageSize))
	parsed.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cannot build releases request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", u.userAgent())

	resp, err := u.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch releases: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("cannot fetch releases: %d %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out []githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4*1024*1024)).Decode(&out); err != nil {
		return nil, fmt.Errorf("cannot decode releases: %w", err)
	}

	return out, nil
}

func (u *Updater) downloadFile(ctx context.Context, src, dst string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return fmt.Errorf("cannot build download request: %w", err)
	}

	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", u.userAgent())

	resp, err := u.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("cannot fetch %s: %w", src, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("cannot fetch %s: %d %s", src, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	tmp := dst + ".part"

	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("cannot create %s: %w", tmp, err)
	}

	if _, err := io.Copy(f, io.LimitReader(resp.Body, defaultDownloadLimit+1)); err != nil {
		_ = f.Close()
		return fmt.Errorf("cannot stream %s: %w", src, err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("cannot close %s: %w", tmp, err)
	}

	stat, err := os.Stat(tmp)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", tmp, err)
	}

	if stat.Size() > defaultDownloadLimit {
		_ = os.Remove(tmp)
		return fmt.Errorf("download %s exceeds %d bytes", src, defaultDownloadLimit)
	}

	if err := os.Rename(tmp, dst); err != nil {
		return fmt.Errorf("cannot move %s into place: %w", tmp, err)
	}

	return nil
}

func (u *Updater) userAgent() string {
	if u.UserAgent != "" {
		return u.UserAgent
	}

	return "probo-agent-updater"
}

func (u *Updater) goos() string {
	if u.GOOS != "" {
		return u.GOOS
	}

	return runtime.GOOS
}

func (u *Updater) goarch() string {
	if u.GOARCH != "" {
		return u.GOARCH
	}

	return runtime.GOARCH
}

// parseTag returns the version (e.g. "0.2.0") for a tag whose value
// starts with prefix (e.g. "probo-agent/v").
func parseTag(tag, prefix string) (string, bool) {
	if !strings.HasPrefix(tag, prefix) {
		return "", false
	}

	v := strings.TrimPrefix(tag, prefix)
	if v == "" {
		return "", false
	}

	if !semver.IsValid("v" + v) {
		return "", false
	}

	return v, true
}

// normalizeSemver returns the canonical form expected by golang.org/x/mod/semver
// (a "v" prefix), or "" when the input is empty / invalid.
func normalizeSemver(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}

	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}

	if !semver.IsValid(v) {
		return ""
	}

	return v
}

func findAssetURL(assets []githubAsset, name string) (string, bool) {
	for _, a := range assets {
		if a.Name == name {
			return a.BrowserDownloadURL, true
		}
	}

	return "", false
}

// verifyChecksum checks that the SHA-256 digest of archivePath
// matches the entry for archiveName in the checksums.txt file
// produced by `sha256sum *.tar.gz *.zip`.
func verifyChecksum(archivePath, checksumPath, archiveName string) error {
	expected, err := readChecksum(checksumPath, archiveName)
	if err != nil {
		return err
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("cannot open archive: %w", err)
	}

	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("cannot hash archive: %w", err)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf(
			"update: checksum mismatch for %s (expected %s, got %s)",
			archiveName,
			expected,
			actual,
		)
	}

	return nil
}

func readChecksum(path, archiveName string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot read checksums: %w", err)
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// `sha256sum` output is `<hex>  <name>`; the GNU tool also
		// supports a single-space separator and a leading `*` flag
		// for binary mode. Handle both.
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := strings.TrimPrefix(fields[1], "*")
		if name != archiveName {
			continue
		}

		return strings.ToLower(fields[0]), nil
	}

	return "", fmt.Errorf("update: %s missing from checksums file", archiveName)
}
