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

package update

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/sigstore/sigstore-go/pkg/bundle"
	"github.com/sigstore/sigstore-go/pkg/root"
	"github.com/sigstore/sigstore-go/pkg/tuf"
	"github.com/sigstore/sigstore-go/pkg/verify"
)

const (
	// expectedSignerIssuer is the OIDC issuer Fulcio embeds in the
	// signing certificate when the workflow uses GitHub Actions'
	// OIDC token. This is the public-good Sigstore configuration.
	expectedSignerIssuer = "https://token.actions.githubusercontent.com"

	// expectedWorkflowPath is the path of the release workflow that
	// is allowed to produce signed probo-agent artifacts. Anything
	// signed by a different workflow (or a workflow run outside of
	// a tagged commit) is rejected.
	expectedWorkflowPath = ".github/workflows/release-probo-agent.yaml"
)

// Verifier verifies that a Sigstore bundle (`checksums.txt.bundle`)
// attests an artifact (`checksums.txt`) was produced by the expected
// signer identity. Implementations MUST hard-fail on any error;
// callers do not interpret the error type.
type Verifier interface {
	// Verify returns nil iff bundlePath is a valid Sigstore bundle
	// for the artifact at artifactPath, and the signer identity
	// matches the verifier's pinned issuer / SAN regex.
	Verify(ctx context.Context, artifactPath, bundlePath string) error
}

// AllowAllVerifier accepts every input. It exists strictly for tests
// of the surrounding download / extract pipeline. Production callers
// must wire a real Verifier (e.g. CosignVerifier).
type AllowAllVerifier struct{}

// Verify always returns nil.
func (AllowAllVerifier) Verify(_ context.Context, _, _ string) error { return nil }

// rejectAllVerifier is exposed for tests that need to assert Apply
// hard-fails on signature problems.
type rejectAllVerifier struct{ err error }

func (v rejectAllVerifier) Verify(_ context.Context, _, _ string) error { return v.err }

// CosignVerifier verifies cosign sign-blob bundles using sigstore-go
// against the Sigstore public-good trust root.
//
// The verifier pins the signer identity to the probo-agent release
// workflow on a tagged commit:
//
//	issuer:  https://token.actions.githubusercontent.com
//	SAN:     https://github.com/<repo>/<workflow>@refs/tags/<tag-prefix><version>
//
// where <repo>, <workflow> and <tag-prefix> default to the values
// hard-coded in the release pipeline. Both fields can be overridden
// for testing or for repository forks.
type CosignVerifier struct {
	Issuer      string
	SANRegex    string
	trustedRoot *root.TrustedRoot
}

// CosignVerifierConfig configures a CosignVerifier.
type CosignVerifierConfig struct {
	// Repo identifies the GitHub repository (e.g. "getprobo/probo").
	Repo string
	// WorkflowPath is the path within the repo to the workflow file
	// allowed to produce signed releases.
	WorkflowPath string
	// TagPrefix is the tag prefix the release workflow signs against
	// (e.g. "probo-agent/v"). The verifier matches anything after
	// this prefix.
	TagPrefix string
	// CacheDir is the on-disk directory used to cache the Sigstore
	// TUF metadata. Required.
	CacheDir string
}

// NewCosignVerifier loads the Sigstore public-good trust root via
// TUF (cached under cfg.CacheDir) and returns a Verifier that pins
// signatures to the configured GitHub Actions workflow on a tagged
// commit.
func NewCosignVerifier(cfg CosignVerifierConfig) (*CosignVerifier, error) {
	if cfg.Repo == "" {
		return nil, fmt.Errorf("update: cosign verifier requires Repo")
	}

	if cfg.WorkflowPath == "" {
		cfg.WorkflowPath = expectedWorkflowPath
	}

	if cfg.TagPrefix == "" {
		cfg.TagPrefix = DefaultTagPrefix
	}

	if cfg.CacheDir == "" {
		return nil, fmt.Errorf("update: cosign verifier requires CacheDir")
	}

	if err := os.MkdirAll(cfg.CacheDir, 0o700); err != nil {
		return nil, fmt.Errorf("cannot create sigstore cache dir: %w", err)
	}

	opts := tuf.DefaultOptions()
	opts.CachePath = cfg.CacheDir

	tufClient, err := tuf.New(opts)
	if err != nil {
		return nil, fmt.Errorf("cannot init sigstore TUF client: %w", err)
	}

	trustedRoot, err := root.GetTrustedRoot(tufClient)
	if err != nil {
		return nil, fmt.Errorf("cannot load sigstore trusted root: %w", err)
	}

	sanRegex := buildSANRegex(cfg.Repo, cfg.WorkflowPath, cfg.TagPrefix)

	return &CosignVerifier{
		Issuer:      expectedSignerIssuer,
		SANRegex:    sanRegex,
		trustedRoot: trustedRoot,
	}, nil
}

// Verify validates that bundlePath attests artifactPath was signed by
// the expected GitHub Actions workflow on a tagged release.
func (v *CosignVerifier) Verify(_ context.Context, artifactPath, bundlePath string) error {
	b, err := bundle.LoadJSONFromPath(bundlePath)
	if err != nil {
		return fmt.Errorf("cannot load sigstore bundle: %w", err)
	}

	identity, err := verify.NewShortCertificateIdentity(v.Issuer, "", "", v.SANRegex)
	if err != nil {
		return fmt.Errorf("cannot build signer identity: %w", err)
	}

	sev, err := verify.NewVerifier(
		v.trustedRoot,
		verify.WithSignedCertificateTimestamps(1),
		verify.WithTransparencyLog(1),
		verify.WithObserverTimestamps(1),
	)
	if err != nil {
		return fmt.Errorf("cannot build sigstore verifier: %w", err)
	}

	artifact, err := os.Open(artifactPath)
	if err != nil {
		return fmt.Errorf("cannot open artifact for verification: %w", err)
	}

	defer func() { _ = artifact.Close() }()

	policy := verify.NewPolicy(
		verify.WithArtifact(artifact),
		verify.WithCertificateIdentity(identity),
	)

	if _, err := sev.Verify(b, policy); err != nil {
		return fmt.Errorf("sigstore verification failed: %w", err)
	}

	return nil
}

func buildSANRegex(repo, workflowPath, tagPrefix string) string {
	return `^https://github\.com/` +
		regexp.QuoteMeta(repo) +
		`/` +
		regexp.QuoteMeta(workflowPath) +
		`@refs/tags/` +
		regexp.QuoteMeta(tagPrefix) +
		`.+$`
}
