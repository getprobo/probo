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

package visitor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
)

const oauthStateTTL = 15 * time.Minute

type OAuthState struct {
	ContinueURL  string
	CodeVerifier string
	Nonce        string
}

func (s *Service) InitiateOAuthAuthorizeURL(
	ctx context.Context,
	authorizeURL string,
	clientID string,
	redirectURI string,
	scopes []string,
	continueURL string,
) (string, error) {
	state, err := generateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("cannot generate state: %w", err)
	}

	nonce, err := generateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("cannot generate nonce: %w", err)
	}

	codeVerifier, err := generateRandomString(64)
	if err != nil {
		return "", fmt.Errorf("cannot generate code verifier: %w", err)
	}

	err = s.persistOAuthState(
		ctx,
		state,
		continueURL,
		oauthSession{
			CodeVerifier: codeVerifier,
			Nonce:        nonce,
		},
	)
	if err != nil {
		return "", err
	}

	authorizeURL, err = buildAuthorizeURL(
		authorizeURL,
		clientID,
		redirectURI,
		scopes,
		state,
		nonce,
		computeCodeChallenge(codeVerifier),
	)
	if err != nil {
		return "", err
	}

	return authorizeURL, nil
}

func (s *Service) ConsumeOAuthState(ctx context.Context, stateID string) (OAuthState, error) {
	var oidcState coredata.OIDCState

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := oidcState.LoadByIDForUpdate(ctx, tx, stateID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrOAuthStateNotFound
				}

				return fmt.Errorf("cannot load oidc state: %w", err)
			}

			if err := oidcState.Delete(ctx, tx); err != nil {
				return fmt.Errorf("cannot delete oidc state: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return OAuthState{}, err
	}

	if time.Now().After(oidcState.ExpiresAt) {
		return OAuthState{}, ErrOAuthStateExpired
	}

	if oidcState.Provider != coredata.OIDCProviderCompliancePortal {
		return OAuthState{}, ErrOAuthStateNotFound
	}

	return OAuthState{
		ContinueURL:  oidcState.ContinueURL,
		CodeVerifier: oidcState.CodeVerifier,
		Nonce:        oidcState.Nonce,
	}, nil
}

type oauthSession struct {
	CodeVerifier string
	Nonce        string
}

func (s *Service) persistOAuthState(
	ctx context.Context,
	stateID string,
	continueURL string,
	session oauthSession,
) error {
	now := time.Now()
	oidcState := &coredata.OIDCState{
		ID:           stateID,
		Provider:     coredata.OIDCProviderCompliancePortal,
		Nonce:        session.Nonce,
		CodeVerifier: session.CodeVerifier,
		ContinueURL:  continueURL,
		CreatedAt:    now,
		ExpiresAt:    now.Add(oauthStateTTL),
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := oidcState.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert oidc state: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("cannot persist oauth state: %w", err)
	}

	return nil
}

func buildAuthorizeURL(
	authorizeURL string,
	clientID string,
	redirectURI string,
	scopes []string,
	state string,
	nonce string,
	codeChallenge string,
) (string, error) {
	u, err := url.Parse(authorizeURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse authorize URL: %w", err)
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func computeCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))

	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", fmt.Errorf("cannot generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
