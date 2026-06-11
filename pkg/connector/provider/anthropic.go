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

package provider

import (
	"context"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func anthropicRegistration() *Registration {
	return &Registration{
		Provider:       coredata.ConnectorProviderAnthropic,
		DisplayName:    "Anthropic",
		SupportsAPIKey: true,
		// Anthropic's Admin API authenticates with the admin key in the
		// x-api-key header; it rejects Authorization: Bearer and returns
		// 400 when both headers are present. APIKeyHeader makes the
		// APIKeyConnection send x-api-key instead of Bearer. There is no
		// third-party OAuth2 flow for the Admin API, so this is API-key
		// only and takes a single admin key (sk-ant-admin...) per org.
		APIKeyHeader: "x-api-key",
		Probe:        probeAnthropic,
		NewDriver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			return drivers.NewAnthropicDriver(c), nil
		},
		NewNameResolver: func(_ context.Context, c *http.Client, _ *coredata.Connector, _ *log.Logger) drivers.NameResolver {
			return drivers.NewAnthropicNameResolver(c)
		},
	}
}
