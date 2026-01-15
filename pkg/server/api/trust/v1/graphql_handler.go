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

package trust_v1

import (
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/trust/v1/schema"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/server/gqlutils/directives/session"
	"go.probo.inc/probo/pkg/trust"
)

func NewGraphQLHandler(iamSvc *iam.Service, trustSvc *trust.Service, logger *log.Logger, baseURL *baseurl.BaseURL, cookieConfig securecookie.Config) http.Handler {
	config := schema.Config{
		Resolvers: &Resolver{
			iam:           iamSvc,
			trust:         trustSvc,
			logger:        logger,
			baseURL:       baseURL,
			sessionCookie: authn.NewCookie(&cookieConfig),
		},
		Directives: schema.DirectiveRoot{
			Session: session.Directive,
		},
	}

	es := schema.NewExecutableSchema(config)
	gqlh := gqlutils.NewHandler(es, logger)

	return gqlh
}
