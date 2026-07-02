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

package connect_v1

import (
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/authz"
	"go.probo.inc/probo/pkg/server/api/connect/v1/schema"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/server/gqlutils/directives/authentication"
	"go.probo.inc/probo/pkg/server/gqlutils/directives/session"
)

func NewGraphQLHandler(svc *iam.Service, logger *log.Logger, fileManagerSvc *filemanager.Service, baseURL *baseurl.BaseURL, cookieConfig securecookie.Config, limits gqlutils.Limits) http.Handler {
	config := schema.Config{
		Resolvers: &Resolver{
			authorize:      authz.NewAuthorizeFunc(svc, logger),
			batchAuthorize: authz.NewBatchAuthorizeFunc(svc, logger),
			logger:         logger,
			iam:            svc,
			scopeRegistry:  svc.OAuth2ScopeRegistry,
			fileManager:    fileManagerSvc,
			baseURL:        baseURL,
			sessionCookie:  authn.NewCookie(&cookieConfig),
		},
		Directives: schema.DirectiveRoot{
			Authentication: authentication.Directive,
			SessionOnly:    session.Directive,
		},
	}

	es := schema.NewExecutableSchema(config)
	gqlh := gqlutils.NewHandler(es, logger, limits)

	return gqlh
}
