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

//go:generate go tool github.com/99designs/gqlgen generate

// Copyright (c) 2025 Probo Inc <hello@probo.com>.
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

package trust_v1

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	trust "go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

type (
	TrustAuthConfig struct {
		CookieName        string
		CookieDomain      string
		CookieDuration    time.Duration
		TokenDuration     time.Duration
		ReportURLDuration time.Duration
		Scope             string
		TokenType         string
		CookieSecure      bool
	}

	Resolver struct {
		trust         *trust.Service
		resourceAlias *resourcealias.Service
		fileManager   *filemanager.Service
		esign         *esign.Service
		mailman       *mailman.Service
		logger        *log.Logger
		iam           *iam.Service
		sessionCookie *authn.Cookie
		baseURL       *baseurl.BaseURL
	}
)

func NewMux(
	logger *log.Logger,
	iamSvc *iam.Service,
	trustSvc *trust.Service,
	resourceAliasSvc *resourcealias.Service,
	fileManagerSvc *filemanager.Service,
	esignSvc *esign.Service,
	mailmanSvc *mailman.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	baseURL *baseurl.BaseURL,
	graphqlLimits gqlutils.Limits,
) *chi.Mux {
	r := chi.NewMux()

	r.Use(complianceportal.NewCompliancePagePresenceMiddleware())

	sessionTransferHandler := NewSessionTransferHandler(
		iamSvc,
		cookieConfig,
		func(ctx context.Context, host string) bool {
			_, err := trustSvc.GetPortalByDomainName(ctx, host)
			return err == nil
		},
		logger,
	)
	r.Method(http.MethodGet, "/session-transfer", sessionTransferHandler)

	graphqlHandler := NewGraphQLHandler(
		iamSvc,
		trustSvc,
		resourceAliasSvc,
		fileManagerSvc,
		esignSvc,
		mailmanSvc,
		logger,
		baseURL,
		cookieConfig,
		tokenSecret,
		graphqlLimits,
	)

	r.Group(
		func(r chi.Router) {
			r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
			r.Use(complianceportal.NewMemberProvisioningMiddleware(trustSvc, logger))
			r.Handle("/graphql", graphqlHandler)
		},
	)

	return r
}
