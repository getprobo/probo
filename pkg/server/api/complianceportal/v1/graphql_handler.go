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

package complianceportal_v1

import (
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/complianceportal/v1/schema"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/server/gqlutils/directives/authentication"
	"go.probo.inc/probo/pkg/server/gqlutils/directives/session"
)

func NewGraphQLHandler(
	iamSvc *iam.Service,
	visitorSvc *visitor.Service,
	resourceAliasSvc *resourcealias.Service,
	fileManagerSvc *filemanager.Service,
	esignSvc *esign.Service,
	mailmanSvc *mailman.Service,
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	cookieConfig securecookie.Config,
	tokenSecret string,
	limits gqlutils.Limits,
) http.Handler {
	config := schema.Config{
		Resolvers: &Resolver{
			iam:           iamSvc,
			visitor:       visitorSvc,
			resourceAlias: resourceAliasSvc,
			fileManager:   fileManagerSvc,
			esign:         esignSvc,
			mailman:       mailmanSvc,
			logger:        logger,
			baseURL:       baseURL,
			sessionCookie: authn.NewCookie(&cookieConfig),
		},
		Directives: schema.DirectiveRoot{
			Nda:            newNDADirective(logger, visitorSvc, esignSvc),
			Authentication: authentication.Directive,
			SessionOnly:    session.Directive,
		},
	}

	es := schema.NewExecutableSchema(config)
	gqlh := gqlutils.NewHandler(es, logger, limits)

	return gqlh
}
