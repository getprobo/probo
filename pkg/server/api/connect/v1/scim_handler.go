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

package connect_v1

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	scimservice "go.probo.inc/probo/pkg/iam/scim"
)

type SCIMHandler struct {
	iam    *iam.Service
	logger *log.Logger
}

func NewSCIMHandler(iam *iam.Service, logger *log.Logger) *SCIMHandler {
	return &SCIMHandler{iam: iam, logger: logger}
}

// Context key for SCIM configuration
type scimCtxKey struct{ name string }

var scimConfigCtxKey = &scimCtxKey{name: "scim_config"}

func scimConfigFromContext(ctx context.Context) *coredata.SCIMConfiguration {
	config, _ := ctx.Value(scimConfigCtxKey).(*coredata.SCIMConfiguration)
	return config
}

// NewSCIMServer creates a new SCIM server using elimity-com/scim
func NewSCIMServer(h *SCIMHandler) http.Handler {
	resourceTypes := []scim.ResourceType{
		{
			ID:          optional.NewString("User"),
			Name:        "User",
			Endpoint:    "/Users",
			Description: optional.NewString("User Account"),
			Schema:      scimservice.UserSchema(),
			Handler:     &scimResourceHandler{handler: h},
		},
	}

	serverConfig := scim.ServiceProviderConfig{
		SupportFiltering: true,
		SupportPatch:     true,
		AuthenticationSchemes: []scim.AuthenticationScheme{
			{
				Type:        scim.AuthenticationTypeOauthBearerToken,
				Name:        "OAuth Bearer Token",
				Description: "Authentication using OAuth Bearer Token",
			},
		},
	}

	server, err := scim.NewServer(
		&scim.ServerArgs{
			ServiceProviderConfig: &serverConfig,
			ResourceTypes:         resourceTypes,
		},
	)
	if err != nil {
		panic(err)
	}

	return server
}

// BearerTokenMiddleware validates the bearer token and sets the SCIM configuration in context
func (h *SCIMHandler) BearerTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpserver.RenderError(w, http.StatusUnauthorized, errors.New("authorization header required"))
			return
		}

		token, err := bearertoken.Parse(authHeader)
		if err != nil {
			httpserver.RenderError(w, http.StatusUnauthorized, errors.New("invalid authorization header"))
			return
		}

		config, err := h.iam.SCIMService.ValidateToken(r.Context(), token)
		if err != nil {
			var invalidToken *scimservice.ErrSCIMInvalidToken
			if errors.As(err, &invalidToken) {
				httpserver.RenderError(w, http.StatusUnauthorized, errors.New("invalid token"))
				return
			}

			h.logger.ErrorCtx(r.Context(), "SCIM token validation error", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		ctx := context.WithValue(r.Context(), scimConfigCtxKey, config)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// scimResourceHandler implements the elimity-com/scim ResourceHandler interface
type scimResourceHandler struct {
	handler *SCIMHandler
}

func (h *scimResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	config := scimConfigFromContext(ctx)

	resource, err := h.handler.iam.SCIMService.CreateUser(ctx, config, attributes, getIPAddress(r))
	if err != nil {
		var scimErr scimerrors.ScimError
		if errors.As(err, &scimErr) {
			return scim.Resource{}, err
		}
		h.handler.logger.ErrorCtx(ctx, "cannot create user", log.Error(err))
		return scim.Resource{}, scimerrors.ScimErrorInternal
	}

	return resource, nil
}

func (h *scimResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx := r.Context()
	config := scimConfigFromContext(ctx)

	membershipID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
	}

	resource, err := h.handler.iam.SCIMService.GetUser(ctx, config, membershipID, getIPAddress(r))
	if err != nil {
		var scimErr scimerrors.ScimError
		if errors.As(err, &scimErr) {
			return scim.Resource{}, err
		}
		h.handler.logger.ErrorCtx(ctx, "cannot get user", log.Error(err))
		return scim.Resource{}, scimerrors.ScimErrorInternal
	}

	return resource, nil
}

func (h *scimResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	ctx := r.Context()
	config := scimConfigFromContext(ctx)

	if err := params.FilterValidator.Validate(); err != nil {
		return scim.Page{}, scimerrors.ScimErrorBadRequest(err.Error())
	}

	// Parse SCIM filter AST into our filter type
	filter, err := scimservice.ParseUserFilter(params.FilterValidator.GetFilter())
	if err != nil {
		var scimErr scimerrors.ScimError
		if errors.As(err, &scimErr) {
			return scim.Page{}, err
		}
		h.handler.logger.ErrorCtx(ctx, "cannot parse filter", log.Error(err))
		return scim.Page{}, scimerrors.ScimErrorInternal
	}

	resources, totalCount, err := h.handler.iam.SCIMService.ListUsers(ctx, config, filter, params.StartIndex, params.Count, getIPAddress(r))
	if err != nil {
		var scimErr scimerrors.ScimError
		if errors.As(err, &scimErr) {
			return scim.Page{}, err
		}
		h.handler.logger.ErrorCtx(ctx, "cannot list users", log.Error(err))
		return scim.Page{}, scimerrors.ScimErrorInternal
	}

	return scim.Page{
		TotalResults: totalCount,
		Resources:    resources,
	}, nil
}

func (h *scimResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	config := scimConfigFromContext(ctx)

	membershipID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
	}

	resource, err := h.handler.iam.SCIMService.ReplaceUser(ctx, config, membershipID, attributes, getIPAddress(r))
	if err != nil {
		var scimErr scimerrors.ScimError
		if errors.As(err, &scimErr) {
			return scim.Resource{}, err
		}
		h.handler.logger.ErrorCtx(ctx, "cannot update user", log.Error(err))
		return scim.Resource{}, scimerrors.ScimErrorInternal
	}

	return resource, nil
}

func (h *scimResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	config := scimConfigFromContext(ctx)

	membershipID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, scimerrors.ScimErrorResourceNotFound(id)
	}

	resource, err := h.handler.iam.SCIMService.PatchUser(ctx, config, membershipID, operations, getIPAddress(r))
	if err != nil {
		var scimErr scimerrors.ScimError
		if errors.As(err, &scimErr) {
			return scim.Resource{}, err
		}
		h.handler.logger.ErrorCtx(ctx, "cannot patch user", log.Error(err))
		return scim.Resource{}, scimerrors.ScimErrorInternal
	}

	return resource, nil
}

func (h *scimResourceHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	config := scimConfigFromContext(ctx)

	membershipID, err := gid.ParseGID(id)
	if err != nil {
		return scimerrors.ScimErrorResourceNotFound(id)
	}

	err = h.handler.iam.SCIMService.DeleteUser(ctx, config, membershipID, getIPAddress(r))
	if err != nil {
		var scimErr scimerrors.ScimError
		if errors.As(err, &scimErr) {
			return err
		}
		h.handler.logger.ErrorCtx(ctx, "cannot delete user", log.Error(err))
		return scimerrors.ScimErrorInternal
	}

	return nil
}

func getIPAddress(r *http.Request) net.IP {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}

	if ip := net.ParseIP(host); ip != nil {
		return ip
	}

	return net.IPv4(127, 0, 0, 1)
}
