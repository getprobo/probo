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

//go:generate go run github.com/99designs/gqlgen generate

package complianceportal_v1

import (
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	trust "go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/resourcealias"
	"go.probo.inc/probo/pkg/server/api/authn"
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
