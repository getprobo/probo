// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package coredata

import (
	"strings"

	"github.com/jackc/pgx/v5"
)

type (
	ConnectorFilter struct {
		provider       *ConnectorProvider
		providers      []ConnectorProvider
		query          *string
		queryProviders []ConnectorProvider
	}
)

func NewConnectorProviderFilter(provider *ConnectorProvider) *ConnectorFilter {
	return &ConnectorFilter{
		provider: provider,
	}
}

// NewConnectorListFilter restricts a connector list. An empty providers
// slice does not constrain the provider. query matches connectors.name;
// queryProviders are the providers whose display name or slug matched query,
// resolved by the caller because display names are not stored.
func NewConnectorListFilter(
	providers []ConnectorProvider,
	query *string,
	queryProviders []ConnectorProvider,
) *ConnectorFilter {
	filter := &ConnectorFilter{
		queryProviders: queryProviders,
	}

	if len(providers) > 0 {
		filter.providers = providers
	}

	if query != nil {
		trimmed := strings.TrimSpace(*query)
		if trimmed != "" {
			filter.query = &trimmed
		}
	}

	return filter
}

func (f *ConnectorFilter) SQLArguments() pgx.NamedArgs {
	var filterProviders []string
	if len(f.providers) > 0 {
		filterProviders = connectorProviderStrings(f.providers)
	}

	args := pgx.NamedArgs{
		"provider":               f.provider,
		"filter_providers":       filterProviders,
		"filter_query":           nil,
		"filter_query_providers": connectorProviderStrings(f.queryProviders),
	}

	if f.query != nil && *f.query != "" {
		args["filter_query"] = escapeLikePattern(*f.query)
	}

	return args
}

func (f *ConnectorFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @provider::connector_provider IS NULL THEN
			TRUE
		ELSE
			provider = @provider::connector_provider
	END
	AND CASE
		WHEN @filter_providers::connector_provider[] IS NULL THEN
			TRUE
		ELSE
			provider = ANY(@filter_providers::connector_provider[])
	END
	AND CASE
		WHEN @filter_query::text IS NOT NULL AND @filter_query::text <> '' THEN
			name ILIKE '%' || @filter_query || '%' ESCAPE '\'
			OR provider = ANY(@filter_query_providers::connector_provider[])
		ELSE TRUE
	END
)`
}

func connectorProviderStrings(providers []ConnectorProvider) []string {
	out := make([]string, len(providers))
	for i, provider := range providers {
		out[i] = string(provider)
	}

	return out
}
