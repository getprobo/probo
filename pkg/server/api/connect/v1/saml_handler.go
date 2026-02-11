package connect_v1

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

type SAMLHandler struct {
	iam           *iam.Service
	sessionCookie *authn.Cookie
	baseURL       *baseurl.BaseURL
	logger        *log.Logger
}

func NewSAMLHandler(iam *iam.Service, cookieConfig securecookie.Config, baseURL *baseurl.BaseURL, logger *log.Logger) *SAMLHandler {
	return &SAMLHandler{
		iam:           iam,
		sessionCookie: authn.NewCookie(&cookieConfig),
		baseURL:       baseURL,
		logger:        logger,
	}
}

func (h *SAMLHandler) renderInternalServerError(w http.ResponseWriter, r *http.Request) {
	httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))
}

func (h *SAMLHandler) MetadataHandler(w http.ResponseWriter, r *http.Request) {
	metadataXML, err := h.iam.SAMLService.GenerateSpMetadata()
	if err != nil {
		panic(fmt.Errorf("cannot generate metadata: %w", err))
	}

	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(metadataXML)
}

func (h *SAMLHandler) ConsumeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("cannot parse form"))
		return
	}

	samlResponse := r.FormValue("SAMLResponse")
	relayState := r.FormValue("RelayState")

	values, err := url.ParseQuery(relayState)
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid relay state"))
		return
	}

	configIDStr := values.Get("config-id")
	if configIDStr == "" {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing config ID"))
		return
	}

	configID, err := gid.ParseGID(configIDStr)
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid config ID"))
		return
	}

	redirectPath := values.Get("redirect-path")

	user, membership, err := h.iam.SAMLService.HandleAssertion(ctx, samlResponse, configID)
	if err != nil {
		httpserver.RenderError(w, http.StatusUnauthorized, err)
		return
	}

	if redirectPath == "" {
		redirectPath = "/organizations/" + membership.OrganizationID.String()
	}

	rootSession := authn.SessionFromContext(ctx)

	switch {
	case rootSession == nil:
		rootSession, err = h.iam.AuthService.OpenSessionWithSAML(ctx, user.ID)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot open root session", log.Error(err))
			h.renderInternalServerError(w, r)
			return
		}
	case rootSession.IdentityID != user.ID:
		err = h.iam.SessionService.CloseSession(ctx, rootSession.ID)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot close session", log.Error(err))
			h.renderInternalServerError(w, r)
			return
		}

		rootSession, err = h.iam.AuthService.OpenSessionWithSAML(ctx, user.ID)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot open root session", log.Error(err))
			h.renderInternalServerError(w, r)
			return
		}
	}

	_, _, err = h.iam.SessionService.OpenSAMLChildSessionForOrganization(ctx, rootSession.ID, membership.OrganizationID)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot open SAML child session", log.Error(err))
		h.renderInternalServerError(w, r)
		return
	}

	h.sessionCookie.Set(w, rootSession)

	redirectURL := h.baseURL.WithPath(redirectPath).MustString()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *SAMLHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	samlConfigIDParam := chi.URLParam(r, "samlConfigID")
	if samlConfigIDParam == "" {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing SAML config ID"))
		return
	}

	redirectPathQueryParam := r.URL.Query().Get("redirect-path")

	samlConfigID, err := gid.ParseGID(samlConfigIDParam)
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid SAML config ID"))
		return
	}

	url, err := h.iam.SAMLService.InitiateLogin(ctx, samlConfigID, redirectPathQueryParam)
	if err != nil {
		panic(fmt.Errorf("cannot initiate SAML login: %w", err))
	}

	http.Redirect(w, r, url.String(), http.StatusFound)
}
