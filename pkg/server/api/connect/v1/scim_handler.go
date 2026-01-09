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
	scimfilter "github.com/scim2/filter-parser/v2"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	scimservice "go.probo.inc/probo/pkg/iam/scim"
)

type (
	ctxKey struct{ name string }

	SCIMHandler struct {
		iam    *iam.Service
		logger *log.Logger
	}

	scimResourceHandler struct {
		handler *SCIMHandler
	}

	scimRequestContext struct {
		ctx          context.Context
		config       *coredata.SCIMConfiguration
		ipAddress    net.IP
		method       string
		path         string
		membershipID *gid.GID
		handler      *scimResourceHandler
	}
)

var (
	scimConfigCtxKey = &ctxKey{name: "scim_config"}
)

func NewSCIMHandler(iam *iam.Service, logger *log.Logger) *SCIMHandler {
	return &SCIMHandler{iam: iam, logger: logger}
}

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

func (rc *scimRequestContext) logAndWrapError(err error, logMsg string) error {
	var scimErr scimerrors.ScimError
	if errors.As(err, &scimErr) {
		errMsg := scimErr.Detail

		// Don't reference membershipID for 404 errors - the resource doesn't exist
		membershipID := rc.membershipID
		if scimErr.Status == http.StatusNotFound {
			membershipID = nil
		}
		rc.handler.handler.iam.SCIMService.LogEvent(rc.ctx, rc.config, rc.method, rc.path, membershipID, rc.ipAddress, scimErr.Status, &errMsg)
		return err
	}

	rc.handler.handler.logger.ErrorCtx(rc.ctx, logMsg, log.Error(err))
	errMsg := "internal server error"
	rc.handler.handler.iam.SCIMService.LogEvent(rc.ctx, rc.config, rc.method, rc.path, rc.membershipID, rc.ipAddress, 500, &errMsg)
	return scimerrors.ScimErrorInternal
}

func (rc *scimRequestContext) logSuccess(statusCode int) {
	rc.handler.handler.iam.SCIMService.LogEvent(rc.ctx, rc.config, rc.method, rc.path, rc.membershipID, rc.ipAddress, statusCode, nil)
}

func (h *scimResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	rc := &scimRequestContext{
		ctx:       r.Context(),
		config:    scimConfigFromContext(r.Context()),
		ipAddress: getIPAddress(r),
		method:    "POST",
		path:      "/Users",
		handler:   h,
	}

	resource, err := h.handler.iam.SCIMService.CreateUser(rc.ctx, rc.config, attributes)
	if err != nil {
		return scim.Resource{}, rc.logAndWrapError(err, "cannot create user")
	}

	membershipID, _ := gid.ParseGID(resource.ID)
	rc.membershipID = &membershipID
	rc.logSuccess(201)
	return resource, nil
}

func (h *scimResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	rc := &scimRequestContext{
		ctx:       r.Context(),
		config:    scimConfigFromContext(r.Context()),
		ipAddress: getIPAddress(r),
		method:    "GET",
		path:      "/Users/" + id,
		handler:   h,
	}

	membershipID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, rc.logAndWrapError(scimerrors.ScimErrorResourceNotFound(id), "invalid membership ID")
	}
	rc.membershipID = &membershipID

	resource, err := h.handler.iam.SCIMService.GetUser(rc.ctx, rc.config, membershipID)
	if err != nil {
		return scim.Resource{}, rc.logAndWrapError(err, "cannot get user")
	}

	rc.logSuccess(200)
	return resource, nil
}

func (h *scimResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	path := "/Users"
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}

	rc := &scimRequestContext{
		ctx:       r.Context(),
		config:    scimConfigFromContext(r.Context()),
		ipAddress: getIPAddress(r),
		method:    "GET",
		path:      path,
		handler:   h,
	}

	var filterExpr scimfilter.Expression
	if params.FilterValidator != nil {
		if err := params.FilterValidator.Validate(); err != nil {
			return scim.Page{}, rc.logAndWrapError(scimerrors.ScimErrorBadRequest(err.Error()), "invalid filter")
		}

		filterExpr = params.FilterValidator.GetFilter()
	}

	resources, totalCount, err := h.handler.iam.SCIMService.ListUsers(rc.ctx, rc.config, filterExpr, params.StartIndex, params.Count)
	if err != nil {
		return scim.Page{}, rc.logAndWrapError(err, "cannot list users")
	}

	rc.logSuccess(200)
	return scim.Page{
		TotalResults: totalCount,
		Resources:    resources,
	}, nil
}

func (h *scimResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	rc := &scimRequestContext{
		ctx:       r.Context(),
		config:    scimConfigFromContext(r.Context()),
		ipAddress: getIPAddress(r),
		method:    "PUT",
		path:      "/Users/" + id,
		handler:   h,
	}

	membershipID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, rc.logAndWrapError(scimerrors.ScimErrorResourceNotFound(id), "invalid membership ID")
	}
	rc.membershipID = &membershipID

	resource, err := h.handler.iam.SCIMService.ReplaceUser(rc.ctx, rc.config, membershipID, attributes)
	if err != nil {
		return scim.Resource{}, rc.logAndWrapError(err, "cannot update user")
	}

	rc.logSuccess(200)
	return resource, nil
}

func (h *scimResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	rc := &scimRequestContext{
		ctx:       r.Context(),
		config:    scimConfigFromContext(r.Context()),
		ipAddress: getIPAddress(r),
		method:    "PATCH",
		path:      "/Users/" + id,
		handler:   h,
	}

	membershipID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, rc.logAndWrapError(scimerrors.ScimErrorResourceNotFound(id), "invalid membership ID")
	}
	rc.membershipID = &membershipID

	resource, err := h.handler.iam.SCIMService.PatchUser(rc.ctx, rc.config, membershipID, operations)
	if err != nil {
		return scim.Resource{}, rc.logAndWrapError(err, "cannot patch user")
	}

	rc.logSuccess(200)
	return resource, nil
}

func (h *scimResourceHandler) Delete(r *http.Request, id string) error {
	rc := &scimRequestContext{
		ctx:       r.Context(),
		config:    scimConfigFromContext(r.Context()),
		ipAddress: getIPAddress(r),
		method:    "DELETE",
		path:      "/Users/" + id,
		handler:   h,
	}

	membershipID, err := gid.ParseGID(id)
	if err != nil {
		return rc.logAndWrapError(scimerrors.ScimErrorResourceNotFound(id), "invalid membership ID")
	}
	rc.membershipID = &membershipID

	err = h.handler.iam.SCIMService.DeleteUser(rc.ctx, rc.config, membershipID)
	if err != nil {
		return rc.logAndWrapError(err, "cannot delete user")
	}

	rc.membershipID = nil
	rc.logSuccess(204)
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
