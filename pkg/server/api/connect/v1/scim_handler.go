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

package connect_v1

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/elimity-com/scim"
	scimerrors "github.com/elimity-com/scim/errors"
	"github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"
	scimfilter "github.com/scim2/filter-parser/v2"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	scimservice "go.probo.inc/probo/pkg/iam/scim"
	"go.probo.inc/probo/pkg/server/api/clientip"
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

	// scimEventRecorder collects the parts of a SCIM audit event that only the
	// resource handler knows about. The event itself is written by
	// EventLoggingMiddleware once the response has been served.
	scimEventRecorder struct {
		handled      bool
		userName     string
		errorMessage *string
	}

	// scimResponseRecorder captures the status code and body served to the
	// provisioning client so the audit event stores the exact response.
	scimResponseRecorder struct {
		http.ResponseWriter

		statusCode int
		body       bytes.Buffer
	}
)

var (
	scimConfigCtxKey        = &ctxKey{name: "scim_config"}
	scimEventRecorderCtxKey = &ctxKey{name: "scim_event_recorder"}
)

func NewSCIMHandler(iam *iam.Service, logger *log.Logger) *SCIMHandler {
	return &SCIMHandler{iam: iam, logger: logger}
}

func scimConfigFromContext(ctx context.Context) *coredata.SCIMConfiguration {
	config, _ := ctx.Value(scimConfigCtxKey).(*coredata.SCIMConfiguration)
	return config
}

func scimEventRecorderFromContext(ctx context.Context) *scimEventRecorder {
	recorder, _ := ctx.Value(scimEventRecorderCtxKey).(*scimEventRecorder)
	return recorder
}

// markHandled flags the request as one of our resource operations, which is what
// makes it eligible for a SCIM audit event. Discovery endpoints served directly
// by the SCIM server never reach a resource handler and stay unlogged.
func (rec *scimEventRecorder) markHandled() {
	if rec == nil {
		return
	}

	rec.handled = true
}

func (rec *scimEventRecorder) setUserName(userName string) {
	if rec == nil {
		return
	}

	rec.userName = userName
}

func (rec *scimEventRecorder) setErrorMessage(errorMessage string) {
	if rec == nil {
		return
	}

	rec.errorMessage = &errorMessage
}

func (rr *scimResponseRecorder) WriteHeader(statusCode int) {
	rr.statusCode = statusCode
	rr.ResponseWriter.WriteHeader(statusCode)
}

func (rr *scimResponseRecorder) Write(b []byte) (int, error) {
	rr.body.Write(b)
	return rr.ResponseWriter.Write(b)
}

// NewSCIMServer creates a new SCIM server using elimity-com/scim
func NewSCIMServer(h *SCIMHandler) http.Handler {
	schema.SetAllowStringValues(true)

	resourceTypes := []scim.ResourceType{
		{
			ID:          optional.NewString("User"),
			Name:        "User",
			Endpoint:    "/Users",
			Description: optional.NewString("User Account"),
			Schema:      scimservice.UserSchema(),
			SchemaExtensions: []scim.SchemaExtension{
				{Schema: scimservice.EnterpriseUserSchema()},
			},
			Handler: &scimResourceHandler{handler: h},
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
			if _, ok := errors.AsType[*scimservice.ErrSCIMInvalidToken](err); ok {
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

// EventLoggingMiddleware records a SCIM audit event for every resource
// operation, storing the request and response bodies exactly as they were
// exchanged with the provisioning client. It must run after
// BearerTokenMiddleware, which resolves the SCIM configuration.
func (h *SCIMHandler) EventLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config := scimConfigFromContext(r.Context())
		if config == nil {
			next.ServeHTTP(w, r)
			return
		}

		requestBody, err := drainRequestBody(r)
		if err != nil {
			h.logger.ErrorCtx(r.Context(), "cannot read SCIM request body", log.Error(err))
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("cannot read request body"))

			return
		}

		recorder := &scimEventRecorder{}
		responseRecorder := &scimResponseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		ctx := context.WithValue(r.Context(), scimEventRecorderCtxKey, recorder)
		next.ServeHTTP(responseRecorder, r.WithContext(ctx))

		if !recorder.handled {
			return
		}

		h.iam.SCIMService.LogEvent(
			r.Context(),
			config,
			r.Method,
			requestPath(r),
			recorder.userName,
			getIPAddress(r),
			responseRecorder.statusCode,
			recorder.errorMessage,
			bodyOrNil(requestBody),
			bodyOrNil(responseRecorder.body.Bytes()),
		)
	})
}

// recordError turns a service error into the error returned to the provisioning
// client, and records its detail on the audit event. Unexpected errors are
// logged and reported as an opaque internal error.
func (h *scimResourceHandler) recordError(
	ctx context.Context,
	recorder *scimEventRecorder,
	err error,
	logMsg string,
) error {
	if scimErr, ok := errors.AsType[scimerrors.ScimError](err); ok {
		recorder.setErrorMessage(scimErr.Detail)

		return err
	}

	h.handler.logger.ErrorCtx(ctx, logMsg, log.Error(err))
	recorder.setErrorMessage("internal server error")

	return scimerrors.ScimErrorInternal
}

