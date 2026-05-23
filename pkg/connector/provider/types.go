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

package provider

import (
	"context"
	"encoding/json"
	"net/http"

	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// Registration is the per-provider metadata + factory bundle. Each
// provider returns one of these from a private constructor (e.g.
// slackRegistration) that NewBuiltinRegistry assembles into the
// runtime *Registry. Fields are grouped by concern: identity, OAuth2
// metadata, supported protocols, extra settings, and factory closures.
type Registration struct {
	// Identity.
	Provider    coredata.ConnectorProvider
	DisplayName string

	// OAuth2 metadata.
	AuthURL                 string
	TokenURL                string
	ExtraAuthParams         map[string]string
	TokenEndpointAuth       string // "post-form" (default), "basic-form", or "basic-json"
	SupportsIncrementalAuth bool
	OAuth2Scopes            []string
	ProbeURL                string
	// RequiresPKCE enables RFC 7636 PKCE (S256) on the authorization
	// request and replays the verifier on the token exchange. Default
	// false; non-PKCE providers are unaffected.
	RequiresPKCE bool
	// AuthURLParams are operator-supplied placeholders substituted
	// into the static provider AuthURL (e.g. Vercel's
	// "{integration_slug}"). Empty for the vast majority of providers.
	AuthURLParams map[string]string

	// Protocol support / GraphQL surface.
	SupportsAPIKey            bool
	SupportsClientCredentials bool
	ExtraSettings             []ExtraSetting

	// Factory closures — wired by Stages 2 and 3.
	NewDriver               func(context.Context, *http.Client, *coredata.Connector, *log.Logger) (drivers.Driver, error)
	NewNameResolver         func(context.Context, *http.Client, *coredata.Connector, *log.Logger) drivers.NameResolver
	SetOrganizationSettings func(*coredata.Connector, string) error
	// MarshalSettings normalises the per-provider extra settings into
	// the JSON blob persisted on coredata.Connector.RawSettings.
	//
	// SECURITY CONTRACT: returned errors are surfaced verbatim to the
	// client via gqlutils.Invalid. They must contain only field names
	// and structural information — never user-supplied values,
	// secrets, or driver-internal details. Use a static string per
	// validation failure ("sentryOrganizationSlug is required",
	// "onePasswordRegion must be one of …"), never interpolate input.
	MarshalSettings func(*SettingsInput) (json.RawMessage, error)
}

// SettingsInput is the union of every optional per-provider field
// available on the GraphQL CreateAPIKeyConnectorInput and
// CreateClientCredentialsConnectorInput types. The resolver populates
// it once from the gqlgen input; each provider's MarshalSettings
// reads only the fields it cares about.
//
// Adding a new provider with extra settings: add the optional field
// here + the corresponding optional field on the GraphQL input + the
// read in the per-provider MarshalSettings closure.
type SettingsInput struct {
	TallyOrganizationID      *string
	SentryOrganizationSlug   *string
	SupabaseOrganizationSlug *string
	GitHubOrganization       *string
	OnePasswordSCIMBridgeURL *string
	OnePasswordAccountID     *string
	OnePasswordRegion        *string
}

// ExtraSetting describes one extra per-provider settings field
// surfaced on ConnectorProviderInfo for the frontend to render.
type ExtraSetting struct {
	Key      string
	Label    string
	Required bool
}
