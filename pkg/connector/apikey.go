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

package connector

import (
	"context"
	"encoding/json"
	"net/http"

	"go.gearno.de/kit/httpclient"
)

type APIKeyConnection struct {
	APIKey string `json:"api_key"`
	// Header selects how the API key is presented on outbound requests.
	// Empty (the default) sends it as `Authorization: Bearer <key>`,
	// which every OAuth-style and standard API-key connector uses. A
	// non-empty value (e.g. "x-api-key") sends the raw key in that
	// request header instead and omits Authorization entirely —
	// required by providers such as Anthropic that reject Bearer auth
	// and return 400 when both x-api-key and Authorization are present.
	// It is populated from the provider Registration at connector
	// creation time.
	Header string `json:"header,omitempty"`
	// BasicAuth, when true, presents the API key as the username of an
	// HTTP Basic credential with an empty password (`Authorization:
	// Basic base64(<key>:)`) — required by providers such as Cursor
	// whose Admin API documents Basic auth and rejects Bearer tokens.
	// It is mutually exclusive with Header and is populated from the
	// provider Registration at connector creation time.
	BasicAuth bool `json:"basic_auth,omitempty"`
}

var _ Connection = (*APIKeyConnection)(nil)

func (c *APIKeyConnection) Type() ProtocolType {
	return ProtocolAPIKey
}

func (c *APIKeyConnection) Scopes() []string {
	return nil
}

func (c *APIKeyConnection) Client(ctx context.Context) (*http.Client, error) {
	underlying := httpclient.DefaultPooledTransport(httpclient.WithSSRFProtection())

	if c.BasicAuth {
		return &http.Client{
			Transport: &basicAuthTransport{
				username:   c.APIKey,
				underlying: underlying,
			},
		}, nil
	}

	if c.Header != "" {
		return &http.Client{
			Transport: &apiKeyHeaderTransport{
				header:     c.Header,
				value:      c.APIKey,
				underlying: underlying,
			},
		}, nil
	}

	return &http.Client{
		Transport: &oauth2Transport{
			token:      c.APIKey,
			tokenType:  "Bearer",
			underlying: underlying,
		},
	}, nil
}

// apiKeyHeaderTransport injects the API key into a custom request header
// (for example "x-api-key") and, unlike oauth2Transport, never sets
// Authorization. Providers such as Anthropic require the key in their
// own header and reject requests that carry both that header and
// Authorization.
type apiKeyHeaderTransport struct {
	header     string
	value      string
	underlying http.RoundTripper
}

func (t *apiKeyHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.Header.Set(t.header, t.value)

	return t.underlying.RoundTrip(req2)
}

// basicAuthTransport presents the API key as the username of an HTTP
// Basic credential with an empty password. Providers such as Cursor
// document `-u <key>:` Basic auth for their Admin API and reject Bearer
// tokens, so neither oauth2Transport nor apiKeyHeaderTransport fits.
type basicAuthTransport struct {
	username   string
	underlying http.RoundTripper
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.SetBasicAuth(t.username, "")

	return t.underlying.RoundTrip(req2)
}

func (c APIKeyConnection) MarshalJSON() ([]byte, error) {
	type Alias APIKeyConnection

	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  string(ProtocolAPIKey),
		Alias: Alias(c),
	})
}

func (c *APIKeyConnection) UnmarshalJSON(data []byte) error {
	type Alias APIKeyConnection

	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	return json.Unmarshal(data, &aux)
}
