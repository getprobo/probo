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
	"fmt"
	"net/http"
	"time"

	authsvc "github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/go-chi/chi/v5"
)

func OrganizationLogoHandler(authSvc *authsvc.Service, fileManager interface {
	GenerateFileUrl(ctx context.Context, file *coredata.File, duration time.Duration) (string, error)
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		user := UserFromContext(ctx)
		session := SessionFromContext(ctx)

		organizationIDStr := chi.URLParam(r, "organizationID")
		organizationID, err := gid.ParseGID(organizationIDStr)
		if err != nil {
			http.Error(w, "Invalid organization ID", http.StatusBadRequest)
			return
		}

		logoFile, err := authSvc.GetOrganizationLogoFile(ctx, user, organizationID, session)
		if err != nil {
			panic(fmt.Errorf("cannot get organization logo: %w", err))
		}

		presignedURL, err := fileManager.GenerateFileUrl(ctx, logoFile, 1*time.Hour)
		if err != nil {
			panic(fmt.Errorf("cannot generate presigned URL: %w", err))
		}

		w.Header().Set("Cache-Control", "public, max-age=3600")

		http.Redirect(w, r, presignedURL, http.StatusFound)
	}
}
