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

package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/cachecontrol"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/net"
)

const (
	cimdMaxDocumentBytes = 5120
	cimdFetchTimeout     = 10 * time.Second
	cimdDefaultCacheTTL  = 24 * time.Hour
)

type (
	ClientMetadataDocument struct {
		ClientID                string   `json:"client_id"`
		ClientName              string   `json:"client_name"`
		ClientURI               string   `json:"client_uri"`
		LogoURI                 string   `json:"logo_uri"`
		RedirectURIs            []string `json:"redirect_uris"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	}

	cimdCacheEntry struct {
		document  *ClientMetadataDocument
		expiresAt time.Time
	}

	cimdFetcher struct {
		httpClient *http.Client
		logger     *log.Logger
		cache      sync.Map
	}
)

func cimdClientIDAllowed(clientID string, allowed []string) bool {
	if len(allowed) == 0 {
		return false
	}

	return slices.Contains(allowed, clientID)
}

func isCIMDClientID(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}

	if parsed.Scheme != "https" {
		return false
	}

	if parsed.Host == "" {
		return false
	}

	if parsed.Path == "" || parsed.Path == "/" {
		return false
	}

	if parsed.User != nil {
		return false
	}

	if parsed.Fragment != "" {
		return false
	}

	return true
}

func newCIMDFetcher(logger *log.Logger) *cimdFetcher {
	return &cimdFetcher{
		httpClient: httpclient.DefaultClient(
			httpclient.WithLogger(logger),
			httpclient.WithSSRFProtection(),
		),
		logger: logger,
	}
}

func (f *cimdFetcher) fetch(ctx context.Context, clientIDURL string) (*ClientMetadataDocument, error) {
	if entry, ok := f.loadCache(clientIDURL); ok {
		return entry, nil
	}

	reqCtx, cancel := context.WithTimeout(ctx, cimdFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, clientIDURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create cimd request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, NewError(
			ErrInvalidClient,
			WithDescription("cannot fetch client metadata document"),
		)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, NewError(
			ErrInvalidClient,
			WithDescription("client metadata document unavailable"),
		)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, cimdMaxDocumentBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read cimd response: %w", err)
	}

	if len(body) > cimdMaxDocumentBytes {
		return nil, NewError(
			ErrInvalidClient,
			WithDescription("client metadata document too large"),
		)
	}

	var doc ClientMetadataDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, NewError(
			ErrInvalidClient,
			WithDescription("client metadata document is not valid JSON"),
		)
	}

	if err := validateClientMetadataDocument(clientIDURL, &doc); err != nil {
		return nil, err
	}

	f.storeCache(clientIDURL, &doc, resp.Header.Get("Cache-Control"))

	return &doc, nil
}

func validateClientMetadataDocument(clientIDURL string, doc *ClientMetadataDocument) error {
	if doc.ClientID != clientIDURL {
		return NewError(
			ErrInvalidClient,
			WithDescription("client metadata client_id does not match document URL"),
		)
	}

	if strings.TrimSpace(doc.ClientName) == "" {
		return NewError(
			ErrInvalidClient,
			WithDescription("client metadata document missing client_name"),
		)
	}

	if len(doc.RedirectURIs) == 0 {
		return NewError(
			ErrInvalidClient,
			WithDescription("client metadata document missing redirect_uris"),
		)
	}

	for _, redirectURI := range doc.RedirectURIs {
		if err := validateCIMDRedirectURI(redirectURI); err != nil {
			return err
		}
	}

	authMethod := doc.TokenEndpointAuthMethod
	if authMethod == "" {
		authMethod = string(coredata.OAuth2ClientTokenEndpointAuthMethodNone)
	}

	switch coredata.OAuth2ClientTokenEndpointAuthMethod(authMethod) {
	case coredata.OAuth2ClientTokenEndpointAuthMethodNone:
		// Public MCP clients (ChatGPT, Claude) authenticate with PKCE.
	default:
		return NewError(
			ErrInvalidClient,
			WithDescription("unsupported token_endpoint_auth_method in client metadata document"),
		)
	}

	return nil
}

func validateCIMDRedirectURI(redirectURI string) error {
	parsed, err := url.Parse(redirectURI)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return NewError(
			ErrInvalidClient,
			WithDescription("client metadata document contains invalid redirect_uri"),
		)
	}

	if parsed.User != nil || parsed.Fragment != "" {
		return NewError(
			ErrInvalidClient,
			WithDescription("client metadata document contains invalid redirect_uri"),
		)
	}

	switch parsed.Scheme {
	case "https":
	case "http":
		if !net.IsLoopback(parsed.Hostname()) {
			return NewError(
				ErrInvalidClient,
				WithDescription("client metadata document contains invalid redirect_uri"),
			)
		}
	default:
		return NewError(
			ErrInvalidClient,
			WithDescription("client metadata document contains invalid redirect_uri"),
		)
	}

	return nil
}

func cimdRedirectURIAllowed(doc *ClientMetadataDocument, redirectURI string) bool {
	for _, allowed := range doc.RedirectURIs {
		if redirectURI == allowed {
			return true
		}

		if cimdLoopbackRedirectMatches(allowed, redirectURI) {
			return true
		}
	}

	return false
}

func cimdLoopbackRedirectMatches(registered, requested string) bool {
	registeredURL, err := url.Parse(registered)
	if err != nil {
		return false
	}

	requestedURL, err := url.Parse(requested)
	if err != nil {
		return false
	}

	if registeredURL.Scheme != requestedURL.Scheme {
		return false
	}

	if !net.IsLoopback(registeredURL.Hostname()) || !net.IsLoopback(requestedURL.Hostname()) {
		return false
	}

	if registeredURL.Hostname() != requestedURL.Hostname() {
		return false
	}

	return registeredURL.Path == requestedURL.Path &&
		registeredURL.RawQuery == requestedURL.RawQuery
}

func (f *cimdFetcher) loadCache(clientIDURL string) (*ClientMetadataDocument, bool) {
	raw, ok := f.cache.Load(clientIDURL)
	if !ok {
		return nil, false
	}

	entry := raw.(cimdCacheEntry)
	if time.Now().After(entry.expiresAt) {
		f.cache.Delete(clientIDURL)
		return nil, false
	}

	return entry.document, true
}

func (f *cimdFetcher) storeCache(clientIDURL string, doc *ClientMetadataDocument, cacheControl string) {
	dir, err := cachecontrol.ParseResponse(cacheControl)
	if err == nil && dir.NoStore() {
		return
	}

	ttl := cimdDefaultCacheTTL

	if err == nil {
		if maxAge, ok := dir.MaxAgeDuration(); ok {
			ttl = min(ttl, maxAge)
		}
	}

	f.cache.Store(
		clientIDURL,
		cimdCacheEntry{
			document:  doc,
			expiresAt: time.Now().Add(ttl),
		},
	)
}

func (s *Service) resolveClient(
	ctx context.Context,
	tx pg.Tx,
	clientIDRaw string,
	redirectURI string,
) (*coredata.OAuth2Client, error) {
	if clientID, err := gid.ParseGID(clientIDRaw); err == nil {
		if tx != nil {
			client := coredata.OAuth2Client{}
			if err := client.LoadByID(ctx, tx, coredata.NewNoScope(), clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil, NewError(ErrInvalidClient, WithDescription("client not found"))
				}

				return nil, fmt.Errorf("cannot load oauth2 client: %w", err)
			}

			return &client, nil
		}

		return s.GetClientByID(ctx, clientID)
	}

	if !isCIMDClientID(clientIDRaw) {
		return nil, NewError(ErrInvalidClient, WithDescription("invalid client_id"))
	}

	if !cimdClientIDAllowed(clientIDRaw, s.cimdAllowedClientIDs) {
		return nil, NewError(
			ErrInvalidClient,
			WithDescription("client_id is not allowed for client metadata documents"),
		)
	}

	doc, err := s.cimd.fetch(ctx, clientIDRaw)
	if err != nil {
		return nil, err
	}

	if redirectURI != "" && !cimdRedirectURIAllowed(doc, redirectURI) {
		return nil, ErrInvalidRedirectURI
	}

	client, err := s.upsertCIMDClient(ctx, tx, clientIDRaw, doc)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *Service) upsertCIMDClient(
	ctx context.Context,
	tx pg.Tx,
	externalClientID string,
	doc *ClientMetadataDocument,
) (*coredata.OAuth2Client, error) {
	var logoURI, clientURI *string
	if doc.LogoURI != "" {
		logoURI = &doc.LogoURI
	}

	if doc.ClientURI != "" {
		clientURI = &doc.ClientURI
	}

	scopes := coredata.OAuth2Scopes(authorizationServerScopes(s.registry.AllWriteScopes()))

	now := time.Now()

	candidate, err := coredata.NewCIMDClient(
		externalClientID,
		doc.ClientName,
		doc.RedirectURIs,
		scopes,
		logoURI,
		clientURI,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot build cimd client: %w", err)
	}

	var client coredata.OAuth2Client

	upsert := func(ctx context.Context, conn pg.Tx) error {
		client = *candidate

		if err := client.UpsertCIMD(ctx, conn); err != nil {
			return fmt.Errorf("cannot upsert cimd oauth2 client: %w", err)
		}

		return nil
	}

	if tx != nil {
		if err := upsert(ctx, tx); err != nil {
			return nil, err
		}

		return &client, nil
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			return upsert(ctx, conn)
		},
	)
	if err != nil {
		return nil, err
	}

	return &client, nil
}