// drainRequestBody reads the request body so it can be stored on the audit
// event, and restores it for the SCIM server to consume.
func drainRequestBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read request body: %w", err)
	}

	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body))

	return body, nil
}

func requestPath(r *http.Request) string {
	if r.URL.RawQuery == "" {
		return r.URL.Path
	}

	return r.URL.Path + "?" + r.URL.RawQuery
}

func bodyOrNil(body []byte) *string {
	if len(body) == 0 {
		return nil
	}

	s := string(body)

	return &s
}

func (h *scimResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	recorder := scimEventRecorderFromContext(ctx)
	recorder.markHandled()

	resource, err := h.handler.iam.SCIMService.CreateUser(ctx, scimConfigFromContext(ctx), attributes)
	if err != nil {
		return scim.Resource{}, h.recordError(ctx, recorder, err, "cannot create user")
	}

	recorder.setUserName(resource.Attributes["userName"].(string))

	return resource, nil
}

func (h *scimResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	ctx := r.Context()
	recorder := scimEventRecorderFromContext(ctx)
	recorder.markHandled()

	profileID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, h.recordError(ctx, recorder, scimerrors.ScimErrorResourceNotFound(id), "invalid profile ID")
	}

	resource, err := h.handler.iam.SCIMService.GetUser(ctx, scimConfigFromContext(ctx), profileID)
	if err != nil {
		return scim.Resource{}, h.recordError(ctx, recorder, err, "cannot get user")
	}

	recorder.setUserName(resource.Attributes["userName"].(string))

	return resource, nil
}

func (h *scimResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	ctx := r.Context()
	recorder := scimEventRecorderFromContext(ctx)
	recorder.markHandled()

	var filterExpr scimfilter.Expression

	if params.FilterValidator != nil {
		if err := params.FilterValidator.Validate(); err != nil {
			return scim.Page{}, h.recordError(ctx, recorder, scimerrors.ScimErrorBadRequest(err.Error()), "invalid filter")
		}

		filterExpr = params.FilterValidator.GetFilter()
	}

	resources, totalCount, err := h.handler.iam.SCIMService.ListUsers(ctx, scimConfigFromContext(ctx), filterExpr, params.StartIndex, params.Count)
	if err != nil {
		return scim.Page{}, h.recordError(ctx, recorder, err, "cannot list users")
	}

	return scim.Page{
		TotalResults: totalCount,
		Resources:    resources,
	}, nil
}

func (h *scimResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	ctx := r.Context()
	recorder := scimEventRecorderFromContext(ctx)
	recorder.markHandled()

	profileID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, h.recordError(ctx, recorder, scimerrors.ScimErrorResourceNotFound(id), "invalid profile ID")
	}

	resource, err := h.handler.iam.SCIMService.ReplaceUser(ctx, scimConfigFromContext(ctx), profileID, attributes)
	if err != nil {
		return scim.Resource{}, h.recordError(ctx, recorder, err, "cannot update user")
	}

	recorder.setUserName(resource.Attributes["userName"].(string))

	return resource, nil
}

func (h *scimResourceHandler) Patch(r *http.Request, id string, operations []scim.PatchOperation) (scim.Resource, error) {
	ctx := r.Context()
	recorder := scimEventRecorderFromContext(ctx)
	recorder.markHandled()

	profileID, err := gid.ParseGID(id)
	if err != nil {
		return scim.Resource{}, h.recordError(ctx, recorder, scimerrors.ScimErrorResourceNotFound(id), "invalid profile ID")
	}

	resource, err := h.handler.iam.SCIMService.PatchUser(ctx, scimConfigFromContext(ctx), profileID, operations)
	if err != nil {
		return scim.Resource{}, h.recordError(ctx, recorder, err, "cannot patch user")
	}

	recorder.setUserName(resource.Attributes["userName"].(string))

	return resource, nil
}

func (h *scimResourceHandler) Delete(r *http.Request, id string) error {
	ctx := r.Context()
	recorder := scimEventRecorderFromContext(ctx)
	recorder.markHandled()

	profileID, err := gid.ParseGID(id)
	if err != nil {
		return h.recordError(ctx, recorder, scimerrors.ScimErrorResourceNotFound(id), "invalid profile ID")
	}

	if err := h.handler.iam.SCIMService.DeleteUser(ctx, scimConfigFromContext(ctx), profileID); err != nil {
		return h.recordError(ctx, recorder, err, "cannot delete user")
	}

	return nil
}

func getIPAddress(r *http.Request) net.IP {
	if ip := net.ParseIP(clientip.Extract(r)); ip != nil {
		return ip
	}

	return net.IPv4(127, 0, 0, 1)
}
