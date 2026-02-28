// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package auth

import (
	"context"

	"go.probo.inc/probo/pkg/coredata"
)

type (
	ctxKey struct{ name string }
)

var (
	consoleIdentityContextKey = &ctxKey{name: "console_identity"}
	consoleSessionContextKey  = &ctxKey{name: "console_session"}
	consoleAPIKeyContextKey   = &ctxKey{name: "console_api_key"}
	consoleOrganizationKey    = &ctxKey{name: "console_organization"}
)

func ConsoleSessionFromContext(ctx context.Context) *coredata.Session {
	session, _ := ctx.Value(consoleSessionContextKey).(*coredata.Session)
	return session
}

func ConsoleContextWithSession(ctx context.Context, session *coredata.Session) context.Context {
	return context.WithValue(ctx, consoleSessionContextKey, session)
}

func ConsoleIdentityFromContext(ctx context.Context) *coredata.Identity {
	identity, _ := ctx.Value(consoleIdentityContextKey).(*coredata.Identity)
	return identity
}

func ConsoleContextWithIdentity(ctx context.Context, identity *coredata.Identity) context.Context {
	return context.WithValue(ctx, consoleIdentityContextKey, identity)
}

func ConsoleAPIKeyFromContext(ctx context.Context) *coredata.PersonalAPIKey {
	apiKey, _ := ctx.Value(consoleAPIKeyContextKey).(*coredata.PersonalAPIKey)
	return apiKey
}

func ConsoleContextWithAPIKey(ctx context.Context, apiKey *coredata.PersonalAPIKey) context.Context {
	return context.WithValue(ctx, consoleAPIKeyContextKey, apiKey)
}

func ConsoleOrganizationFromContext(ctx context.Context) *coredata.Organization {
	org, _ := ctx.Value(consoleOrganizationKey).(*coredata.Organization)
	return org
}

func ConsoleContextWithOrganization(ctx context.Context, org *coredata.Organization) context.Context {
	return context.WithValue(ctx, consoleOrganizationKey, org)
}
