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
}

var _ Connection = (*APIKeyConnection)(nil)

func (c *APIKeyConnection) Type() ProtocolType {
	return ProtocolAPIKey
}

func (c *APIKeyConnection) Scopes() []string {
	return nil
}

func (c *APIKeyConnection) Client(ctx context.Context) (*http.Client, error) {
	transport := &oauth2Transport{
		token:      c.APIKey,
		tokenType:  "Bearer",
		underlying: httpclient.DefaultPooledTransport(httpclient.WithSSRFProtection()),
	}
	return &http.Client{Transport: transport}, nil
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
