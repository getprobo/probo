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

package githubsecretscanning_v1

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	publicKeyCacheDuration = time.Hour
	maxPublicKeyBodyBytes  = 1 << 20
)

type (
	HTTPClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	KeyProvider interface {
		PublicKey(ctx context.Context, identifier string) (*ecdsa.PublicKey, error)
	}

	GitHubKeyProvider struct {
		httpClient HTTPClient
		mu         sync.Mutex
		keys       map[string]*ecdsa.PublicKey
		expiresAt  time.Time
	}

	publicKeyResponse struct {
		PublicKeys []publicKeyEntry `json:"public_keys"`
	}

	publicKeyEntry struct {
		Identifier string `json:"key_identifier"`
		Key        string `json:"key"`
	}
)

var (
	ErrPublicKeyNotFound = errors.New("GitHub secret scanning public key not found")
)

func NewGitHubKeyProvider(httpClient HTTPClient) *GitHubKeyProvider {
	return &GitHubKeyProvider{
		httpClient: httpClient,
		keys:       make(map[string]*ecdsa.PublicKey),
	}
}

func (p *GitHubKeyProvider) PublicKey(ctx context.Context, identifier string) (*ecdsa.PublicKey, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if key, ok := p.keys[identifier]; ok && time.Now().Before(p.expiresAt) {
		return key, nil
	}

	if err := p.refresh(ctx); err != nil {
		return nil, err
	}

	key, ok := p.keys[identifier]
	if !ok {
		return nil, ErrPublicKeyNotFound
	}

	return key, nil
}

func (p *GitHubKeyProvider) refresh(ctx context.Context) error {
	endpoint := (&url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   "/meta/public_keys/secret_scanning",
	}).String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create GitHub public key request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot fetch GitHub public keys: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cannot fetch GitHub public keys: unexpected status %d", resp.StatusCode)
	}

	var payload publicKeyResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxPublicKeyBodyBytes)).Decode(&payload); err != nil {
		return fmt.Errorf("cannot decode GitHub public keys: %w", err)
	}

	keys := make(map[string]*ecdsa.PublicKey, len(payload.PublicKeys))
	for _, entry := range payload.PublicKeys {
		key, err := parsePublicKey(entry.Key)
		if err != nil {
			return fmt.Errorf("cannot parse GitHub public key: %w", err)
		}

		keys[entry.Identifier] = key
	}

	p.keys = keys
	p.expiresAt = time.Now().Add(publicKeyCacheDuration)

	return nil
}

func parsePublicKey(value string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(value))
	if block == nil {
		return nil, errors.New("invalid PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse PKIX public key: %w", err)
	}

	ecdsaKey, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not ECDSA")
	}

	return ecdsaKey, nil
}
