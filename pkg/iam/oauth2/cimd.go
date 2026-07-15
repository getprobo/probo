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
	"go.probo.inc/probo/pkg/netx"
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
		Scope                   string   `json:"scope,omitempty"`
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

	CIMDAllowance string

	CIMDAllowFunc func(ctx context.Context, clientIDURL string) (CIMDAllowance, error)
)

const (
	CIMDAllowanceDenied             CIMDAllowance = "denied"
	CIMDAllowanceAllowed            CIMDAllowance = "allowed"
	CIMDAllowanceAllowedSkipConsent CIMDAllowance = "allowed_skip_consent"
)

func (a CIMDAllowance) Allowed() bool {
	return a != CIMDAllowanceDenied
}

func (a CIMDAllowance) SkipsConsent() bool {
	return a == CIMDAllowanceAllowedSkipConsent
}

func CIMDAllowFromClientIDs(clientIDs []string) CIMDAllowFunc {
	allowed := slices.Clone(clientIDs)

	return func(_ context.Context, clientIDURL string) (CIMDAllowance, error) {
		if slices.Contains(allowed, clientIDURL) {
			return CIMDAllowanceAllowed, nil
		}

		return CIMDAllowanceDenied, nil
	}
}

func CIMDClientIDHost(raw string) (string, bool) {
	if !IsCIMDClientID(raw) {
		return "", false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	host := parsed.Hostname()
	if host == "" {
		return "", false
	}

	return host, true
}

func IsCIMDClientID(raw string) bool {
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
	// CIMD URLs are allowlisted in resolveClient before fetch runs.
	return &cimdFetcher{
		httpClient: httpclient.DefaultClient(
			httpclient.WithLogger(logger),
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

	return f.finishFetch(clientIDURL, &doc, resp.Header.Get("Cache-Control"))
}

func (f *cimdFetcher) finishFetch(
	clientIDURL string,
	doc *ClientMetadataDocument,
	cacheControl string,
) (*ClientMetadataDocument, error) {
	if err := validateClientMetadataDocument(clientIDURL, doc); err != nil {
		return nil, err
	}

	f.storeCache(clientIDURL, doc, cacheControl)

	return doc, nil
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
		if !netx.IsLoopback(parsed.Hostname()) {
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

	if !IsCIMDClientID(clientIDRaw) {
		return nil, NewError(ErrInvalidClient, WithDescription("invalid client_id"))
	}

	if allowance, err := s.cimdAllowance(ctx, clientIDRaw); err != nil || !allowance.Allowed() {
		if err != nil {
			s.logger.WarnCtx(ctx, "cannot check cimd client allowance", log.Error(err))
		}

		return nil, NewError(
			ErrInvalidClient,
			WithDescription("client_id is not allowed for client metadata documents"),
		)
	}

	doc, err := s.cimd.fetch(ctx, clientIDRaw)
	if err != nil {
		return nil, err
	}

	scopes, err := s.cimdScopes(doc)
	if err != nil {
		return nil, err
	}

	client, err := s.upsertCIMDClient(ctx, tx, clientIDRaw, doc, scopes)
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
	scopes coredata.OAuth2Scopes,
) (*coredata.OAuth2Client, error) {
	var logoURI, clientURI *string
	if doc.LogoURI != "" {
		logoURI = &doc.LogoURI
	}

	if doc.ClientURI != "" {
		clientURI = &doc.ClientURI
	}

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

func (s *Service) cimdAllowance(ctx context.Context, clientIDRaw string) (CIMDAllowance, error) {
	if !IsCIMDClientID(clientIDRaw) {
		return CIMDAllowanceDenied, nil
	}

	if s.cimdAllow == nil {
		return CIMDAllowanceDenied, nil
	}

	return s.cimdAllow(ctx, clientIDRaw)
}

func (s *Service) cimdScopes(doc *ClientMetadataDocument) (coredata.OAuth2Scopes, error) {
	if strings.TrimSpace(doc.Scope) == "" {
		return coredata.OAuth2Scopes(authorizationServerScopes(s.registry.AllWriteScopes())), nil
	}

	scopes, err := parseCIMDMetadataScopes(doc.Scope)
	if err != nil {
		return nil, NewError(ErrInvalidScope, WithDescription(err.Error()))
	}

	if err := s.validateCIMDScopes(scopes); err != nil {
		return nil, err
	}

	return scopes, nil
}

func parseCIMDMetadataScopes(raw string) (coredata.OAuth2Scopes, error) {
	fields := strings.Fields(strings.TrimSpace(raw))
	if len(fields) == 0 {
		return nil, nil
	}

	scopes := make(coredata.OAuth2Scopes, len(fields))
	for i, field := range fields {
		scopes[i] = coredata.OAuth2Scope(field)
	}

	return scopes, nil
}

func (s *Service) validateCIMDScopes(scopes coredata.OAuth2Scopes) error {
	for _, scope := range scopes {
		if IsStandardScope(scope) {
			continue
		}

		if err := s.registry.ValidateScopes(coredata.OAuth2Scopes{scope}); err != nil {
			return NewError(
				ErrInvalidScope,
				WithDescription(fmt.Sprintf("invalid scope in client metadata document: %s", scope)),
			)
		}
	}

	return nil
}

func (s *Service) CIMDClientMetadata(ctx context.Context, clientIDRaw string) (*ClientMetadataDocument, error) {
	if !IsCIMDClientID(clientIDRaw) {
		return nil, nil
	}

	doc, err := s.cimd.fetch(ctx, clientIDRaw)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch cimd client metadata: %w", err)
	}

	return doc, nil
}

func (s *Service) CIMDClientDisplayName(ctx context.Context, clientIDRaw string) (*string, error) {
	doc, err := s.CIMDClientMetadata(ctx, clientIDRaw)
	if err != nil {
		return nil, err
	}

	if doc == nil || doc.ClientName == "" {
		return nil, nil
	}

	return &doc.ClientName, nil
}
