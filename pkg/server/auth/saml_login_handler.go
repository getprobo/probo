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
	"fmt"
	"net/http"

	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/gid"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
)

func SAMLLoginHandler(samlSvc *authsvc.SAMLService, authSvc *authsvc.Service, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		samlConfigIDStr := chi.URLParam(r, "samlConfigID")
		if samlConfigIDStr == "" {
			logger.WarnCtx(ctx, "missing SAML config ID in URL")
			http.Error(w, "missing SAML config ID", http.StatusBadRequest)
			return
		}

		samlConfigID, err := gid.ParseGID(samlConfigIDStr)
		if err != nil {
			logger.ErrorCtx(ctx, "invalid SAML config ID", log.Error(err), log.String("saml_config_id", samlConfigIDStr))
			http.Error(w, "invalid SAML config ID", http.StatusBadRequest)
			return
		}

		tenantID := samlConfigID.TenantID()

		config, err := authSvc.WithTenant(tenantID).GetSAMLConfigurationByID(ctx, samlConfigID)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot load SAML configuration", log.Error(err), log.String("saml_config_id", samlConfigID.String()))
			http.Error(w, "SAML configuration not found", http.StatusNotFound)
			return
		}

		redirectURL, err := samlSvc.InitiateSAMLLogin(ctx, config.OrganizationID, tenantID, config.EmailDomain)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot initiate SAML login", log.Error(err), log.String("saml_config_id", samlConfigID.String()), log.String("org_id", config.OrganizationID.String()))
			http.Error(w, fmt.Sprintf("SAML login failed: %v", err), http.StatusInternalServerError)
			return
		}

		logger.InfoCtx(ctx, "SAML login initiated", log.String("saml_config_id", samlConfigID.String()), log.String("org_id", config.OrganizationID.String()))

		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}
